package qiniu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/backend"
)

type Wagoner = backend.Wagoner
type JsonKV = backend.JsonKV

type ListItem = storage.ListItem

type ListOption struct {
	Prefix    string
	Delimiter string
	Limit     int
}

type bucketWrap struct {
	Name   string // 存储空间名
	Prefix string // bucket prefix
	URI    string // bucket cdn uri
	cfg    *storage.Config
	mgr    *storage.BucketManager
}

var (
	conf struct {
		AccessKey string            `envconfig:"ACCESS_KEY"`
		SecretKey string            `envconfig:"SECRET_KEY"`
		Prefix    string            `envconfig:"PREFIX"` // example: imsto/
		Zones     map[string]string `envconfig:"ZONES"`  // [roof]zone
		URIs      map[string]string `envconfig:"URIS"`   // [roof]uri
	}

	mac     *qbox.Mac
	buckets map[string]*bucketWrap // replace it with Bucketer soon

	httpClient = http.Client{
		Timeout:   time.Second * 20,
		Transport: http.DefaultTransport,
	}
)

func init() {
	envconfig.MustProcess("qiniu", &conf)

	mac = qbox.NewMac(conf.AccessKey, conf.SecretKey)
	buckets = map[string]*bucketWrap{}
	for roof, name := range config.Current.Buckets {
		zone := conf.Zones[roof]
		cfg := storage.Config{}
		cfg.Zone = getZone(zone)
		cfg.UseHTTPS = false
		buckets[roof] = &bucketWrap{
			cfg:    &cfg,
			mgr:    storage.NewBucketManager(mac, &cfg),
			Name:   name,
			URI:    conf.URIs[roof],
			Prefix: conf.Prefix,
		}
	}

	backend.RegisterEngine("qiniu", qnDial)
}

func getZone(zone string) *storage.Region {
	switch zone {
	case "Huadong":
		return &storage.ZoneHuadong
	case "Huabei":
		return &storage.ZoneHuabei
	case "Huanan":
		return &storage.ZoneHuanan
	case "Beimei":
		return &storage.ZoneBeimei
	}
	logger().Errorw("qn: invalid zone", "zone", zone)
	return &storage.ZoneHuabei
}

func qnDial(roof string) (Wagoner, error) {
	if b, ok := buckets[roof]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("QN: bucket of %s not found", roof)

	// cdnurl.RegisterPrivateFunc(func(s string) string {
	// 	if bu, ok := buckets[BUCKET_PRIVATES]; ok {
	// 		return bu.MakePrivateURL(s)
	// 	}
	// 	return s
	// })
	// if b, ok := buckets[BUCKET_MEDIA]; ok {
	// 	cdnurl.SetPrefix(b.GetUrl())
	// }
	// return c, nil
}

// MakeURL Make a public url
func (bu *bucketWrap) MakeURL(key string) string {
	return bu.makeURL(key, false)
}

// MakePrivateURL
func (bu *bucketWrap) MakePrivateURL(key string) string {
	return bu.makeURL(key, true)
}

// makeURL ...
func (bu *bucketWrap) makeURL(key string, private bool) string {
	if bu == nil {
		log.Printf("bucket wrap is nil for key %v, private %v", key, private)
		return ""
	}
	if key == "" {
		log.Printf("empty key of %s", bu.Name)
		return ""
	}
	if private {
		deadline := time.Now().Add(time.Second * 3600).Unix() //1小时有效期
		return storage.MakePrivateURL(mac, bu.URI, key, deadline)
	}
	return storage.MakePublicURL(bu.URI, key)
}

func (bu *bucketWrap) id2key(id string) string {
	return path.Join(bu.Prefix, backend.ID2Path(id))
}

//get the name with prefix file，if prefix is nil,will get all
func (bu *bucketWrap) List(opt ListOption) (entries []storage.ListItem, err error) {
	if opt.Limit == 0 {
		opt.Limit = 100
	}
	entries, _, _, _, err = bu.mgr.ListFiles(bu.Name, opt.Prefix, opt.Delimiter, "", opt.Limit)
	if err != nil {
		log.Println("get qiniu List error:", err)
		return
	}
	return
}

func (bu *bucketWrap) Get(id string) (data []byte, err error) {
	key := bu.id2key(id)
	uri := bu.makeURL(key, false)
	res, err := httpClient.Get(uri)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err = ioutil.ReadAll(res.Body)

	return
}

func (bu *bucketWrap) Put(id string, data []byte, meta JsonKV) (sev JsonKV, err error) {
	key := bu.id2key(id)
	rd := bytes.NewReader(data)
	size := len(data)

	upToken, ret, putExtra := bu.defaultUpload()
	formUploader := storage.NewFormUploader(bu.cfg)
	putExtra.Params = metaToMaps(meta)

	err = formUploader.Put(context.Background(), &ret, upToken, key, rd, int64(size), &putExtra)
	if err != nil {
		log.Printf("upload file err: %s, fileName: %s", err, key)
		return
	}
	return
}

func (bu *bucketWrap) defaultUpload() (upToken string, ret storage.PutRet, putExtra storage.PutExtra) {
	putPolicy := storage.PutPolicy{
		Scope: bu.Name,
	}
	upToken = putPolicy.UploadToken(mac)
	ret = storage.PutRet{}
	putExtra = storage.PutExtra{}
	return
}

func (bu *bucketWrap) Upload(rd io.Reader, key string, size int) (err error) {
	upToken, ret, putExtra := bu.defaultUpload()
	formUploader := storage.NewFormUploader(bu.cfg)

	err = formUploader.Put(context.Background(), &ret, upToken, key, rd, int64(size), &putExtra)
	if err != nil {
		log.Printf("upload file err: %s, fileName: %s", err, key)
		return
	}
	return
}

func metaToMaps(h JsonKV) (m map[string]string) {
	m = make(map[string]string)
	for k, v := range h {
		if k == "name" {
			m["x:"+k] = url.QueryEscape(fmt.Sprint(v))
		} else {
			m["x:"+k] = fmt.Sprint(v)
		}
	}
	return
}

func (bu *bucketWrap) Exists(id string) (bool, error) {
	fi, err := bu.mgr.Stat(bu.Name, bu.id2key(id))
	if err != nil {
		return false, err
	}
	return fi.Fsize > 0, nil
}

func (bu *bucketWrap) Del(id string) error {
	return bu.Delete(bu.id2key(id))
}

func (bu *bucketWrap) Delete(keys ...string) (err error) {
	deleteOps := make([]string, 0, len(keys))
	for _, key := range keys {
		deleteOps = append(deleteOps, storage.URIDelete(bu.Name, key))
	}
	_, err = bu.mgr.Batch(deleteOps)
	if err != nil {
		log.Println("delete list file error:", err)
		return
	}
	return
}

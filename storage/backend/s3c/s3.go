package backend

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/kelseyhightower/envconfig"

	"github.com/go-imsto/aws4"

	"github.com/go-imsto/imsto/storage/backend"
)

// Wagoner ...
type Wagoner = backend.Wagoner

// JsonKV ...
type JsonKV = backend.JsonKV

const (
	emptySum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type s3Conn struct {
	name     string // bucket
	endpoint string
	region   string
	prefix   string
	ac       *aws4.Client
	uri      string
}

// vars
var (
	ErrEndpoint   = errors.New("need Endpoint in environment")
	ErrBucketName = errors.New("need Bucket in environment")
	ErrRequest    = errors.New("request error")

	protocol  = "http"
	uriFormat = "https://s3.%s.amazonaws.com/%s"
	buckets   = map[string]*s3Conn{} // replace it with Bucketer soon
	dft       *s3Conn
)

func init() {
	var conf struct {
		AccessKey string            `envconfig:"ACCESS_KEY"`
		SecretKey string            `envconfig:"SECRET_KEY"`
		Region    string            `envconfig:"REGION"`                  // only one region for now
		Bucket    string            `envconfig:"BUCKET"`                  // default bucket
		Protocol  string            `envconfig:"PROTOCOL" default:"http"` // example: https
		Prefix    string            `envconfig:"PREFIX" default:"imsto/"` // example: imsto/
		Buckets   map[string]string `envconfig:"BUCKETS"`                 // [roof]bucket
		Endpoints map[string]string `envconfig:"ENDPOINTS"`               // [roof]endpoint
		URIs      map[string]string `envconfig:"URIS"`                    // [roof]uri
	}
	envconfig.MustProcess("aws_s3", &conf)

	ac := &aws4.Client{Keys: &aws4.Keys{conf.AccessKey, conf.SecretKey}}
	ac.Name = "s3"
	ac.Region = conf.Region
	for roof, name := range conf.Buckets {
		endpoint := conf.Endpoints[roof]

		buckets[roof] = &s3Conn{
			name:     name,
			endpoint: endpoint,
			region:   conf.Region,
			prefix:   conf.Prefix,
			ac:       ac,
			uri:      conf.URIs[roof],
		}
	}
	if len(buckets) == 0 {
		if conf.Bucket != "" && conf.Region != "" {
			dft = &s3Conn{name: conf.Bucket, region: conf.Region, prefix: conf.Prefix, ac: ac}
		}
	}

	logger().Infow("init done ", "buckets", buckets)

	backend.RegisterEngine("s3", s3Dial)
}

func s3Dial(roof string) (Wagoner, error) {
	if c, ok := buckets[roof]; ok {
		return c, nil
	}
	if dft != nil {
		return dft, nil
	}
	return nil, ErrBucketName
}

func (c *s3Conn) id2key(id string) string {
	return path.Join(c.prefix, backend.ID2Path(id))
}

func (c *s3Conn) getURL(key string) (uri string) {
	if len(c.endpoint) > 0 {
		uri = protocol + "://" + c.endpoint + "/" + c.name
	} else {
		uri = fmt.Sprintf(uriFormat, c.region, c.name)
	}

	if len(key) > 0 {
		uri = uri + "/" + key
	}
	return
}

// Exists ...
func (c *s3Conn) Exists(id string) (exist bool, err error) {
	var req *http.Request
	req, err = http.NewRequest("HEAD", c.getURL(c.id2key(id)), nil)
	if err != nil {
		return
	}
	req.Header.Set("x-amz-content-sha256", emptySum)
	var resp *http.Response
	resp, err = c.ac.Do(req)
	if err != nil {
		logger().Infow("exists fail", "id", id, "err", err)
		return
	}
	exist = resp.StatusCode == 200
	return
}

// Get ...
func (c *s3Conn) Get(id string) (data []byte, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", c.getURL(c.id2key(id)), nil)
	if err != nil {
		return
	}

	req.Header.Set("x-amz-content-sha256", emptySum)
	var resp *http.Response
	resp, err = c.ac.Do(req)
	if err != nil {
		logger().Infow("get fail", "id", id, "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger().Infow("get status", "code", resp.StatusCode)
		err = ErrRequest
		return
	}
	data, err = ioutil.ReadAll(resp.Body)

	return
}

func metaToMaps(h JsonKV) (m map[string][]string) {
	m = make(map[string][]string)
	for k, v := range h {
		if k == "name" {
			m[k] = []string{url.QueryEscape(fmt.Sprint(v))}
		} else {
			m[k] = []string{fmt.Sprint(v)}
		}

	}
	return
}

// Put ...
func (c *s3Conn) Put(id string, data []byte, meta JsonKV) (sev JsonKV, err error) {
	key := c.id2key(id)
	uri := c.getURL(key)
	var req *http.Request
	req, err = http.NewRequest("PUT", uri, bytes.NewReader(data))
	if err != nil {
		return
	}

	h := sha256.New()
	h.Write(data)
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", h.Sum(nil)))
	req.Header.Set("content-type", fmt.Sprint(meta.Get("mime")))
	req.Header.Set("content-length", fmt.Sprint(len(data)))
	log.Printf("s3 Put %s: %s %s size %d\n", c.name, key, meta, len(data))

	var resp *http.Response
	resp, err = c.ac.Do(req)
	if err != nil {
		logger().Infow("put fail", "id", id, "meta", meta, "err", err)
		return
	}
	defer resp.Body.Close()
	logger().Infow("put ", "header", req.Header, "err", err)

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		logger().Infow("put ", "uri", uri, "code", resp.StatusCode, "result", buf)
		err = ErrRequest
		return
	}

	sev = JsonKV{"engine": "s3", "bucket": c.name, "key": key, "host": c.endpoint}
	logger().Infow("s3 Put done", "id", id)

	return
}

// Delete ...
func (c *s3Conn) Delete(id string) (err error) {
	var req *http.Request
	req, err = http.NewRequest("DELETE", c.getURL(c.id2key(id)), nil)
	if err != nil {
		return
	}
	req.Header.Set("x-amz-content-sha256", emptySum)
	var resp *http.Response
	resp, err = c.ac.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		logger().Infow("delete fail", "id", id, "code", resp.StatusCode)
		err = ErrRequest
		return
	}
	logger().Infow("s3 Delete done", "id", id)
	return
}

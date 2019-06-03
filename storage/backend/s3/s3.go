package backend

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/backend"
)

type Wagoner = backend.Wagoner
type JsonKV = backend.JsonKV

type s3Conn struct {
	b *s3.Bucket
}

var (
	ErrBucketName = errors.New("need bucket_name in settings")
)

func init() {
	backend.RegisterEngine("s3", s3Dial)
}

func s3Dial(roof string) (Wagoner, error) {

	var (
		access, secret, bucket string
		err                    error
	)
	access = config.GetValue(roof, "s3_access_key")
	if access == "" {
		access = os.Getenv("S3_ACCESS_KEY")
	}
	secret = config.GetValue(roof, "s3_secret_key")
	if secret == "" {
		secret = os.Getenv("S3_SECRET_KEY")
	}
	bucket = config.GetValue(roof, "bucket_name")
	if bucket == "" {
		err = ErrBucketName
		log.Print(err)
		return nil, err
	}

	auth := aws.Auth{AccessKey: access, SecretKey: secret}

	b := s3.New(auth, aws.USEast).Bucket(bucket)
	c := &s3Conn{
		b: b,
	}

	if _, err = c.list("", "/", 10); err != nil {
		log.Print("s3Dial:", err)
		return nil, err
	}

	return c, nil
}

func (c *s3Conn) list(prefix, delim string, max int) (*s3.ListResp, error) {
	resp, err := c.b.List(prefix, delim, "", max)
	if err != nil {
		return nil, err
	}
	ret := resp
	for max == 0 && resp.IsTruncated {
		last := resp.Contents[len(resp.Contents)-1].Key
		resp, err = c.b.List(prefix, delim, last, max)
		if err != nil {
			return ret, err
		}
		ret.Contents = append(ret.Contents, resp.Contents...)
		ret.CommonPrefixes = append(ret.CommonPrefixes, resp.CommonPrefixes...)
	}
	return ret, nil
}

func (c *s3Conn) Exists(id string) (exist bool, err error) {
	exist, err = c.b.Exists(backend.Id2Path(id))
	return
}

func (c *s3Conn) Get(id string) (data []byte, err error) {
	key := backend.Id2Path(id)
	for i := 0; ; {
		data, err = c.b.Get(key)
		if err == nil {
			break
		}
		if i++; i >= 3 {
			return
		}
		log.Printf("error: s3 Get %s: %s", key, err)
	}
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

func (c *s3Conn) Put(id string, data []byte, meta JsonKV) (sev JsonKV, err error) {
	key := backend.Id2Path(id)
	log.Printf("s3 Put %s: %s %s size %d\n", c.b.Name, key, meta, len(data))
	err = c.b.Put(key, data, fmt.Sprint(meta.Get("mime")), s3.Private, s3.Options{Meta: metaToMaps(meta)})
	if err != nil {
		log.Print("s3 Put:", err)
	}
	sev = JsonKV{"engine": "s3", "bucket": c.b.Name, "key": key}
	log.Print("s3 Put done")

	return
}

func (c *s3Conn) Del(id string) error {
	return c.b.Del(backend.Id2Path(id))
}

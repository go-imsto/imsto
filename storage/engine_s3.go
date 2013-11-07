package storage

import (
	"calf/config"
	"errors"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"os"
)

type s3Conn struct {
	b *s3.Bucket
}

var (
	ErrBucketName = errors.New("need bucket_name in settings")
)

func init() {
	RegisterEngine("s3", s3Dial)
}

func s3Dial(sn string) (Wagoner, error) {

	var (
		access, secret, bucket string
		err                    error
	)
	access = config.GetValue(sn, "s3_access_key")
	if access == "" {
		access = os.Getenv("S3_ACCESS_KEY")
	}
	secret = config.GetValue(sn, "s3_secret_key")
	if secret == "" {
		secret = os.Getenv("S3_SECRET_KEY")
	}
	bucket = config.GetValue(sn, "bucket_name")
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

func (c *s3Conn) Exists(key string) bool {
	return false
}

func (c *s3Conn) Get(key string) (data []byte, err error) {
	for i := 0; ; {
		data, err = c.b.Get(key)
		if err == nil {
			break
		}
		if i++; i >= 3 {
			return nil, err
		}
		log.Print("s3 Get:", err)
	}
	return data, nil
}

func (c *s3Conn) Put(entry *Entry, data []byte) (err error) {

	err = c.b.Put(entry.Path, data, entry.Mime, s3.Private)
	if err != nil {
		log.Print("s3 Put:", err)
	}
	log.Print("s3 Put done")

	return
}

func (c *s3Conn) Del(key string) error {
	return c.b.Del(key) // key = path
}

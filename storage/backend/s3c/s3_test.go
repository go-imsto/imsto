package backend

import (
	"log"
	"testing"

	"go.uber.org/zap"

	zlog "github.com/go-imsto/imsto/log"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var logger *zap.Logger
	logger, _ = zap.NewDevelopment()
	logger.Debug("logger start")
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	zlog.Set(sugar)
}

func TestS3(t *testing.T) {

	s3, err := s3Dial("demo")
	if err != nil {
		t.Fatal(err)
	}

	id := Key{ID: "test001.txt", Cat: "imsto"}
	text := "hello world"

	meta := Meta{"mime": "text/plain"}
	_, err = s3.Put(id, []byte(text), meta)
	if err != nil {
		t.Fatalf("put %s err %s", id, err)
	}

	var ok bool
	ok, err = s3.Exists(id)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("exists %s, %v", id, ok)

	var data []ListItem
	data, err = s3.List(ListSpec{Prefix: "imsto", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data %v", data)

	err = s3.Delete(id)
	if err != nil {
		t.Fatal(err)
	}

}

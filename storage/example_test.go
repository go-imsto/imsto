package storage

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	iimg "github.com/go-imsto/imsto/image"
	zlog "github.com/go-imsto/imsto/log"

	_ "github.com/go-imsto/imsto/storage/backend/file"
	_ "github.com/go-imsto/imsto/storage/backend/s3c"
)

const (
	jpegWidth   = iimg.Dimension(124)
	jpegHeight  = iimg.Dimension(144)
	jpegQuality = iimg.Quality(88)
	jpegSize    = 5642 // 5642, 5562
)
const jpegData = `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`

func TestMain(m *testing.M) {

	var logger *zap.Logger
	logger, _ = zap.NewDevelopment()
	logger.Debug("logger start")
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	zlog.Set(sugar)

	m.Run()
}

// TestEntry ...
func TestEntry(t *testing.T) {

	// rd := base64.NewDecoder(base64.StdEncoding, strings.NewReader(jpegData))
	buf, err := base64.StdEncoding.DecodeString(jpegData)
	assert.NoError(t, err)
	assert.NotEmpty(t, buf)

	entry, err := NewEntryReader(bytes.NewReader(buf), "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, entry)
	assert.Equal(t, jpegWidth, entry.im.Width)
	assert.Equal(t, jpegHeight, entry.im.Height)
	assert.Equal(t, int(jpegQuality), int(entry.im.Quality))

	var (
		_hash = "709e291268aea5f67a3397679b6fd9cd"
		_id   = "1kvyfwpt4u9l9"
	)

	IID := entry.Id

	if entry.h != _hash {
		t.Fatalf("unexpected result from HashContent:\n+ %v\n- %v", entry.h, _hash)
	}
	if IID.String() != _id {
		t.Fatalf("unexpected result from HashContent:\n+ %v\n- %v", IID, _id)
	}

	ch := entry.Store("demo")

	<-ch
}

const (
	t_salt  = "abcd"
	t_value = "test"
)

func TestApiToken(t *testing.T) {
	var (
		ver   = VerID(0)
		appid = AppID(0)
		vc    = valueCate(0)
	)
	token, err := newToken(ver, appid, []byte(t_salt))
	if err != nil {
		t.Fatal(err)
	}

	token.SetValue([]byte(t_value), vc)
	// t.Logf("api token bins: %x", token.Binary())
	str := token.String()
	t.Logf("api token strs: %s", str)
	t.Logf("api token hash: %x, stamp: %d, value: %s", token.hash, token.stamp, token.GetValue())

	token, err = newToken(ver, appid, []byte(t_salt))
	if err != nil {
		t.Fatal(err)
	}

	err = token.VerifyString(str)
	if err != nil {
		t.Fatal(err)
	}
	value := string(token.GetValue())
	t.Logf("token value: %s", value)

	if value != t_value {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", value, t_value)
	}
}

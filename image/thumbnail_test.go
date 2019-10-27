package image

import (
	"bytes"
	"encoding/base64"
	_ "image/jpeg" // test
	// "strings"
	"log"
	"testing"
)

var (
	topts = []ThumbOption{
		{Width: 160, Height: 160, IsFit: true, IsCrop: false},
		{Width: 120, Height: 100, IsFit: true, IsCrop: false},
		{Width: 200, Height: 60, IsFit: true, IsCrop: false},
		{Width: 60, Height: 200, IsFit: true, IsCrop: false},
		{Width: 60, Height: 60, IsFit: true, IsCrop: false},
		{Width: 60, Height: 60, IsFit: true, IsCrop: true},
		{Width: 32, Height: 60, IsFit: true, IsCrop: true},
		{MaxWidth: 60, IsFit: true},
		{MaxHeight: 60, IsFit: true, WriteOption: WriteOption{Format: ".gif"}},
		{MaxHeight: 60, IsFit: true, WriteOption: WriteOption{Format: ".png"}},
	}
)

func TestThumbnails(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	// rd := base64.NewDecoder(base64.StdEncoding, strings.NewReader(jpegData))
	data, err := base64.StdEncoding.DecodeString(jpegData)
	if err != nil {
		t.Errorf("decode err %s", err)
		return
	}

	var buf bytes.Buffer

	// var err error
	for i, topt := range topts {
		err = Thumbnail(bytes.NewReader(data), &buf, topt)
		if err != nil {
			t.Fatalf("Thumbnail '%s' error: %s", topt, err)
		}
		t.Logf("%d thumbnail ok, d %dx%d, f %s", i, topt.Width, topt.Height, topt.Format)
	}
	// t.Fatal("fail")
}

package image

import (
	"encoding/base64"
	"image"
	_ "image/jpeg"
	"strings"
	"testing"
)

var (
	topts = []ThumbOption{
		{Width: 60, Height: 60, IsFit: true, IsCrop: false},
		{Width: 60, Height: 60, IsFit: true, IsCrop: true},
		{Width: 60, Height: 60, IsFit: false, IsCrop: false},
	}
)

func TestThumbnails(t *testing.T) {
	rd := base64.NewDecoder(base64.StdEncoding, strings.NewReader(jpeg_data))
	im, _, err := image.Decode(rd)
	if err != nil {
		t.Fatalf("image decode error: %s", err)
	}

	for _, topt := range topts {
		m, err := ThumbnailImage(im, &topt)
		if err != nil {
			t.Fatalf("ThumbnailImage '%s' error: %s", topt, err)
		}
		mb := m.Bounds()
		t.Logf("thumbnail ok, %dx%d", mb.Dx(), mb.Dy())
	}
	// t.Fatal("fail")
}

package image

import (
	// "imsto"
	"os"
	// "errors"
	"io"
	"log"
)

type KVMapper interface {
	Maps() map[string]interface{}
}

type ImageAttr struct {
	Width   uint32
	Height  uint32
	Quality uint8
	Size    uint32
	Ext     string
}

var attr_keys = []string{"width", "height", "quality", "size", "ext"}

func (ia *ImageAttr) Maps() (m map[string]interface{}) {
	m = map[string]interface{}{
		"width":   ia.Width,
		"height":  ia.Height,
		"quality": ia.Quality,
		"size":    ia.Size,
		"ext":     ia.Ext,
	}
	return
}

type WriteOption struct {
	StripAll bool
	Quality  uint8
}

// export NewImageAttr
func NewImageAttr(w, h uint, q uint8) *ImageAttr {
	return &ImageAttr{uint32(w), uint32(h), uint8(q), uint32(0), ""}
}

// func NewImageAttrByMap(m map[string]interface{}) *ImageAttr {
// 	ia := &ImageAttr{}
// 	for key := range attr_keys {
// 		if v, ok = m[key]; ok {

// 		}
// 	}

// }

type ThumbOption struct {
	Width, Height int
	IsFit         bool
	IsCrop        bool
	wopt          WriteOption
}

type ImageReader interface {
	Open(r io.Reader) error
	GetAttr() *ImageAttr
	Format() string
	Blob(length *uint) []byte
}

type ImageWriter interface {
	SetOption(wopt WriteOption)
	Write(w io.Writer) error
}

type Image interface {
	ImageReader
	ImageWriter
	io.Closer
}

func Open(r io.Reader) (im Image, err error) {

	rr := asReader(r)
	var (
		t   TypeId
		ext string
	)
	t, ext, err = GuessType(r)

	log.Printf("GuessType: %d ext: %s\n", t, ext)

	if t == TYPE_NONE {
		return nil, ErrorFormat
	}

	im = getImageImpl(t)

	if _, ok := r.(*os.File); ok {
		err = im.Open(r)
	} else {
		err = im.Open(rr)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	attr := im.GetAttr()
	attr.Ext = ext

	return im, nil
}

func getImageImpl(t TypeId) (im Image) {
	if t == TYPE_JPEG {
		im = newSimpJPEG()
	} else {
		im = newWandImage()
	}

	return
}

func Thumbnail(r io.Reader, w io.Writer, topt ThumbOption) error {
	im := newWandImage()
	im.Open(r)
	err := im.Thumbnail(topt)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

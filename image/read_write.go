package image

import (
	// "imsto"
	"os"
	// "errors"
	"io"
	"log"
)

type ImageAttr struct {
	Width   uint32
	Height  uint32
	Quality uint8
}

type WriteOption struct {
	StripAll bool
	Quality  uint8
}

// export NewImageAttr
func NewImageAttr(w, h uint, q uint8) *ImageAttr {
	return &ImageAttr{uint32(w), uint32(h), uint8(q)}
}

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
	var data []byte
	data, err = readHead(rr)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(data)
	t := GuessType(&data)

	log.Printf("GuessType: %d\n", t)

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

	return im, nil
}

func getImageImpl(t int) (im Image) {
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

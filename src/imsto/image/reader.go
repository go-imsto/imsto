package image

import (
	// "imsto"
	// "os"
	// "errors"
	"fmt"
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

type Image interface {
	Open(filename string) error
	OpenBlob(blob []byte, length uint) error
	GetAttr() *ImageAttr
	Write(filename string) error
	GetImageBlob() ([]byte, error)
	Close()
}

func Open(filename string) (Image, error) {

	data, err := readHeadFile(filename)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// fmt.Print(data)

	t := GuessType(&data)

	fmt.Printf("GuessType: %d\n", t)

	if t == TYPE_NONE {
		return nil, ErrorFormat
	}

	var im Image
	if t == TYPE_JPEG {
		im = newSimpJPEG()
	} else {
		im = newWandImage()
	}
	err = im.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return im, nil
}

func Thumbnail(src, dest string, topt ThumbOption) error {
	im := newWandImage()
	im.Open(src)
	err := im.Thumbnail(topt)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

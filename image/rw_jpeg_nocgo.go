// +build !cgo

package image

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"log"
)

const jpegFormat = "JPEG"

// jpeg simp_image
type simpJPEG struct {
	m    image.Image
	wopt WriteOption
	// size Size
	*Attr
}

func newSimpJPEG() *simpJPEG {
	o := &simpJPEG{}
	return o
}

func (self *simpJPEG) Format() string {
	return jpegFormat
}

func (self *simpJPEG) Open(r io.Reader) error {
	var (
		format string
		err    error
	)
	self.m, format, err = image.Decode(r)
	if format != "png" {
		log.Println("not a png file")
	}

	rec := self.m.Bounds()
	self.Attr = NewAttr(uint(rec.Max.X), uint(rec.Max.Y), 0) //&Attr{Width: uint32(), Height: uint32()}

	return err
}

func (self *simpJPEG) GetAttr() *Attr {
	return self.Attr
}

func (self *simpJPEG) GetBlob() ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, self.m, &jpeg.Options{
		Quality: int(self.wopt.Quality),
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (self *simpJPEG) SetOption(wopt WriteOption) {
	log.Printf("setOption: q %d, s %v", wopt.Quality, wopt.StripAll)
	if wopt.Quality < MIN_JPEG_QUALITY {
		self.wopt.Quality = MIN_JPEG_QUALITY
	} else {
		self.wopt.Quality = wopt.Quality
	}
	log.Printf("set quality: %d", self.wopt.Quality)

	self.wopt.StripAll = wopt.StripAll
}

func (self *simpJPEG) WriteTo(out io.Writer) error {

	data, err := self.GetBlob()
	if err != nil {
		return err
	}
	// log.Printf("blob %d bytes\n", len(data))

	ret, err := out.Write(data)
	if err != nil {
		// log.Println(err)
		return err
	}

	log.Printf("writed %d\n", ret)

	return nil
}

func (self *simpJPEG) Close() error {
	return nil
}

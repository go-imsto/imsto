package image

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"log"
)

const png_format = "PNG"

type simpPNG struct {
	m    image.Image
	attr *ImageAttr
}

func newSimpPNG() *simpPNG {
	o := &simpPNG{}
	return o
}

func (self *simpPNG) Format() string {
	return png_format
}

func (self *simpPNG) Open(r io.Reader) (err error) {
	var (
		format string
	)
	self.m, format, err = image.Decode(r)
	if format != "png" {
		log.Println("not a png file")
	}

	rec := self.m.Bounds()
	self.attr = NewImageAttr(uint(rec.Max.X), uint(rec.Max.Y), 0) //&ImageAttr{Width: uint32(), Height: uint32()}

	return
}

func (self *simpPNG) GetAttr() *ImageAttr {
	return self.attr
}

func (self *simpPNG) Blob(length *uint) []byte {
	var buf bytes.Buffer
	err := png.Encode(&buf, self.m)
	if err != nil {
		log.Println(err)
	}

	*length = uint(buf.Len())

	return buf.Bytes()
}

func (self *simpPNG) SetOption(wopt WriteOption) {

}

func (self *simpPNG) Write(w io.Writer) error {
	var length uint
	data := self.Blob(&length)
	size, err := w.Write(data)
	if err != nil {
		log.Println(err)
		return err
	}

	if size != int(length) {
		log.Println("write error", size)
	}
	return nil

}

func (self *simpPNG) Close() error {
	return nil
}

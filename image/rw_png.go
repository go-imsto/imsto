package image

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"log"
)

const pngFormat = "PNG"

type simpPNG struct {
	m image.Image
	*Attr
}

func newSimpPNG() *simpPNG {
	o := &simpPNG{}
	return o
}

func (self *simpPNG) Format() string {
	return pngFormat
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
	self.Attr = NewAttr(uint(rec.Max.X), uint(rec.Max.Y), 0) //&Attr{Width: uint32(), Height: uint32()}

	return
}

func (self *simpPNG) GetAttr() *Attr {
	return self.Attr
}

func (self *simpPNG) GetBlob() ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, self.m)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (self *simpPNG) SetOption(wopt WriteOption) {

}

func (self *simpPNG) WriteTo(w io.Writer) error {
	data, err := self.GetBlob()
	if err != nil {
		log.Println(err)
		return err
	}

	size, err := w.Write(data)
	if err != nil {
		log.Println(err)
		return err
	}

	if size != len(data) {
		log.Println("write error", size)
	}
	return nil

}

func (self *simpPNG) Close() error {
	return nil
}

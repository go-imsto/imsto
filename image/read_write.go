package image

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"

	"github.com/liut/jpegquality"
)

const (
	formatGIF  = "gif"
	formatJPEG = "jpeg"
	formatPNG  = "png"
)

var (
	mtypes = map[string]string{
		formatGIF:  "image/gif",
		formatJPEG: "image/jpeg",
		formatPNG:  "image/png",
	}
)

// Image ...
type Image struct {
	m image.Image
	*Attr
	Format string
}

// Open ...
func Open(rs io.ReadSeeker) (*Image, error) {

	cw := new(CountWriter)
	m, format, err := image.Decode(io.TeeReader(rs, cw))
	if err != nil {
		return nil, err
	}

	pt := m.Bounds().Max
	attr := NewAttr(uint(pt.X), uint(pt.Y), format)
	if mt, ok := mtypes[format]; ok {
		attr.Mime = mt
	}
	attr.Size = Size(cw.Len())
	if format == formatJPEG {
		jr, err := jpegquality.New(rs)
		if err != nil {
			return nil, err
		}
		attr.Quality = Quality(jr.Quality())
	}
	return &Image{
		m:      m,
		Attr:   attr,
		Format: format,
	}, nil
}

// WriteOption ...
type WriteOption struct {
	Format   string
	StripAll bool
	Quality  Quality
}

// SaveTo ...
func (im *Image) SaveTo(w io.Writer, opt WriteOption) error {
	if opt.Format == "" {
		opt.Format = im.Format
	}
	return SaveTo(w, im.m, opt)
}

// SaveTo ...
func SaveTo(w io.Writer, m image.Image, opt WriteOption) error {
	switch opt.Format {
	case formatJPEG:
		qlt := int(opt.Quality)
		if qlt == 0 {
			qlt = MinJPEGQuality
		}
		return jpeg.Encode(w, m, &jpeg.Options{Quality: qlt})
	case formatGIF:
		return gif.Encode(w, m, &gif.Options{
			NumColors: 256,
			Quantizer: nil,
			Drawer:    nil,
		})
	case formatPNG:
		return png.Encode(w, m)
	}

	log.Printf("opt %v", opt)
	return ErrorFormat
}

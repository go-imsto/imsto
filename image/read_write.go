package image

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"

	"github.com/chai2010/webp"

	"github.com/liut/jpegquality"
)

// consts
const (
	MinJPEGQuality = jpeg.DefaultQuality // 75
	MinWebpQuality = 80
)

const (
	formatGIF  = "gif"
	formatJPEG = "jpeg"
	formatPNG  = "png"
	formatWEBP = "webp"
)

var (
	mtypes = map[string]string{
		formatGIF:  "image/gif",
		formatJPEG: "image/jpeg",
		formatPNG:  "image/png",
		formatWEBP: "image/webp",
	}
)

// Image ...
type Image struct {
	m image.Image
	*Attr
	Format string
	rs     io.ReadSeeker
	rn     int // read length
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
		rs:     rs,
		rn:     cw.Len(),
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
	var buf bytes.Buffer
	n, err := SaveTo(&buf, im.m, opt)
	if err != nil {
		return err
	}
	var nn int64
	if n > im.rn {
		log.Printf("saved %d, im size %d", n, im.rn)
		im.rs.Seek(0, 0)
		nn, err = io.Copy(w, im.rs)
	} else {
		nn, err = io.Copy(w, &buf)
	}
	log.Printf("copied %d bytes", nn)
	return err
}

// SaveTo ...
func SaveTo(w io.Writer, m image.Image, opt WriteOption) (n int, err error) {
	cw := new(CountWriter)
	defer func() { n = cw.Len() }()
	w = io.MultiWriter(w, cw)
	switch opt.Format {
	case formatJPEG:
		qlt := int(opt.Quality)
		if qlt == 0 {
			qlt = MinJPEGQuality
		}
		err = jpeg.Encode(w, m, &jpeg.Options{Quality: qlt})
		return
	case formatGIF:
		err = gif.Encode(w, m, &gif.Options{
			NumColors: 256,
			Quantizer: nil,
			Drawer:    nil,
		})
		return
	case formatPNG:
		err = png.Encode(w, m)
		return
	case formatWEBP:
		qlt := int(opt.Quality)
		if qlt == 0 {
			qlt = MinWebpQuality
		}
		err = webp.Encode(w, m, &webp.Options{Quality: float32(qlt)})
		return
	}

	log.Printf("opt %v", opt)
	err = ErrorFormat
	return
}

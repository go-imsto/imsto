package image

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"io"
	"log"
	"mime"
	"os"
)

type Dimension uint32
type Size uint32
type Quality uint8

type Attr struct {
	Width   Dimension `json:"width"`
	Height  Dimension `json:"height"`
	Quality Quality   `json:"quality,omitempty"`
	Size    Size      `json:"size"`
	Ext     string    `json:"ext,omitempty"`
	Mime    string    `json:"mime,omitempty"`
	Name    string    `json:"-"`
}

func (a Attr) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"width":  a.Width,
		"height": a.Height,
		"ext":    a.Ext,
		"mime":   a.Mime,
	}
	if a.Quality > 0 {
		m["quality"] = a.Quality
	}
	return m
}

func (a *Attr) FromMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	if a == nil {
		*a = Attr{}
	}
	if v, ok := m["width"]; ok {
		if vv, ok := v.(uint32); ok {
			a.Width = Dimension(vv)
		}
	}
	if v, ok := m["height"]; ok {
		if vv, ok := v.(uint32); ok {
			a.Height = Dimension(vv)
		}
	}
	if v, ok := m["ext"]; ok {
		if vv, ok := v.(string); ok {
			a.Ext = vv
		}
	}
}

// var attr_keys = []string{"width", "height", "quality", "size", "ext"}

func (a *Attr) Scan(b interface{}) error {
	if b == nil {
		return nil
	}
	return json.Unmarshal(b.([]byte), a)
}

func (a Attr) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type WriteOption struct {
	StripAll bool
	Quality  Quality
}

// export NewAttr
func NewAttr(w, h uint, q uint8) *Attr {
	return &Attr{
		Width:   Dimension(w),
		Height:  Dimension(h),
		Quality: Quality(q),
	}
}

type ImageReader interface {
	Open(r io.Reader) error
	GetAttr() *Attr
	Format() string
}

type ImageWriter interface {
	SetOption(wopt WriteOption)
	GetBlob() ([]byte, error)
	WriteTo(w io.Writer) error
}

type Image interface {
	ImageReader
	ImageWriter
	io.Closer
}

func Open(r io.Reader) (im Image, err error) {

	var (
		t    TypeId
		ext  string
		size Size
	)
	t, ext, err = GuessType(r)

	log.Printf("GuessType: %d ext: %s\n", t, ext)

	if t == TYPE_NONE {
		return nil, ErrorFormat
	}

	im = getImageImpl(t)

	if rr, ok := r.(io.Seeker); ok {
		rr.Seek(0, 0)
	}

	if f, ok := r.(*os.File); ok {
		log.Println("rw: open from file")
		// f.Seek(0, 0)
		var fi os.FileInfo
		if fi, err = f.Stat(); err != nil {
			size = Size(fi.Size())
		}
		err = im.Open(f)
	} else if rr, ok := r.(*bytes.Reader); ok {
		// rr.Seek(0, 0)
		size = Size(rr.Len())
		log.Printf("rw: open from buf, size: %d", rr.Len())
		err = im.Open(rr)
	} else { // 目前只支持从文件或二进制数据读取
		// log.Println("open from other", reflect.TypeOf(r))
		// rr := bufio.NewReader(r)
		// rr.Reset()
		// err = im.Open(rr)
		log.Fatalf("rw: unsupport reader %v", r)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	ia := im.GetAttr()
	ia.Ext = ext
	ia.Mime = mime.TypeByExtension(ext)
	if size > Size(0) {
		ia.Size = size
	}

	return
}

func getImageImpl(t TypeId) (im Image) {
	if t == TYPE_JPEG {
		im = newSimpJPEG()
	} else if t == TYPE_PNG {
		im = newSimpPNG()
	} else {
		log.Panicf("rw: unsupport type %s", t)
		// im = newWandImage()
	}

	return
}

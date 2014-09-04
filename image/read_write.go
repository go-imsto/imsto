package image

import (
	"bytes"
	"io"
	"log"
	"mime"
	"os"
	"reflect"
	"wpst.me/calf/db"
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

// var attr_keys = []string{"width", "height", "quality", "size", "ext"}

func (ia *Attr) Hstore() db.Hstore {
	return db.StructToHstore(*ia)
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
		log.Panicln("rw: unsupport reader ", reflect.TypeOf(r))
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

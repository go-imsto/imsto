package image

import (
	"bytes"
	"calf/db"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"reflect"
)

type Dimension uint32
type Size uint32
type Quality uint8

type ImageAttr struct {
	Width   Dimension `json:"width"`
	Height  Dimension `json:"height"`
	Quality Quality   `json:"quality"`
	Size    Size      `json:"size"`
	Ext     string    `json:"ext,omitempty"`
	Mime    string    `json:"mime,omitempty"`
}

// var attr_keys = []string{"width", "height", "quality", "size", "ext"}

func (ia *ImageAttr) Hstore() db.Hstore {
	return db.StructToHstore(*ia)
}

type WriteOption struct {
	StripAll bool
	Quality  Quality
}

// export NewImageAttr
func NewImageAttr(w, h uint, q uint8) *ImageAttr {
	return &ImageAttr{Dimension(w), Dimension(h), Quality(q), Size(0), "", ""}
}

type ThumbOption struct {
	Width, Height       uint
	MaxWidth, MaxHeight uint
	IsFit               bool
	IsCrop              bool
	wopt                WriteOption
}

type ImageReader interface {
	Open(r io.Reader) error
	GetAttr() *ImageAttr
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
		im = newWandImage()
	}

	return
}

func Thumbnail(r io.Reader, w io.Writer, topt ThumbOption) error {
	im := newWandImage()
	im.Open(r)
	err := im.Thumbnail(topt)

	if err != nil {
		return err
	}

	err = im.WriteTo(w)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func ThumbnailFile(src, dest string, topt ThumbOption) (err error) {
	var in *os.File
	in, err = os.Open(src)
	if err != nil {
		log.Print(err)
		return
	}
	defer in.Close()
	im := newWandImage()
	im.Open(in)
	err = im.Thumbnail(topt)
	if err != nil {
		return err
	}

	dir := path.Dir(dest)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}

	return im.WriteFile(dest)

	// out, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	// if err != nil {
	// 	log.Print(err)
	// 	return
	// }
	// defer out.Close()

	// return Thumbnail(in, out, topt)
}

package image

/*
#cgo linux CFLAGS: -I/usr/include
#cgo linux LDFLAGS: -ljpeg -L/usr/lib
#cgo darwin CFLAGS: -I/opt/local/include -DIM_DEBUG
#cgo darwin LDFLAGS: -ljpeg -L/opt/local/lib

// debug CFLAGS:
// -DIM_DEBUG for simp_image output
// -DJPEG_DEBUG for jpeg output, depend IM_DEBUG

#include "c-jpeg.h"

static unsigned char ** makeCharArray(int size) {
    return calloc(sizeof(unsigned char*), size);
}

//static void setArrayString(char **a, char *s, int n) {
//    a[n] = s;
//}

//static void freeCharArray(char **a, int size) {
//    int i;
//    for (i = 0; i < size; i++)
//        free(a[i]);
//    free(a);
//}

*/
import "C"

import (
	// "fmt"
	// "imsto"
	"io"
	"os"
	// "strings"
	// "bufio"
	"errors"
	"io/ioutil"
	"log"
	"unsafe"
)

const jpeg_format = "JPEG"

// jpeg simp_image
type simpJPEG struct {
	si   *C.Simp_Image
	attr *ImageAttr
	wopt WriteOption
	size uint32
}

func newSimpJPEG() *simpJPEG {
	o := &simpJPEG{}
	return o
}

func (self *simpJPEG) Format() string {
	return jpeg_format
}

func (self *simpJPEG) Open(r io.Reader) error {

	var si *C.Simp_Image

	if f, ok := r.(*os.File); ok {
		log.Println("open file reader")
		f.Seek(0, 0)
		cmode := C.CString("rb")
		defer C.free(unsafe.Pointer(cmode))
		infile := C.fdopen(C.int(f.Fd()), cmode)
		//defer C.fclose(infile)
		si = C.simp_open_stdio(infile)
		if si == nil {
			return errors.New("simp_open_stdio failed")
		}
		fi, _ := f.Stat()
		self.size = uint32(fi.Size())
	} else {
		blob, err := ioutil.ReadAll(r)

		if err != nil {
			// log.Println(err)
			return err
		}

		size := len(blob)
		log.Printf("jpeg blob head: %x", blob[0:8])
		log.Printf("jpeg blob tail: %x", blob[size-2:size])

		log.Printf("open mem buf len %d\n", size)
		p := (*C.uchar)(unsafe.Pointer(&blob[0]))

		si = C.simp_open_mem(p, C.uint(size))
		if si == nil {
			return errors.New("simp_open_mem failed")
		}

		self.size = uint32(size)
	}

	self.si = si
	w := C.simp_get_width(self.si)
	h := C.simp_get_height(self.si)
	q := C.simp_get_quality(self.si)
	log.Printf("image open, w: %d, h: %d, q: %d", w, h, q)

	self.attr = NewImageAttr(uint(w), uint(h), uint8(q))
	return nil
}

func (self *simpJPEG) Close() error {
	if self.si != nil {
		C.simp_close(self.si)
		self.si = nil
	}
	return nil
}

func (self *simpJPEG) GetAttr() *ImageAttr {
	return self.attr
}

func (self *simpJPEG) SetOption(wopt WriteOption) {
	self.wopt = wopt
	log.Printf("setOption: q %d, s %v", self.wopt.Quality, self.wopt.StripAll)
	if self.wopt.Quality > 0 && self.attr.Quality > self.wopt.Quality {
		C.simp_set_quality(self.si, C.int(self.wopt.Quality))
		self.attr.Quality = self.wopt.Quality
		log.Printf("set quality: %d", self.wopt.Quality)
	}

}

func (self *simpJPEG) WriteTo(out io.Writer) error {
	if f, ok := out.(*os.File); ok {
		log.Printf("write a file %s\n", f.Name())

		ocmode := C.CString("wb")
		defer C.free(unsafe.Pointer(ocmode))
		outfile := C.fdopen(C.int(f.Fd()), ocmode)

		r := C.simp_output_file(self.si, outfile)
		if !r {
			log.Println("simp out file error")
			return errors.New("output error")
		}

	} else {
		log.Println("write to buf")

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
	}

	return nil
}

func (self *simpJPEG) GetBlob() ([]byte, error) {
	cblob := (**C.uchar)(C.makeCharArray(C.int(self.size)))
	*cblob = nil
	defer C.free(unsafe.Pointer(cblob))

	var size = 0
	r := C.simp_output_mem(self.si, cblob, (*C.ulong)(unsafe.Pointer(&size)))

	if !r || *cblob == nil {
		log.Println("simp out mem error")
		// return errors.New("output error")
		return nil, errors.New("output error")
	}
	log.Printf("c output %d bytes\n", size)
	return C.GoBytes(unsafe.Pointer(*cblob), C.int(size)), nil
}

func OptimizeJpeg(src, dest *os.File, wopt *WriteOption) error {
	var (
		im              *simpJPEG
		st_i, st_o      os.FileInfo
		err             error
		insize, outsize int64
		ratio           float64
	)

	st_i, err = src.Stat()
	if err != nil {
		// log.Println(err)
		return err
	}

	im = newSimpJPEG()

	err = im.Open(src)
	if err != nil {
		// log.Println(err)
		return err
	}
	defer im.Close()
	im.SetOption(*wopt)

	err = im.WriteTo(dest)
	if err != nil {
		// log.Println(err)
		return err
	}

	st_o, err = dest.Stat()
	if err != nil {
		// log.Println(err)
		return err
	}

	insize = st_i.Size()
	outsize = st_o.Size()
	ratio = float64(insize-outsize) * 100.0 / float64(insize)
	log.Printf("%d --> %d bytes (%0.2f%%), optimized.\n", insize, outsize, ratio)

	return nil
}

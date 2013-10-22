package image

/*
#cgo CFLAGS: -I/opt/local/include -DIM_DEBUG
#cgo LDFLAGS: -ljpeg -L/opt/local/lib
// liut: add CFLAGS "-DIM_DEBUG" for debug output

#include "c-jpeg.h"
*/
import "C"

import (
	"fmt"
	// "imsto"
	"os"
	// "strings"
	"errors"
	"unsafe"
)

// jpeg simp_image
type simpJPEG struct {
	si   *C.Simp_Image
	attr *ImageAttr
	opt  *WriteOption
}

func newSimpJPEG() *simpJPEG {
	o := &simpJPEG{}
	return o
}

func (self *simpJPEG) Open(filename string) error {
	// fmt.Printf("simpJPEG.Open %s\n", filename)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	cmode := C.CString("rb")
	defer C.free(unsafe.Pointer(cmode))
	infile := C.fdopen(C.int(file.Fd()), cmode)

	// var ia C.struct_jpeg_attr
	// r := C.read_jpeg_file(infile, &ia)

	var si *C.Simp_Image
	si = C.simp_open_stdio(infile)

	if si == nil {
		// fmt.Printf("simp_open_stdio failed\n")
		return errors.New("simp_open_stdio failed")
	}

	self.si = si

	// ia_ptr->width = (UINT16)im->in.w;
	// ia_ptr->height = (UINT16)im->in.h;
	// ia_ptr->quality = (UINT8)im->in.q;
	// simp_close(im);
	// fmt.Println(ia)
	// fmt.Printf("C.Read_jpeg_file %d\n", r)
	self.attr = NewImageAttr(uint(si.in.w), uint(si.in.h), uint8(si.in.q))
	return nil
}

func (self *simpJPEG) Close() {
	C.simp_close(self.si)
	self.si = nil
}

func (self *simpJPEG) OpenBlob(blob []byte, length uint) error {
	// TODO:

	return nil
}

func (self *simpJPEG) GetAttr() *ImageAttr {
	return self.attr
}

func (self *simpJPEG) Write(filename string) error {
	// TODO:

	return nil
}

func (self *simpJPEG) GetImageBlob() ([]byte, error) {
	// TODO:

	return nil, nil
}

func ReadJpegImage(file *os.File) (*ImageAttr, error) {
	cmode := C.CString("rb")
	defer C.free(unsafe.Pointer(cmode))
	infile := C.fdopen(C.int(file.Fd()), cmode)

	var ia C.struct_jpeg_attr
	r := C.read_jpeg_file(infile, &ia)
	// fmt.Println(ia)
	fmt.Printf("C.Read_jpeg_file %d\n", r)
	return NewImageAttr(uint(ia.width), uint(ia.height), uint8(ia.quality)), nil
}

func ReadJpeg(filename string) (*ImageAttr, error) {
	file, err := os.Open(filename)

	if err != nil {
		return &ImageAttr{}, err
	}

	defer file.Close()

	return ReadJpegImage(file)

	// fmt.Println("ReadJpeg: " + filename)
	// csfilename := C.CString(filename)
	// defer C.free(unsafe.Pointer(csfilename))
	// // var cinfo C.j_decompress_ptr
	// var ia C.jpeg_attr
	// r := C.read_jpeg_file(csfilename, &ia)
	// fmt.Println(ia)
	// fmt.Printf("C.Read_jpeg_file %d\n", r)
	// return NewImageAttr(ia.width, ia.height, ia.quality), nil
}

func RewriteJpeg(src, dest *os.File, wo *WriteOption) error {
	var (
		st_i, st_o      os.FileInfo
		err             error
		insize, outsize int64
		ratio           float64
	)
	st_i, err = src.Stat()
	if err != nil {
		return err
	}
	icmode := C.CString("rb")
	defer C.free(unsafe.Pointer(icmode))
	infile := C.fdopen(C.int(src.Fd()), icmode)

	ocmode := C.CString("wb")
	defer C.free(unsafe.Pointer(ocmode))
	outfile := C.fdopen(C.int(dest.Fd()), ocmode)

	var opt C.struct_jpeg_option

	opt.quality = C.UINT8(wo.Quality)
	if wo.StripAll {
		opt.strip_all = C.boolean(1)
	} else {
		opt.strip_all = C.boolean(0)
	}

	r := C.write_jpeg_file(infile, outfile, &opt)

	fmt.Printf("C.write_jpeg_file %d\n", r)
	st_o, err = dest.Stat()
	if err != nil {
		return err
	}
	insize = st_i.Size()
	outsize = st_o.Size()
	ratio = float64(insize-outsize) * 100.0 / float64(insize)
	fmt.Printf("%d --> %d bytes (%0.2f%%), optimized.\n", insize, outsize, ratio)
	// fmt.Printf("src size: %d, dest size: %d \n", st_i.Size(), st_o.Size())

	return nil
}

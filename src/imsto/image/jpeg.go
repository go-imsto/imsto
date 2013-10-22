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
	"unsafe"
)

type ImageAttr struct {
	Width   uint32
	Height  uint32
	Quality uint8
}

type WriteOption struct {
	StripAll bool
	Quality  uint8
}

// export NewImageAttr
func NewImageAttr(w, h C.UINT16, q C.UINT8) *ImageAttr {
	return &ImageAttr{uint32(w), uint32(h), uint8(q)}
}

// custom error output, unfinished
func my_error_exit(cinfo C.j_common_ptr) {
	fmt.Println(cinfo.err)
}

func ReadJpegImage(file *os.File) (*ImageAttr, error) {
	cmode := C.CString("rb")
	defer C.free(unsafe.Pointer(cmode))
	infile := C.fdopen(C.int(file.Fd()), cmode)

	var ia C.struct_jpeg_attr
	r := C.read_jpeg_file(infile, &ia)
	// fmt.Println(ia)
	fmt.Printf("C.Read_jpeg_file %d\n", r)
	return NewImageAttr(ia.width, ia.height, ia.quality), nil
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

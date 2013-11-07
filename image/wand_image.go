package image

/*
#cgo linux pkg-config: MagickWand
#cgo darwin pkg-config: MagickWand-6.Q16

#include <wand/MagickWand.h>

char *MagickGetPropertyName(char **properties, size_t index) {
  return properties[index];
}

*/
import "C"

import (
	"fmt"
	// "imsto"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unsafe"
)

func cargv(b [][]byte) **C.char {
	outer := make([]*C.char, len(b)+1)
	for i, inner := range b {
		outer[i] = C.CString(string(inner))
	}
	return (**C.char)(unsafe.Pointer(&outer[0])) // void bar(**char) {...}
}

// Image object
type wandImpl struct {
	wand *C.MagickWand

	filename string
	width    string
	height   string

	wopt WriteOption
}

func init() {
	C.MagickWandGenesis()
}

// Returns a new image object.
func newWandImage() *wandImpl {
	self := &wandImpl{}

	self.wand = C.NewMagickWand()

	return self
}

// Opens an image file, returns nil on success, error otherwise.
func (self *wandImpl) Open(r io.Reader) error {
	// stat, err := os.Stat(filename)

	// if err != nil {
	// 	return err
	// }

	// if stat.IsDir() == true {
	// 	return fmt.Errorf(`Could not open file "%s": it's a directory!`, filename)
	// }
	var status C.MagickBooleanType

	if f, ok := r.(*os.File); ok {
		f.Seek(0, 0)
		// cfilename := C.CString(filename)
		// status := C.MagickReadImage(self.wand, cfilename)
		// C.free(unsafe.Pointer(cfilename))
		cmode := C.CString("rb")
		defer C.free(unsafe.Pointer(cmode))
		file := C.fdopen(C.int(f.Fd()), cmode)
		defer C.fclose(file)
		status = C.MagickReadImageFile(self.wand, file)
		self.filename = f.Name()
	} else {
		// var blob []byte
		blob, err := ioutil.ReadAll(f)

		if err != nil {
			log.Println(err)
			return err
		}

		status = C.MagickReadImageBlob(self.wand, unsafe.Pointer(&blob[0]), C.size_t(len(blob)))
	}
	if status == C.MagickFalse {
		return fmt.Errorf(`Could not open image: %s`, self.Error())
	}

	// self.filename = filename

	// _im := C.GetImageFromMagickWand(self.wand)

	// fmt.Println(_im.columns)

	return nil
}

// Reads an image or image sequence from a blob.
func (self *wandImpl) OpenBlob(blob []byte) error {
	status := C.MagickReadImageBlob(self.wand, unsafe.Pointer(&blob[0]), C.size_t(len(blob)))

	if status == C.MagickFalse {
		return fmt.Errorf(`Could not open image from blob: %s`, self.Error())
	}

	return nil
}

func (self *wandImpl) GetAttr() *ImageAttr {
	return NewImageAttr(self.Width(), self.Height(), self.Quality())
}

// Returns the format of a particular image in a sequence.
func (self *wandImpl) Format() string {
	return C.GoString(C.MagickGetImageFormat(self.wand))
}

// Sets the format of a particular image
func (self *wandImpl) SetFormat(format string) error {
	cformat := C.CString(format)
	defer C.free(unsafe.Pointer(cformat))

	if C.MagickSetImageFormat(self.wand, cformat) == C.MagickFalse {
		return fmt.Errorf("Could not set format: %s", self.Error())
	}

	return nil
}

func (self *wandImpl) SetOption(wopt WriteOption) {
	self.wopt = wopt
	self.SetQuality(wopt.Quality)
}

// Get image data as a byte array.
func (self *wandImpl) GetImageBlob() ([]byte, error) {
	size := C.size_t(0)
	p := unsafe.Pointer(C.MagickGetImageBlob(self.wand, &size))
	if size == 0 {
		return nil, fmt.Errorf("Could not get image blob \n")
	}

	blob := C.GoBytes(p, C.int(size))

	C.MagickRelinquishMemory(p)

	return blob, nil
}

// Converts the current image into a thumbnail of the specified
// width and height preserving ratio. It uses Crop() to clip the
// image to the specified area.
//
// If width or height are bigger than the current image, a centered
// thumbnail will be produced.
//
// Is width and height are smaller than the current image, the image
// will be resized and cropped, if needed.
func (self *wandImpl) Thumbnail(topt ThumbOption) error {

	// TODO:

	return nil
}

// Returns image' width.
func (self *wandImpl) Width() uint {
	return uint(C.MagickGetImageWidth(self.wand))
}

// Returns image' height.
func (self *wandImpl) Height() uint {
	return uint(C.MagickGetImageHeight(self.wand))
}

// Writes image to a file, returns nil on success.
func (self *wandImpl) Write(out io.Writer) error {
	if f, ok := out.(*os.File); ok {

		cmode := C.CString("w+")
		defer C.free(unsafe.Pointer(cmode))
		file := C.fdopen(C.int(f.Fd()), cmode)
		defer C.fclose(file)
		success := C.MagickWriteImageFile(self.wand, file)

		if success == C.MagickFalse {
			return fmt.Errorf("Could not write: %s", self.Error())
		}

	} else {
		var size *uint
		blob := self.Blob(size)

		log.Printf("blob %d bytes\n", size)
		out.Write(blob)
	}
	return nil
}

// Implements direct to memory image formats. It returns the image as a blob
func (self *wandImpl) Blob(length *uint) []byte {
	ptr := unsafe.Pointer(C.MagickGetImageBlob(self.wand, (*C.size_t)(unsafe.Pointer(length))))
	data := C.GoBytes(ptr, C.int(*length))
	C.MagickRelinquishMemory(ptr)
	return data
}

// Changes the size of the image, returns true on success.
func (self *wandImpl) Resize(width uint, height uint) error {
	success := C.MagickResizeImage(self.wand, C.size_t(width), C.size_t(height), C.GaussianFilter, C.double(1.0))

	if success == C.MagickFalse {
		return fmt.Errorf("Could not resize: %s", self.Error())
	}

	return nil
}

// Returns the compression quality of the image. Ranges from 1 (lowest) to 100 (highest).
func (self *wandImpl) Quality() uint8 {
	return uint8(C.MagickGetImageCompressionQuality(self.wand))
}

// Changes the compression quality of the canvas. Ranges from 1 (lowest) to 100 (highest).
func (self *wandImpl) SetQuality(quality uint8) error {
	success := C.MagickSetImageCompressionQuality(self.wand, C.size_t(quality))

	if success == C.MagickFalse {
		return fmt.Errorf("Could not set compression quality: %s", self.Error())
	}

	return nil
}

// Destroys image.
func (self *wandImpl) Destroy() error {

	if self.wand == nil {
		return fmt.Errorf("Nothing to destroy")
	} else {
		C.DestroyMagickWand(self.wand)
		self.wand = nil
	}

	return nil
}

func (self *wandImpl) Close() error {
	self.Destroy()
	return nil
}

// Returns all metadata keys from the currently loaded image.
func (self *wandImpl) Metadata() map[string]string {
	var n C.size_t
	var i C.size_t

	var value *C.char
	var key *C.char

	data := make(map[string]string)

	cplist := C.CString("*")

	properties := C.MagickGetImageProperties(self.wand, cplist, &n)

	C.free(unsafe.Pointer(cplist))

	for i = 0; i < n; i++ {
		key = C.MagickGetPropertyName(properties, C.size_t(i))
		value = C.MagickGetImageProperty(self.wand, key)

		data[strings.Trim(C.GoString(key), " ")] = strings.Trim(C.GoString(value), " ")

		C.MagickRelinquishMemory(unsafe.Pointer(value))
		C.MagickRelinquishMemory(unsafe.Pointer(key))
	}

	return data
}

// Returns the latest error reported by the MagickWand API.
func (self *wandImpl) Error() error {
	var t C.ExceptionType
	ptr := C.MagickGetException(self.wand, &t)
	message := C.GoString(ptr)
	C.MagickClearException(self.wand)
	C.MagickRelinquishMemory(unsafe.Pointer(ptr))
	return fmt.Errorf(message)
}

func Finalize() {
	C.MagickWandTerminus()
}

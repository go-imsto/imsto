package image

import (
	"imsto"
)

func GetImageAttr(filename string) *imsto.ImageAttr {

	//args := [][]byte{[]byte("-format"), []byte("%m %w %h %Q"), []byte(filename)}
	// args := [][]byte{[]byte("-verbose"), []byte(filename)}
	//argc := len(args)

	//argv := cargv(args)

	// defer C.free(unsafe.Pointer(argv))
	//r := C.IdentifyMain(C.int(argc), argv)

	//fmt.Println(r)

	// imagick.Initialize()
	// Schedule cleanup
	// defer imagick.Terminate()
	// var err error

	// image := New()

	// Opening some image from disk.
	// err := image.Open(filename)

	// mw := imagick.NewMagickWand()
	// Schedule cleanup
	// defer image.Destroy()

	// if err != nil {
	// 	panic(err)
	// }

	// Get original logo size
	// width := mw.GetImageWidth()
	// height := mw.GetImageHeight()
	// quality := mw.GetImageCompressionQuality()
	// format := mw.GetImageFormat()

	// fmt.Println(quality)
	// fmt.Println(image.Metadata())
	// fmt.Println(image.Error())

	// ia := &imsto.ImageAttr{image.Width(), image.Height(), image.Quality(), image.Format()}

	// return ia
	return &imsto.ImageAttr{}
}

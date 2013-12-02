package main

import (
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

var (
	resCases = []struct {
		interp resize.InterpolationFunction
		name   string
	}{
		{resize.NearestNeighbor, "test_resize.NearestNeighbor"},
		{resize.Bilinear, "test_resize.Bilinear"},
		{resize.Bicubic, "test_resize.Bicubic"},
		{resize.MitchellNetravali, "test_resize.MitchellNetravali"},
		{resize.Lanczos2Lut, "test_resize.Lanczos2Lut"},
		{resize.Lanczos2, "test_resize.Lanczos2"},
		{resize.Lanczos3Lut, "test_resize.Lanczos3Lut"},
		{resize.Lanczos3, "test_resize.Lanczos3"},
	}
)

func main() {
	// open "test.jpg"
	file, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("format: %s", format)
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	for i, c := range resCases {
		log.Printf("%d resize to: %s", i, c.name)
		m := resize.Resize(400, 0, img, c.interp)
		encode(m, format, c.name+"."+format)
	}

}

func encode(m image.Image, format, filename string) {

	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	switch format {
	case "jpeg":
		err = jpeg.Encode(out, m, &jpeg.Options{88})
		break
	case "png":
		err = png.Encode(out, m)
		break
	}

	if err != nil {
		log.Fatal(err)
	}
}

package image

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path"
)

const (
	MIN_JPEG_QUALITY = jpeg.DefaultQuality // 75
)

type Position uint8

const (
	BottomRight Position = iota
	TopLeft
	TopRight
	BottomLeft
	Center
	Golden
)

type ThumbOption struct {
	Width, Height       uint
	MaxWidth, MaxHeight uint
	IsFit               bool
	IsCrop              bool
	CropX, CropY        int
	ctWidth, ctHeight   uint // for crop temporary
	WriteOption
}

func (topt ThumbOption) String() string {
	return fmt.Sprintf("%dx%d q%d %v %v", topt.Width, topt.Height, topt.Quality, topt.IsFit, topt.IsCrop)
}

func (topt *ThumbOption) calc(ow, oh uint) error {
	if topt.Quality < MIN_JPEG_QUALITY {
		topt.Quality = MIN_JPEG_QUALITY
	}

	if topt.Width >= ow && topt.Height >= oh {
		return fmt.Errorf("%dx%d is too big", topt.Width, topt.Height)
	}

	if topt.IsFit {
		if topt.IsCrop {
			ratio_x := float32(topt.Width) / float32(ow)
			ratio_y := float32(topt.Height) / float32(oh)

			if ratio_x > ratio_y {
				topt.ctWidth = topt.Width
				topt.ctHeight = uint(ratio_x * float32(oh))
			} else {
				topt.ctHeight = topt.Height
				topt.ctWidth = uint(ratio_y * float32(ow))
			}
			// :resize

			if topt.ctWidth == topt.Width && topt.ctHeight == topt.Height {
				return fmt.Errorf("crop %dx%d is too big", topt.Width, topt.Height)
			}

			topt.CropX = int(float32(topt.ctWidth-topt.Width) / 2)
			topt.CropY = int(float32(topt.ctHeight-topt.Height) / 2)

			// log.Printf("cropX: %d, cropY: %d", cropX, cropY)
			// :crop

		} else {

			rel := float32(ow) / float32(oh)
			if topt.MaxWidth > 0 && topt.MaxWidth <= ow {
				topt.Width = topt.MaxWidth
				topt.Height = uint(float32(topt.Width) / rel)
			} else if topt.MaxHeight > 0 && topt.MaxHeight <= oh {
				topt.Height = topt.MaxHeight
				topt.Width = uint(float32(topt.Height) * rel)
			} else {
				bounds := float32(topt.Width) / float32(topt.Height)
				if rel >= bounds {
					topt.Height = uint(float32(topt.Width) / rel)
				} else {
					topt.Width = uint(float32(topt.Height) * rel)
				}
			}
		}
	}
	return nil
}

func ThumbnailImage(img image.Image, topt *ThumbOption) (image.Image, error) {

	ob := img.Bounds()
	ow := uint(ob.Dx())
	oh := uint(ob.Dy())

	if topt.Width >= ow && topt.Height >= oh {
		return img, nil
	}

	err := topt.calc(ow, oh)
	if err != nil {
		return nil, err
	}
	// log.Printf("thumb option: %s", topt)
	if topt.IsFit {
		if topt.IsCrop {
			buf := resize.Resize(topt.ctWidth, topt.ctHeight, img, resize.Bicubic)
			dst := image.NewRGBA(image.Rect(0, 0, int(topt.Width), int(topt.Height)))
			pt := image.Point{topt.CropX, topt.CropY}
			draw.Draw(dst, dst.Bounds(), buf, pt, draw.Src)
			return dst, nil
		}
	}
	m := resize.Resize(topt.Width, topt.Height, img, resize.Bicubic)
	return m, nil
}

func Thumbnail(r io.Reader, w io.Writer, topt ThumbOption) error {
	var err error
	im, _, err := image.Decode(r)
	if err != nil {
		log.Printf("ThumbnailIo image decode error: %s", err)
		return err
	}

	m, err := ThumbnailImage(im, &topt)
	if err != nil {
		return err
	}
	// log.Printf("thumb option: %s", topt)
	err = jpeg.Encode(w, m, &jpeg.Options{int(topt.Quality)})

	// err = im.WriteTo(w)

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
	// im := newWandImage()
	// im.Open(in)
	// err = im.Thumbnail(topt)
	// if err != nil {
	// 	log.Printf("im.Thumbnail error: %s", err)
	// 	return err
	// }

	dir := path.Dir(dest)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}

	// return im.WriteFile(dest)

	var out *os.File
	out, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		log.Print("openfile error: %s", err)
		return
	}
	defer out.Close()

	err = Thumbnail(in, out, topt)

	// return Thumbnail(in, out, topt)
	return
}

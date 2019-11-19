package image

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"log"
	"os"
	"path"

	"github.com/nfnt/resize"
)

// ThumbOption ...
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
	if topt.Width >= ow && topt.Height >= oh {
		return fmt.Errorf("%dx%d is too big, orig is %dx%d", topt.Width, topt.Height, ow, oh)
	}

	if topt.IsFit {
		if topt.IsCrop {
			ratioX := float32(topt.Width) / float32(ow)
			ratioY := float32(topt.Height) / float32(oh)

			if ratioX > ratioY {
				topt.ctWidth = topt.Width
				topt.ctHeight = uint(ratioX * float32(oh))
			} else {
				topt.ctHeight = topt.Height
				topt.ctWidth = uint(ratioY * float32(ow))
			}
			// :resize

			if topt.ctWidth == topt.Width && topt.ctHeight == topt.Height {
				return nil
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

// ThumbnailImage ...
func ThumbnailImage(img image.Image, topt *ThumbOption) (image.Image, error) {

	ob := img.Bounds()
	ow := uint(ob.Dx())
	oh := uint(ob.Dy())

	if ow <= topt.Width && oh <= topt.Height {
		log.Printf("ThumbnailImage %dx%d <= %dx%d", ow, oh, topt.Width, topt.Height)
		return nil, ErrOrigTooSmall
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

// Thumbnail ...
func Thumbnail(r io.Reader, w io.Writer, topt ThumbOption) error {
	var err error
	im, format, err := image.Decode(r)
	if err != nil {
		log.Printf("Thumbnail image decode error: %s", err)
		return err
	}

	m, err := ThumbnailImage(im, &topt)
	if err != nil {
		if err == ErrOrigTooSmall {
			if rr, ok := r.(io.Seeker); ok {
				rr.Seek(0, 0)
			}
			var written int64
			written, err = io.Copy(w, r)
			if err == nil {
				log.Printf("written %d", written)
				return nil
			}
			log.Printf("copy error %s", err)
		}
		return err
	}

	opt := topt.WriteOption
	if opt.Format != "" {
		opt.Format = Ext2Format(opt.Format)
	} else {
		opt.Format = format
	}

	_, err = SaveTo(w, m, opt)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// ThumbnailFile ...
func ThumbnailFile(src, dest string, topt ThumbOption) (err error) {
	var in *os.File
	in, err = os.Open(src)
	if err != nil {
		log.Print(err)
		return
	}
	defer in.Close()

	dir := path.Dir(dest)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}

	var out *os.File
	out, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		log.Printf("openfile error: %s", err)
		return
	}
	defer out.Close()

	err = Thumbnail(in, out, topt)

	return
}

package image

import (
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"os"
	"path"
)

// Position ...
type Position uint8

// Opacity ...
type Opacity uint8

// Position
const (
	BottomRight Position = iota
	BottomLeft
	TopLeft
	TopRight
	Center
	Golden
)

// WaterOption ...
type WaterOption struct {
	Pos      Position
	Opacity  Opacity
	Filename string
	WriteOption
}

// GetPoint ...
func GetPoint(sm, wm image.Point, pos Position) (pt image.Point) {

	switch pos {
	case BottomRight:
		pt.X = int(sm.X-wm.X) - 10
		pt.Y = int(sm.Y-wm.Y) - 10
		break
	case TopRight:
		pt.X = int(sm.X-wm.X) - 10
		pt.Y = 10
		break
	case BottomLeft:
		pt.X = 10
		pt.Y = int(sm.Y-wm.Y) - 10
		break
	case Center:
		pt.X = int(sm.X-wm.X) / 2
		pt.Y = int(sm.Y-wm.Y) / 2
		break
	default:
		// left = sm.X * 0.382 - wm.X / 2
		pt.X = int(sm.X-wm.X) / 2
		pt.Y = int(float64(sm.Y)*0.618 - float64(wm.Y)/2)

	}
	return
}

type grayMask struct {
	rect  image.Rectangle
	alpha uint8
}

func newGrayMask(rect image.Rectangle, opacity Opacity) *grayMask {
	if opacity < 0 {
		opacity = 0
	} else if opacity > 100 {
		opacity = 100
	}
	return &grayMask{rect, uint8(255.0 * float64(opacity) / float64(100))}
}

func (g *grayMask) ColorModel() color.Model {
	return color.AlphaModel
}

func (g *grayMask) Bounds() image.Rectangle {
	return g.rect
}

func (g *grayMask) At(x, y int) color.Color {
	return color.Alpha{g.alpha}
}

// WatermarkImage add a watermark and copyright into a image with position and opacity
func WatermarkImage(img, water image.Image, pos Position, opacity Opacity) (image.Image, error) {
	sm := img.Bounds().Max
	wm := water.Bounds().Max
	offset := GetPoint(sm, wm, pos)
	// log.Printf("watermark offset %s", offset)
	b := img.Bounds()
	m := image.NewRGBA(b)
	wb := water.Bounds()

	if opacity == 0 {
		opacity = 15
	}
	// log.Printf("set watermark opacity: %.2f", float64(opacity)/float64(100))

	draw.Draw(m, b, img, image.ZP, draw.Src)

	draw.DrawMask(m, wb.Add(offset), water, image.ZP, newGrayMask(water.Bounds(), opacity), image.ZP, draw.Over)

	return m, nil
}

// Watermark ...
func Watermark(r, wr io.Reader, w io.Writer, wo WaterOption) error {

	im, format, err := image.Decode(r)
	if err != nil {
		log.Printf("Watermark: decode src error: %s", err)
		return err
	}

	water, _, err := image.Decode(wr)
	if err != nil {
		log.Printf("Watermark: decode water error: %s", err)
		return err
	}

	m, err := WatermarkImage(im, water, wo.Pos, wo.Opacity)
	if err != nil {
		return err
	}

	opt := wo.WriteOption
	if opt.Format != "" {
		opt.Format = Ext2Format(opt.Format)
	} else {
		opt.Format = format
	}

	_, err = SaveTo(w, m, opt)
	if err != nil {
		return err
	}

	return nil
}

// WatermarkFile ...
func WatermarkFile(src, dest string, wo WaterOption) (err error) {
	var in, wr, out *os.File
	in, err = os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	wr, err = os.Open(wo.Filename)
	if err != nil {
		return
	}
	defer wr.Close()

	// log.Printf("copyright: %s", wo.Copyright)

	dir := path.Dir(dest)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}

	out, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		log.Printf("openfile error: %s", err)
		return
	}
	defer out.Close()

	err = Watermark(in, wr, out, wo)

	return
}

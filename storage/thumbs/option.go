package thumbs

import (
	imagi "github.com/go-imsto/imagi"
	"github.com/go-imsto/imsto/storage/imagio"
)

type Option func(*thumber)

func WithLoader(fn LoadFunc) func(*thumber) {
	return func(s *thumber) {
		s.loader = fn
	}
}

func WithWalker(fn WalkFunc) func(*thumber) {
	return func(s *thumber) {
		s.walker = fn
	}
}

func WithSizes(ss ...uint) func(*thumber) {
	return func(s *thumber) {
		s.okSizes = imagio.Sizes(ss)
	}
}

func WithWatermark(filename string) func(*thumber) {
	return func(s *thumber) {
		s.watermark = filename
	}
}

func WithWaterOpacity(opacity uint8) func(*thumber) {
	return func(s *thumber) {
		if opacity > 0 && opacity < 100 {
			s.waterOpacity = opacity
		}
	}
}

func ThumbOptionFromParam(p *imagio.Param) *imagi.ThumbOption {
	topt := ThumbOptionFrom(p.Mode, p.Width, p.Height)
	topt.Format = p.Ext
	return topt
}

// MakeThumbOption 根据给定的模式、宽度和高度创建并返回图像缩略图选项
func ThumbOptionFrom(mode rune, width, height uint) *imagi.ThumbOption {
	topt := &imagi.ThumbOption{
		Width:  width,
		Height: height,
		IsFit:  true,
	}
	if mode == imagio.ModeCrop {
		topt.IsCrop = true
	} else if mode == imagio.ModeWidth {
		topt.MaxWidth = width
	} else if mode == imagio.ModeHeight {
		topt.MaxHeight = height
	}
	return topt
}

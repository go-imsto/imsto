package thumbs

import (
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

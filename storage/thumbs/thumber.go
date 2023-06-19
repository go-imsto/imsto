package thumbs

import (
	"fmt"
	"os"
	"path"
	"time"

	imagi "github.com/go-imsto/imagi"
	"github.com/go-imsto/imid"
	"github.com/go-imsto/imsto/storage/imagio"
	"github.com/go-imsto/imsto/utils"
)

// consts Cate of Key
const (
	catOrig  = "orig"
	CatThumb = "thumb"
)

const (
	defaultOpacity uint8 = 30
)

func New(root string, opts ...Option) (Thumber, error) {
	if fi, err := os.Stat(root); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", root)
	}
	s := &thumber{root: root, waterOpacity: defaultOpacity}
	for _, opt := range opts {
		opt(s)
	}
	if s.loader == nil {
		return nil, fmt.Errorf("nil loader func")
	}

	return s, nil
}

type thumber struct {
	root         string
	watermark    string
	waterOpacity uint8
	loader       LoadFunc
	walker       WalkFunc
	okSizes      imagio.Sizes
}

func (s *thumber) Thumbnail(u string) error {
	p, err := imagio.ParseFromPath(u)
	if err != nil {
		logger().Infow("bad url", "url", u, "err", err)
		return err
	}
	if p.SizeOp != catOrig {
		dimension := p.SizeOp[1:]
		if len(s.okSizes) > 0 && !p.ValidSizes(s.okSizes...) {
			return NewCodeError(400, fmt.Sprintf("unsupported size: %s", dimension))
		}
	}

	if p.Mop == "w" && p.Width < 100 {
		return NewCodeError(400, "bad size with watermark")
	}

	logger().Debugw("parsed", "param", p)
	root := path.Join(s.root, CatThumb)
	oi := &outItem{
		p:        p,
		id:       p.ID,
		src:      p.Path,
		isOrig:   p.IsOrig,
		root:     root,
		origFile: path.Join(root, catOrig, p.Path),
	}

	if oi.isOrig {
		oi.dst = oi.origFile
	} else {
		dstPath := fmt.Sprintf("%s/%s", p.SizeOp, oi.src)
		oi.dst = path.Join(oi.root, dstPath)
	}

	err = utils.ReadyDir(oi.origFile)
	if err != nil {
		logger().Infow("ready dir fail", "err", err)
		return err
	}

	oi.lock, err = utils.NewFLock(oi.origFile + ".lock")
	if err != nil {
		logger().Infow("create lock fail", "err", err)
		return err
	}
	err = s.prepare(oi)
	if err != nil {
		logger().Warnw("prepare fail", "param", oi.p, "err", err)
		return err
	}
	if s.walker != nil {
		return oi.Walk(s.walker)
	}
	return nil
}

type file struct {
	dst      string
	name     string
	length   int64
	modified time.Time
	*os.File
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.length
}

func (f *file) Modified() time.Time {
	return f.modified
}

// temporary item for http read
type outItem struct {
	p        *imagio.Param
	roof     string
	src      string
	dst      string
	id       imid.IID
	isOrig   bool
	lock     utils.FLock
	length   int64
	modified time.Time
	root     string
	origFile string
}

func (o *outItem) GetID() string {
	return o.id.String()
}
func (o *outItem) GetName() string {
	return o.p.Name
}
func (o *outItem) GetRoof() string {
	return o.p.Roof
}
func (o *outItem) IsOrigin() bool {
	return o.isOrig
}
func (o *outItem) GetOrigin() string {
	return o.origFile
}
func (o *outItem) Walk(c WalkFunc) error {
	fp, err := os.Open(o.dst)
	if err != nil {
		return err
	}
	if fp == nil {
		return fmt.Errorf("Fatal error: open %s failed", o.p.Path)
	}
	defer fp.Close()
	c(&file{
		File:     fp,
		name:     o.GetName(),
		length:   o.length,
		modified: o.modified,
	})
	return nil
}

func (s *thumber) prepare(o *outItem) (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 && o.p.Mop == "" {
		o.length = fi.Size()
		o.modified = fi.ModTime()
		return
	}

	// var roof string
	logger().Infow("prepare", "orig", o.origFile)
	if fi, fe := os.Stat(o.origFile); fe != nil && os.IsNotExist(fe) || fe == nil && fi.Size() == 0 {
		logger().Infow("get mapping", "id", o.id)
		err = s.loader(o)
		if err != nil {
			return err
		}
		if fi, fe := os.Stat(o.origFile); fe != nil {
			if os.IsNotExist(fe) || fi.Size() == 0 {
				err = fmt.Errorf("write file fail: %w", fe)
				return
			}
		}
	}

	err = o.thumbnail()
	if err != nil {
		return
	}

	if o.p.Mop != "" {
		if o.p.Mop == "w" && s.watermark != "" {
			orgFile := path.Join(o.root, o.p.SizeOp, o.src)
			dstFile := path.Join(o.root, o.p.SizeOp+"w", o.src)
			waterOption := imagi.WaterOption{
				Pos:     imagi.Golden,
				Opacity: imagi.Opacity(s.waterOpacity),
			}
			err = imagi.WatermarkFile(orgFile, s.watermark, dstFile, waterOption)
			if err != nil {
				logger().Infow("watermark fail", "err", err)
			}
			o.dst = dstFile
		}
	}
	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		o.length = fi.Size()
		o.modified = fi.ModTime()
		return
	}

	return
}

func (o *outItem) thumbnail() (err error) {
	if o.isOrig {
		return
	}

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		// log.Print("thumbnail already done")
		return
	}
	// // mode := o.m["size"][0:1]
	// dimension := o.p.SizeOp[1:]
	// // log.Printf("mode %s, dimension %s", mode, dimension)
	// supportSize := config.Current.SupportSizes
	// if !supportSize.Has(o.p.Width) || !supportSize.Has(o.p.Height) {
	// 	err = NewCodeError(400, fmt.Sprintf("Unsupported size: %s", dimension))
	// 	return
	// }

	var topt = &imagi.ThumbOption{Width: o.p.Width, Height: o.p.Height, IsFit: true}
	topt.Format = o.p.Ext
	if o.p.Mode == "c" {
		topt.IsCrop = true
	} else if o.p.Mode == "w" {
		topt.MaxWidth = o.p.Width
	} else if o.p.Mode == "h" {
		topt.MaxHeight = o.p.Height
	}
	logger().Infow("thumbnail starting", "roof", o.roof, "name", o.GetName(), "opt", topt)
	err = imagi.ThumbnailFile(o.origFile, o.dst, topt)
	if err != nil {
		logger().Infow("imagi.ThumbnailFile fail", "src", o.src, "name", o.GetName(), "opt", topt, "err", err)
		return
	}

	return
}

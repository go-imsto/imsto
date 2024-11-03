package imagio

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	imagi "github.com/go-imsto/imagi"
	"github.com/go-imsto/imid"
)

const (
	ModeScale  rune = 's'
	ModeCrop   rune = 'c'
	ModeWidth  rune = 'w'
	ModeHeight rune = 'h'
)

const (
	ptImagePath  = `(?P<tp>[a-z_][a-z0-9_-]*)/(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/?(?P<t2>[a-z0-9]{2})/?(?P<t3>[a-z0-9]{5,36})\.(?P<ext>gif|jpg|jpeg|png|webp)$`
	ptImageSize  = `(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?)(?P<mop>[a-z])?`
	minDimension = 20   // 最小尺寸
	maxDimension = 9999 // 最大尺寸
)

// ErrInvalidSize 表示无效的尺寸格式
var ErrInvalidSize = errors.New("invalid image size format")

var (
	ire = regexp.MustCompile(ptImagePath)
	sre = regexp.MustCompile(ptImageSize)
)

type harg map[string]string

// Param ...
type Param struct {
	ID     imid.IID `json:"id"`
	IsOrig bool     `json:"isOrig"`
	Path   string   `json:"path"`
	SizeOp string   `json:"size"`
	Mop    string   `json:"mop,omitempty"`
	Mode   rune     `json:"mode"`
	Ext    string   `json:"ext"`
	Name   string   `json:"name,omitempty"`
	Roof   string   `json:"roof,omitempty"`

	Width  uint `json:"width"`
	Height uint `json:"height"`

	m harg
}

func (p *Param) ValidSizes(ss ...uint) bool {
	a := Sizes(ss)
	return a.Has(p.Width) && a.Has(p.Height)
}

// StoredPath 计算存储路径
func StoredPath(r string) string {
	if len(r) < 7 {
		return r
	}
	return r[0:2] + "/" + r[2:4] + "/" + r[4:]
}

// ParseFromPath ...
func ParseFromPath(uri string) (p *Param, err error) {
	var m harg
	m, err = parsePath(uri)
	if err != nil {
		return
	}
	idstr := m["t1"] + m["t2"] + m["t3"]
	var id imid.IID
	id, err = imid.ParseID(idstr)
	if err != nil {
		log.Printf("invalid id: %s", err)
		return
	}
	name := idstr + "." + m["ext"]
	p = &Param{m: m,
		ID:     id,
		Path:   StoredPath(name),
		SizeOp: m["size"],
		Mop:    m["mop"],
		Ext:    m["ext"],
		IsOrig: m["size"] == "orig",
		Name:   name,
		Roof:   m["tp"],
	}
	if !p.IsOrig {
		p.Mode, p.Width, p.Height = parseSizeOp(p.SizeOp)
	}

	return
}

func parsePath(s string) (m harg, err error) {
	match := ire.FindStringSubmatch(s)
	if len(match) == 0 {
		err = fmt.Errorf("Invalid Path: %s", s)
		return
	}
	m = make(harg)
	for i, n := range ire.SubexpNames() {
		if n != "" {
			m[n] = match[i]
		}
	}
	return
}

// ParseSize 解析输入的字符串格式，返回模式、宽度和高度
// 格式示例: "s100", "w800x600", "h500x300"
// 格式示例说明：
// - c100    (正方形，100x100)
// - c800x600 (宽800高600)
// - h500    (限高500)
// - w300    (限宽500)
func ParseSize(s string) (mode rune, width, height uint, err error) {
	// 基础格式验证
	if len(s) < 2 {
		err = fmt.Errorf("%w: %q is too short", ErrInvalidSize, s)
		return
	}

	// 验证模式字符
	mode = rune(s[0])
	if !strings.ContainsRune("scwh", rune(mode)) {
		err = fmt.Errorf("%w: invalid mode %q", ErrInvalidSize, mode)
		return
	}

	// 使用正则表达式验证完整格式
	if !sre.MatchString(s) {
		err = fmt.Errorf("%w: %q", ErrInvalidSize, s)
		return
	}

	mode, width, height = parseSizeOp(s)
	// 验证尺寸范围
	if !isValidDimension(int(width)) || !isValidDimension(int(height)) {
		err = fmt.Errorf("%w: dimensions must be between %d and %d",
			ErrInvalidSize, minDimension, maxDimension)
		return
	}
	return
}

// parseSizeOp 从输入字符串解析并返回模式、宽度和高度
func parseSizeOp(s string) (mode rune, width, height uint) {
	mode = rune(s[0])
	sz := s[1:]
	if i := strings.Index(sz, "x"); i > 1 {
		dw, _ := strconv.Atoi(sz[0:i])
		dh, _ := strconv.Atoi(sz[i+1:])
		width = uint(dw)
		height = uint(dh)
	} else {
		d, _ := strconv.Atoi(sz)
		width = uint(d)
		height = uint(d)
	}
	return
}

// isValidDimension 检查维度是否在有效范围内
func isValidDimension(d int) bool {
	return d >= minDimension && d <= maxDimension
}

func (p *Param) ToThumbOption() *imagi.ThumbOption {
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
	if mode == ModeCrop {
		topt.IsCrop = true
	} else if mode == ModeWidth {
		topt.MaxWidth = width
	} else if mode == ModeHeight {
		topt.MaxHeight = height
	}
	return topt
}

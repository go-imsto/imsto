package imagio

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-imsto/imid"
)

const (
	ptImagePath = `(?P<tp>[a-z_][a-z0-9_-]*)/(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/?(?P<t2>[a-z0-9]{2})/?(?P<t3>[a-z0-9]{5,36})\.(?P<ext>gif|jpg|jpeg|png|webp)$`
	ptImageSize = `(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?)(?P<mop>[a-z])?`
)

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
	Mode   string   `json:"mode"`
	Ext    string   `json:"ext"`
	Name   string   `json:"name,omitempty"`

	Width  uint `json:"width"`
	Height uint `json:"height"`

	m harg
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

// ParseSize ...
func ParseSize(s string) (mode string, width, height uint, err error) {
	if !sre.MatchString(s) {
		err = fmt.Errorf("invalid size %q", s)
		return
	}
	mode, width, height = parseSizeOp(s)
	return
}

func parseSizeOp(s string) (mode string, width, height uint) {
	mode = s[0:1]
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

package imagio

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-imsto/imagid"
)

const (
	ptImagePath = `(?P<tp>[a-z_][a-z0-9_-]*)/(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/?(?P<t2>[a-z0-9]{2})/?(?P<t3>[a-z0-9]{5,36})\.(?P<ext>gif|jpg|jpeg|png)$`
)

var (
	ire = regexp.MustCompile(ptImagePath)
)

type harg map[string]string

// Param ...
type Param struct {
	ID     imagid.IID `json:"id"`
	IsOrig bool       `json:"isOrig"`
	Path   string     `json:"path"`
	SizeOp string     `json:"size"`
	Mop    string     `json:"mop,omitempty"`
	Mode   string     `json:"mode"`
	Ext    string     `json:"ext"`
	Name   string     `json:"name,omitempty"`

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
	var id imagid.IID
	id, err = imagid.ParseID(idstr)
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
		p.splitSize()
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

func (p *Param) splitSize() {

	mode := p.SizeOp[0:1]
	dimension := p.SizeOp[1:]
	p.Mode = mode

	if p.m["x"] == "" {
		var d uint64
		d, _ = strconv.ParseUint(dimension, 10, 32)
		p.Width = uint(d)
		p.Height = uint(d)
	} else {
		a := strings.Split(dimension, "x")
		var dw, dh uint64
		dw, _ = strconv.ParseUint(a[0], 10, 32)
		dh, _ = strconv.ParseUint(a[1], 10, 32)
		p.Width = uint(dw)
		p.Height = uint(dh)
	}

}

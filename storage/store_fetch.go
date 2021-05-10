package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"
)

var (
	dfthc = &http.Client{
		Timeout: 25 * time.Second,
	}
)

// FetchInput ...
type FetchInput struct {
	URI     string
	Roof    string
	Referer string
	AppID   int
	UserID  int
}

// Fetch ...
func Fetch(in FetchInput) (entry *Entry, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, in.URI, nil)
	if err != nil {
		return
	}
	name := path.Base(req.URL.Path)
	if len(in.Referer) > 0 {
		req.Header.Set("Referer", in.Referer)
	}
	req.Header.Set("User-Agent", userAgent)

	logger().Infow("fetching", "in", in, "name", name)

	var res *http.Response
	res, err = dfthc.Do(req)
	if err != nil {
		logger().Warnw("fetch fail", "err", err)
		return
	}

	defer res.Body.Close()
	logger().Infow("fetched", "code", res.StatusCode, "len", res.ContentLength, "content-type", res.Header.Get("Content-Type"))

	var data []byte
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	// Check the response
	if res.StatusCode != 200 {
		err = fmt.Errorf("status code %d: %s", res.StatusCode, res.Status)
		return
	}

	entry, err = NewEntryReader(bytes.NewReader(data), name)
	if err != nil {
		return
	}
	if in.AppID > 0 {
		entry.AppId = AppID(in.AppID)
		entry.Author = Author(in.UserID)
	}
	err = <-entry.Store(in.Roof)

	return
}

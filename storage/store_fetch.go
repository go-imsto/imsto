package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
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

	logger().Infow("fetching", "in", in, "name", name)

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
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

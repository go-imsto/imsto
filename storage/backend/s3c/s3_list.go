package backend

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ListOutput ...
type ListOutput struct {
	Items      []ListItem `xml:"Contents"`
	Marker     string
	NextMarker string
	Name       string
	Delimiter  string
	Prefix     string
	MaxKeys    int
}

func (c *s3Conn) List(ls ListSpec) (items []ListItem, err error) {
	q := url.Values{}
	if len(ls.Delimiter) > 0 {
		q.Add("delimiter", ls.Delimiter)
	}
	if len(ls.Marker) > 0 {
		q.Add("marker", ls.Marker)
	}
	if ls.Limit > 0 {
		q.Add("max-keys", fmt.Sprintf("%d", ls.Limit))
	}
	if len(ls.Prefix) > 0 {
		q.Add("prefix", ls.Prefix)
	}

	var req *http.Request
	req, err = http.NewRequest("GET", c.getURL("")+"/", nil)
	if err != nil {
		return
	}

	req.URL.RawQuery = q.Encode()

	logger().Infow("listing", "url", req.URL)
	req.Header.Set("x-amz-content-sha256", emptySum)
	var resp *http.Response
	resp, err = c.ac.Do(req)
	if err != nil {
		logger().Infow("get fail", "spec", ls, "err", err)
		return
	}
	defer resp.Body.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		logger().Infow("get status", "code", resp.StatusCode, "result", string(buf))
		err = ErrRequest
		return
	}
	// logger().Infow("result", "data", string(buf))

	var lo ListOutput
	err = xml.Unmarshal(buf, &lo)
	if err == nil {
		items = lo.Items
	}
	return
}

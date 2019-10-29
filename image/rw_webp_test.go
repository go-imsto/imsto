package image

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebp(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(webpData)
	if err != nil {
		t.Errorf("decode err %s", err)
		return
	}

	im, err := Open(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.NotNil(t, im)
	assert.NotNil(t, im.Attr)
	assert.Equal(t, webpWidth, int(im.Attr.Width))
	assert.Equal(t, webpHeight, int(im.Attr.Height))
	assert.Equal(t, ".webp", im.Attr.Ext)
	assert.Equal(t, "image/webp", im.Attr.Mime)
	// assert.Equal(t, int(jpegQuality), int(im.Attr.Quality))
	assert.Equal(t, webpOrgSize, int(im.Attr.Size))

	var buf bytes.Buffer
	var n int
	n, err = SaveTo(&buf, im.m, WriteOption{Format: "webp", Quality: webpQuality})
	assert.NoError(t, err)

	assert.Equal(t, int(webpNewSize), n)
	assert.Equal(t, int(webpNewSize), buf.Len())

	meta := im.Attr.ToMap()
	assert.NotNil(t, meta)
	assert.Equal(t, ".webp", meta["ext"])

	var a Attr
	a.FromMap(meta)
	assert.Equal(t, ".webp", a.Ext)
}

const (
	webpWidth   = 150
	webpHeight  = 100
	webpOrgSize = 2450
	webpNewSize = 2288
	webpQuality = 75
)
const webpData = `UklGRooJAABXRUJQVlA4IH4JAAAyLwCdASqWAGQAPo04lUeioYwGg0EUBGJZQDWZNsO1JP5n64fB3Ub8v+o9XG3d53fTuYGQ1jfacgmJk6PthcDb6Dzr0z01vyivwkv9E+4YUHpwH65k6mVaIXqrC4XTBQLqkGc1rCgGes04gKPUimJAmZhyXeG4DBy84mCBiQxZwB3y0ABGURmcpLymWUnJPKhGe8RujkoIZ810gU5JiH/ONuOi9qprpe/D+igi50G8+UMtjxOVM02DI1cX8zoiV/kuK0uHAtfQaqJC3eZhgnSAYCLcn9Ayj9pMHbfk0JJRbdF9H8CZCBhGkcIF2K+oGIR+gVW/6aRvmPtQNcIoyzgdFht+Z31lpq0fF/hD2JlqnmNkvDsKndGC/A+Z2BH0/lTXb9S/ELRNKqWKIrfN/h2Pao6wijZj9KO5+yNBvsnvdZyNmFFB2O/JHVnn3XtYdzUEN9w2Fz52sfwOAMXrMIYRcnd3nEYDHjHyw4dvT5hhr5BZSlFAHgz7I3AWw5qCIy/YgAD+/p/aKcNszx8Xd4P/UesGZn9KuPC+bM7Ftm6MOeuMu5ZP3zWAOb4skNx3FFcXmMGazmsEGYQxHgVSeNZ9FqVet4IUc/QajL5J2TiLIyAJvSnMT8RZXKotqUhB00DSZ+hzvShMFGwcuPBTvvABW6TRYILdCUNwFVOGZl4yyQ9vGi5HuKDPa0gsLt5fP8S4wD8dJy27Tx3P+MUlS3egiSAY3YnlR2aEQlh866iFPPH4U06y6+VGtvbq+N2Kh7z14i/+zoWLhk9pvfWkhXMMm9MPLtl0gOEtc9xLInLyeKbb3kbUD60/vRcE6d6302jpc9f8tyCW9qcg2LhBeY7H2akeaH/6bXaX5rc45emsA+58yiAwxbaGeWFF5DyCqUMWbwX+BIHTlFjTeatv8Sl8hjklsykxPbFpSt0dCM//nn9ZdFg3tLTgq6vbkKhe7dWNNSGsafAD2DYlAyGH4P0320LlribfnfkovkKWxuky/vTuGOBnAr/wZcL2eyg+34CczxKpZKh77B+Y/VUqH4LnC2BOEdIPmgkE/zOGa2S5wyNqfdctPTNhYMOtR/HFOukVkv/v5q5E4alohykAaUISEiLqRbaZkkTkhypdOAfifqYybCSqlCYCwsaqVL4sDhbep/dRfVog8tmgSoeh+MIl+D/ePR8eD56tz6EYEKnHVN8AaXvpFK50F3SFMLcnaEMJhUJc8PvGF/Ez/q7X+HKQ3J84TzFJbscxr7eozHeFMMH9/h0l5DXY4aTv+iIR4gM0dnbjbEHgsKP61ALeNz28bvmwhHa0rFBa9uJVskb1W3k5cAMzC1RkxfJ/sT++R+7jjNsNc3XkIKUrwP6nZrwzhpXAvJx4tFQoPhnq1k754EgDJG16UBY37s0uMFIAsT+slJHQ0OGmnE8aWr9oXIe02bm/CAIXyEWvxnHcHTTxIqhFD4XzHGJ+MCe7opIY/dEkOPcKUD6zNQ8IPOT/XpDvWGw/ok/UcILHoyQ3QdCgUqFDr1rm1Jt5wWbZcWSNhXjf1LuKzwJrkf1rHf/Ktf4tAtfWr/rQhun31sTg2rbi5fWKOZul29eY3ngUYVPHgSb/vvPisPqqXdfnJLpilz0u/VvCPFVofn1VADfJKNOJkXtCIxQ/v7RKZMrtdUkmRJZswSHIZ9950jHQi7KTbIozDSPMHUy5aw2VzfNS47v6d4YOvGY2pMW5Q4B6vh1E+FirurI+iYkoyBWyfAJxKHc58PQ5CYOUXqC7pdinZViLBy70kYbbPbgU16bokAP2zAR/UBW9+GTlfdAUNdGskU7nGzUbge7rt45UqaTZcKTu5GbN3YflPxU2uS7NxWGOuCgtGCu1O2zX12MVXSKgHlYv6pddY6EWs6QeEpvgtKAdQv8MJmISVUS1y3+Xom1owgeGRyqpgO9dulNmPt3UdK4PBiVz97n4QbX6QRooaAlPvFmUgYNO1CJXiWGkzP0vNP2JNu7Uymtxhz7UZCAHjW2aFJ7s6cNOZIdKmCkq2bw5iA4jZSPs+snTiT8gFyUKnSK7Lvc7rUOJNq9qvNoZtAgC9mMDhOA4WVjNR2cBJKQgzfgtRsVLMpbbb3FJroDkux5VhZ2nB16TBoEqgYqQVo5WdivKjNOLGJqAfcX2hMjE3UK1Yu8DT5qfQrRuAK2Y5rGut0SWSmpLZX5oCO+KXgbeh9mzk1ULYJcaWCVCKBy5bGrsnEUN/9L8qkahJ2sF4FJp/aG7WATm/9nNfvkAi8lhzP/kWYQa9HUJpUL3gBJcTRd4yFXMPA3Sd0p4NQBULx1DdGWxeCP9eyvD39+nWkHgYJIfrEqlpXaD5TrpJt5O/XbQa7IlyHjPimjnSQsjf5WXdfIGOL21QHGApYH6u54q1jf6WGac8ynys7oJFEM36NyUp+UTsX4pxn1RaWGRKJ8EUk3XJ5GXVGQO9BB7YID2vR537253t2WNYirDbju9w8bfBJ02jDOgEjZRSMP4VZLutXotpzQo8259eLACChjWX8ODdcvEf/OwTwQNoI8hZGimdkNEgILhM0zYlOe/fvv4VdkxvU3oMXhie9v16T7kN9sL2GPzB//P0FLhM+QKOxbdguTnMDvH6VbEoeuX6wO/o2wec6id/QOZb0e5hbVrq2UgcXpbHm0bYn+DuA5e4FT50zUUlVe7cTvmnstNZ+p0Xx903zN/WMjmLvgt1/b2syHzhRgTFraP9Do3ny9hNZnIrCrGjzHRtrVqe+tXgR6ZNZ+m/jwyFG7P2rXbjnkhpHuEMADn/iT25su+3c8AO61vNksZfCsbt4+kycfRJ+i4XiIP08ro2g7+yyYpRvnPmyJ9ojMVMcGtxMwqA0xX3VFsDzaRi/RhJi2Oq3eNKX0RmNIU44qhLkJGgqfqTFv0iHdegTV1gC1kmG5RjbTZPcOHR1p1IeEJIHrfBdSa+gUASMi1S7t3crCjBtIcmPizFbEGi/EzDV7APO+vG523/eSG2foiiRUUpHFULlY4/yxOcoewd28fFY+vXN5LbUbZKIV5H8GiUlX6VfPimYmh0cF/plxhM9cKZjyoroVYgx2CS0YDaRx1cXFEkQfP5XjpzRKf4p907sdYKFcqEPQlbmHFjEidsSKEpR+BW97QhyY5R5qhJLhYeW9z8daTdGUKGE35tq2v7YKorbyGClrbuz8sZhepmUhbcbem+wvwkD1AW074/fSpNdTCIzlajUOz++SHPqSubek2AAAAAAA=`

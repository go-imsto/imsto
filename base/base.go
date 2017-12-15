// see also: php.net/base_convert
package base

import (
	"bytes"
	"errors"
	"math/big"
)

const (
	BASE62_STR = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	base62         = []byte(BASE62_STR)
	ErrInvalidBase = errors.New("The base number must be between 2 and 36.")
)

func Convert(num string, frombase int, tobase int) (string, error) {
	if num == "" {
		return "", nil
	}

	if frombase == tobase {
		return num, nil
	}

	if 2 > frombase || frombase > 62 || 2 > tobase || tobase > 62 {
		return "", ErrInvalidBase
	}

	var fromdigits = base62[0:frombase]
	var todigits = base62[0:tobase]
	from_b := big.NewInt(int64(frombase))
	to_b := big.NewInt(int64(tobase))

	x := big.NewInt(0)
	for _, digit := range []byte(num) {
		x.Mul(x, from_b)
		i := bytes.IndexByte(fromdigits, digit)
		if i < 0 {
			return "", errors.New("the number string is invalid")
		}
		x.Add(x, big.NewInt(int64(i)))
	}

	var res []byte
	for x.Cmp(big.NewInt(0)) > 0 {
		digit := new(big.Int).Mod(x, to_b).Uint64()
		res = append([]byte{todigits[digit]}, res...)
		x.Div(x, to_b)
	}

	return string(res), nil
}

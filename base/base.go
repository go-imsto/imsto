// see also: php.net/base_convert
package base

import (
	"bytes"
	"errors"
	"math/big"
)

const (
	BASE62Text = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	base62         = []byte(BASE62Text)
	ErrInvalidBase = errors.New("The base number must be between 2 and 36.")
)

// Convert Convert a number between arbitrary bases
func Convert(num string, frombase int, tobase int) (string, error) {
	return ConvertBytes([]byte(num), frombase, tobase)
}

// ConvertBytes ...
func ConvertBytes(num []byte, frombase int, tobase int) (string, error) {
	if len(num) == 0 {
		return "", nil
	}

	if frombase == tobase {
		return string(num), nil
	}

	if 2 > frombase || frombase > 62 || 2 > tobase || tobase > 62 {
		return "", ErrInvalidBase
	}

	var fromdigits = base62[0:frombase]
	var todigits = base62[0:tobase]
	fromBi := big.NewInt(int64(frombase))
	toBi := big.NewInt(int64(tobase))

	x := big.NewInt(0)
	for _, digit := range num {
		x.Mul(x, fromBi)
		i := bytes.IndexByte(fromdigits, digit)
		if i < 0 {
			return "", errors.New("the number string is invalid")
		}
		x.Add(x, big.NewInt(int64(i)))
	}

	var res []byte
	for x.Cmp(big.NewInt(0)) > 0 {
		digit := new(big.Int).Mod(x, toBi).Uint64()
		res = append([]byte{todigits[digit]}, res...)
		x.Div(x, toBi)
	}

	return string(res), nil
}

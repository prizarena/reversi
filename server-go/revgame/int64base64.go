package revgame

import (
	"strings"
	"fmt"
	)

// https://codereview.stackexchange.com/questions/71272/convert-int64-to-custom-base64-number-string
var codes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-"

func EncodeIntToBase64(v int64) string {
	s := make([]byte, 0, 12)
	if v == 0 {
		return "0"
	}
	for v > 0 {
		ch := codes[v%64]
		s = append(s, byte(ch))
		v /= 64
	}
	return string(s)
	// return strings.TrimLeft(string(s), "0")
}

func DecodeIntFromBase64(s string) (int64, error) {
	res := int64(0)

	for i := len(s); i > 0; i-- {
		ch := s[i-1]
		res *= 64
		mod := strings.IndexByte(codes, ch)
		if mod == -1 {
			return -1, fmt.Errorf("invalid character: '%c'", ch)
		}
		res += int64(mod)
	}
	return res, nil
}

func EncodeIntToString(v int64) string {
	return EncodeIntToBase64(v)
}

func DecodeStringToInt(s string) (int64, error) {
	return DecodeIntFromBase64(s)
}
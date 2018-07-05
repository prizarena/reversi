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

const intToStrBase = 36

func EncodeIntToString(v int64) string {
	// return strconv.FormatInt(v, intToStrBase)
	return EncodeIntToBase64(v)
	// 0000gw_0000801   - base64
	// fwopxxc_vmppmo0  - base=36

}

func DecodeStringToInt(s string) (int64, error) {
	// return strconv.ParseInt(s, intToStrBase, 64)
	return DecodeIntFromBase64(s)
}
package revgame

import (
	"github.com/pkg/errors"
	"strings"
)

type Transcript []byte

var ErrNotValidTranscript = errors.New("not valid transcript")

func EmptyTranscript() Transcript{
	return Transcript([]byte{})
}

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func NewTranscript(s string) (transcript Transcript) {
	if len(s) == 0 {
		return
	}
	transcript = make(Transcript, len(s))
	for i, v := range []byte(s) {
		transcript[i] = byte(strings.Index(encodeURL, string(v)))
	}
	return
}

func (t Transcript) ToBase64() string {
	v := make([]byte, len(t))
	for i, a := range t {
		v[i] = encodeURL[a]
	}
	return string(v)
}

func (t Transcript) Pop() (Move, Transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	return t.LastMove(), t[:len(t)-1]
}

func (t Transcript) LastMove() Move {
	return Move(t[len(t)-1])
}

type Move byte

func (m Move) Address() Address {
	i := int8(m)
	return Address{i % BoardSize, i / BoardSize}
}

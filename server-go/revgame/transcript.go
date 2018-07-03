package revgame

import (
	"github.com/pkg/errors"
	"encoding/base64"
)

type Transcript []byte

var ErrNotValidTranscript = errors.New("not valid transcript")

func EmptyTranscript() Transcript{
	return Transcript([]byte{})
}

func NewTranscript(s string) (transcript Transcript) {
	if len(s) == 0 {
		return
	}
	var (
		v   []byte
		err error
	)
	if v, err = base64.RawURLEncoding.DecodeString(s); err != nil {
		panic(ErrNotValidTranscript)
		return
	}
	transcript = Transcript(v)
	return
}

func (t Transcript) ToBase64() string {
	return base64.RawURLEncoding.EncodeToString([]byte(t))
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

func (m Move) Address() address {
	i := int8(m)
	return address{i % BoardSize, i / BoardSize}
}

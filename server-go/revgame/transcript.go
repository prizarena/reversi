package revgame

import (
	"github.com/pkg/errors"
		"github.com/prizarena/turn-based"
	"bytes"
	"fmt"
)

type Transcript []byte

func (t Transcript) Equal(t2 Transcript) bool {
	return bytes.Equal([]byte(t), []byte(t2))
}

var ErrNotValidTranscript = errors.New("not valid transcript")

func EmptyTranscript() Transcript{
	return Transcript([]byte{})
}

var encodeURL = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

func NewTranscript(s string) (transcript Transcript) {
	if len(s) == 0 {
		return
	}
	transcript = make(Transcript, len(s))
	for i, b := range []byte(s) {
		if v := bytes.IndexByte(encodeURL, b); v < 0 {
			panic(fmt.Sprintf("unkonw transcript code: " + string(v)))
		} else {
			transcript[i] = byte(v)
		}
	}
	return
}

func NewTranscriptFromHumanReadable(s string) (transcript Transcript) {
	if len(s) == 0 {
		return
	}
	count := len(s)/2
	if count*2 != len(s) {
		panic("len of transcript is not event")
	}

	transcript = make(Transcript, count)
	for i := 0; i<count; i++ {
		cell := turnbased.CellAddress(s[i*2:i*2+2])
		transcript[i] = byte(Address{X: int8(cell.X()), Y: int8(cell.Y())}.ToMove())
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

func (t Transcript) String() string {
	s := make([]byte, len(t)*2)
	for i, v := range t {
		a := Move(v).Address()
		ca := turnbased.NewCellAddress(int(a.X), int(a.Y))
		j := i*2
		s[j] = ca[0]
		s[j+1] = ca[1]
	}
	return string(s)
}

func (t Transcript) Pop() (Move, Transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	return t.LastMove(), t[:len(t)-1]
}

func (t Transcript) NextMove() (Move, Transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	return Move(t[0]), t[1:]
}

func (t Transcript) LastMove() Move {
	return Move(t[len(t)-1])
}

type Move byte

func (m Move) Address() Address {
	i := int8(m)
	return Address{i % BoardSize, i / BoardSize}
}

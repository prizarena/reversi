package revgame

import (
	"github.com/prizarena/turn-based"
	"github.com/pkg/errors"
)

type Transcript string

func NewTranscript(s string) (Transcript, error) {
	if len(s) % 2 != 0 {
		return "", errors.New("transcript length should be even")
	}
	return Transcript(s), nil
}

func (t Transcript) Pop() (address, Transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	last := turnbased.CellAddress(t[len(t)-2:])
	x, y := last.XY()

	return address{x, y}, t[:len(t)-2]
}

func (t Transcript) String() string {
	return string(t)
}

func (t Transcript) LastMove() turnbased.CellAddress {
	if t == "" {
		return ""
	}
	return turnbased.CellAddress(t[len(t)-2:])
}
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

func (t Transcript) Pop() (Move, Transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	last := Move(t[len(t)-3:])
	return last, t[:len(t)-3]
}

func (t Transcript) String() string {
	return string(t)
}

func (t Transcript) LastMove() Move {
	if t == "" {
		return ""
	}
	return Move(t[len(t)-3:])
}

type Move string

func (m Move) Address() turnbased.CellAddress {
	return turnbased.CellAddress(m[1:])
}

func (m Move) Player() Disk {
	return Disk(m[0])
}

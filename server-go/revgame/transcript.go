package revgame

import "github.com/prizarena/turn-based"

type transcript string

func NewTranscript(s string) transcript {
	if len(s) % 2 != 0 {
		panic("transcript length should be even")
	}
	return transcript(s)
}

func (t transcript) Pop() (address, transcript) {
	if len(t) == 0 {
		panic("nothing to pop")
	}
	last := turnbased.CellAddress(t[len(t)-2:])
	x, y := last.XY()

	return address{x, y}, t[:len(t)-2]
}

func (t transcript) String() string {
	return string(t)
}

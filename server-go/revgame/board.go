package revgame

import (
	"github.com/prizarena/turn-based"
	"math/bits"
	"bytes"
	"strings"
	"github.com/strongo/emoji/go/emoji"
)

type Board struct {
	Blacks Disks
	Whites Disks
}

var OthelloBoard = Board{
	Whites: (1 << (3*8 + 3)) | (1 << (4*8 + 4)),
	Blacks: (1 << (3*8 + 4)) | (1 << (4*8 + 3)),
}

func (b Board) Flip(cell turnbased.CellAddress) (board Board) {
	if b.Whites& b.Blacks != 0 {
		panic("b.White & b.Black != 0")
	}

	board = b

	whiteIsPlaced := b.Whites.IsPlaced(cell)
	blackIsPlaced := b.Blacks.IsPlaced(cell)
	if whiteIsPlaced && blackIsPlaced {
		panic("whiteIsPlaced && blackIsPlaced")
	} else if !whiteIsPlaced && !blackIsPlaced {
		panic("!whiteIsPlaced && !blackIsPlaced")
	} else {
		var err error
		var whiteResult, blackResult Disks
		switch true {
		case whiteIsPlaced:
			if whiteResult, err = b.Whites.Remove(cell); err != nil {
				panic(err)
			}
			if blackResult, err = b.Blacks.Add(cell); err != nil {
				panic(err)
			}
		case blackIsPlaced:
			if blackResult, err = b.Blacks.Remove(cell); err != nil {
				panic(err)
			}
			if whiteResult, err = b.Whites.Add(cell); err != nil {
				panic(err)
			}
		default:
			panic("!whiteIsPlaced && !blackIsPlaced")
		}
		board = Board{Whites: whiteResult, Blacks: blackResult}
	}

	return
}

func (b Board) NextMove() string {
	switch (bits.OnesCount64(uint64(b.Whites)) + bits.OnesCount64(uint64(b.Blacks))) % 2 {
	case 0:
		return "black"
	case 1:
		return "white"
	default:
		panic("unexpected branch")
	}
}

func (b Board) Rows() (rows [8][8]string) {
	// rows = make([][]string, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			bit := Disks(1 << (uint(y)*8 + uint(x)))
			if bit & b.Whites != 0 {
				rows[y][x] = emoji.WhiteCircle
			} else if bit & b.Blacks != 0 {
				rows[y][x] = emoji.BlackCircle
			} else {
				rows[y][x] = emoji.WhiteLargeSquare
			}
		}
	}

	return
}

func (board Board) DrawBoard(colSeparator, rowSeparator string) string {
	s := new(bytes.Buffer)

	s.WriteRune('\n')
	rows := board.Rows()
	for _, row := range rows {
		s.WriteString(strings.Join(row[:], colSeparator))
		s.WriteString(rowSeparator)
	}
	return s.String()
}
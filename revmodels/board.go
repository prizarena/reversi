package revmodels

import "github.com/prizarena/turn-based"

type Board struct {
	White Disks
	Black Disks
}

func (b Board) Flip(cell turnbased.CellAddress) (board Board) {
	if b.White & b.Black != 0 {
		panic("b.White & b.Black != 0")
	}

	board = b

	whiteIsPlaced := b.White.IsPlaced(cell)
	blackIsPlaced := b.Black.IsPlaced(cell)
	if whiteIsPlaced && blackIsPlaced {
		panic("whiteIsPlaced && blackIsPlaced")
	} else if !whiteIsPlaced && !blackIsPlaced {
		panic("!whiteIsPlaced && !blackIsPlaced")
	} else {
		var err error
		var whiteResult, blackResult Disks
		switch true {
		case whiteIsPlaced:
			if whiteResult, err = b.White.Remove(cell); err != nil {
				panic(err)
			}
			if blackResult, err = b.Black.Add(cell); err != nil {
				panic(err)
			}
		case blackIsPlaced:
			if blackResult, err = b.Black.Remove(cell); err != nil {
				panic(err)
			}
			if whiteResult, err = b.White.Add(cell); err != nil {
				panic(err)
			}
		default:
			panic("!whiteIsPlaced && !blackIsPlaced")
		}
		board = Board{White: whiteResult, Black: blackResult}
	}

	return
}


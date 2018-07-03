package revgame

import (
	"github.com/pkg/errors"
)

type Disks int64

var ErrAlreadyOccupied = errors.New("cell already occupied")
var ErrNotValidMove = errors.New("not valid move")
var ErrNotOccupied = errors.New("cell is not occupied")

func (pd Disks) bit(a Address) Disks {
	bit := Disks(1) << (uint(a.Y) * 8)
	bit = bit << uint(a.X)
	return bit
}

func (pd Disks) isPlaced(a Address) bool {
	bit := pd.bit(a)
	return (bit & pd) != 0
}

func (pd Disks) add(a Address) (result Disks, err error) {
	result = pd
	bit := pd.bit(a)
	if (bit & pd) != 0 {
		err = errors.WithMessage(ErrAlreadyOccupied, a.String())
		return
	}
	result |= bit
	return
}

func (pd Disks) mustAdd(a Address) Disks {
	result, err := pd.add(a)
	if err != nil {
		panic(err)
	}
	return result
}

func (pd Disks) remove(a Address) (result Disks) {
	result = pd
	bit := pd.bit(a)
	if (bit & pd) == 0 {
		panic(errors.WithMessage(ErrNotOccupied, a.String()))
	}
	result ^= bit
	return
}

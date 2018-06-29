package revmodels

import (
	"github.com/prizarena/turn-based"
	"github.com/pkg/errors"
)

type Disks uint64

var ErrAlreadyOccupied = errors.New("cell already occupied")
var ErrNotOccupied = errors.New("cell is not occupied")

func (pd Disks) bit(cell turnbased.CellAddress) Disks {
	x, y := cell.XY()
	bit := Disks(1) << uint(x) * 8
	bit = bit << uint(y)
	return bit
}

func (pd Disks) IsPlaced(cell turnbased.CellAddress) bool {
	bit := pd.bit(cell)
	return bit & pd != 0
}

func (pd Disks) Add(cell turnbased.CellAddress) (result Disks, err error) {
	result = pd
	bit := pd.bit(cell)
	if (bit & pd) != 0 {
		err = ErrAlreadyOccupied
		return
	}
	result |= bit
	return
}

func (pd Disks) Remove(cell turnbased.CellAddress) (result Disks, err error) {
	result = pd
	bit := pd.bit(cell)
	if (bit & pd) == 0 {
		err = ErrNotOccupied
		return
	}
	result ^= bit
	return
}



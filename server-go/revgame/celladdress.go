package revgame

// CellAddress is coded as "CharDigit" where Char is X and Digit is Y, e.g. A1, B2, C3.
// Inlined from the former github.com/prizarena/turn-based dependency (only the small
// coordinate helper revgame actually used), keeping the engine dependency-light.
type CellAddress string

// NewCellAddress builds a CellAddress from zero-based x,y (x -> 'A'.., y -> '1'..).
func NewCellAddress(x, y int) CellAddress {
	return CellAddress([]rune{
		'A' + rune(x),
		'1' + rune(y),
	})
}

// X returns the zero-based column.
func (ca CellAddress) X() int {
	return int(rune(ca[0]) - 'A')
}

// Y returns the zero-based row (single-digit rows, as on an 8x8 board).
func (ca CellAddress) Y() int {
	return int(rune(ca[1]) - '1')
}

// XY returns the zero-based column and row.
func (ca CellAddress) XY() (x, y int) {
	return ca.X(), ca.Y()
}

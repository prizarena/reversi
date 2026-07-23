package revgame

import "testing"

func addrSetEqual(got []Address, want []Address) bool {
	if len(got) != len(want) {
		return false
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestBoard_ValidMoves_StartPosition(t *testing.T) {
	// Black's four standard Othello opening moves: d3, c4, f5, e6.
	want := []Address{{X: 3, Y: 2}, {X: 2, Y: 3}, {X: 5, Y: 4}, {X: 4, Y: 5}}
	got := OthelloBoard.ValidMoves(Black)
	if !addrSetEqual(got, want) {
		t.Errorf("ValidMoves(Black) = %v, want the four openings %v", got, want)
	}
	if !OthelloBoard.HasValidMoves(Black) {
		t.Error("HasValidMoves(Black) = false, want true on the start board")
	}
}

func TestBoard_ValidMoves_NoMoves(t *testing.T) {
	// A board with a single black disk and no white disks: White cannot move.
	board := Board{Blacks: Disks(1 << 0), Whites: 0, Last: Address{X: 0, Y: 0}}
	if got := board.ValidMoves(White); len(got) != 0 {
		t.Errorf("ValidMoves(White) = %v, want empty", got)
	}
	if board.HasValidMoves(White) {
		t.Error("HasValidMoves(White) = true, want false when White has no move")
	}
}

package revgame

import "testing"

// A board where White's single legal move is the corner (0,0): black at (1,0)
// flanked between white (2,0) and the corner. Exercises the corner-preference
// branch and the single-move fast path in bestMove.
func TestSimpleAI_GetMove_PrefersCorner(t *testing.T) {
	board := Board{Blacks: Disks(1 << 1), Whites: Disks(1 << 2)}
	move := SimpleAI{}.GetMove(board, White)
	if (move != Address{X: 0, Y: 0}) {
		t.Fatalf("expected the corner move (0,0), got %v", move)
	}
}

// Two symmetric corner moves that yield identical scores force the tie-break
// path (bestMoves accumulates >1 candidate, then rnd picks one). Whichever is
// returned must be one of the two legal corner moves.
func TestSimpleAI_GetMove_TieBreak(t *testing.T) {
	board := Board{
		Blacks: Disks(1<<1) | Disks(1<<(7*8+1)), // (1,0) and (1,7)
		Whites: Disks(1<<2) | Disks(1<<(7*8+2)), // (2,0) and (2,7)
	}
	valid := board.getValidMoves(White)
	if !addrSetEqual(valid, []Address{{X: 0, Y: 0}, {X: 0, Y: 7}}) {
		t.Fatalf("precondition: expected the two corner moves, got %v", valid)
	}
	for i := 0; i < 50; i++ {
		move := SimpleAI{}.GetMove(board, White)
		if (move != Address{X: 0, Y: 0}) && (move != Address{X: 0, Y: 7}) {
			t.Fatalf("iter %d: got %v, want one of the two corners", i, move)
		}
	}
}

func TestSimpleAI_GetMove(t *testing.T) {
	board := OthelloBoard
	board.Blacks, _ = board.Blacks.add(Address{3, 5})
	board.Blacks, _ = board.Blacks.add(Address{3, 6})
	board.Blacks, _ = board.Blacks.add(Address{2, 4})
	//
	move := SimpleAI{}.GetMove(board, White)
	if move.X != 3 || move.Y != 7 {
		t.Log(board.DrawBoard("*", "O", ".", "", "\n"))
		t.Fatalf("unexpected move: %v", move)
	}
}

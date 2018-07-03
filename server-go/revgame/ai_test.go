package revgame

import "testing"

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

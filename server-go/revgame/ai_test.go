package revgame

import "testing"

func TestSimpleAI_GetMove(t *testing.T) {
	board := OthelloBoard
	board.Blacks, _ = board.Blacks.add(address{3, 5})
	board.Blacks, _ = board.Blacks.add(address{3, 6})
	board.Blacks, _ = board.Blacks.add(address{2, 4})
	t.Log(board.DrawBoard("*", "O", ".", "", "\n"))
	move := SimpleAI{}.GetMove(board, White)
	if move.X != 3 || move.Y != 7 {
		t.Fatalf("unexpected move: %v", move)
	}
}

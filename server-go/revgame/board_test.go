package revgame

import "testing"

func TestBoard_DrawBoard(t *testing.T) {
	board := OthelloBoard

	t.Log(board.DrawBoard("", "\n"))
}

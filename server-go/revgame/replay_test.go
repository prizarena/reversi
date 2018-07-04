package revgame

import "testing"

func TestRewind(t *testing.T) {
	move := Address{5,4}
	board, err := OthelloBoard.MakeMove(Black, move)
	if err != nil {
		t.Fatal(err)
	}
	transcript := make(Transcript, 0, 1)
	transcript, _ = AddMoveToTranscript(transcript, 0, move)

	board, nextMove := Rewind(board, transcript, 1)
	if board != OthelloBoard {
		t.Errorf("board != OthelloBoard:%v", board.DrawBoardAsText("."))
	}
	if !nextMove.IsOnBoard() {
		t.Errorf("nextMove is out of board: %v", nextMove)
	}
	if nextMove != move {
		t.Errorf("nextMove != move: %v != %v", nextMove, move)
	}
}

package revgame

import "testing"

func TestReplay(t *testing.T) {
	move := Address{5,4}
	board := OthelloBoard
	transcript := make(Transcript, 0, 1)
	transcript, _ = AddMoveToTranscript(transcript, 0, move)

	board = Replay(board, transcript, 1)
	if board != OthelloBoard {
		t.Errorf("board != OthelloBoard:%v", board.DrawBoardAsText("."))
	}

	currentBoard := Replay(board, transcript, 1)
	var err error
	var expectedBoard Board
	expectedBoard, err = OthelloBoard.MakeMove(Black, move)
	if err != nil {
		t.Fatal(err)
	}
	if currentBoard != board {
		t.Errorf("currentBoard != expectedBoard:%v", expectedBoard.DrawBoardAsText("."))
	}

	// if !nextMove.IsOnBoard() {
	// 	t.Errorf("nextMove is out of board: %v", nextMove)
	// }
	// if nextMove != move {
	// 	t.Errorf("nextMove != move: %v != %v", nextMove, move)
	// }
}

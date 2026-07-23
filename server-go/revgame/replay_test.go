package revgame

import "testing"

// TestReplay_ExecutesMoves drives Replay with backSteps=0 so the replay loop
// actually applies moves (the existing TestReplay uses backSteps=1, which
// leaves the loop body unentered).
func TestReplay_ExecutesMoves(t *testing.T) {
	move := Address{X: 3, Y: 2} // legal Black opening on the Othello board
	transcript, _ := AddMove(EmptyTranscript(), 0, move)

	got := Replay(OthelloBoard, transcript, 0)

	want, err := OthelloBoard.MakeMove(Black, move)
	if err != nil {
		t.Fatalf("setup MakeMove: %v", err)
	}
	if got != want {
		t.Errorf("Replay result mismatch:\n got %v\nwant %v",
			got.DrawBoardAsText("."), want.DrawBoardAsText("."))
	}
}

func TestReplay_IllegalMovePanics(t *testing.T) {
	// (0,0) flanks nothing on the Othello board -> MakeMove errors -> Replay panics.
	transcript, _ := AddMove(EmptyTranscript(), 0, Address{X: 0, Y: 0})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Replay to panic on an illegal transcript move")
		}
	}()
	Replay(OthelloBoard, transcript, 0)
}

func TestAddMove_TruncatesOnBackSteps(t *testing.T) {
	// Two recorded moves, then AddMove with backSteps=1: the last move is
	// dropped and replaced, backSteps resets to 0.
	tr := EmptyTranscript()
	tr, _ = AddMove(tr, 0, Address{X: 3, Y: 2})
	tr, _ = AddMove(tr, 0, Address{X: 2, Y: 4})

	newMove := Address{X: 5, Y: 4}
	got, backSteps := AddMove(tr, 1, newMove)

	if backSteps != 0 {
		t.Errorf("backSteps = %d, want 0", backSteps)
	}
	want := EmptyTranscript()
	want, _ = AddMove(want, 0, Address{X: 3, Y: 2})
	want, _ = AddMove(want, 0, newMove)
	if !got.Equal(want) {
		t.Errorf("AddMove truncation = %v, want %v", []byte(got), []byte(want))
	}
}

func TestAddMove_DuplicatePanics(t *testing.T) {
	tr := EmptyTranscript()
	tr, _ = AddMove(tr, 0, Address{X: 3, Y: 2})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected AddMove to panic on a duplicate move")
		}
	}()
	AddMove(tr, 0, Address{X: 3, Y: 2})
}

func TestVerifyBoardTranscript(t *testing.T) {
	// A move that references an empty cell (0,0) on the Othello board is fine.
	okTranscript := Transcript{byte(Address{X: 0, Y: 0}.Index())}
	if err := VerifyBoardTranscript(OthelloBoard, okTranscript); err != nil {
		t.Errorf("expected nil error for a transcript over empty cells, got: %v", err)
	}

	// A move that references the occupied cell d4 (3,3) must be rejected.
	badTranscript := Transcript{byte(Address{X: 3, Y: 3}.Index())}
	err := VerifyBoardTranscript(OthelloBoard, badTranscript)
	if err == nil {
		t.Fatal("expected an error for a transcript referencing an occupied cell")
	}
	if got := err.Error(); got == "" {
		t.Error("error message should name the occupied cell")
	}
}

func TestReplay(t *testing.T) {
	move := Address{5, 4}
	board := OthelloBoard
	transcript := make(Transcript, 0, 1)
	transcript, _ = AddMove(transcript, 0, move)

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

package revgame

import (
	"testing"
	"github.com/pkg/errors"
)

func TestBoard_DrawBoard(t *testing.T) {
	board := OthelloBoard

	t.Log(board.DrawBoard("*", "O", "", "", "\n"))
}

func TestBoard_MakeMove(t *testing.T) {
	testSteps := []struct {
		player Disk
		move address
		err error
		cause error
	}{
		{
			player: Black,
			move: address{1, 1},
			cause: ErrNotValidMove,
		},
		{
			player: Black,
			move: address{3, 2},
		},
		{
			player: White,
			move: address{2, 4},
		},
		{
			player: Black,
			move: address{2, 5},
		},
	}

	board := OthelloBoard

	var err error

	for i, step := range testSteps {
		var newBoard Board
		newBoard, err = board.MakeMove(step.player, step.move.x, step.move.y)
		if err != nil {
			if newBoard != board {
				t.Fatalf("err != nil && newBoard != board")
			}
		}
		if step.err != nil {
			if step.err != err {
				t.Errorf("Step #%v: expected error [%v], got: %v", i+1, step.err, err)
			}
			continue
		} else if step.cause != nil {
			if errors.Cause(err) != step.cause {
				t.Errorf("Step #%v: expected error cause [%v], got: %v", i+1, step.cause, errors.Cause(err))
			}
			continue
		} else if err != nil {
			t.Errorf("Step #%v: unexpected error: %v", i+1, err)
			continue
		}

		if newBoard.Whites == 0 && newBoard.Blacks == 0 {
			t.Fatalf("Step %v: board == 0", i+1)
		}
		if newBoard == board {
			t.Fatalf("Step #%v: newBoard == board", i+1)
		}
		t.Logf("Step #%v:%v", i+1, newBoard.DrawBoard("*", "O", ".", "", "\n"))
		board = newBoard
	}
}

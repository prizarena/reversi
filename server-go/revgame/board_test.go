package revgame

import (
	"testing"
	"github.com/pkg/errors"
	"github.com/prizarena/turn-based"
)

func TestBoard_DrawBoard(t *testing.T) {
	board := OthelloBoard

	t.Log(board.DrawBoard("*", "O", "", "", "\n"))
}

func TestBoard_MakeMove(t *testing.T) {
	testSteps := []struct {
		player Disk
		move   Address
		err    error
		cause  error
	}{
		{
			player: Black,
			move:   Address{1, 1},
			cause:  ErrNotValidMove,
		},
		{
			player: Black,
			move:   Address{3, 2},
		},
		{
			player: White,
			move:   Address{2, 4},
		},
		{
			player: Black,
			move:   Address{2, 5},
		},
	}

	board := OthelloBoard

	var err error

	for i, step := range testSteps {
		var newBoard Board
		newBoard, err = board.MakeMove(step.player, step.move)
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
		//t.Logf("Step #%v:%v", i+1, newBoard.DrawBoard("*", "O", ".", "", "\n"))
		board = newBoard
	}
}

func TestBoard_UndoMove(t *testing.T) {
	validateBoard := func(b Board) {
		if (b.Whites & b.Blacks) != 0 {
			whites := b
			whites.Blacks = 0
			blacks := b
			blacks.Whites = 0
			t.Fatalf("board.Whites | board.Blacks:\nblacks:\n%v\nwhites:\n%v",
				blacks.DrawBoard("*", "O", "", "", "\n"),
				whites.DrawBoard("*", "O", "", "", "\n"),
			)
		}
	}

	steps := []struct {
		p Disk
		ca turnbased.CellAddress
		board Board
	} {
		{p: White, board: OthelloBoard},
		{p: Black, ca: "F5"},
		{p: White, ca: "F4"},
	}
	var err error
	var board Board

	for i, step := range steps {
		if i == 0 {
			board = step.board
			continue
		}
		a := CellAddressToRevAddress(step.ca)
		if board, err = board.MakeMove(step.p, a); err != nil {
			t.Fatalf("uexpeced err at step %v(%v=%v): %v", i+1, step.p, step.ca, err)
		}
		validateBoard(board)
		steps[i].board = board
		//t.Logf("Step #%v%v", i, board.DrawBoard("*", "O", "", "", "\n"))
	}
	for i := len(steps)-1; i > 0; i-- {
		step := steps[i]
		a := CellAddressToRevAddress(step.ca)
		prevStep := steps[i-1]
		board = board.UndoMove(a, prevStep.board.Last)
		validateBoard(board)
		if board.Last != prevStep.board.Last {
			t.Errorf("Invalid undo at step %v: Expected.Last:%+v != board.Last:%+v", i+1, prevStep.board.Last, board.Last)
		}
		if board != prevStep.board {
			t.Fatalf("Invalid undo at step %v:\nExpected: %v\n Got: %v",
				i+1,
				prevStep.board.DrawBoard("*", "O", "", "", "\n"),
				board.DrawBoard("*", "O", "", "", "\n"))
		}
	}
}

func TestNewBoardFromBase64(t *testing.T) {
	board := OthelloBoard
	s := board.ToBase64()
	t.Log(s)
	board2, err := NewBoardFromBase64(s)
	if err != nil {
		t.Fatal(err)
	}
	if board2.Blacks != board.Blacks {
		t.Errorf("board2.Blacks != board.Blacks")
	}
	if board2.Whites != board.Whites {
		t.Errorf("board2.Whites != board.Whites")
	}
}

func TestOthelloBoard(t *testing.T) {
	if p := OthelloBoard.NextPlayer(); p != Black {
		t.Fatalf("OthelloBoard expected to return Black  (%v) as NextPlayer(), got: %v", Black, p)
	}
}
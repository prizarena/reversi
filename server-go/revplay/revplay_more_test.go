package revplay

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/prizarena/reversi/server-go/revgame"
)

// --- statusText branches -----------------------------------------------------

func TestStatusText_WhiteToMove(t *testing.T) {
	// OthelloBoard with Last set to a black cell (4,3) makes White the next player.
	board := revgame.Board{Blacks: revgame.OthelloBoard.Blacks, Whites: revgame.OthelloBoard.Whites, Last: revgame.Address{X: 4, Y: 3}}
	if board.NextPlayer() != revgame.White {
		t.Fatalf("precondition: NextPlayer = %v, want White", board.NextPlayer())
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI})
	if !strings.Contains(got, "Opponent (White) to move") {
		t.Errorf("statusText = %q, want it to mention White to move", got)
	}
}

func TestStatusText_Draw(t *testing.T) {
	// Two isolated single disks, one each colour: game over, equal score => draw.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: revgame.Disks(1 << 7), Last: revgame.Address{X: 0, Y: 0}}
	if !board.IsCompleted() {
		t.Fatalf("precondition: board should be completed")
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI})
	if !strings.Contains(strings.ToLower(got), "draw") {
		t.Errorf("statusText = %q, want it to announce a draw", got)
	}
}

func TestStatusText_WhiteWins(t *testing.T) {
	// One black disk, two isolated white disks: game over, White ahead.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: revgame.Disks(1<<7) | revgame.Disks(1<<15), Last: revgame.Address{X: 0, Y: 0}}
	if !board.IsCompleted() {
		t.Fatalf("precondition: board should be completed")
	}
	black, white := board.Scores()
	if white <= black {
		t.Fatalf("precondition: expected white>black, got black=%d white=%d", black, white)
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI})
	if !strings.Contains(got, "White wins") {
		t.Errorf("statusText = %q, want it to announce White wins", got)
	}
}

// --- HumanMove: not Black's turn is a no-op -----------------------------------

func TestHumanMove_NotBlacksTurn(t *testing.T) {
	board := revgame.Board{Blacks: revgame.OthelloBoard.Blacks, Whites: revgame.OthelloBoard.Whites, Last: revgame.Address{X: 4, Y: 3}}
	if board.NextPlayer() != revgame.White {
		t.Fatalf("precondition: NextPlayer = %v, want White", board.NextPlayer())
	}
	s := Snapshot{Board: board, Opponent: OpponentAI}
	next, applied, err := s.HumanMove(revgame.Address{X: 2, Y: 3}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if applied {
		t.Error("HumanMove should not apply when it is not Black's turn")
	}
	if next != s {
		t.Errorf("snapshot changed when it was not Black's turn:\n got %+v\nwant %+v", next, s)
	}
}

// --- HumanMove: drive a full game to completion -------------------------------

func TestHumanMove_PlaysToCompletion(t *testing.T) {
	s := NewGame(OpponentRandom)
	rnd := rand.New(rand.NewSource(20240723))

	// After each HumanMove it is either game over or Black's turn (NextPlayer
	// resolves passes), so Black always has a legal move to play here.
	for i := 0; i < 100 && !s.IsGameOver(); i++ {
		moves := s.Board.ValidMoves(revgame.Black)
		if len(moves) == 0 {
			t.Fatalf("iteration %d: Black has no move but game is not over", i)
		}
		next, applied, err := s.HumanMove(moves[0], rnd)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if !applied {
			t.Fatalf("iteration %d: legal move %v was not applied", i, moves[0])
		}
		s = next
	}

	if !s.IsGameOver() {
		t.Fatal("game did not reach completion within the move budget")
	}
	black, white := s.Board.Scores()
	if black+white > 64 {
		t.Errorf("final disk total %d exceeds 64", black+white)
	}
	// A completed snapshot renders without panicking and announces a result.
	txt := statusText(s)
	if !strings.Contains(txt, "Game over") {
		t.Errorf("completed status = %q, want it to say the game is over", txt)
	}
}

// --- opponentMove: AI with no legal move --------------------------------------

func TestOpponentMoveAI_NoMove(t *testing.T) {
	// No white disks at all: the AI opponent has no legal move.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: 0, Last: revgame.Address{X: 0, Y: 0}}
	if _, ok := opponentMove(OpponentAI, board, revgame.White, nil); ok {
		t.Error("opponentMove(AI) reported a move where White has none")
	}
}

// --- intn: nil rnd falls back to the global source ----------------------------

func TestIntn_NilRnd(t *testing.T) {
	if got := intn(nil, 1); got != 0 {
		t.Errorf("intn(nil, 1) = %d, want 0", got)
	}
	for i := 0; i < 100; i++ {
		if got := intn(nil, 5); got < 0 || got >= 5 {
			t.Fatalf("intn(nil, 5) = %d, out of range [0,5)", got)
		}
	}
}

// --- DecodeSnapshot: remaining error branches ---------------------------------

func TestDecodeSnapshot_ErrorBranches(t *testing.T) {
	// Missing separator.
	if _, err := DecodeSnapshot("abc"); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("missing separator: err = %v, want ErrInvalidSnapshot", err)
	}
	// Unknown opponent.
	if _, err := DecodeSnapshot("z." + revgame.OthelloBoard.ToBase64()); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("unknown opponent: err = %v, want ErrInvalidSnapshot", err)
	}
	// Valid opponent, invalid board base64.
	if _, err := DecodeSnapshot("a.###"); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("bad board base64: err = %v, want ErrInvalidSnapshot", err)
	}
}

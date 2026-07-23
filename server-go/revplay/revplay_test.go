package revplay

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-go-core/botkb"
	"github.com/prizarena/reversi/server-go/revgame"
)

func containsAddr(addrs []revgame.Address, a revgame.Address) bool {
	for _, x := range addrs {
		if x == a {
			return true
		}
	}
	return false
}

// bitAt / isBlack / isWhite query a single cell using the engine's bit layout
// (bit = 1 << (y*8 + x)); the engine keeps its per-cell disk query unexported.
func bitAt(a revgame.Address) revgame.Disks           { return revgame.Disks(1) << uint(a.Y*8+a.X) }
func isBlack(b revgame.Board, a revgame.Address) bool { return b.Blacks&bitAt(a) != 0 }
func isWhite(b revgame.Board, a revgame.Address) bool { return b.Whites&bitAt(a) != 0 }

// --- Snapshot round-trip -----------------------------------------------------

func TestSnapshotRoundTrip(t *testing.T) {
	midBoard, err := revgame.OthelloBoard.MakeMove(revgame.Black, revgame.Address{X: 3, Y: 2})
	if err != nil {
		t.Fatalf("failed to build mid-game board: %v", err)
	}

	// A completely full board: rows 0-3 black, rows 4-7 white, last on a white cell.
	fullBoard := revgame.Board{
		Blacks: revgame.Disks(0xFFFFFFFF),
		Whites: revgame.Disks(^int64(0xFFFFFFFF)),
		Last:   revgame.Address{X: 7, Y: 7},
	}

	cases := []struct {
		name string
		snap Snapshot
	}{
		{"start-ai", Snapshot{Board: revgame.OthelloBoard, Opponent: OpponentAI}},
		{"mid-random", Snapshot{Board: midBoard, Opponent: OpponentRandom}},
		{"full-ai", Snapshot{Board: fullBoard, Opponent: OpponentAI}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := tc.snap.Encode()
			if len(encoded) > 40 {
				t.Errorf("encoded snapshot too long: %d chars (%q)", len(encoded), encoded)
			}
			decoded, err := DecodeSnapshot(encoded)
			if err != nil {
				t.Fatalf("DecodeSnapshot(%q) error: %v", encoded, err)
			}
			if decoded.Board != tc.snap.Board {
				t.Errorf("board mismatch after round-trip:\n got %+v\nwant %+v", decoded.Board, tc.snap.Board)
			}
			if decoded.Opponent != tc.snap.Opponent {
				t.Errorf("opponent mismatch: got %q want %q", decoded.Opponent, tc.snap.Opponent)
			}
		})
	}
}

func TestDecodeSnapshotRejectsGarbage(t *testing.T) {
	if _, err := DecodeSnapshot(""); err == nil {
		t.Error("expected error decoding empty string")
	}
	if _, err := DecodeSnapshot("x.notbase64!!!"); err == nil {
		t.Error("expected error decoding unknown opponent / bad board")
	}
}

// --- New game ----------------------------------------------------------------

func TestNewGameStartPosition(t *testing.T) {
	s := NewGame(OpponentAI)
	if s.Opponent != OpponentAI {
		t.Errorf("opponent = %q, want %q", s.Opponent, OpponentAI)
	}
	if s.Board != revgame.OthelloBoard {
		t.Errorf("board = %+v, want OthelloBoard %+v", s.Board, revgame.OthelloBoard)
	}
	black, white := s.Board.Scores()
	if black != 2 || white != 2 {
		t.Errorf("start scores = (%d,%d), want (2,2)", black, white)
	}
	// d5/e4 black, d4/e5 white (x: d=3,e=4 ; y: row5=4, row4=3).
	blackCells := []revgame.Address{{X: 3, Y: 4}, {X: 4, Y: 3}}
	whiteCells := []revgame.Address{{X: 3, Y: 3}, {X: 4, Y: 4}}
	for _, a := range blackCells {
		if !isBlack(s.Board, a) {
			t.Errorf("expected black disk at %v", a)
		}
	}
	for _, a := range whiteCells {
		if !isWhite(s.Board, a) {
			t.Errorf("expected white disk at %v", a)
		}
	}
	// Human (Black) moves first.
	if s.Board.NextPlayer() != revgame.Black {
		t.Errorf("NextPlayer = %v, want Black", s.Board.NextPlayer())
	}
}

// --- Legal human move flips and triggers exactly one opponent reply ----------

func TestLegalHumanMoveFlipsAndTriggersOneReply(t *testing.T) {
	s := NewGame(OpponentRandom)
	rnd := rand.New(rand.NewSource(42))

	// d3 = (x=3,y=2) is a legal Black opening that flanks the white on d4.
	next, applied, err := s.HumanMove(revgame.Address{X: 3, Y: 2}, rnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Fatal("expected the legal move to be applied")
	}
	// Start had 4 disks. Black placement (+1) and one White reply placement (+1) => 6.
	if got := next.Board.DisksCount(); got != 6 {
		t.Errorf("disk count = %d, want 6 (one black move + exactly one white reply)", got)
	}
	// After the single white reply it is the human's turn again.
	if next.Board.NextPlayer() != revgame.Black {
		t.Errorf("NextPlayer = %v, want Black", next.Board.NextPlayer())
	}
	// The board must actually have changed.
	if next.Board == s.Board {
		t.Error("board unchanged after a legal move")
	}
}

// --- Illegal / occupied tap is a no-op ---------------------------------------

func TestIllegalTapIsNoop(t *testing.T) {
	s := NewGame(OpponentAI)

	// Occupied cell d4 = (3,3) (a white disk).
	next, applied, err := s.HumanMove(revgame.Address{X: 3, Y: 3}, nil)
	if err != nil {
		t.Fatalf("unexpected error on occupied tap: %v", err)
	}
	if applied {
		t.Error("occupied tap should not be applied")
	}
	if next != s {
		t.Errorf("snapshot changed on occupied tap:\n got %+v\nwant %+v", next, s)
	}

	// Empty but illegal cell a1 = (0,0).
	next2, applied2, err2 := s.HumanMove(revgame.Address{X: 0, Y: 0}, nil)
	if err2 != nil {
		t.Fatalf("unexpected error on illegal tap: %v", err2)
	}
	if applied2 {
		t.Error("illegal empty tap should not be applied")
	}
	if next2 != s {
		t.Error("snapshot changed on illegal empty tap")
	}
}

// --- Auto-pass: White has no reply, stays Black's turn, clean flip -----------

func TestAutoPassWhenOpponentHasNoMove(t *testing.T) {
	// Row 0: (0,0) empty, (1,0) white, (2,0..7,0) black -> Black can play (0,0),
	// flipping the white at (1,0); white cannot move (blacks run to the edge).
	// Row 2: same shape with a second white at (1,2) that survives, so after
	// Black's move it is still Black's turn (White passes).
	board := revgame.Board{
		Blacks: revgame.Disks(0xFC | (0xFC << 16)),
		Whites: revgame.Disks((1 << 1) | (1 << 17)),
		Last:   revgame.Address{X: 1, Y: 0}, // a white cell => Black to move
	}
	if board.NextPlayer() != revgame.Black {
		t.Fatalf("precondition: NextPlayer = %v, want Black", board.NextPlayer())
	}
	if board.HasValidMoves(revgame.White) {
		t.Fatal("precondition: White should have no move on this board")
	}

	s := Snapshot{Board: board, Opponent: OpponentRandom}
	rnd := rand.New(rand.NewSource(7))

	beforeWhite := board.Score(revgame.White) // 2
	next, applied, err := s.HumanMove(revgame.Address{X: 0, Y: 0}, rnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Fatal("expected the Black move to be applied")
	}
	// The flanked white at (1,0) must have flipped to black (clean, no reply).
	if !isBlack(next.Board, revgame.Address{X: 1, Y: 0}) {
		t.Error("flanked white at (1,0) should have flipped to black")
	}
	if isWhite(next.Board, revgame.Address{X: 1, Y: 0}) {
		t.Error("cell (1,0) should no longer be white")
	}
	// No White disk was added: white count only went DOWN (one flipped), never up.
	if afterWhite := next.Board.Score(revgame.White); afterWhite >= beforeWhite {
		t.Errorf("white score = %d, want < %d (white passed, one flipped)", afterWhite, beforeWhite)
	}
	if next.Board.HasValidMoves(revgame.White) {
		t.Error("White still has no move after Black's move")
	}
	// It remains the human's turn.
	if next.Board.NextPlayer() != revgame.Black {
		t.Errorf("NextPlayer = %v, want Black (White auto-passed)", next.Board.NextPlayer())
	}
}

// --- Opponent move helper: AI routes to SimpleAI, Random is always legal ------

func TestOpponentMoveAI(t *testing.T) {
	// Board where White has exactly one legal move: (0,0) flanking black (1,0)
	// against white (2,0). SimpleAI must return that unique move.
	board := revgame.Board{
		Blacks: revgame.Disks(1 << 1), // (1,0)
		Whites: revgame.Disks(1 << 2), // (2,0)
		Last:   revgame.Address{X: 2, Y: 0},
	}
	legal := board.ValidMoves(revgame.White)
	if len(legal) != 1 || legal[0] != (revgame.Address{X: 0, Y: 0}) {
		t.Fatalf("precondition: expected the single white move (0,0), got %v", legal)
	}
	got, ok := opponentMove(OpponentAI, board, revgame.White, nil)
	if !ok {
		t.Fatal("opponentMove(AI) reported no move")
	}
	if got != (revgame.Address{X: 0, Y: 0}) {
		t.Errorf("opponentMove(AI) = %v, want (0,0)", got)
	}
	if !containsAddr(legal, got) {
		t.Errorf("opponentMove(AI) returned illegal move %v", got)
	}
}

func TestOpponentMoveRandomAlwaysLegal(t *testing.T) {
	board, err := revgame.OthelloBoard.MakeMove(revgame.Black, revgame.Address{X: 3, Y: 2})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	legal := board.ValidMoves(revgame.White)
	if len(legal) < 2 {
		t.Fatalf("precondition: expected White to have several moves, got %v", legal)
	}
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		got, ok := opponentMove(OpponentRandom, board, revgame.White, rnd)
		if !ok {
			t.Fatalf("iter %d: reported no move", i)
		}
		if !containsAddr(legal, got) {
			t.Fatalf("iter %d: random move %v is not legal", i, got)
		}
	}
}

func TestOpponentMoveNoMove(t *testing.T) {
	// A board with no white disks at all: White has no move.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: 0, Last: revgame.Address{X: 0, Y: 0}}
	if _, ok := opponentMove(OpponentRandom, board, revgame.White, rand.New(rand.NewSource(1))); ok {
		t.Error("expected no move for White on a board with no white targets")
	}
}

// --- Rendering ---------------------------------------------------------------

func TestRenderProduces8x8Keyboard(t *testing.T) {
	s := NewGame(OpponentAI)
	cellData := func(a revgame.Address) string { return fmt.Sprintf("d%02d", a.Index()) }

	r := Render(s, cellData)

	kb, ok := r.Keyboard.(*botkb.MessageKeyboard)
	if !ok {
		t.Fatalf("keyboard is %T, want *botkb.MessageKeyboard", r.Keyboard)
	}
	if len(kb.Buttons) != 8 {
		t.Fatalf("keyboard has %d rows, want 8", len(kb.Buttons))
	}
	for y := 0; y < 8; y++ {
		if len(kb.Buttons[y]) != 8 {
			t.Fatalf("row %d has %d buttons, want 8", y, len(kb.Buttons[y]))
		}
		for x := 0; x < 8; x++ {
			btn, ok := kb.Buttons[y][x].(*botkb.DataButton)
			if !ok {
				t.Fatalf("button [%d][%d] is %T, want *botkb.DataButton", y, x, kb.Buttons[y][x])
			}
			want := cellData(revgame.Address{X: int8(x), Y: int8(y)})
			if btn.Data != want {
				t.Errorf("button [%d][%d] data = %q, want %q", y, x, btn.Data, want)
			}
			if btn.Text == "" {
				t.Errorf("button [%d][%d] has empty face text", y, x)
			}
		}
	}
	if r.Text == "" {
		t.Error("rendered text is empty")
	}
	// Score line should mention the current score.
	if !strings.Contains(r.Text, "2") {
		t.Errorf("rendered text %q should include the start score", r.Text)
	}
}

func TestRenderCompletedBoardNoPanicAndShowsResult(t *testing.T) {
	// A completed board that still has empty cells (only black disks present).
	// This exercises the game-over path without tripping the engine's
	// getValidMoves panic on a Completed next-player.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: 0, Last: revgame.Address{X: 0, Y: 0}}
	if !board.IsCompleted() {
		t.Fatalf("precondition: board should be completed")
	}
	s := Snapshot{Board: board, Opponent: OpponentAI}
	r := Render(s, func(a revgame.Address) string { return "x" })
	if kb, ok := r.Keyboard.(*botkb.MessageKeyboard); !ok || len(kb.Buttons) != 8 {
		t.Fatalf("expected an 8-row keyboard, got %T", r.Keyboard)
	}
	if !strings.Contains(strings.ToLower(r.Text), "win") {
		t.Errorf("completed text %q should announce a winner", r.Text)
	}
	if !s.IsGameOver() {
		t.Error("IsGameOver() should be true for a completed board")
	}
}

func TestNewGameNotOver(t *testing.T) {
	if NewGame(OpponentAI).IsGameOver() {
		t.Error("a fresh game must not be over")
	}
}

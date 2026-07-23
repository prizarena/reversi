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
		{"start-ai-black", Snapshot{Board: revgame.OthelloBoard, Opponent: OpponentAI, Human: revgame.Black}},
		{"start-ai-white", Snapshot{Board: revgame.OthelloBoard, Opponent: OpponentAI, Human: revgame.White}},
		{"mid-random-white", Snapshot{Board: midBoard, Opponent: OpponentRandom, Human: revgame.White}},
		{"full-ai-black", Snapshot{Board: fullBoard, Opponent: OpponentAI, Human: revgame.Black}},
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
			if decoded.Human != tc.snap.Human {
				t.Errorf("human mismatch: got %v want %v", decoded.Human, tc.snap.Human)
			}
		})
	}
}

func TestDecodeSnapshotRejectsGarbage(t *testing.T) {
	if _, err := DecodeSnapshot(""); err == nil {
		t.Error("expected error decoding empty string")
	}
	if _, err := DecodeSnapshot("xy.notbase64!!!"); err == nil {
		t.Error("expected error decoding unknown opponent / bad board")
	}
}

// --- New game ----------------------------------------------------------------

// assertStartPosition checks the four Othello opening disks are in place.
func assertStartPosition(t *testing.T, b revgame.Board) {
	t.Helper()
	black, white := b.Scores()
	if black != 2 || white != 2 {
		t.Errorf("start scores = (%d,%d), want (2,2)", black, white)
	}
	// d5/e4 black, d4/e5 white (x: d=3,e=4 ; y: row5=4, row4=3).
	for _, a := range []revgame.Address{{X: 3, Y: 4}, {X: 4, Y: 3}} {
		if !isBlack(b, a) {
			t.Errorf("expected black disk at %v", a)
		}
	}
	for _, a := range []revgame.Address{{X: 3, Y: 3}, {X: 4, Y: 4}} {
		if !isWhite(b, a) {
			t.Errorf("expected white disk at %v", a)
		}
	}
}

func TestNewGameHumanBlack(t *testing.T) {
	s := NewGame(OpponentAI, revgame.Black)
	if s.Opponent != OpponentAI {
		t.Errorf("opponent = %q, want %q", s.Opponent, OpponentAI)
	}
	if s.Board != revgame.OthelloBoard {
		t.Errorf("board = %+v, want OthelloBoard %+v", s.Board, revgame.OthelloBoard)
	}
	assertStartPosition(t, s.Board)
	if s.HumanColor() != revgame.Black {
		t.Errorf("HumanColor = %v, want Black", s.HumanColor())
	}
	if s.OpponentColor() != revgame.White {
		t.Errorf("OpponentColor = %v, want White", s.OpponentColor())
	}
	// Human (Black) moves first.
	if !s.IsHumanTurn() {
		t.Error("IsHumanTurn should be true for human=Black at the start")
	}
	if s.IsOpponentTurn() {
		t.Error("IsOpponentTurn should be false for human=Black at the start")
	}
}

func TestNewGameHumanWhite(t *testing.T) {
	s := NewGame(OpponentRandom, revgame.White)
	if s.Board != revgame.OthelloBoard {
		t.Errorf("board = %+v, want OthelloBoard %+v", s.Board, revgame.OthelloBoard)
	}
	assertStartPosition(t, s.Board)
	if s.HumanColor() != revgame.White {
		t.Errorf("HumanColor = %v, want White", s.HumanColor())
	}
	if s.OpponentColor() != revgame.Black {
		t.Errorf("OpponentColor = %v, want Black", s.OpponentColor())
	}
	// Black moves first, so the opponent is to move — not the human.
	if !s.IsOpponentTurn() {
		t.Error("IsOpponentTurn should be true for human=White at the start (Black moves first)")
	}
	if s.IsHumanTurn() {
		t.Error("IsHumanTurn should be false for human=White at the start")
	}
}

// --- ApplyHumanMove: applies only the human move, opponent NOT yet played -----

func TestApplyHumanMoveBlackFlipsAndYieldsOpponentTurn(t *testing.T) {
	s := NewGame(OpponentRandom, revgame.Black)

	// d3 = (x=3,y=2) is a legal Black opening that flanks the white on d4.
	before := s.Board.DisksCount() // 4
	next, applied, err := s.ApplyHumanMove(revgame.Address{X: 3, Y: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Fatal("expected the legal move to be applied")
	}
	// Exactly the human's placement was added (no opponent disc yet): +1 disc.
	if got := next.Board.DisksCount(); got != before+1 {
		t.Errorf("disk count = %d, want %d (only the human's move, no opponent reply)", got, before+1)
	}
	// The placed black disc is at d3, and the flanked white on d4 flipped to black.
	if !isBlack(next.Board, revgame.Address{X: 3, Y: 2}) {
		t.Error("expected the human's black disc at (3,2)")
	}
	if !isBlack(next.Board, revgame.Address{X: 3, Y: 3}) {
		t.Error("expected the flanked disc at (3,3) to have flipped to black")
	}
	// The opponent has NOT moved yet: it is now the opponent's turn.
	if !next.IsOpponentTurn() {
		t.Error("after the human's move it should be the opponent's turn")
	}
	if next.IsHumanTurn() {
		t.Error("after the human's move it should not still be the human's turn")
	}
}

func TestApplyHumanMoveWhite(t *testing.T) {
	// Drive Black's opening so it becomes White's turn, then let the human (White) move.
	board, err := revgame.OthelloBoard.MakeMove(revgame.Black, revgame.Address{X: 3, Y: 2})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	s := Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.White}
	if !s.IsHumanTurn() {
		t.Fatalf("precondition: expected human (White) to be to move")
	}
	moves := s.Board.ValidMoves(revgame.White)
	if len(moves) == 0 {
		t.Fatal("precondition: White should have a legal move")
	}
	before := s.Board.DisksCount()
	next, applied, err := s.ApplyHumanMove(moves[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Fatalf("legal white move %v was not applied", moves[0])
	}
	if got := next.Board.DisksCount(); got != before+1 {
		t.Errorf("disk count = %d, want %d (only the human's move)", got, before+1)
	}
	if !isWhite(next.Board, moves[0]) {
		t.Errorf("expected a white disc at the played cell %v", moves[0])
	}
	// Opponent (Black) has not replied yet.
	if !next.IsOpponentTurn() {
		t.Error("after the human's white move it should be the opponent's (Black) turn")
	}
}

// --- ApplyHumanMove: illegal / occupied / not-your-turn are no-ops -----------

func TestApplyHumanMoveIllegalTapIsNoop(t *testing.T) {
	s := NewGame(OpponentAI, revgame.Black)

	// Occupied cell d4 = (3,3) (a white disk).
	next, applied, err := s.ApplyHumanMove(revgame.Address{X: 3, Y: 3})
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
	next2, applied2, err2 := s.ApplyHumanMove(revgame.Address{X: 0, Y: 0})
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

func TestApplyHumanMoveNotHumansTurn(t *testing.T) {
	// Human is White but Black (the opponent) is to move on the fresh board.
	s := NewGame(OpponentAI, revgame.White)
	if !s.IsOpponentTurn() {
		t.Fatalf("precondition: expected the opponent to be to move")
	}
	// (2,3) is a legal Black opening, but the human plays White — must be a no-op.
	next, applied, err := s.ApplyHumanMove(revgame.Address{X: 2, Y: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if applied {
		t.Error("ApplyHumanMove should not apply when it is not the human's turn")
	}
	if next != s {
		t.Errorf("snapshot changed when it was not the human's turn:\n got %+v\nwant %+v", next, s)
	}
}

// --- PlayOpponent: exactly one reply in the normal case ----------------------

func TestPlayOpponentPlaysOneReply(t *testing.T) {
	s := NewGame(OpponentRandom, revgame.Black)
	rnd := rand.New(rand.NewSource(42))

	afterHuman, applied, err := s.ApplyHumanMove(revgame.Address{X: 3, Y: 2})
	if err != nil || !applied {
		t.Fatalf("setup ApplyHumanMove: applied=%v err=%v", applied, err)
	}
	before := afterHuman.Board.DisksCount() // 5

	next, moved, err := afterHuman.PlayOpponent(rnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !moved {
		t.Fatal("expected the opponent to move")
	}
	// Exactly one opponent placement: +1 disc.
	if got := next.Board.DisksCount(); got != before+1 {
		t.Errorf("disk count = %d, want %d (exactly one opponent reply)", got, before+1)
	}
	// Control returns to the human (or the game is over — not here).
	if !next.IsHumanTurn() {
		t.Errorf("after the opponent's reply it should be the human's turn; NextPlayer=%v", next.Board.NextPlayer())
	}
}

func TestPlayOpponentNoopOnHumansTurn(t *testing.T) {
	s := NewGame(OpponentAI, revgame.Black) // human (Black) to move
	next, moved, err := s.PlayOpponent(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if moved {
		t.Error("PlayOpponent should not move on the human's turn")
	}
	if next != s {
		t.Errorf("snapshot changed on a PlayOpponent no-op:\n got %+v\nwant %+v", next, s)
	}
}

func TestPlayOpponentWhiteHumanFirstReply(t *testing.T) {
	// Human is White: the opponent (Black) makes the opening move.
	s := NewGame(OpponentRandom, revgame.White)
	rnd := rand.New(rand.NewSource(99))
	if !s.IsOpponentTurn() {
		t.Fatalf("precondition: opponent (Black) should move first")
	}
	next, moved, err := s.PlayOpponent(rnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !moved {
		t.Fatal("expected the opponent to open")
	}
	if next.Board.DisksCount() != 5 {
		t.Errorf("disk count = %d, want 5 (one Black opening move)", next.Board.DisksCount())
	}
	if !next.IsHumanTurn() {
		t.Errorf("after Black's opening it should be the human's (White) turn; NextPlayer=%v", next.Board.NextPlayer())
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
	s := NewGame(OpponentAI, revgame.Black)
	cellData := func(a revgame.Address) string { return fmt.Sprintf("d%02d", a.Index()) }

	r := Render(s, cellData)

	kb, ok := r.Keyboard.(*botkb.MessageKeyboard)
	if !ok {
		t.Fatalf("keyboard is %T, want *botkb.MessageKeyboard", r.Keyboard)
	}
	if len(kb.Buttons) != 8 {
		t.Fatalf("keyboard has %d rows, want 8", len(kb.Buttons))
	}
	sawHint := false
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
			switch btn.Text {
			case glyphBlack, glyphWhite, glyphEmpty, glyphMove:
				// expected faces
			default:
				t.Errorf("button [%d][%d] face = %q, not a known glyph", y, x, btn.Text)
			}
			if btn.Text == glyphMove {
				sawHint = true
			}
		}
	}
	// Human (Black) is to move, so legal-move hints must be present.
	if !sawHint {
		t.Error("expected 🟢 legal-move hints on the human's turn")
	}
	if r.Text == "" {
		t.Error("rendered text is empty")
	}
	// Score line should mention the current score.
	if !strings.Contains(r.Text, "2") {
		t.Errorf("rendered text %q should include the start score", r.Text)
	}
}

func TestRenderNoHintsOnOpponentTurn(t *testing.T) {
	// Human is White; on the fresh board it is the opponent's (Black) turn, so
	// there must be no 🟢 hints.
	s := NewGame(OpponentAI, revgame.White)
	r := Render(s, func(a revgame.Address) string { return "x" })
	kb, ok := r.Keyboard.(*botkb.MessageKeyboard)
	if !ok {
		t.Fatalf("keyboard is %T, want *botkb.MessageKeyboard", r.Keyboard)
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			btn := kb.Buttons[y][x].(*botkb.DataButton)
			if btn.Text == glyphMove {
				t.Errorf("unexpected 🟢 hint at [%d][%d] on the opponent's turn", y, x)
			}
		}
	}
	if !strings.Contains(r.Text, "Opponent (Black) to move") {
		t.Errorf("rendered text %q should say the opponent (Black) is to move", r.Text)
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
	s := Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.Black}
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
	if NewGame(OpponentAI, revgame.Black).IsGameOver() {
		t.Error("a fresh game must not be over")
	}
}

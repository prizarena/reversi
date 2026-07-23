package revplay

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/prizarena/reversi/server-go/revgame"
)

// --- statusText branches -----------------------------------------------------

func TestStatusText_YourTurnBlack(t *testing.T) {
	got := statusText(NewGame(OpponentAI, revgame.Black))
	if !strings.Contains(got, "Your turn ("+glyphBlack+" Black)") {
		t.Errorf("statusText = %q, want it to say it is the human's (Black) turn", got)
	}
}

func TestStatusText_YourTurnWhite(t *testing.T) {
	// OthelloBoard with Last on a black cell (4,3) makes White the next player.
	board := revgame.Board{Blacks: revgame.OthelloBoard.Blacks, Whites: revgame.OthelloBoard.Whites, Last: revgame.Address{X: 4, Y: 3}}
	if board.NextPlayer() != revgame.White {
		t.Fatalf("precondition: NextPlayer = %v, want White", board.NextPlayer())
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.White})
	if !strings.Contains(got, "Your turn ("+glyphWhite+" White)") {
		t.Errorf("statusText = %q, want it to say it is the human's (White) turn", got)
	}
}

func TestStatusText_OpponentWhiteToMove(t *testing.T) {
	// White to move, but the human plays Black => the opponent (White) is to move.
	board := revgame.Board{Blacks: revgame.OthelloBoard.Blacks, Whites: revgame.OthelloBoard.Whites, Last: revgame.Address{X: 4, Y: 3}}
	if board.NextPlayer() != revgame.White {
		t.Fatalf("precondition: NextPlayer = %v, want White", board.NextPlayer())
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.Black})
	if !strings.Contains(got, "Opponent (White) to move") {
		t.Errorf("statusText = %q, want it to mention the White opponent to move", got)
	}
}

func TestStatusText_OpponentBlackToMove(t *testing.T) {
	// Fresh board (Black to move) but the human plays White => opponent (Black) to move.
	got := statusText(NewGame(OpponentAI, revgame.White))
	if !strings.Contains(got, "Opponent (Black) to move") {
		t.Errorf("statusText = %q, want it to mention the Black opponent to move", got)
	}
}

func TestStatusText_Draw(t *testing.T) {
	// Two isolated single disks, one each colour: game over, equal score => draw.
	board := revgame.Board{Blacks: revgame.Disks(1 << 0), Whites: revgame.Disks(1 << 7), Last: revgame.Address{X: 0, Y: 0}}
	if !board.IsCompleted() {
		t.Fatalf("precondition: board should be completed")
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.Black})
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
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.White})
	if !strings.Contains(got, "White wins") {
		t.Errorf("statusText = %q, want it to announce White wins", got)
	}
}

func TestStatusText_BlackWins(t *testing.T) {
	// Two isolated black disks, one white disk: game over, Black ahead.
	board := revgame.Board{Blacks: revgame.Disks(1<<0) | revgame.Disks(1<<8), Whites: revgame.Disks(1 << 15), Last: revgame.Address{X: 0, Y: 0}}
	if !board.IsCompleted() {
		t.Fatalf("precondition: board should be completed")
	}
	got := statusText(Snapshot{Board: board, Opponent: OpponentAI, Human: revgame.Black})
	if !strings.Contains(got, "Black wins") {
		t.Errorf("statusText = %q, want it to announce Black wins", got)
	}
}

// --- Auto-pass in ApplyHumanMove: opponent has no reply, stays the human's turn ---

func TestApplyHumanMove_OpponentAutoPasses(t *testing.T) {
	// Row 0: (0,0) empty, (1,0) white, (2,0..7,0) black -> Black can play (0,0),
	// flipping the white at (1,0); White then has no move. Row 2 carries a second
	// white at (1,2) that survives, so Black still has a move afterwards and it
	// remains the human's (Black) turn.
	board := revgame.Board{
		Blacks: revgame.Disks(0xFC | (0xFC << 16)),
		Whites: revgame.Disks((1 << 1) | (1 << 17)),
		Last:   revgame.Address{X: 1, Y: 0}, // a white cell => Black to move
	}
	s := Snapshot{Board: board, Opponent: OpponentRandom, Human: revgame.Black}
	if !s.IsHumanTurn() {
		t.Fatalf("precondition: expected the human (Black) to be to move")
	}
	if board.HasValidMoves(revgame.White) {
		t.Fatal("precondition: White should have no move on this board")
	}

	beforeWhite := board.Score(revgame.White) // 2
	next, applied, err := s.ApplyHumanMove(revgame.Address{X: 0, Y: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Fatal("expected the Black move to be applied")
	}
	// The flanked white at (1,0) must have flipped to black.
	if !isBlack(next.Board, revgame.Address{X: 1, Y: 0}) {
		t.Error("flanked white at (1,0) should have flipped to black")
	}
	// No White disk was added; white count only went DOWN (one flipped).
	if afterWhite := next.Board.Score(revgame.White); afterWhite >= beforeWhite {
		t.Errorf("white score = %d, want < %d (white passed, one flipped)", afterWhite, beforeWhite)
	}
	// The opponent auto-passed, so it remains the human's turn.
	if !next.IsHumanTurn() {
		t.Error("it should remain the human's turn after the opponent auto-passed")
	}
	if next.IsOpponentTurn() {
		t.Error("IsOpponentTurn should be false after the opponent auto-passed")
	}
	// A follow-up PlayOpponent is a no-op (it is not the opponent's turn).
	afterPlay, moved, err := next.PlayOpponent(rand.New(rand.NewSource(7)))
	if err != nil {
		t.Fatalf("unexpected error from PlayOpponent: %v", err)
	}
	if moved {
		t.Error("PlayOpponent should be a no-op while it is the human's turn")
	}
	if afterPlay != next {
		t.Error("PlayOpponent changed the snapshot on a no-op")
	}
}

// --- Auto-pass in PlayOpponent: human passes, opponent keeps playing ----------

func TestPlayOpponent_AutoPassesHumanNoMove(t *testing.T) {
	// Human plays White; opponent plays Black. Two independent groups:
	//   row 0: whites at x=1..6, black at x=7, (0,0) empty
	//   row 2: whites at x=1..6, black at x=7, (0,2) empty
	// It is Black's turn. After Black plays one group's (0,y), that whole row
	// flips to black and White still has no move, so Black must play the second
	// group too. PlayOpponent must not stop on White's forced pass.
	board := revgame.Board{
		Blacks: revgame.Disks(0x80 | (0x80 << 16)),
		Whites: revgame.Disks(0x7E | (0x7E << 16)),
		Last:   revgame.Address{X: 1, Y: 0}, // a white cell => Black to move
	}
	s := Snapshot{Board: board, Opponent: OpponentRandom, Human: revgame.White}
	if !s.IsOpponentTurn() {
		t.Fatalf("precondition: expected the opponent (Black) to be to move")
	}
	if board.HasValidMoves(revgame.White) {
		t.Fatal("precondition: White should have no move on this board")
	}

	next, moved, err := s.PlayOpponent(rand.New(rand.NewSource(3)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !moved {
		t.Fatal("expected the opponent to move")
	}
	// BOTH Black moves ran (White auto-passed between them): every white flipped.
	if w := next.Board.Score(revgame.White); w != 0 {
		t.Errorf("white score = %d, want 0 — PlayOpponent must not stop on the human's forced pass", w)
	}
	if b := next.Board.Score(revgame.Black); b != 16 {
		t.Errorf("black score = %d, want 16 (2 originals + 12 flipped + 2 placed)", b)
	}
	// With no white disks left, neither side can move: the game is over.
	if !next.IsGameOver() {
		t.Error("expected the game to be over once every white was flipped")
	}
	if next.IsOpponentTurn() {
		t.Error("PlayOpponent must return with it no longer being the opponent's turn")
	}
}

// --- Full game: alternate ApplyHumanMove + PlayOpponent to completion ---------

func TestFullGamePlaysToCompletion(t *testing.T) {
	cases := []struct {
		name  string
		opp   Opponent
		human revgame.Disk
		seed  int64
	}{
		{"random-black", OpponentRandom, revgame.Black, 20240723},
		{"random-white", OpponentRandom, revgame.White, 11111},
		{"ai-black", OpponentAI, revgame.Black, 7},
		{"ai-white", OpponentAI, revgame.White, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGame(tc.opp, tc.human)
			rnd := rand.New(rand.NewSource(tc.seed))
			for i := 0; i < 200 && !s.IsGameOver(); i++ {
				if s.IsHumanTurn() {
					moves := s.Board.ValidMoves(s.HumanColor())
					if len(moves) == 0 {
						t.Fatalf("iteration %d: human has no move but IsHumanTurn is true", i)
					}
					next, applied, err := s.ApplyHumanMove(moves[0])
					if err != nil {
						t.Fatalf("iteration %d: ApplyHumanMove error: %v", i, err)
					}
					if !applied {
						t.Fatalf("iteration %d: legal move %v was not applied", i, moves[0])
					}
					s = next
				}
				next, _, err := s.PlayOpponent(rnd)
				if err != nil {
					t.Fatalf("iteration %d: PlayOpponent error: %v", i, err)
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
			if !strings.Contains(statusText(s), "Game over") {
				t.Errorf("completed status = %q, want it to say the game is over", statusText(s))
			}
		})
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

// --- DecodeSnapshot: error branches -------------------------------------------

func TestDecodeSnapshot_ErrorBranches(t *testing.T) {
	board := revgame.OthelloBoard.ToBase64()
	// Missing separator.
	if _, err := DecodeSnapshot("ab"); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("missing separator: err = %v, want ErrInvalidSnapshot", err)
	}
	// Prefix not exactly two chars.
	if _, err := DecodeSnapshot("a." + board); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("short prefix: err = %v, want ErrInvalidSnapshot", err)
	}
	// Unknown opponent (valid 2-char prefix, bad opponent).
	if _, err := DecodeSnapshot("zb." + board); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("unknown opponent: err = %v, want ErrInvalidSnapshot", err)
	}
	// Unknown human colour (valid opponent, bad colour char).
	if _, err := DecodeSnapshot("az." + board); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("unknown human colour: err = %v, want ErrInvalidSnapshot", err)
	}
	// Valid prefix, invalid board base64.
	if _, err := DecodeSnapshot("ab.###"); !errors.Is(err, ErrInvalidSnapshot) {
		t.Errorf("bad board base64: err = %v, want ErrInvalidSnapshot", err)
	}
}

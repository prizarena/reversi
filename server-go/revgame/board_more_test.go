package revgame

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// expectPanic runs fn and fails the test if it does not panic.
func expectPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected a panic, got none", name)
		}
	}()
	fn()
}

func TestOtherPlayer(t *testing.T) {
	if got := OtherPlayer(White); got != Black {
		t.Errorf("OtherPlayer(White) = %v, want Black", got)
	}
	if got := OtherPlayer(Black); got != White {
		t.Errorf("OtherPlayer(Black) = %v, want White", got)
	}
	expectPanic(t, "OtherPlayer(Empty)", func() { OtherPlayer(Empty) })
}

func TestBoard_IsEmpty(t *testing.T) {
	if !(Board{}).IsEmpty() {
		t.Error("empty board IsEmpty() = false, want true")
	}
	if OthelloBoard.IsEmpty() {
		t.Error("OthelloBoard IsEmpty() = true, want false")
	}
}

func TestBoard_DisksCount_Turns(t *testing.T) {
	if got := OthelloBoard.DisksCount(); got != 4 {
		t.Errorf("OthelloBoard.DisksCount() = %d, want 4", got)
	}
	if got := OthelloBoard.Turns(); got != 0 {
		t.Errorf("OthelloBoard.Turns() = %d, want 0", got)
	}
}

func TestNewBoardFromBase64_Error(t *testing.T) {
	if _, err := NewBoardFromBase64("!!! not base64 !!!"); err == nil {
		t.Error("expected an error decoding invalid base64")
	}
}

func TestBoard_flip_OverlapPanics(t *testing.T) {
	// The same cell owned by both colors is an impossible state -> flip panics.
	board := Board{Blacks: Disks(1 << 0), Whites: Disks(1 << 0)}
	expectPanic(t, "flip on overlapping board", func() { board.flip(Address{X: 0, Y: 0}) })
}

func TestBoard_flip_EmptyCellPanics(t *testing.T) {
	// Flipping a cell owned by neither color hits the default panic.
	expectPanic(t, "flip on empty cell", func() { OthelloBoard.flip(Address{X: 0, Y: 0}) })
}

func TestBoard_NextPlayer_Branches(t *testing.T) {
	// autopass board: White has no move, Black does (from revplay's auto-pass case).
	blackToMoveWhitePasses := Board{
		Blacks: Disks(0xFC) | Disks(0xFC<<16),
		Whites: Disks(1<<1) | Disks(1<<17),
	}
	// Colour-swapped: Black has no move, White does.
	whiteToMoveBlackPasses := Board{
		Whites: Disks(0xFC) | Disks(0xFC<<16),
		Blacks: Disks(1<<1) | Disks(1<<17),
	}

	cases := []struct {
		name  string
		board Board
		want  Disk
	}{
		{"white-last->black", OthelloBoard, Black}, // Last {4,4} is white; Black to move
		{"black-last->white", Board{Blacks: OthelloBoard.Blacks, Whites: OthelloBoard.Whites, Last: Address{X: 4, Y: 3}}, White},
		{"black-last->black (white passes)", withLast(blackToMoveWhitePasses, Address{X: 2, Y: 0}), Black},
		{"white-last->white (black passes)", withLast(whiteToMoveBlackPasses, Address{X: 2, Y: 0}), White},
		{"black-last->completed", Board{Blacks: Disks(1 << 0), Whites: 0, Last: Address{X: 0, Y: 0}}, Completed},
		{"white-last->completed", Board{Blacks: 0, Whites: Disks(1 << 0), Last: Address{X: 0, Y: 0}}, Completed},
		{"empty-last->black (parity 0)", Board{Blacks: OthelloBoard.Blacks, Whites: OthelloBoard.Whites, Last: Address{X: 0, Y: 0}}, Black},
		{"empty-last->white (parity 1)", Board{Blacks: OthelloBoard.Blacks | Disks(1<<(2*8+2)), Whites: OthelloBoard.Whites, Last: Address{X: 0, Y: 0}}, White},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.board.NextPlayer(); got != tc.want {
				t.Errorf("NextPlayer() = %v, want %v", got, tc.want)
			}
		})
	}
}

func withLast(b Board, last Address) Board {
	b.Last = last
	return b
}

func TestBoard_NextPlayer_EmptyLastTooManyTurnsPanics(t *testing.T) {
	// 14 disks (turns == 10) with Last pointing at an empty cell: the engine
	// cannot infer whose turn it is and panics.
	board := Board{Blacks: Disks(0x3FFF), Whites: 0, Last: Address{X: 7, Y: 7}}
	if board.DisksCount() != 14 {
		t.Fatalf("precondition: DisksCount = %d, want 14", board.DisksCount())
	}
	expectPanic(t, "NextPlayer with empty Last and too many turns", func() { board.NextPlayer() })
}

func TestBoard_DrawBoardAsText(t *testing.T) {
	s := OthelloBoard.DrawBoardAsText(".")
	if s == "" {
		t.Fatal("DrawBoardAsText returned empty string")
	}
	if !strings.Contains(s, "O") || !strings.Contains(s, "*") {
		t.Errorf("DrawBoardAsText output missing disk glyphs:\n%s", s)
	}
}

func TestBoard_DrawBoardAsEmoji_ShowsPossibleMoves(t *testing.T) {
	// DrawBoardAsEmoji passes a non-empty possibleMove glyph, so Rows consults
	// NextPlayer/getValidMoves and marks the legal cells.
	s := OthelloBoard.DrawBoardAsEmoji("", "", "")
	if s == "" {
		t.Fatal("DrawBoardAsEmoji returned empty string")
	}
	if !strings.Contains(s, "⚫️") || !strings.Contains(s, "⚪️") {
		t.Errorf("DrawBoardAsEmoji output missing disk glyphs:\n%s", s)
	}
	if !strings.Contains(s, ".") {
		t.Errorf("DrawBoardAsEmoji output missing possible-move markers:\n%s", s)
	}
}

func TestAddress_ToCellAddress_ToMove(t *testing.T) {
	a := Address{X: 3, Y: 3}
	if ca := a.ToCellAddress(); ca != "D4" {
		t.Errorf("ToCellAddress() = %q, want D4", ca)
	}
	if m := a.ToMove(); m != Move(27) {
		t.Errorf("ToMove() = %d, want 27", m)
	}
}

func TestAddress_IsOnBoard(t *testing.T) {
	if !(Address{X: 0, Y: 0}).IsOnBoard() {
		t.Error("(0,0).IsOnBoard() = false, want true")
	}
	if !(Address{X: 7, Y: 7}).IsOnBoard() {
		t.Error("(7,7).IsOnBoard() = false, want true")
	}
	if (Address{X: -1, Y: 0}).IsOnBoard() {
		t.Error("(-1,0).IsOnBoard() = true, want false")
	}
}

func TestBoard_IsCompleted(t *testing.T) {
	completed := Board{Blacks: Disks(1 << 0), Whites: 0, Last: Address{X: 0, Y: 0}}
	if !completed.IsCompleted() {
		t.Error("expected completed board to report IsCompleted() = true")
	}
	if OthelloBoard.IsCompleted() {
		t.Error("OthelloBoard.IsCompleted() = true, want false")
	}
}

func TestBoard_Scores_And_Score(t *testing.T) {
	black, white := OthelloBoard.Scores()
	if black != 2 || white != 2 {
		t.Errorf("Scores() = (%d,%d), want (2,2)", black, white)
	}
	if got := OthelloBoard.Score(Black); got != 2 {
		t.Errorf("Score(Black) = %d, want 2", got)
	}
	if got := OthelloBoard.Score(White); got != 2 {
		t.Errorf("Score(White) = %d, want 2", got)
	}
	expectPanic(t, "Score(Empty)", func() { OthelloBoard.Score(Empty) })
}

func TestBoard_MakeMove_OccupiedCell(t *testing.T) {
	// d4 (3,3) already holds a white disk: MakeMove must reject it as occupied
	// and leave the board unchanged.
	got, err := OthelloBoard.MakeMove(Black, Address{X: 3, Y: 3})
	if errors.Cause(err) != ErrAlreadyOccupied {
		t.Fatalf("MakeMove on occupied cell: err = %v, want cause %v", err, ErrAlreadyOccupied)
	}
	if got != OthelloBoard {
		t.Error("board changed after a rejected move on an occupied cell")
	}
}

func TestBoard_getDisksToFlip_OpponentRunToEdge(t *testing.T) {
	// A run of white disks from (6,0) to the edge (7,0) with no black terminator:
	// scanning that direction reaches the board edge and contributes no flips.
	board := Board{Whites: Disks(1<<(0*8+6)) | Disks(1<<(0*8+7))}
	flips, err := board.getDisksToFlip(Address{X: 5, Y: 0}, Black)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flips) != 0 {
		t.Errorf("expected no flips for an opponent run that reaches the edge, got %v", flips)
	}
}

func TestBoard_getDisksToFlip_Panics(t *testing.T) {
	expectPanic(t, "getDisksToFlip off-board", func() {
		_, _ = OthelloBoard.getDisksToFlip(Address{X: -1, Y: -1}, Black)
	})
	expectPanic(t, "getDisksToFlip unknown player", func() {
		_, _ = OthelloBoard.getDisksToFlip(Address{X: 2, Y: 2}, Empty)
	})
}

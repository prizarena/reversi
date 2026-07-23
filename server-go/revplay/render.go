package revplay

import (
	"fmt"
	"strings"

	"github.com/bots-go-framework/bots-go-core/botkb"
	"github.com/prizarena/reversi/server-go/revgame"
)

// Cell face glyphs used on the inline-keyboard buttons.
const (
	glyphBlack = "⚫" // the human's disks
	glyphWhite = "⚪" // the opponent's disks
	glyphMove  = "🟢" // a legal move for the human (Black), when it is their turn
	glyphEmpty = "🟩" // an empty board cell
)

// Rendered is a host-agnostic rendering of a snapshot: a status line and the
// 8x8 inline keyboard. The host wraps these into its messenger message
// (e.g. a botmsg.MessageFromBot with IsEdit set).
type Rendered struct {
	Text     string
	Keyboard botkb.Keyboard
}

// Render builds the status text and an 8x8 inline keyboard for the snapshot —
// one button per cell, its face a disk/empty/legal-move glyph.
//
// cellData supplies each button's callback_data for the cell it targets. The
// play layer never hardcodes the host's command code, so the host is expected
// to build the full callback_data itself (embedding s.Encode() and its own
// command prefix) inside cellData.
func Render(s Snapshot, cellData func(cell revgame.Address) string) Rendered {
	b := s.Board

	// possibleMove="" so Rows never consults NextPlayer/getValidMoves (which
	// panics for a Completed next-player on a board that still has empty cells);
	// legal-move hints for the human are overlaid below, only when it is Black's
	// turn — and ValidMoves(Black) is always safe to call.
	faces := b.Rows(glyphBlack, glyphWhite, "", glyphEmpty)
	if !b.IsCompleted() && b.NextPlayer() == revgame.Black {
		for _, m := range b.ValidMoves(revgame.Black) {
			faces[m.Y][m.X] = glyphMove
		}
	}

	rows := make([][]botkb.Button, revgame.BoardSize)
	for y := 0; y < revgame.BoardSize; y++ {
		row := make([]botkb.Button, revgame.BoardSize)
		for x := 0; x < revgame.BoardSize; x++ {
			addr := revgame.Address{X: int8(x), Y: int8(y)}
			row[x] = botkb.NewDataButton(faces[y][x], cellData(addr))
		}
		rows[y] = row
	}

	return Rendered{
		Text:     statusText(s),
		Keyboard: botkb.NewMessageKeyboard(botkb.KeyboardTypeInline, rows...),
	}
}

// IsGameOver reports whether the game has finished (neither side can move).
// Hosts use it to decide whether to offer a "start a new game" affordance.
func (s Snapshot) IsGameOver() bool {
	return s.Board.IsCompleted()
}

// statusText renders the score plus whose turn it is, or — when the game is
// over — the final score and the winner (or a draw).
func statusText(s Snapshot) string {
	b := s.Board
	black, white := b.Scores()

	var sb strings.Builder
	fmt.Fprintf(&sb, "Reversi  %s %d : %d %s", glyphBlack, black, white, glyphWhite)

	if b.IsCompleted() {
		sb.WriteString("\nGame over — ")
		switch {
		case black > white:
			fmt.Fprintf(&sb, "%s Black wins!", glyphBlack)
		case white > black:
			fmt.Fprintf(&sb, "%s White wins!", glyphWhite)
		default:
			sb.WriteString("it's a draw.")
		}
		return sb.String()
	}

	switch b.NextPlayer() {
	case revgame.Black:
		sb.WriteString("\nYour turn (Black).")
	case revgame.White:
		sb.WriteString("\nOpponent (White) to move.")
	}
	return sb.String()
}

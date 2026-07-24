package revplay

import (
	"fmt"
	"strings"

	"github.com/bots-go-framework/bots-go-core/botkb"
	"github.com/sneat-games/reversi/server-go/revgame"
)

// Cell face glyphs used on the inline-keyboard buttons. Black disks always show
// ⚫ and White disks always show ⚪, regardless of which colour the human plays.
const (
	glyphBlack = "⚫" // a Black disk
	glyphWhite = "⚪" // a White disk
	glyphMove  = "🟢" // a legal move for the human, shown only on the human's turn
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
// one button per cell, its face a disk/empty/legal-move glyph. Legal-move hints
// (🟢) are overlaid only when it is the human's turn, for the human's colour.
//
// cellData supplies each button's callback_data for the cell it targets. The
// play layer never hardcodes the host's command code, so the host is expected
// to build the full callback_data itself (embedding s.Encode() and its own
// command prefix) inside cellData.
func Render(s Snapshot, cellData func(cell revgame.Address) string) Rendered {
	b := s.Board

	// possibleMove="" so Rows never consults NextPlayer/getValidMoves (which
	// panics for a Completed next-player on a board that still has empty cells);
	// legal-move hints for the human are overlaid below, only when it is the
	// human's turn — and ValidMoves(human) is always safe to call.
	faces := b.Rows(glyphBlack, glyphWhite, "", glyphEmpty)
	if s.IsHumanTurn() {
		for _, m := range b.ValidMoves(s.Human) {
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

// diskName returns the human-readable colour name for a disk.
func diskName(d revgame.Disk) string {
	if d == revgame.White {
		return "White"
	}
	return "Black"
}

// diskGlyph returns the board glyph for a disk colour.
func diskGlyph(d revgame.Disk) string {
	if d == revgame.White {
		return glyphWhite
	}
	return glyphBlack
}

// statusText renders the score plus whose turn it is, or — when the game is
// over — the final score and the winner (or a draw). The "your turn" line uses
// the human's colour; the "opponent to move" line names the opponent's colour.
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

	if s.IsHumanTurn() {
		fmt.Fprintf(&sb, "\nYour turn (%s %s).", diskGlyph(s.Human), diskName(s.Human))
	} else {
		opp := s.OpponentColor()
		fmt.Fprintf(&sb, "\nOpponent (%s) to move.", diskName(opp))
	}
	return sb.String()
}

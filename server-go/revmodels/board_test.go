package revmodels

import (
	"testing"
	"github.com/prizarena/reversi/server-go/revgame"
)

func TestBoard_Kind(t *testing.T) {
	b := RevBoard{}
	if b.Kind() != BoardKind {
		t.Fatal("b.Kind() != BoardKind")
	}
}

func TestRevBoardEntity_GetBoard(t *testing.T) {
	entity := RevBoardEntity{}
	if board, err := entity.GetBoard(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if board != revgame.OthelloBoard {
		t.Errorf("Expected to get Othello board")
	}
}
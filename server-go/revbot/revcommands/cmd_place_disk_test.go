package revcommands

import (
	"testing"
	"github.com/prizarena/reversi/server-go/revgame"
)

func TestSwitch(t *testing.T) {
	switch "+1"[0] {
	case '+', '-':
		t.Log("OK")
	default:
		t.Fatal("Not OK!")
	}
}

func TestGetPlaceDiskSinglePlayerCallbackData(t *testing.T) {
	lastMoves := revgame.NewTranscript("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")
	s := getPlaceDiskSinglePlayerCallbackData(revgame.OthelloBoard, revgame.SinglePlayer, "F5", lastMoves, 2, "ru-RU", "TOUR123" )
	if len(s) > 64 {
		t.Errorf("too long, should be < = 64, got: %v", len(s))
	}
	t.Log(s)
}

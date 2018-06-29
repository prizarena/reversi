package revsecrets

import "testing"

func TestSecrets(t *testing.T) {
	if TelegramProdBot == "" {
		t.Error("TelegramProdBot is required")
	}
	if TelegramProdToken == "" {
		t.Error("TelegramProdToken is required")
	}
	if PrizarenaGameID == "" {
		t.Error("PrizarenaGameID is required")
	}
	if PrizarenaToken == "" {
		t.Error("PrizarenaToken is required")
	}
	if GaTrackingID == "" {
		t.Error("GaTrackingID is required")
	}
}

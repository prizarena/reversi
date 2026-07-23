package revgame

import (
	"math"
	"testing"
)

func TestEncodeDecodeIntBase64_RoundTrip(t *testing.T) {
	// Non-negative values round-trip exactly through the base64-int codec.
	cases := []int64{
		0,
		1,
		5,
		63, // last single-code value
		64, // first two-code value ("01")
		1000,
		1 << 20,
		math.MaxInt64,
	}
	for _, v := range cases {
		encoded := EncodeIntToBase64(v)
		if encoded == "" {
			t.Errorf("EncodeIntToBase64(%d) returned empty string", v)
		}
		decoded, err := DecodeIntFromBase64(encoded)
		if err != nil {
			t.Errorf("DecodeIntFromBase64(%q) error: %v", encoded, err)
			continue
		}
		if decoded != v {
			t.Errorf("round-trip mismatch for %d: encoded %q decoded %d", v, encoded, decoded)
		}
	}
}

func TestEncodeIntToBase64_Zero(t *testing.T) {
	if s := EncodeIntToBase64(0); s != "0" {
		t.Errorf("EncodeIntToBase64(0) = %q, want %q", s, "0")
	}
}

func TestEncodeIntToBase64_KnownValues(t *testing.T) {
	// codes = "0123456789abc...". v=64 => low code '0', high code '1' => "01".
	if s := EncodeIntToBase64(64); s != "01" {
		t.Errorf("EncodeIntToBase64(64) = %q, want %q", s, "01")
	}
	if s := EncodeIntToBase64(5); s != "5" {
		t.Errorf("EncodeIntToBase64(5) = %q, want %q", s, "5")
	}
}

// The codec only supports non-negative integers: a negative value encodes to
// the empty string and therefore does NOT round-trip (it decodes back to 0).
// This documents the real, current behavior rather than asserting a fictional
// round-trip. See the coverage report note.
func TestEncodeIntToBase64_NegativeIsNotSupported(t *testing.T) {
	for _, v := range []int64{-1, -64, math.MinInt64} {
		if s := EncodeIntToBase64(v); s != "" {
			t.Errorf("EncodeIntToBase64(%d) = %q, want empty string (negatives unsupported)", v, s)
		}
	}
	// Decoding an empty string yields 0.
	if got, err := DecodeIntFromBase64(""); err != nil || got != 0 {
		t.Errorf("DecodeIntFromBase64(\"\") = (%d, %v), want (0, nil)", got, err)
	}
}

func TestDecodeIntFromBase64_InvalidChar(t *testing.T) {
	got, err := DecodeIntFromBase64("!")
	if err == nil {
		t.Fatalf("expected error for invalid character, got nil (value %d)", got)
	}
	if got != -1 {
		t.Errorf("value on error = %d, want -1", got)
	}
}

func TestEncodeDecodeIntToString_AliasRoundTrip(t *testing.T) {
	// EncodeIntToString/DecodeStringToInt are thin aliases of the base64 codec.
	for _, v := range []int64{0, 42, 1 << 30} {
		s := EncodeIntToString(v)
		if s != EncodeIntToBase64(v) {
			t.Errorf("EncodeIntToString(%d) = %q, want %q", v, s, EncodeIntToBase64(v))
		}
		got, err := DecodeStringToInt(s)
		if err != nil {
			t.Fatalf("DecodeStringToInt(%q) error: %v", s, err)
		}
		if got != v {
			t.Errorf("DecodeStringToInt round-trip mismatch: got %d want %d", got, v)
		}
	}
}

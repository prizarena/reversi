package revgame

import "testing"

func TestTranscript_String(t *testing.T) {
	transcript := Transcript{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		60, 61, 62, 63,
	}
	s := transcript.String()
	t.Log(len(s), s)

}

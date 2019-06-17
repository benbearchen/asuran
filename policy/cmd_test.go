package policy

import "testing"

func TestQuote(t *testing.T) {
	str := `` + "`" + ``
	if str != "`" {
		t.Errorf("quote not equal")
	}
}

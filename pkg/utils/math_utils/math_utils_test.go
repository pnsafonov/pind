package math_utils

import "testing"

func TestRound2(t *testing.T) {
	if Round2(77.1504342243) != 77.15 {
		t.FailNow()
	}
	if Round2(77.155555) != 77.16 {
		t.FailNow()
	}
	if Round2(77.144444) != 77.14 {
		t.FailNow()
	}
}

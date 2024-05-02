package pkg

import (
	"strings"
	"testing"
)

func TestIndex0(t *testing.T) {
	version0 := "v1.0.1"
	in0 := strings.Index(version0, "v")
	if in0 < 0 {
		t.Fatalf("version0 does not contain v")
	}

	in1 := in0 + 1
	version1 := version0[in1:]
	if version1 != "1.0.1" {
		t.Fatalf("version1 should be 1.0.1, but got %s", version1)
	}
}

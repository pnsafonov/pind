package math_utils

import (
	"fmt"
	"reflect"
	"testing"
)

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

func TestIntDivideCeil(t *testing.T) {
	type testCaseIntDivideCeil1 struct {
		left   int
		right  int
		result int
	}
	cases := []testCaseIntDivideCeil1{
		{left: 3, right: 2, result: 2},
		{left: 5, right: 2, result: 3},
		{left: 1, right: 2, result: 1},
		{left: 0, right: 2, result: 0},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test_case_%d", i), func(t *testing.T) {
			result := IntDivideCeil(c.left, c.right)
			if reflect.DeepEqual(result, c.result) != true {
				t.Errorf("got %v, want %v", result, c.result)
			}
		})
	}
}

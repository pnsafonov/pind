package core_utils

import (
	"fmt"
	"testing"
)

func TestIsIntSliceEqual(t *testing.T) {
	type testCase struct {
		left  []int
		right []int
		equal bool
	}
	testCases := []testCase{
		{
			left:  []int{1, 2, 3},
			right: []int{1, 2, 3},
			equal: true,
		},
		{
			left:  []int{1, 2, 3},
			right: []int{1, 2, 3, 4},
			equal: false,
		},
		{
			left:  []int{1, 2, 3},
			right: nil,
			equal: false,
		},
	}
	for i, c := range testCases {
		t.Run(fmt.Sprintf("test_case_%d", i), func(t *testing.T) {
			if IsIntSliceEqual(c.left, c.right) != c.equal {
				t.FailNow()
			}
		})
	}
}

package pkg

import (
	"pind/pkg/numa"
	"testing"
)

type isMaskInSetTestCase0 struct {
	mask   []int
	set    []int
	result bool
}

func TestIsMaskInSet0(t *testing.T) {
	cases := []isMaskInSetTestCase0{
		{
			mask:   []int{1, 2, 3, 4, 5},
			set:    []int{1, 2, 3, 4, 5},
			result: true,
		},
		{
			mask:   []int{1, 2, 3, 4},
			set:    []int{1, 2, 3, 4, 5},
			result: true,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 6},
			set:    []int{1, 2, 3, 4, 5},
			result: false,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 68, 69, 597},
			set:    []int{1, 2, 3, 4, 5, 68, 69, 597},
			result: true,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 68, 69},
			set:    []int{1, 2, 3, 4, 5, 68, 69, 597},
			result: true,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 68, 69, 597},
			set:    []int{1, 2, 3, 4, 5, 68, 69},
			result: false,
		},
		{
			mask:   []int{2, 3, 4, 5, 68, 69, 597},
			set:    []int{1, 2, 3, 4, 5, 68, 69, 597},
			result: true,
		},
		{
			mask:   []int{1, 3, 4, 5, 68, 69, 597},
			set:    []int{1, 2, 3, 4, 5, 68, 69, 597},
			result: true,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 68, 69, 597},
			set:    []int{2, 3, 4, 5, 68, 69, 597},
			result: false,
		},
		{
			mask:   []int{1, 2, 3, 4, 5, 68, 69, 597},
			set:    []int{1, 2, 3, 4, 5, 68, 597},
			result: false,
		},
	}
	l0 := len(cases)
	for i := 0; i < l0; i++ {
		case0 := cases[i]
		mask := numa.CpusToMask(case0.mask)
		set := numa.CpusToMask(case0.set)
		result0 := isMaskInSet(mask, set)
		if result0 != case0.result {
			t.FailNow()
		}
	}

}

package config

import (
	"reflect"
	"testing"
)

type intervalTestCase0 struct {
	str      string
	expected Intervals
}

func TestInervals0(t *testing.T) {
	cases := []intervalTestCase0{
		{"1, 5-8, 19-20", Intervals{Values: []int{1, 5, 6, 7, 8, 19, 20}}},
		{" 1 , 5 - 8 , 19 - 20 ", Intervals{Values: []int{1, 5, 6, 7, 8, 19, 20}}},
		{"9,8, 5,  3,   1", Intervals{Values: []int{1, 3, 5, 8, 9}}},
		{"5-8,   1-2,     115-     119", Intervals{Values: []int{1, 2, 5, 6, 7, 8, 115, 116, 117, 118, 119}}},
	}

	l0 := len(cases)
	for i := 0; i < l0; i++ {
		case0 := cases[i]
		val0, err := ParseIntervals(case0.str)
		if err != nil {
			t.FailNow()
		}
		if !reflect.DeepEqual(case0.expected, val0) {
			t.FailNow()
		}
	}
}

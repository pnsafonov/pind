package numa

import (
	"reflect"
	"testing"
)

func TestGetCpuTopologyInfo0(t *testing.T) {
	tis, err := GetCpuTopologyInfo()
	if err != nil {
		t.Fatal(err)
	}
	_ = tis
}

func TestParseIntList(t *testing.T) {
	type testCaseParseIntList struct {
		in  string
		out []int
	}
	cases := []testCaseParseIntList{
		{in: "0,1,77", out: []int{0, 1, 77}},
		{in: " 0 , 1 ,	77	", out: []int{0, 1, 77}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			out, err := parseIntList(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(out, c.out) != true {
				t.Errorf("got %v, want %v", out, c.out)
			}
		})
	}
}

func TestParseIntList0(t *testing.T) {
	type testCaseParseIntList0 struct {
		in  string
		out []int
	}
	cases := []testCaseParseIntList0{
		{in: "3-6", out: []int{3, 4, 5, 6}},
		{in: " 3	-   6  ", out: []int{3, 4, 5, 6}},
		{in: "", out: nil},
		{in: "	 ", out: nil},
		{in: "	8-8 ", out: []int{8}},
		{in: "	9	- 9 ", out: []int{9}},
		{in: "	10	- 11 ", out: []int{10, 11}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			out, err := parseIntList0(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(out, c.out) != true {
				t.Errorf("got %v, want %v", out, c.out)
			}
		})
	}
}

package pkg

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetIdleCoresCountDefault0(t *testing.T) {
	type testGetIdleCoresCountDefault0 struct {
		numaCount  int
		coresCount int
		result     int
	}
	cases := []testGetIdleCoresCountDefault0{
		{numaCount: 1, coresCount: 4, result: 1},
		{numaCount: 1, coresCount: 6, result: 1},
		{numaCount: 1, coresCount: 8, result: 2},
		{numaCount: 1, coresCount: 16, result: 4},
		{numaCount: 1, coresCount: 24, result: 6},
		{numaCount: 1, coresCount: 32, result: 8},
		{numaCount: 2, coresCount: 4, result: 1},
		{numaCount: 2, coresCount: 6, result: 2},
		{numaCount: 2, coresCount: 8, result: 2},
		{numaCount: 2, coresCount: 16, result: 5},
		{numaCount: 2, coresCount: 24, result: 8},
		{numaCount: 2, coresCount: 32, result: 10},
		{numaCount: 3, coresCount: 4, result: 2},
		{numaCount: 3, coresCount: 6, result: 3},
		{numaCount: 3, coresCount: 8, result: 4},
		{numaCount: 3, coresCount: 16, result: 8},
		{numaCount: 3, coresCount: 24, result: 12},
		{numaCount: 3, coresCount: 32, result: 16},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test_case_%d", i), func(t *testing.T) {
			result := getIdleCoresCountDefault(c.numaCount, c.coresCount)
			if reflect.DeepEqual(result, c.result) != true {
				t.Errorf("got %v, want %v", result, c.result)
			}
		})
	}
}

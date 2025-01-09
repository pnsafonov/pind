package pkg

import (
	"testing"

	"github.com/pnsafonov/pind/pkg/numa"
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

// type testPool struct {
// 	Nodes     []*PoolNodeInfo
// 	NodeIndex int
// }

// func getAvailableCoresCountAlgo0(map0 map[int]*PoolCore) int {
// 	return 0
// }

// func (x *testPool) algo0(requiredCountPhys int, requiredCount int) (*PoolNodeInfo, bool) {
// 	l0 := len(x.Nodes)
// 	var freeNode *PoolNodeInfo
// 	counter := 0
// 	i := x.NodeIndex
// 	for {
// 		if i >= l0 {
// 			i = 0
// 		}
// 		if counter >= l0 {
// 			break
// 		}

// 		node := x.Nodes[i]
// 		freeCountPhys := len(node.LoadFree)
// 		freeCount := getAvailableCoresCountAlgo0(node.LoadFree)
// 		if freeCountPhys >= requiredCountPhys && freeCount >= requiredCount {
// 			freeNode = node
// 			i++
// 			break
// 		}

// 		i++
// 		counter++
// 	}
// 	x.NodeIndex = i
// 	return freeNode, freeNode != nil
// }

// func TestAlgo0(t *testing.T) {
// 	node0 := &PoolNodeInfo{}

// 	node1 := &PoolNodeInfo{}

// 	nodes0 := []*PoolNodeInfo{
// 		node0, node1,
// 	}

// 	pool0 := &testPool{
// 		Nodes:     nodes0,
// 		NodeIndex: 1,
// 	}

// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// 	pool0.algo0(2, 2)
// }

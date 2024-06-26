package core_utils

func CopyIntSlice(slice []int) []int {
	l0 := len(slice)
	sl0s := make([]int, l0)
	copy(sl0s, slice)
	return sl0s
}

func CopyStringSlice(slice []string) []string {
	l0 := len(slice)
	sl0s := make([]string, l0)
	copy(sl0s, slice)
	return sl0s
}

func IsIntSliceEqual(left []int, right []int) bool {
	l0 := len(left)
	l1 := len(right)
	if l0 != l1 {
		return false
	}
	for i := 0; i < l0; i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func IsSliceContains(sl []int, val int) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}
	return false
}

package core_utils

func CopyIntSlice(slice []int) []int {
	l0 := len(slice)
	sl0s := make([]int, l0)
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

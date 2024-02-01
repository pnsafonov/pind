package core_utils

func CopyIntSlice(slice []int) []int {
	l0 := len(slice)
	sl0s := make([]int, l0)
	copy(sl0s, slice)
	return sl0s
}

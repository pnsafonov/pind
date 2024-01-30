package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Intervals struct {
	Values []int
}

func (x *Intervals) UnmarshalText(text []byte) error {
	l, err := ParseIntervals(string(text))
	if err != nil {
		return err
	}

	*x = l

	return nil
}

// ParseIntervals - "1, 7-20, 3-4" -> [1, 7, 8...20]
func ParseIntervals(str0 string) (Intervals, error) {
	result := Intervals{}
	// 1, 7-20, 3-4
	splits0 := strings.Split(str0, ",")
	l0 := len(splits0)
	for i := 0; i < l0; i++ {
		// 7-20
		split0 := splits0[i]
		// 7 20
		splits1 := strings.Split(split0, "-")
		l1 := len(splits1)
		// 1,
		if l1 == 1 {
			str1 := splits1[0]
			str10 := strings.TrimSpace(str1)
			val0, err := strconv.ParseInt(str10, 10, 64)
			if err != nil {
				return Intervals{}, err
			}
			result.Values = append(result.Values, int(val0))
			continue
		}
		// 7-20,
		if l1 == 2 {
			str1 := splits1[0]
			str10 := strings.TrimSpace(str1)
			val0, err := strconv.ParseInt(str10, 10, 64)
			if err != nil {
				return Intervals{}, err
			}

			str2 := splits1[1]
			str20 := strings.TrimSpace(str2)
			val1, err := strconv.ParseInt(str20, 10, 64)
			if err != nil {
				return Intervals{}, err
			}

			if val0 == val1 {
				result.Values = append(result.Values, int(val0))
				continue
			}

			min0 := val0
			max0 := val1
			if val0 > val1 {
				return Intervals{}, fmt.Errorf("left must be less, but %d > %d", val0, val1)
			}

			for j := min0; j <= max0; j++ {
				result.Values = append(result.Values, int(j))
			}
		}
	}

	sort.Slice(result.Values, func(i, j int) bool {
		return result.Values[i] < result.Values[j]
	})

	return result, nil
}

func IsCpuInSlice(cpu int, cpus []int) bool {
	l0 := len(cpus)
	for i := 0; i < l0; i++ {
		if cpus[i] == cpu {
			return true
		}
	}
	return false
}

func IsCpuInInterval(cpu int, interval Intervals) bool {
	return IsCpuInSlice(cpu, interval.Values)
}

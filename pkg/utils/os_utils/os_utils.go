package os_utils

import "os"

// Exists - is file exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func SameFiles(path0 string, path1 string) (bool, error) {
	if path0 == path1 {
		return true, nil
	}
	fi0, err := os.Stat(path0)
	if err != nil {
		return false, err
	}
	fi1, err := os.Stat(path0)
	if err != nil {
		return false, err
	}
	result := os.SameFile(fi0, fi1)
	return result, nil
}

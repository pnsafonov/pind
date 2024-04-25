package os_utils

import (
	"os"
	"path/filepath"
	"strings"
)

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

func getPath() []string {
	return strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
}

func Which(program ...string) (string, bool) {
	for _, prog := range program {
		for _, p := range getPath() {
			candidate := filepath.Join(p, prog)
			if Exists(candidate) {
				return candidate, true
			}
		}
	}
	return "", false
}

func Which0(programs ...string) ([]string, bool) {
	result := make([]string, 0, len(programs))
	for _, prog := range programs {
		path0, ok := Which(prog)
		if ok {
			result = append(result, path0)
		}
	}
	return result, len(result) != 0
}

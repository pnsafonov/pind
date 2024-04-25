package os_utils

import "testing"

func TestWhich(t *testing.T) {
	path0, ok := Which("abcdesd1111")
	if ok || path0 != "" {
		t.FailNow()
	}
}

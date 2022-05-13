package test

import "testing"

func MustCleanup(t *testing.T, f func() error) {
	t.Cleanup(func() {
		if err := f(); err != nil {
			t.Log("failed to clean up after test:", err)
		}
	})
}

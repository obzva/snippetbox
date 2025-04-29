package assert

import "testing"

func Equal[U comparable](t *testing.T, got, want U) {
	t.Helper()

	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

package session

import "testing"

// Without the guard, pid 0 would signal this test's own process group and
// kill the test binary.
func TestKillProcessGroupIgnoresNonPositivePID(t *testing.T) {
	killProcessGroup(0)
	killProcessGroup(-1)
}

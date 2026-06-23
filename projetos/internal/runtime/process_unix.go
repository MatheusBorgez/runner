//go:build !windows

package runtime

import (
	"os"
	"syscall"
)

func isRunning(p *os.Process) bool {
	return p.Signal(syscall.Signal(0)) == nil
}

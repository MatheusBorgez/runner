//go:build windows

package runtime

import "os"

func isRunning(p *os.Process) bool {
	h, err := os.FindProcess(p.Pid)
	return err == nil && h != nil
}

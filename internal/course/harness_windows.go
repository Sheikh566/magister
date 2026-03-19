//go:build windows

package course

import (
	"os/exec"
)

func setupProcessGroup(cmd *exec.Cmd) {
	// No process group on Windows; child processes are terminated when parent exits.
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}

func killProcessGroupForce(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}

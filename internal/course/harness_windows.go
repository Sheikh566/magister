//go:build windows

package course

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

func setupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// setCmdLine bypasses Go's default argument escaping so that cmd.exe
// receives the user's command string verbatim. Go escapes arguments
// using C-runtime rules which do not match cmd.exe's parsing
// (see https://github.com/golang/go/issues/69939).
func setCmdLine(cmd *exec.Cmd, rawCommand string) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CmdLine = fmt.Sprintf(`/c %s`, rawCommand)
}

// killProcessGroup kills the entire process tree rooted at cmd using
// taskkill /F /T, which recursively terminates all child processes.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	kill := exec.Command("taskkill", "/F", "/T", "/PID", pid)
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = kill.Run()
}

func killProcessGroupForce(cmd *exec.Cmd) {
	killProcessGroup(cmd)
}

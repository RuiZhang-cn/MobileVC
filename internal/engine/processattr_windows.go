package engine

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"mobilevc/internal/logx"
)

const windowsCreateNewProcessGroup = 0x00000200

func hideCommandWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}

func isolateCommandProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= windowsCreateNewProcessGroup
}

func killCommandProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	if pid <= 0 {
		return
	}
	// taskkill /T kills the entire process tree including grandchildren
	// (bash.exe → winpty.exe → node.exe + any tool subprocesses).
	taskCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	taskCmd := exec.CommandContext(taskCtx, "taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	if taskCmd.SysProcAttr == nil {
		taskCmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	taskCmd.SysProcAttr.HideWindow = true
	output, err := taskCmd.CombinedOutput()
	if err == nil {
		return
	}
	logx.Warn("engine", "taskkill /T failed for pid=%d: %v (output=%q), falling back to Process.Kill", pid, err, strings.TrimSpace(string(output)))
	_ = cmd.Process.Kill()
}

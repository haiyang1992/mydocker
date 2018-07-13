package container

import (
	"os"
	"os/exec"
	"syscall"
)

/*
	NewParentProcess
	This is executed by the parent process
	1. /proc/self refers to the env of the current process (mydocker), exec just runs itself to initialize a child proc
	2. args is the parameters, with "init" being the first argument passed to the process
	3. the clone arguments forks a new process and uses namespace for isolation
	4. if user specifies "-ti", then I/O of the process is redirected to std I/O
*/
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

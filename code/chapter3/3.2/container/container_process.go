package container

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
	NewParentProcess
	This is executed by the parent process
	1. /proc/self refers to the env of the current process (mydocker), exec just runs itself to initialize a child proc
	2. args is the parameters, with "init" being the first argument passed to the process
	3. the clone arguments forks a new process and uses namespace for isolation
	4. if user specifies "-ti", then I/O of the process is redirected to std I/O
*/
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// pass handle for the read end of the pipe
	// child process will be created with the readPipe as the 4th file descriptor (after Stdin, Stdout, Stderr)
	cmd.ExtraFiles = []*os.File{readPipe}

	return cmd, writePipe
}

// NewPipe creates an anonymous pipe and returns two files: read and write
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}

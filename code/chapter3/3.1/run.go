package main

import (
	"os"
	"syscall"

	"./container"
	log "github.com/sirupsen/logrus"
)

// Run Actually runs the created command. Clones a process with namespace isolation, and runs /proc/self/exe in child process, sends parameters for init, and runs init to initialize the container's resources
func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	parent.Wait()

	// adding the following lines will solve a bug which causes terminal to not accept some commands (i.e. sudo) after exiting
	sysMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_RELATIME
	syscall.Mount("proc", "/proc", "proc", uintptr(sysMountFlags), "")
	os.Exit(-1)
}

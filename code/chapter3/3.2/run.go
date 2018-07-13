package main

import (
	"os"
	"strings"

	"github.com/haiyang1992/mydocker/code/chapter3/3.2/cgroups"
	"github.com/haiyang1992/mydocker/code/chapter3/3.2/cgroups/subsystems"
	"github.com/haiyang1992/mydocker/code/chapter3/3.2/container"
	log "github.com/sirupsen/logrus"
)

// Run Actually runs the created command. Clones a process with namespace isolation, and runs /proc/self/exe in child process, sends parameters for init, and runs init to initialize the container's resources
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// use mydocker-cgroup as cgroup name
	// create cgroup manager, use set() and apply() to set resources of the container
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	// set resource restrictions
	cgroupManager.Set(res)
	// add container process into cgroups mounted by each subsystem
	cgroupManager.Apply(parent.Process.Pid)
	// initialize the container, send the user commands to child
	sendInitCommand(comArray, writePipe)
	parent.Wait()

	// adding the following lines will solve a bug which causes terminal to not accept some commands (i.e. sudo) after exiting
	//sysMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_RELATIME
	//syscall.Mount("proc", "/proc", "proc", uintptr(sysMountFlags), "")
	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("complete command is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

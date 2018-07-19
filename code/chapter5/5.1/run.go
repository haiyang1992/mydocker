package main

import (
	"os"
	"strings"

	"github.com/haiyang1992/mydocker/code/chapter5/5.1/cgroups"
	"github.com/haiyang1992/mydocker/code/chapter5/5.1/cgroups/subsystems"
	"github.com/haiyang1992/mydocker/code/chapter5/5.1/container"
	log "github.com/sirupsen/logrus"
)

// Run Actually runs the created command. Clones a process with namespace isolation, and runs /proc/self/exe in child process, sends parameters for init, and runs init to initialize the container's resources
func Run(tty bool, volume string, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("new parent process error")
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
	if tty {
		parent.Wait()
	}
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	if _, err := os.Stat(mntURL); !os.IsNotExist(err) {
		container.DeleteWorkSpace(rootURL, mntURL, volume)
	}

	// this issue is solved in pivotRoot() in init.go, so the method below is no longer needed
	// adding the following lines will solve a bug which causes terminal to not accept some commands (i.e. sudo) after exiting
	/*sysMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_RELATIME
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(sysMountFlags), ""); err != nil {
		log.Errorf("mount /proc error %v", err)
	}*/
	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("complete command is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

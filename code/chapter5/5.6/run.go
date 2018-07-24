package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/haiyang1992/mydocker/code/chapter5/5.6/cgroups"
	"github.com/haiyang1992/mydocker/code/chapter5/5.6/cgroups/subsystems"
	"github.com/haiyang1992/mydocker/code/chapter5/5.6/container"
	log "github.com/sirupsen/logrus"
)

// Run Actually runs the created command. Clones a process with namespace isolation, and runs /proc/self/exe in child process, sends parameters for init, and runs init to initialize the container's resources
func Run(tty bool, volume string, comArray []string, res *subsystems.ResourceConfig, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, containerName, volume)
	if parent == nil {
		log.Errorf("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// record info about the container
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName)
	if err != nil {
		log.Errorf("record container info error %v", err)
		return
	}

	// use mydocker-cgroup as cgroup name
	// create cgroup manager, use set() and apply() to set resources of the container
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	// set resource restrictions
	cgroupManager.Set(res)
	// add container process into cgroups mounted by each subsystem
	cgroupManager.Apply(parent.Process.Pid)
	log.Infof("finished setting up cgroup")
	// initialize the container, send the user commands to child
	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		mntURL := "/root/mnt/"
		rootURL := "/root/"
		if _, err := os.Stat(mntURL); !os.IsNotExist(err) {
			container.DeleteWorkSpace(rootURL, mntURL, volume)
		}
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

// recordContainerInfo writes metadata of the container to the file system
func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	// first we get a 10-digit number as container ID
	id := randStringBytes(10)
	// use current time as container creation time
	creationTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	// if user did not specify a container name, use id instead
	if containerName == "" {
		containerName = id
	}
	log.Infof("using %s as container name", containerName)
	containerInfo := &container.Info{
		Id:           id,
		Pid:          strconv.Itoa(containerPID),
		Command:      command,
		CreationTime: creationTime,
		Status:       container.RUNNING,
		Name:         containerName,
	}

	// convert the containerInfor object into its json encoding
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	// piece together the path of the file to write to
	saveDirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// if the directory does not exist, we need to recursively mkdir all of them
	if err := os.MkdirAll(saveDirURL, 0622); err != nil {
		log.Errorf("mkdir %s error %v", saveDirURL, err)
		return "", err
	}
	saveFileName := path.Join(saveDirURL, container.ConfigName)
	// create the config.json config file
	file, err := os.Create(saveFileName)
	defer file.Close()
	if err != nil {
		log.Errorf("create file %s error %v", saveFileName, err)
		return "", err
	}
	// write the json-ized data into the file
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("file write to %s error %v", file, err)
		return "", err
	}
	log.Infof("written config file for container[Name: %s, ID: %s] to %s", containerName, id, saveFileName)

	return containerName, nil
}

func deleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("remove dir %s error %v", dirURL, err)
	} else {
		log.Infof("deleted container info at %s", path.Join(dirURL, "config.json"))
	}
}

// randStringBytes returns a string representing a random number with 10 digits
func randStringBytes(n int) string {
	letterBytes := "0123456789ABCDEF"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

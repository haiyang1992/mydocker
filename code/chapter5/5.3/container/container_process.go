package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Info stores data about the container
type Info struct {
	Id           string `json:"pid`          // the PID of the init process of the container on the host
	Pid          string `json:"id`           // container id
	Name         string `json:"name`         //container name
	Command      string `json:"command`      //the command of the init process runs inside the container
	CreationTime string `json:"creationTime` //the creation time of the container
	Status       string `json:"status`       // the status of the container
}

// some constants
var (
	RUNNING             = "Running"
	STOP                = "Stopped"
	EXIT                = "Exited"
	DefaultInfoLocation = "/var/run/mydocker/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)

/*
	NewParentProcess
	This is executed by the parent process
	1. /proc/self refers to the env of the current process (mydocker), exec just runs itself to initialize a child proc
	2. args is the parameters, with "init" being the first argument passed to the process
	3. the clone arguments forks a new process and uses namespace for isolation
	4. if user specifies "-ti", then I/O of the process is redirected to std I/O
*/
func NewParentProcess(tty bool, containerName string, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
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
	} else {
		// generate the container.log file under the corresponding container directory
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			log.Errorf("NewParentProcess() mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := path.Join(dirURL, ContainerLogFile)
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess() create log file %s error %v", stdLogFilePath, err)
		}
		cmd.Stdout = stdLogFile
	}

	// pass handle for the read end of the pipe
	// child process will be created with the readPipe as the 4th file descriptor (after Stdin, Stdout, Stderr)
	cmd.ExtraFiles = []*os.File{readPipe}
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL

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

// NewWorkSpace create an AUFS filesystem as the container root workspace
func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
	// determines if we will mount the data volume depending on "volume"
	if volume != "" {
		volumeURLs := volumeURLExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(mntURL, volumeURLs)
			log.Infof("mounted volumes: %q", volumeURLs)
		} else {
			log.Infof("volume parameter not correctly set!")
		}
	}
}

// CreateReadOnlyLayer untars busybox.tar to busybox to use as the container's read-only layer
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := path.Join(rootURL, "busybox")
	busyboxTarURL := path.Join(rootURL, "busybox.tar")
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("failed to tell whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("untar file %s error. %v", busyboxTarURL, err)
		}
		log.Infof("untared busybox.tar to %s", busyboxURL)
	} else {
		log.Infof("%s already exists, skipping untar process", busyboxURL)
	}
}

// CreateWriteLayer creates a writeLayer folder as the container's only write layer
func CreateWriteLayer(rootURL string) {
	writeURL := path.Join(rootURL, "writeLayer")
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", writeURL, err)
	} else {
		log.Infof("created directory %s", writeURL)
	}

}

// CreateMountPoint mounts writeLayer and busybox under mnt
func CreateMountPoint(rootURL string, mntURL string) {
	// create mnt folder as the mount point
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error %v", mntURL, err)
	} else {
		log.Infof("created directory %s", mntURL)
	}

	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount %v", err)
	} else {
		log.Infof("\"mount -t aufs -o %s none %s\" successful", dirs, mntURL)
	}
}

// volumeUrlExtract analyzes the volume string
func volumeURLExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

// MountVolume mounts volumeURLS[0], which is a directory on the host onto volumeURL[1] inside the container
func MountVolume(mntURL string, volumeURLs []string) {
	// creates parentURL on the host
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		log.Infof("mkdir parent dir %s error. %v", parentURL, err)
	} else {
		log.Infof("created directory %s as volume dir on host", parentURL)
	}
	// create a mount point inside the container fs
	containerVolumeURL := path.Join(mntURL, volumeURLs[1])
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("mkdir container dir %s error. %v", containerVolumeURL, err)
	} else {
		log.Infof("created directory %s as volume dir on container", containerVolumeURL)
	}
	// mount parentURL to the container mount point containerVolumeURL
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume failed %v", err)
	} else {
		log.Infof("\"mount -t aufs -o %s none %s\" successful", dirs, containerVolumeURL)
	}
}

// DeleteWorkSpace deletes the AUFS filesystem at container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := volumeURLExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(mntURL)
		}
	} else {
		DeleteMountPoint(mntURL)
	}
	DeleteWriteLayer(rootURL)
}

// DeleteMountPointWithVolume deletes mount points along with volumes
func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeURLs []string) {
	// unmount the fs on the volume mount point
	containerVolumeURL := path.Join(mntURL, volumeURLs[1])
	cmd := exec.Command("umount", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("unmount volume failed %v", err)
	} else {
		log.Infof("\"umount %s\" successful", containerVolumeURL)
	}
	// umount the mount point of the container and delete it
	DeleteMountPoint(mntURL)
}

// DeleteMountPoint unmounts and removes mnt
func DeleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("unmount error %v", err)
	} else {
		log.Infof("\"umount %s\" successful", mntURL)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove dir %s error %v", mntURL, err)
	} else {
		log.Infof("deleted directory %s", mntURL)
	}
}

// DeleteWriteLayer deletes writeLayer
func DeleteWriteLayer(rootURL string) {
	writeURL := path.Join(rootURL, "writeLayer")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove dir %s error %v", writeURL, err)
	} else {
		log.Infof("deleted directory %s", writeURL)
	}
}

// PathExists returns if the given path exists in the system
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

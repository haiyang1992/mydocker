package container

import (
	"os"
	"os/exec"
	"path"
	"strings"
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
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
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
	}
}

// CreateWriteLayer creates a writeLayer folder as the container's only write layer
func CreateWriteLayer(rootURL string) {
	writeURL := path.Join(rootURL, "writeLayer")
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

// CreateMountPoint mounts writeLayer and busybox under mnt
func CreateMountPoint(rootURL string, mntURL string) {
	// create mnt folder as the mount point
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error %v", mntURL, err)
	}

	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount %v", err)
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
	}
	// create a mount point inside the container fs
	containerVolumeURL := path.Join(mntURL, volumeURLs[1])
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// mount parentURL to the container mount point containerVolumeURL
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume failed %v", err)
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
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove dir %s error %v", mntURL, err)
	}
}

// DeleteWriteLayer deletes writeLayer
func DeleteWriteLayer(rootURL string) {
	writeURL := path.Join(rootURL, "writeLayer")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove dir %s error %v", writeURL, err)
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

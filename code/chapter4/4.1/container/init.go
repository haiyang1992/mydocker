package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
	RunContainerInitProcess
	The init function runs inside a container. Now the process which holds the container
	has been created.
	Use mount to mount proc fs, so that we can use ps, etc. to check process resources
*/
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}

	setupMount()

	// use exec.LookPath to get abs path for commands
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	// uintptr(3) is a file descriptor with index=3, which is the one end of the pipe passed in
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setupMount() {
	// get cwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		log.Errorf("%v", err)
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {
	// this is necessary for pivot_root to work
	// gets rid of a bug which causes terminal to not accept some commands (i.e. sudo) after exiting
	// and the system not displaying correctly after exiting
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount --make-rprivate / %v", err)
	}
	/*
		We need to remount root s.t. the old root and new root will be on different fs
		bind mount is used to replicate an already mounted dir tree
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	log.Infof("bind mounted %s", root)

	// create rootfs/.pivot_root to store old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if _, err := os.Stat(pivotDir); os.IsNotExist(err) {
		if err := os.Mkdir(pivotDir, 0777); err != nil {
			return fmt.Errorf("mkdir ./pivot_root %v", err)
		}
	}
	// pivot_root to the new rootfs, old_root is mounted on rootfs/.pivot_root
	// we can still see the mount point with mount
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	log.Infof("pivot_root to %s, old root at %s", root, pivotDir)
	// change cwd to "/"
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	//umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}
	log.Infof("unmounted %s", pivotDir)

	// delete temporary dir
	if err := os.Remove(pivotDir); err != nil {
		return fmt.Errorf("remove pivot_root dir %v", err)
	}
	log.Infof("removed %s", pivotDir)
	return nil
}

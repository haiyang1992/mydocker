package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CPUsetSubSystem struct
type CPUsetSubSystem struct {
}

// Set will configure the CPU limit of the cgroup designated by cgroupPath
func (c *CPUsetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// GetCgroupPath gets the path of the current subsystem in the virtual fs
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err == nil {
		if res.CPUShare != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CPUSet), 0644); err != nil {
				return fmt.Errorf("set cgroup CPUset fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove removes the cgroup specified by cgroupPath
func (c *CPUsetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err == nil {
		// deleting the correspoinding cgroupPath will delete the cgroup
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply adds a process to the cgroup specified by cgroupPath
func (c *CPUsetSubSystem) Apply(cgroupPath string, pid int) error {
	// GetCgroupPath gets the path of the current subsystem in the virtual fs
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error %v", cgroupPath, err)
	}
}

// Name returns cgroup's name
func (c *CPUsetSubSystem) Name() string {
	return "cpuset"
}

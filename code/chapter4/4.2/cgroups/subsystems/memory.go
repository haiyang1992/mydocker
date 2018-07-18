package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// MemorySubSystem struct
type MemorySubSystem struct {
}

// Set will configure the memory limit of the cgroup designated by cgroupPath
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// GetCgroupPath gets the path of the current subsystem in the virtual fs
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		subsysName, _ := path.Split(subsysCgroupPath)
		log.Infof("Found subsystem's cgroupPath at %s", subsysName)
		if res.MemoryLimit != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove removes the cgroup specified by cgroupPath
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// deleting the correspoinding cgroupPath will delete the cgroup
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply adds a process to the cgroup specified by cgroupPath
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	// GetCgroupPath gets the path of the current subsystem in the virtual fs
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}

// Name returns cgroup's name
func (s *MemorySubSystem) Name() string {
	return "memory"
}

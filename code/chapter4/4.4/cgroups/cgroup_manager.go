package cgroups

import (
	"github.com/haiyang1992/mydocker/code/chapter4/4.4/cgroups/subsystems"
	log "github.com/sirupsen/logrus"
)

// CgroupManager struct
type CgroupManager struct {
	// path of cgroup in hierarchy
	Path string
	// resource allocation
	Resource *subsystems.ResourceConfig
}

// NewCgroupManager constructor
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply adds pid to every cgroup
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set cgroup resource limits mounted on each subsystem
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

// Destroy releases cgroups mounted on each subsystem
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}

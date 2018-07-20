package subsystems

// ResourceConfig : struct for passing resouce limit config
type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

// Subsystem interfaces
// cgroup is represented by path, becausethe path of cgroup under
// hierarchy is the virtual path in the virtual fs
type Subsystem interface {
	// returns name of subsystem
	Name() string

	// sets the resouce limits of a cgroup in this subsystem
	Set(path string, res *ResourceConfig) error

	// add a process to a cgroup
	Apply(path string, pid int) error

	// remove a cgroup
	Remove(path string) error
}

// use different subsystems to initialize an array of resource limit instances
var (
	SubsystemsIns = []Subsystem{
		&CPUsetSubSystem{},
		&MemorySubSystem{},
		&CPUSubSystem{},
	}
)

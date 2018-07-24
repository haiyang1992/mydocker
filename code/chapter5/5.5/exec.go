package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/haiyang1992/mydocker/code/chapter5/5.5/container"
	_ "github.com/haiyang1992/mydocker/code/chapter5/5.5/nsenter"
	log "github.com/sirupsen/logrus"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func execContainer(containerName string, cmdArray []string) {
	// get PID of the corresponding container with containerName
	pid, err := getContainerPIDByName(containerName)
	if err != nil {
		log.Errorf("exec container: getContainerPIDbyName(%s) error %v", containerName, err)
		return
	}
	// format the command to be space separated
	cmdString := strings.Join(cmdArray, " ")
	log.Infof("exec container: container PID is %s", pid)
	log.Infof("exec container: command is %s", cmdString)

	// using "exec" here invokes the C code in nsenter.go
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdString)

	if err := cmd.Run(); err != nil {
		log.Errorf("exec container: exec %s command error %v", containerName, err)
	}
}

func getContainerPIDByName(containerName string) (string, error) {
	// piece togeghet the container's location
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := path.Join(configFileDir, container.ConfigName)
	// read file contents
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("read file %s error %v", configFilePath, err)
		return "", err
	}
	var containerInfo container.Info
	// unmarshall the json metadata into an object of the container.Info struct
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("json unmarshall error %v", err)
		return "", err
	}
	return containerInfo.Pid, nil
}

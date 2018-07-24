package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/haiyang1992/mydocker/code/chapter5/5.6/container"
	log "github.com/sirupsen/logrus"
)

func stopContainer(containerName string) {
	// get the containrt's PID
	pid, err := getContainerPIDByName(containerName)
	if err != nil {
		log.Errorf("get container pid by name %s error %v", containerName, err)
		return
	}
	// convert PID from string to int
	intPid, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("error converting pid %s from string to int %v", pid, err)
		return
	}
	// use the kill syscall to send SIGTERM to the process
	if err := syscall.Kill(intPid, syscall.SIGTERM); err != nil {
		log.Errorf("stop container %s error %v", containerName, err)
		return
	}
	log.Infof("sent kill command to container %s with pid %s", containerName, pid)
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("get container %s's info error %v", containerName, err)
		return
	}
	// now we need to modify the container's status and set its PID to empty
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("json marchall container %s'info error %v", containerName, err)
		return
	}
	saveDirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	saveFileName := path.Join(saveDirURL, container.ConfigName)
	if err := ioutil.WriteFile(saveFileName, jsonBytes, 0622); err != nil {
		log.Errorf("write to file %s error %v", saveFileName, err)
	}
	log.Infof("overwritten config file %s", saveFileName)
}

func getContainerInfoByName(containerName string) (*container.Info, error) {
	// piece togeghet the container's location
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := path.Join(configFileDir, container.ConfigName)
	// read file contents
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.Info
	// unmarshall the json metadata into an object of the container.Info struct
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("json unmarshall error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

func removeContainer(containerName string, volume string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("get container %s's info error %v", containerName, err)
		return
	}
	// only remove stopped containers
	if containerInfo.Status != container.STOP {
		log.Errorf("can't remove a running container!")
		return
	}
	deleteContainerInfo(containerName)

	mntURL := "/root/mnt/"
	rootURL := "/root/"
	if _, err := os.Stat(mntURL); !os.IsNotExist(err) {
		container.DeleteWorkSpace(rootURL, mntURL, volume)
	}
}

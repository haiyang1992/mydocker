package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/haiyang1992/mydocker/code/chapter5/5.3/container"
	log "github.com/sirupsen/logrus"
)

// ListContainers finds all running containers and prints metadata
func ListContainers() {
	// find path for /var/run/mydocker/""/ -> /var/run/mydocker
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	// read all files under this directory
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		log.Errorf("read dir %s error %v", dirURL, err)
		return
	}

	var containerInfos []*container.Info
	// loop through all files
	for _, file := range files {
		curContainerInfo, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("get container info error %v", err)
			continue
		}
		containerInfos = append(containerInfos, curContainerInfo)
	}

	// use tabwrite.NewWriter to print container infos in the console
	// tabwriter calls the text/tabwriter library to print space aligned tables
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	// tab columns in the console
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containerInfos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreationTime)
	}
	// flush the stdout buffer zone and print the container list
	if err := w.Flush(); err != nil {
		log.Errorf("flush error %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.Info, error) {
	// get file name
	containerName := file.Name()
	// generate the absolute path from the file name
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := configFileDir + container.ConfigName
	// read metadate from config.json
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

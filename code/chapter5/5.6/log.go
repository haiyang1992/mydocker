package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/haiyang1992/mydocker/code/chapter5/5.6/container"
	log "github.com/sirupsen/logrus"
)

func logContainer(containerName string) {
	// find log file location
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFilePath := path.Join(dirURL, container.ContainerLogFile)
	// open the log file
	file, err := os.Open(logFilePath)
	defer file.Close()
	if err != nil {
		log.Errorf("open container log file %s error %v", logFilePath, err)
		return
	}
	// read file contents
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read container log file %s error %v", logFilePath, err)
	}
	// usr fmt.Fprint() to output the contents to stdout, the console
	fmt.Fprint(os.Stdout, string(content))
}

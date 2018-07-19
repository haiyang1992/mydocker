package main

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// commitContainer packages the container fs into ${imageName}.tar
func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("packaging destination %s\n", imageTar)
	if _, err := exec.Command("tar", "czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntURL, err)
	}
}

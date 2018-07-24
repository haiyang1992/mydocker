package main

import (
	"os"

	"github.com/haiyang1992/mydocker/code/chapter5/5.5/container"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `mydocker is a simple container runtime implementation.
	  		   The purpose of this project is to learn how docker works and how to write a docker by ourselves
	           Enjoy it, just for fun.`

func main() {
	log.AddHook(filename.NewHook())

	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
	}

	app.Before = func(context *cli.Context) error {
		//log.SetFormatter(&log.TextFormatter{ForceColors: true})

		log.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		// do cleanup nonetheless
		log.Infof("still try to do cleanup..")
		mntURL := "/root/mnt/"
		rootURL := "/root/"
		container.DeleteWorkSpace(rootURL, mntURL, "")
		log.Fatal(err)
	}

}

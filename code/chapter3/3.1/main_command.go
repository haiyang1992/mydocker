package main

import (
	"fmt"

	"github.com/haiyang1992/mydocker/code/chapter3/3.1/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// defines Flags of runCommand
var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},

	/*
		main func of runCommand
		1. determines if args include command
		2. get user-defined command
		3. invokes Run function to start the container
	*/
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("ti")
		Run(tty, cmd)
		return nil
	},
}

// defines operations for initCommand
var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",

	/*
		1. get command parameter
		2. execute container initialization
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command %s", cmd)
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}

package main

import (
	"fmt"

	"github.com/haiyang1992/mydocker/code/chapter4/4.3/cgroups/subsystems"
	"github.com/haiyang1992/mydocker/code/chapter4/4.3/container"

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
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
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
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("ti")
		volume := context.String("v")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUShare:    context.String("cpushare"),
			CPUSet:      context.String("cpuset"),
		}
		Run(tty, volume, cmdArray, resConf)
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
		err := container.RunContainerInitProcess()
		return err
	},
}

package main

import (
	"fmt"

	"github.com/haiyang1992/mydocker/code/chapter5/5.2/cgroups/subsystems"
	"github.com/haiyang1992/mydocker/code/chapter5/5.2/container"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// defines Flags of runCommanRund
var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach",
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
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
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
		detach := context.Bool("d")
		volume := context.String("v")
		if tty && detach {
			return fmt.Errorf("ti and d parameters cannot be provided at the same time")
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUShare:    context.String("cpushare"),
			CPUSet:      context.String("cpuset"),
		}
		log.Infof("tty enabled: %v", tty)
		// pass container name, null if not specified
		containerName := context.String("name")
		Run(tty, volume, cmdArray, resConf, containerName)
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

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "Commit a container into an image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "List all containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

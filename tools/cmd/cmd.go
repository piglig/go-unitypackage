package cmd

import (
	"flag"
	"fmt"
	"github.com/piglig/go-unitypackage"
)

const (
	compressAct   = "compress"
	decompressAct = "decompress"
)

type Command struct {
	action string
	src    string
	dst    string
}

func newCommand() *Command {
	cmd := &Command{}
	flag.StringVar(&cmd.action, "act", "decompress", "compress or decompress unitypackage")
	flag.StringVar(&cmd.src, "src", "", "compress or decompress the source directory")
	flag.StringVar(&cmd.dst, "dest", "", "compress or decompress the destination directory")

	flag.Parse()
	return cmd
}

func Execute() {
	cmd := newCommand()

	if cmd.action != compressAct && cmd.action != decompressAct {
		fmt.Errorf("action must be compress or decompress")
		flag.Usage()
		return
	}

	if cmd.src == "" || cmd.dst == "" {
		fmt.Println("Missing required flags")
		flag.Usage()
		return
	}

	if cmd.action == compressAct {
		if err := unitypackage.GeneratePackage(cmd.src, cmd.dst); err != nil {
			fmt.Println(err)
			return
		}
	} else if cmd.action == decompressAct {
		if err := unitypackage.UnPackage(cmd.src, cmd.dst); err != nil {
			fmt.Println(err)
			return
		}
	}
}

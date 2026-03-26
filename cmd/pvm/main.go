package main

import (
	"os"

	"github.com/ankele/pvm/internal/cli"
)

func main() {
	root := cli.NewRootCommand(cli.Dependencies{})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

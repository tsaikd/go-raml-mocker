package main

import (
	"os"

	"github.com/tsaikd/go-raml-mocker/cmd"
)

func main() {
	rootCommand := cmd.Module.MustNewRootCommand(nil)
	rootCommand.SilenceUsage = true
	if err := rootCommand.Execute(); err != nil {
		os.Exit(-1)
	}
}

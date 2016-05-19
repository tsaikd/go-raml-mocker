package cmd

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/version"
)

// Action cmd main action
var Action = func(c *cli.Context) (err error) {
	return
}

// Flags list all global flags for application
var Flags = []cli.Flag{}

// Commands list all commands for application
var Commands = []cli.Command{}

// Main entry point
func Main() {
	app := cli.NewApp()
	app.Name = "ramlMocker"
	app.Usage = "Go RAML mock web server"
	app.Version = version.String()
	app.Action = Action
	app.Flags = Flags
	app.Commands = Commands

	errutil.Trace(app.Run(os.Args))
}

package mocker

import (
	"fmt"
	"path"

	"github.com/codegangsta/cli"
	"github.com/tsaikd/go-raml-mocker/cmd"
	"github.com/tsaikd/go-raml-parser/parser"
	"github.com/tsaikd/go-raml-parser/parser/parserConfig"
)

func init() {
	cmd.Flags = append(cmd.Flags, cli.StringFlag{
		Name:        "f, ramlfile",
		Value:       "api.raml",
		Usage:       "Source RAML file",
		Destination: &ramlFile,
	})
	cmd.Flags = append(cmd.Flags, cli.BoolFlag{
		Name:        "checkRAMLVersion",
		Usage:       "Check RAML Version",
		Destination: &checkRAMLVersion,
	})
	cmd.Flags = append(cmd.Flags, cli.IntFlag{
		Name:        "port",
		Value:       4000,
		Usage:       "Mock server listen port",
		Destination: &port,
	})
	cmd.Action = action
}

var ramlFile string
var checkRAMLVersion bool
var port int

func action(c *cli.Context) (err error) {
	dir, _ := path.Split(ramlFile)
	watch(dir)
	return reload()
}

func reload() (err error) {
	parser := parser.NewParser()

	if err = parser.Config(parserConfig.CheckRAMLVersion, checkRAMLVersion); err != nil {
		return
	}

	rootdoc, err := parser.ParseFile(ramlFile)
	if err != nil {
		return
	}

	addr := fmt.Sprintf(":%d", port)
	if err = start(rootdoc, addr); err != nil {
		return
	}

	return
}

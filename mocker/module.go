package mocker

import (
	"fmt"
	"path"

	"github.com/tsaikd/KDGoLib/cliutil/cmder"
	"github.com/tsaikd/KDGoLib/futil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
	"github.com/tsaikd/go-raml-parser/parser/parserConfig"
	"gopkg.in/urfave/cli.v2"
)

// Module info
var Module = cmder.NewModule("ramlMocker").
	SetUsage("Go RAML mock web server").
	AddFlag(
		&cli.StringFlag{
			Name:        "f",
			Aliases:     []string{"ramlfile"},
			Value:       "api.raml",
			Usage:       "Source RAML file",
			Destination: &ramlFile,
		},
		&cli.BoolFlag{
			Name:        "checkRAMLVersion",
			Usage:       "Check RAML Version",
			Destination: &checkRAMLVersion,
		},
		&cli.IntFlag{
			Name:        "port",
			Value:       4000,
			Usage:       "Mock server listen port",
			Destination: &port,
		},
	).
	SetAction(action)

var ramlFile string
var checkRAMLVersion bool
var port int

// used for reset routes
var router *gin.Engine

func action(c *cli.Context) (err error) {
	if futil.IsDir(ramlFile) {
		watch(ramlFile)
	} else {
		dir, _ := path.Split(ramlFile)
		watch(dir)
	}
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

	if router == nil {
		addr := fmt.Sprintf(":%d", port)
		router = engineFromRootDocument(router, rootdoc)
		if err = router.Run(addr); err != nil {
			return
		}
	} else {
		router = engineFromRootDocument(router, rootdoc)
	}

	return
}

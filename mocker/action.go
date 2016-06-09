package mocker

import (
	"fmt"
	"path"

	"github.com/codegangsta/cli"
	"github.com/tsaikd/KDGoLib/cliutil/cmder"
	"github.com/tsaikd/KDGoLib/futil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
	"github.com/tsaikd/go-raml-parser/parser/parserConfig"
)

func init() {
	cmder.Name = "ramlMocker"
	cmder.Usage = "Go RAML mock web server"
	cmder.Flags = append(cmder.Flags,
		cli.StringFlag{
			Name:        "f, ramlfile",
			Value:       "api.raml",
			Usage:       "Source RAML file",
			Destination: &ramlFile,
		},
		cli.BoolFlag{
			Name:        "checkRAMLVersion",
			Usage:       "Check RAML Version",
			Destination: &checkRAMLVersion,
		},
		cli.IntFlag{
			Name:        "port",
			Value:       4000,
			Usage:       "Mock server listen port",
			Destination: &port,
		},
	)
	cmder.Action = action
}

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

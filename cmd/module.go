package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/go-raml-mocker/mocker"
)

// command line flags
var (
	flagFile = &cobrather.StringFlag{
		Name:      "ramlfile",
		ShortHand: "f",
		Default:   "api.raml",
		Usage:     "Source RAML file or directory path",
	}
	flagCheckRAMLVersion = &cobrather.BoolFlag{
		Name:  "checkRAMLVersion",
		Usage: "Check RAML Version",
	}
	flagPort = &cobrather.Int64Flag{
		Name:    "port",
		Default: 4000,
		Usage:   "Mock web server listen port",
	}
	flagProxy = &cobrather.StringFlag{
		Name:  "proxy",
		Usage: "Proxy for mock request to original server, used when only mock some of APIs in RAML, keep empty to disable, e.g. http://origin.backend.addr:port",
	}
)

// Module info
var Module = &cobrather.Module{
	Use:   "go-raml-mocker",
	Short: "RAML mock web server written in Go",
	Flags: []cobrather.Flag{
		flagFile,
		flagCheckRAMLVersion,
		flagPort,
		flagProxy,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return mocker.Start(mocker.Config{
			RAMLFile:         flagFile.String(),
			CheckRAMLVersion: flagCheckRAMLVersion.Bool(),
			Port:             flagPort.Int64(),
			Proxy:            flagProxy.String(),
		})
	},
}

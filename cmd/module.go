package cmd

import (
	"context"
	"strings"

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
	flagCacheDir = &cobrather.StringFlag{
		Name:  "cache",
		Usage: "Cache parsed RAML file in cache directory",
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
	flagResources = &cobrather.StringSliceFlag{
		Name:      "resource",
		ShortHand: "r",
		Usage:     "Mock resources, keep empty to mock all found resources",
	}
	flagAllowRequiredPropertyToBeEmpty = &cobrather.BoolFlag{
		Name:  "allowRequiredPropertyToBeEmpty",
		Usage: "allow required property to be empty value, but still should be existed",
	}
)

// Module info
var Module = &cobrather.Module{
	Use:   "go-raml-mocker",
	Short: "RAML mock web server written in Go",
	Example: strings.TrimSpace(`
go-raml-mocker --ramlfile "api.raml" --proxy "https://backend.example.com"
go-raml-mocker --ramlfile "./raml/directory/path" --cache ".ramlcache" --proxy "https://backend.example.com" --resource "/mock/resource1" --resource "/mock/resource2"
	`),
	Commands: []*cobrather.Module{
		cobrather.VersionModule,
	},
	Flags: []cobrather.Flag{
		flagFile,
		flagCheckRAMLVersion,
		flagCacheDir,
		flagPort,
		flagProxy,
		flagResources,
		flagAllowRequiredPropertyToBeEmpty,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) error {
		return mocker.Start(mocker.Config{
			RAMLFile:                       flagFile.String(),
			CheckRAMLVersion:               flagCheckRAMLVersion.Bool(),
			CacheDir:                       flagCacheDir.String(),
			Port:                           flagPort.Int64(),
			Proxy:                          flagProxy.String(),
			Resources:                      mocker.BuildResourcesMap(flagResources.StringSlice()),
			AllowRequiredPropertyToBeEmpty: flagAllowRequiredPropertyToBeEmpty.Bool(),
		})
	},
}

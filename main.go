package main

import (
	"github.com/tsaikd/go-raml-mocker/cmd"

	// load mocker module
	_ "github.com/tsaikd/go-raml-mocker/mocker"
)

func main() {
	cmd.Main()
}

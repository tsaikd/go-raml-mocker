package main

import (
	"github.com/tsaikd/KDGoLib/cliutil/cmder"

	// load mocker module
	_ "github.com/tsaikd/go-raml-mocker/mocker"
)

func main() {
	cmder.Main()
}

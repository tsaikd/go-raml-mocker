package main

import (
	"github.com/tsaikd/KDGoLib/cliutil/cmder"
	"github.com/tsaikd/go-raml-mocker/mocker"
)

func main() {
	cmder.Main(
		*mocker.Module,
	)
}

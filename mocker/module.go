package mocker

import (
	"fmt"
	"path"
	"regexp"

	"github.com/tsaikd/KDGoLib/futil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
	"github.com/tsaikd/go-raml-parser/parser/parserConfig"
)

// Config for mock server
type Config struct {
	RAMLFile         string
	CheckRAMLVersion bool
	Port             int64
	Proxy            string
	Resources        map[string]bool
}

// BuildResourcesMap return resource map by resources string slice
func BuildResourcesMap(resources []string) map[string]bool {
	resmap := map[string]bool{}
	regRAMLParam := regexp.MustCompile(`{(\w+)}`)
	regGinParam := regexp.MustCompile(`:(\w+)`)
	for _, respath := range resources {
		resmap[respath] = true
		ramlpath := regGinParam.ReplaceAllString(respath, "{$1}")
		resmap[ramlpath] = true
		ginpath := regRAMLParam.ReplaceAllString(respath, ":$1")
		resmap[ginpath] = true
	}
	return resmap
}

var config = &Config{}

// Start mock server
func Start(conf Config) (err error) {
	config = &conf

	if futil.IsDir(config.RAMLFile) {
		watch(config.RAMLFile)
	} else {
		dir, _ := path.Split(config.RAMLFile)
		watch(dir)
	}

	return reload()
}

var checkValueOptions = []parser.CheckValueOption{
	parser.CheckValueOptionAllowIntegerToBeNumber(true),
}

// used for reset routes
var router *gin.Engine

func reload() (err error) {
	ramlParser := parser.NewParser()

	if err = ramlParser.Config(parserConfig.CheckRAMLVersion, config.CheckRAMLVersion); err != nil {
		return
	}
	if err = ramlParser.Config(parserConfig.CheckValueOptions, checkValueOptions); err != nil {
		return
	}

	rootdoc, err := ramlParser.ParseFile(config.RAMLFile)
	if err != nil {
		return
	}

	if router == nil {
		addr := fmt.Sprintf(":%d", config.Port)
		router = engineFromRootDocument(router, rootdoc)
		if err = router.Run(addr); err != nil {
			return
		}
	} else {
		router = engineFromRootDocument(router, rootdoc)
	}

	return
}

package mocker

import (
	"regexp"

	"github.com/tsaikd/go-raml-parser/parser"
)

var (
	regRAMLParam = regexp.MustCompile(`{(\w+)}`)
	regGinParam  = regexp.MustCompile(`:(\w+)`)
)

func toRAMLResource(resource string) string {
	return regGinParam.ReplaceAllString(resource, "{$1}")
}

func toGinResource(resource string) string {
	return regRAMLParam.ReplaceAllString(resource, ":$1")
}

func checkConfigResource(resources parser.Resources) (err error) {
	for respath := range config.Resources {
		ramlpath := toRAMLResource(respath)
		if _, exist := resources[ramlpath]; !exist {
			return ErrorResourceNotFound1.New(nil, respath)
		}
	}
	return nil
}

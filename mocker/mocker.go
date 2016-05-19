package mocker

import (
	"regexp"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
)

// errors
var (
	ErrorUnsupportedMIMEType1 = errutil.NewFactory("unsupported MIME type: %q")
)

var router *gin.Engine

// start mock web server or reset router
func start(rootdoc parser.RootDocument, addr string) (err error) {
	if router != nil {
		router.ResetRoutes()
		bindRootDocument(router, rootdoc)
		return
	}

	router = gin.Default()
	router.Use(gin.ErrorLogger())
	bindRootDocument(router, rootdoc)
	return router.Run(addr)
}

func bindRoute(router gin.IRouter, method string, path string, code int, mimetype string, body parser.Body) {
	switch mimetype {
	case "application/json":
		router.Handle(method, path, func(c *gin.Context) {
			if !body.Examples.IsEmpty() {
				for _, example := range body.Examples {
					outputJSON(c, code, example.Value)
					return
				}
			}

			outputJSON(c, code, body.Example.Value)
		})
	default:
		errutil.Trace(ErrorUnsupportedMIMEType1.New(nil, mimetype))
	}
}

func bindRootDocument(router gin.IRouter, rootdoc parser.RootDocument) {
	regParam := regexp.MustCompile(`{(\w+)}`)
	for ramlPath, resource := range rootdoc.Resources {
		ginPath := regParam.ReplaceAllString(ramlPath, ":$1")

		for code, response := range resource.Get.Responses {
			for mimetype, body := range response.Bodies.ForMIMEType {
				bindRoute(router, "GET", ginPath, int(code), mimetype, *body)
			}
		}
	}
}

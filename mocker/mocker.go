package mocker

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
)

// errors
var (
	ErrorUnsupportedMIMEType1 = errutil.NewFactory("unsupported MIME type: %q")
	ErrorHeaderRequired1      = errutil.NewFactory("header %q required")
)

var router *gin.Engine

// start mock web server or reset router
func start(rootdoc parser.RootDocument, addr string) (err error) {
	if _, err = json.Marshal(rootdoc); err != nil {
		errutil.Trace(err)
	}

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

func bindRoute(router gin.IRouter, method string, path string, code int, mimetype string, body parser.Body, istraits ...parser.IsTraits) {
	switch mimetype {
	case "application/json":
		router.Handle(method, path, func(c *gin.Context) {
			for _, istrait := range istraits {
				for _, trait := range istrait {
					for name, header := range trait.Headers {
						if header.Required {
							reqHeader := c.Request.Header.Get(name)
							if reqHeader == "" {
								c.AbortWithError(http.StatusBadRequest, ErrorHeaderRequired1.New(nil, name))
								return
							}
						}
					}
				}
			}

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

		for name, method := range resource.Methods {
			for code, response := range method.Responses {
				for mimetype, body := range response.Bodies.ForMIMEType {
					bindRoute(router, strings.ToUpper(name), ginPath, int(code), mimetype, *body, resource.Is, method.Is)
				}
			}
		}
	}
}

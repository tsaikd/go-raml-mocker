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
	ErrorUnsupportedMIMEType1    = errutil.NewFactory("unsupported MIME type: %q")
	ErrorHeaderRequired1         = errutil.NewFactory("header %q required")
	ErrorQueryParameterRequired1 = errutil.NewFactory("query parameter %q required")
)

const (
	mimeTypeJSON = "application/json"
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
	case mimeTypeJSON:
		router.Handle(method, path, func(c *gin.Context) {
			bodyPost := map[string]interface{}{}
			if method == "POST" {
				if err := c.Bind(&bodyPost); err != nil {
					return
				}
			}

			for _, istrait := range istraits {
				for _, trait := range istrait {
					for name, header := range trait.Headers {
						if !header.Required {
							continue
						}
						if reqHeader := c.Request.Header.Get(name); reqHeader != "" {
							continue
						}
						c.AbortWithError(http.StatusBadRequest, ErrorHeaderRequired1.New(nil, name))
						return
					}

					for name, qp := range trait.QueryParameters {
						if !qp.Required {
							continue
						}
						if _, exist := c.Params.Get(name); exist {
							continue
						}
						if _, exist := c.GetQuery(name); exist {
							continue
						}
						if _, exist := c.GetPostForm(name); exist {
							continue
						}
						if _, exist := bodyPost[name]; exist {
							continue
						}
						c.AbortWithError(http.StatusBadRequest, ErrorQueryParameterRequired1.New(nil, name))
						return
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

func bindDefaultResponse(router gin.IRouter, method string, path string, code int, istraits ...parser.IsTraits) {
	mimetype := mimeTypeJSON
	body := parser.Body{}
	bindRoute(router, method, path, code, mimetype, body, istraits...)
}

func bindRootDocument(router gin.IRouter, rootdoc parser.RootDocument) {
	regParam := regexp.MustCompile(`{(\w+)}`)
	for ramlPath, resource := range rootdoc.Resources {
		ginPath := regParam.ReplaceAllString(ramlPath, ":$1")

		for name, method := range resource.Methods {
			if method == nil || len(method.Responses) < 1 {
				bindDefaultResponse(router, strings.ToUpper(name), ginPath, 200, resource.Is)
				continue
			}

			for code, response := range method.Responses {
				if response == nil {
					bindDefaultResponse(router, strings.ToUpper(name), ginPath, int(code), resource.Is, method.Is)
					continue
				}

				for mimetype, body := range response.Bodies.ForMIMEType {
					bindRoute(router, strings.ToUpper(name), ginPath, int(code), mimetype, *body, resource.Is, method.Is)
				}
			}
		}
	}
}

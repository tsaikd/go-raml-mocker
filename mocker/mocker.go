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

func engineFromRootDocument(prevEngine *gin.Engine, rootdoc parser.RootDocument) *gin.Engine {
	if _, err := json.Marshal(rootdoc); err != nil {
		errutil.Trace(err)
	}

	if prevEngine != nil {
		prevEngine.ResetRoutes()
		bindRootDocument(prevEngine, rootdoc)
		return prevEngine
	}

	router := gin.Default()
	router.Use(gin.ErrorLogger())
	bindRootDocument(router, rootdoc)
	return router
}

func checkValueType(apiType parser.APIType, ivalue interface{}, options ...parser.CheckValueOption) error {
	value, err := parser.NewValue(ivalue)
	if err != nil {
		return err
	}
	if err = parser.CheckValueAPIType(apiType, value, options...); err != nil {
		return err
	}
	return nil
}

func checkHeader(req *http.Request, headerName string, header parser.Header, options ...parser.CheckValueOption) error {
	headerValue := req.Header.Get(headerName)
	if header.Required && headerValue == "" {
		return ErrorHeaderRequired1.New(nil, headerName)
	}
	if err := checkValueType(header.APIType, headerValue, options...); err != nil {
		return err
	}
	return nil
}

func checkTrait(trait parser.Trait, c *gin.Context, requestBody map[string]interface{}) error {
	for headerName, header := range trait.Headers {
		if err := checkHeader(c.Request, headerName, *header); err != nil {
			return err
		}
	}

	for name, qp := range trait.QueryParameters {
		param := getParam(c, name, requestBody)
		if qp.Required && param.IsEmpty() {
			return ErrorQueryParameterRequired1.New(nil, name)
		}
		if err := checkValueType(
			qp.APIType,
			param,
			parser.CheckValueOptionAllowIntegerToBeNumber(true),
		); err != nil {
			return err
		}
	}

	for _, inherit := range trait.Is {
		if err := checkTrait(*inherit, c, requestBody); err != nil {
			return err
		}
	}

	return nil
}

func getParam(c *gin.Context, name string, requestBody map[string]interface{}) parser.Value {
	if param, exist := c.Params.Get(name); exist {
		result, err := parser.NewValue(param)
		errutil.Trace(err)
		return result
	}
	if param, exist := c.GetQuery(name); exist {
		result, err := parser.NewValue(param)
		errutil.Trace(err)
		return result
	}
	if param, exist := c.GetPostForm(name); exist {
		result, err := parser.NewValue(param)
		errutil.Trace(err)
		return result
	}
	if param, exist := requestBody[name]; exist {
		result, err := parser.NewValue(param)
		errutil.Trace(err)
		return result
	}
	return parser.Value{}
}

func bindRoute(
	router gin.IRouter,
	methodName string,
	path string,
	code int,
	mimetype string,
	method parser.Method,
	responseBody parser.Body,
	istraits ...parser.IsTraits,
) {
	switch mimetype {
	case mimeTypeJSON:
		router.Handle(methodName, path, func(c *gin.Context) {
			for headerName, header := range method.Headers {
				if err := checkHeader(c.Request, headerName, *header); err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
					return
				}
			}

			requestBody := map[string]interface{}{}
			if methodBody, exist := method.Bodies[mimetype]; exist {
				if err := c.Bind(&requestBody); err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
					return
				}
				if err := checkValueType(
					methodBody.APIType,
					requestBody,
					parser.CheckValueOptionAllowIntegerToBeNumber(true),
				); err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
					return
				}
			}

			for _, istrait := range istraits {
				for _, trait := range istrait {
					if err := checkTrait(*trait, c, requestBody); err != nil {
						c.AbortWithError(http.StatusBadRequest, err)
						return
					}
				}
			}

			if !responseBody.Examples.IsEmpty() {
				for _, example := range responseBody.Examples {
					outputJSON(c, code, example.Value)
					return
				}
			}

			outputJSON(c, code, responseBody.Example.Value)
		})
	default:
		errutil.Trace(ErrorUnsupportedMIMEType1.New(nil, mimetype))
	}
}

func bindDefaultResponse(
	router gin.IRouter,
	methodName string,
	path string,
	code int,
	method parser.Method,
	istraits ...parser.IsTraits,
) {
	mimetype := mimeTypeJSON
	responseBody := parser.Body{}
	bindRoute(router, methodName, path, code, mimetype, method, responseBody, istraits...)
}

func bindRootDocument(router gin.IRouter, rootdoc parser.RootDocument) {
	regParam := regexp.MustCompile(`{(\w+)}`)
	for ramlPath, resource := range rootdoc.Resources {
		ginPath := regParam.ReplaceAllString(ramlPath, ":$1")

		for name, method := range resource.Methods {
			methodName := strings.ToUpper(name)
			if method == nil {
				bindDefaultResponse(router, methodName, ginPath, 200, parser.Method{}, resource.Is)
				continue
			}
			if len(method.Responses) < 1 {
				bindDefaultResponse(router, methodName, ginPath, 200, *method, resource.Is, method.Is)
				continue
			}

			for code, response := range method.Responses {
				if response == nil {
					bindDefaultResponse(router, methodName, ginPath, int(code), *method, resource.Is, method.Is)
					continue
				}

				for mimetype, responseBody := range response.Bodies {
					bindRoute(router, methodName, ginPath, int(code), mimetype, *method, *responseBody, resource.Is, method.Is)
				}
			}
		}
	}
}

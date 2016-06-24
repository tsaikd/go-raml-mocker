package mocker

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
)

// errors
var (
	ErrorUnsupportedMIMEType1    = errutil.NewFactory("unsupported MIME type: %q")
	ErrorHeaderRequired1         = errutil.NewFactory("header %q required")
	ErrorQueryParameterRequired1 = errutil.NewFactory("query parameter %q required")
	ErrorBindFailed              = errutil.NewFactory("bind request body failed")
	ErrorWSDialFailed            = errutil.NewFactory("websocket dial failed")
	ErrorWSUpgrdaeFailed         = errutil.NewFactory("websocket upgrade failed")
	ErrorWSIOFailed              = errutil.NewFactory("websocket IO failed")
)

const (
	mimeTypeJSON = "application/json"
	mimeTypeBMP  = "image/bmp"
	mimeTypeGIF  = "image/gif"
	mimeTypeJPEG = "image/jpeg"
	mimeTypePNG  = "image/png"
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
	router.NoRoute(proxyRoute)
	router.NoMethod(proxyRoute)
	return router
}

func proxyRoute(c *gin.Context) {
	if proxy == "" {
		return
	}

	logger.Debugf("Proxy to: %s %s", c.Request.Method, proxy+c.Request.RequestURI)

	if c.Request.Header.Get("Upgrade") == "websocket" {
		if err := proxyWebSocket(c); err != nil {
			logger.Debugln(err)
			return
		}
		return
	}

	client := http.Client{}

	req, err := http.NewRequest(c.Request.Method, proxy+c.Request.RequestURI, c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req.Header = c.Request.Header

	resp, err := client.Do(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	outputHeader := c.Writer.Header()
	for name, headers := range resp.Header {
		for _, header := range headers {
			outputHeader.Add(name, header)
		}
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.Data(code, contentType, data)
}

func proxyWebSocket(c *gin.Context) (err error) {
	regexpProto := regexp.MustCompile(`^http`)
	origin := c.Request.Header.Get("Origin")
	wsurl := regexpProto.ReplaceAllString(proxy+c.Request.RequestURI, "ws")
	header := http.Header{}
	header.Set("Origin", origin)
	wssrc, _, err := websocket.DefaultDialer.Dial(wsurl, header)
	if err != nil {
		return ErrorWSDialFailed.New(err)
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsdst, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return ErrorWSUpgrdaeFailed.New(err)
	}

	copyfn := func(dst *websocket.Conn, src *websocket.Conn) (err error) {
		srcType, srcData, err := src.ReadMessage()
		if err != nil {
			return err
		}
		if err = dst.WriteMessage(srcType, srcData); err != nil {
			return err
		}
		return nil
	}

	errorchan := make(chan error)

	go func() {
		for {
			if err = copyfn(wssrc, wsdst); err != nil {
				errorchan <- err
				return
			}
		}
	}()
	go func() {
		for {
			if err = copyfn(wsdst, wssrc); err != nil {
				errorchan <- err
				return
			}
		}
	}()

	err = <-errorchan
	if err != nil {
		switch err.(type) {
		case *websocket.CloseError:
			closeError := err.(*websocket.CloseError)
			switch closeError.Code {
			case websocket.CloseGoingAway:
				return nil
			}
		}
		return ErrorWSIOFailed.New(err)
	}
	return nil
}

func checkValueType(apiType parser.APIType, ivalue interface{}) error {
	value, err := parser.NewValue(ivalue)
	if err != nil {
		return err
	}
	if err = parser.CheckValueAPIType(apiType, value, checkValueOptions...); err != nil {
		return err
	}
	return nil
}

func checkHeader(req *http.Request, headerName string, header parser.Header) error {
	headerValue := req.Header.Get(headerName)
	if header.Required && headerValue == "" {
		return ErrorHeaderRequired1.New(nil, headerName)
	}
	if err := checkValueType(header.APIType, headerValue); err != nil {
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
		if err := checkValueType(qp.APIType, param); err != nil {
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
	var outputFunc func(c *gin.Context, code int, data interface{})

	switch mimetype {
	case mimeTypeJSON:
		outputFunc = outputJSON
	case mimeTypeBMP, mimeTypeGIF, mimeTypeJPEG, mimeTypePNG:
		outputFunc = outputData
	default:
		errutil.Trace(ErrorUnsupportedMIMEType1.New(nil, mimetype))
		return
	}

	router.Handle(methodName, path, func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")

		for headerName, header := range method.Headers {
			if err := checkHeader(c.Request, headerName, *header); err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		requestBody := map[string]interface{}{}
		if methodBody, exist := method.Bodies[mimetype]; exist {
			if err := c.Bind(&requestBody); err != nil {
				if err != io.EOF {
					c.AbortWithError(http.StatusBadRequest, ErrorBindFailed.New(err))
					return
				}
			}
			if err := checkValueType(methodBody.APIType, requestBody); err != nil {
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
				outputFunc(c, code, example.Value)
				return
			}
		}

		outputFunc(c, code, responseBody.Example.Value)
	})
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

	if proxy != "" {
		router.OPTIONS("/*path", func(c *gin.Context) {
			if value := c.Request.Header.Get("Access-Control-Request-Method"); value != "" {
				c.Header("Access-Control-Allow-Methods", value)
			}

			if value := c.Request.Header.Get("Access-Control-Request-Headers"); value != "" {
				c.Header("Access-Control-Allow-Headers", value)
			}

			c.Header("Access-Control-Allow-Origin", "*")
			c.Status(http.StatusOK)
		})
	}
}

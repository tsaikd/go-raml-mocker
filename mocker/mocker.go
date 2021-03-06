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
	ErrorResourceNotFound1       = errutil.NewFactory("resource %q not found in RAML file")
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
	if config.Proxy == "" {
		return
	}

	logger.Debugf("Proxy to: %s %s", c.Request.Method, config.Proxy+c.Request.RequestURI)

	if c.Request.Header.Get("Upgrade") == "websocket" {
		if err := proxyWebSocket(c); err != nil {
			logger.Debugln(err)
			return
		}
		return
	}

	client := http.Client{}

	req, err := http.NewRequest(c.Request.Method, config.Proxy+c.Request.RequestURI, c.Request.Body)
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
	wsurl := regexpProto.ReplaceAllString(config.Proxy+c.Request.RequestURI, "ws")
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
	defer wsdst.Close()

	errorchan := make(chan error, 2)
	done := make(chan bool, 2)

	go func() {
		defer func() {
			done <- true
		}()
		for {
			if err = copyWebSocketMessage(wssrc, wsdst); err != nil {
				errorchan <- err
				return
			}
		}
	}()
	go func() {
		defer func() {
			done <- true
		}()
		for {
			if err = copyWebSocketMessage(wsdst, wssrc); err != nil {
				errorchan <- err
				return
			}
		}
	}()
	defer func() {
		<-done
		<-done
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

func copyWebSocketMessage(dst *websocket.Conn, src *websocket.Conn) (err error) {
	srcType, srcData, err := src.ReadMessage()
	if err != nil {
		return err
	}
	if err = dst.WriteMessage(srcType, srcData); err != nil {
		return err
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

func checkHeader(req *http.Request, header parser.Property) error {
	headerName := header.Name
	headerValue := req.Header.Get(headerName)
	if header.Required && headerValue == "" {
		return ErrorHeaderRequired1.New(nil, headerName)
	}
	if err := checkValueType(header.APIType, headerValue); err != nil {
		return err
	}
	return nil
}

func checkTrait(trait parser.Trait, c *gin.Context, requestBody parser.Value) error {
	for _, header := range trait.Headers.Slice() {
		if err := checkHeader(c.Request, *header); err != nil {
			return err
		}
	}

	for _, qp := range trait.QueryParameters.Slice() {
		name := qp.Name
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

func getParam(c *gin.Context, name string, requestBody parser.Value) parser.Value {
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
	if param, exist := requestBody.Map[name]; exist && param != nil {
		return *param
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

		for _, header := range method.Headers.Slice() {
			if err := checkHeader(c.Request, *header); err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		requestBody := parser.Value{}
		if methodBody, exist := method.Bodies[mimetype]; exist {
			var err error
			if requestBody, err = parseRequestBody(c, methodBody.APIType); err != nil {
				c.AbortWithError(http.StatusBadRequest, ErrorBindFailed.New(err))
				return
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

func parseRequestBody(c *gin.Context, apiType parser.APIType) (reqbody parser.Value, err error) {
	if c.Request.Method != "GET" {
		mapbody := map[string]interface{}{}
		if err = c.Bind(&mapbody); err != nil {
			if err != io.EOF {
				return
			}
		}
		return parser.NewValue(mapbody)
	}

	if err = c.Request.ParseForm(); err != nil {
		return
	}
	return parser.NewValueWithAPIType(apiType, c.Request.Form)
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
	for ramlPath, resource := range rootdoc.Resources {
		if !isNeedToBindResource(ramlPath) {
			continue
		}
		ginPath := toGinResource(ramlPath)

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

	if config.Proxy != "" {
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

func isNeedToBindResource(resourcePath string) bool {
	if len(config.Resources) < 1 {
		return true
	}
	return config.Resources[resourcePath]
}

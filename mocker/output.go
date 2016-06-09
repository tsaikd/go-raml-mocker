package mocker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
)

// errors
var (
	ErrorUnsupportedOutputType1 = errutil.NewFactory("unsupported output type %T")
	ErrorUnexpectedOutputType2  = errutil.NewFactory("output type mismatch, expected %q but got %q")
)

func outputJSON(c *gin.Context, code int, data interface{}) {
	pretty := false
	if queryPretty, exist := c.GetQuery("pretty"); exist {
		pretty, _ = strconv.ParseBool(queryPretty)
	}

	if pretty {
		jsondata, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		buffer := bytes.NewBuffer(jsondata)
		if err = buffer.WriteByte('\n'); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.Data(code, "", buffer.Bytes())
	} else {
		c.JSON(code, data)
	}
}

func outputData(c *gin.Context, code int, data interface{}) {
	switch data.(type) {
	case []byte:
		c.Data(code, "", data.([]byte))
		return
	case parser.Value:
		value := data.(parser.Value)
		if value.Type != parser.TypeBinary {
			c.AbortWithError(http.StatusInternalServerError, ErrorUnexpectedOutputType2.New(nil, "[]byte", value.Type))
			return
		}
		c.Data(code, "", value.Binary)
		return
	default:
		c.AbortWithError(http.StatusInternalServerError, ErrorUnsupportedOutputType1.New(nil, data))
		return
	}
}

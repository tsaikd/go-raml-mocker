package mocker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tsaikd/gin"
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

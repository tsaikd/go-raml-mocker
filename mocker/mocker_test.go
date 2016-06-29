package mocker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gin"
	"github.com/tsaikd/go-raml-parser/parser"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func getBodyValueForJSONType(t *testing.T, res *http.Response) parser.Value {
	require := require.New(t)
	require.NotNil(require)

	contentType := res.Header.Get("Content-Type")
	require.Contains(contentType, mimeTypeJSON)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(err)
	err = res.Body.Close()
	require.NoError(err)

	bodyMap := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyMap)
	require.NoError(err)

	bodyValue, err := parser.NewValue(bodyMap)
	require.NoError(err)

	return bodyValue
}

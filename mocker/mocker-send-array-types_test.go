package mocker

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tsaikd/go-raml-parser/parser"
)

func Test_MockServer_SendArrayTypes(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ramlParser := parser.NewParser()
	require.NotNil(ramlParser)

	rootdoc, err := ramlParser.ParseFile("../example/send-array-types.raml")
	require.NoError(err)

	ts := httptest.NewServer(engineFromRootDocument(nil, rootdoc))
	defer ts.Close()
	require.NotNil(ts)

	client := http.DefaultClient

	func() {
		req, err := http.NewRequest("POST", ts.URL+"/send", bytes.NewBufferString(`{
			"bools": [true,false],
			"ints": [1,2,3],
			"nums": [1.1,2.2],
			"strs": ["a","b"]
		}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusOK, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	func() {
		req, err := http.NewRequest("POST", ts.URL+"/send", bytes.NewBufferString(`{
				"bools": [true,false,"a"],
				"ints": [1,2,3,"b"],
				"nums": [1.1,2.2,"c"],
				"strs": ["a","b",0]
			}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()
}

package mocker

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tsaikd/go-raml-parser/parser"
)

func Test_MockServer_OrganisationAPI(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ramlParser := parser.NewParser()
	require.NotNil(ramlParser)

	rootdoc, err := ramlParser.ParseFile("../example/organisation-api.raml")
	require.NoError(err)

	ts := httptest.NewServer(engineFromRootDocument(nil, rootdoc))
	defer ts.Close()
	require.NotNil(ts)

	client := http.DefaultClient

	// test get resource
	func() {
		req, err := http.NewRequest("GET", ts.URL+"/organisation", nil)
		require.NoError(err)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(201, res.StatusCode)

		body := getBodyValueForJSONType(t, res)
		require.Equal(parser.TypeObject, body.Type)
		require.Contains(body.Map, "name")
	}()

	// test get non-exist resource
	func() {
		req, err := http.NewRequest("GET", ts.URL+"/error", nil)
		require.NoError(err)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(404, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	// test post resource
	func() {
		req, err := http.NewRequest("POST", ts.URL+"/organisation", bytes.NewBufferString(`{
			"name": "foo"
		}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)
		req.Header.Set("UserID", "SWED-123")

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusOK, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	// test post resource without valid header
	func() {
		req, err := http.NewRequest("POST", ts.URL+"/organisation", bytes.NewBufferString(`{
				"name": "foo"
			}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	// test post resource with wrong parameter
	func() {
		req, err := http.NewRequest("POST", ts.URL+"/organisation", bytes.NewBufferString(`{
			"name": "foo",
			"value": 9527
		}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)
		req.Header.Set("UserID", "SWED-123")

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	// test post resource without parameter
	func() {
		req, err := http.NewRequest("POST", ts.URL+"/organisation", bytes.NewBufferString(`{
		}`))
		require.NoError(err)
		req.Header.Set("Content-Type", mimeTypeJSON)
		req.Header.Set("UserID", "SWED-123")

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()
}

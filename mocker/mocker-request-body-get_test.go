package mocker

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tsaikd/go-raml-parser/parser"
)

func Test_MockServer_RequestBodyGet(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ramlParser := parser.NewParser()
	require.NotNil(ramlParser)

	rootdoc, err := ramlParser.ParseFile("../example/request-body-get.raml")
	require.NoError(err)

	ts := httptest.NewServer(engineFromRootDocument(nil, rootdoc))
	defer ts.Close()
	require.NotNil(ts)

	client := http.DefaultClient

	func() {
		params := url.Values{
			"bool":  []string{"true"},
			"bools": []string{"false", "true"},
			"int":   []string{"9527"},
			"ints":  []string{"5566", "9527"},
			"num":   []string{"3.14"},
			"nums":  []string{"6.02", "3.14"},
			"str":   []string{"test string value"},
			"strs":  []string{"str1", "str2"},
		}

		requrl, err := url.Parse(ts.URL + "/query")
		require.NoError(err)
		requrl.RawQuery = params.Encode()

		req, err := http.NewRequest("GET", requrl.String(), nil)
		require.NoError(err)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusOK, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	func() {
		params := url.Values{
			"bool":  []string{"true"},
			"bools": []string{"false", "true"},
			"int":   []string{"9527"},
			"ints":  []string{"5566", "9527"},
			"num":   []string{"3.14"},
			"nums":  []string{"6.02", "3.14"},
			"str":   []string{""},
			"strs":  []string{"str1", "str2"},
		}

		requrl, err := url.Parse(ts.URL + "/query")
		require.NoError(err)
		requrl.RawQuery = params.Encode()

		req, err := http.NewRequest("GET", requrl.String(), nil)
		require.NoError(err)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()

	func() {
		params := url.Values{
			"bool":  []string{"true"},
			"bools": []string{"false", "true"},
			"int":   []string{"9527"},
			"ints":  []string{"5566", "9527"},
			"num":   []string{"3.14"},
			"nums":  []string{"6.02", "3.14"},
			"str":   []string{"test string value"},
			"strs":  []string{},
		}

		requrl, err := url.Parse(ts.URL + "/query")
		require.NoError(err)
		requrl.RawQuery = params.Encode()

		req, err := http.NewRequest("GET", requrl.String(), nil)
		require.NoError(err)

		res, err := client.Do(req)
		require.NoError(err)
		require.EqualValues(http.StatusBadRequest, res.StatusCode)

		err = res.Body.Close()
		require.NoError(err)
	}()
}

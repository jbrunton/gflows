package fixtures

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TestRoundTripper struct {
	responses map[string]*http.Response
}

func (roundTripper *TestRoundTripper) StubResponse(url string, response *http.Response) {
	roundTripper.responses[url] = response
}

func (roundTripper *TestRoundTripper) StubBody(url string, responseBody string) {
	roundTripper.StubResponse(url, &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(responseBody)),
		Header:     make(http.Header),
	})
}

func (roundTripper *TestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	url := request.URL.String()
	response := roundTripper.responses[url]
	if response == nil {
		return nil, fmt.Errorf("Missing response for %s", url)
	}
	return response, nil
}

func NewTestRoundTripper() *TestRoundTripper {
	return &TestRoundTripper{
		responses: make(map[string]*http.Response),
	}
}

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

type TestHttpClient struct {
	*http.Client
	roundTripper *TestRoundTripper
}

func (httpClient *TestHttpClient) StubResponse(url string, response *http.Response) {
	fmt.Println("stubbing response for", url)
	httpClient.roundTripper.responses[url] = response
}

func (httpClient *TestHttpClient) StubBody(url string, responseBody string) {
	httpClient.StubResponse(url, &http.Response{
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

func NewTestClient() *TestHttpClient {
	testRoundTripper := &TestRoundTripper{
		responses: make(map[string]*http.Response),
	}
	return &TestHttpClient{
		&http.Client{Transport: testRoundTripper},
		testRoundTripper,
	}
}

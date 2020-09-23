package fixtures

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockRoundTripper struct {
	mock.Mock
}

func (roundTripper *MockRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	args := roundTripper.Called(request)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (roundTripper *MockRoundTripper) StubResponse(url string, response *http.Response) {
	roundTripper.
		On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool { return req.URL.String() == url })).
		Return(response, nil)
}

func (roundTripper *MockRoundTripper) StubBody(url string, responseBody string) {
	roundTripper.StubResponse(url, &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(responseBody)),
		Header:     make(http.Header),
	})
}

func (roundTripper *MockRoundTripper) StubStatusCode(url string, statusCode int) {
	roundTripper.StubResponse(url, &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	})
}

func NewMockRoundTripper() *MockRoundTripper {
	return new(MockRoundTripper)
}

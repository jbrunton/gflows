package content

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestReadRemoteFile(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	roundTripper.StubBody("https://example.com/my-file.txt", "my file")
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	reader := NewReader(fs, &http.Client{Transport: roundTripper})

	content, err := reader.ReadContent("https://example.com/my-file.txt")

	assert.Equal(t, "my file", content)
	assert.Nil(t, err)
}

func TestReadLocalFile(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	reader := NewReader(fs, &http.Client{Transport: roundTripper})
	fs.WriteFile("/my-file", []byte("my file"), 0644)

	content, err := reader.ReadContent("/my-file")

	assert.Equal(t, "my file", content)
	assert.Nil(t, err)
}

func TestHttpError(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	roundTripper.StubResponse("https://example.com/my-file.txt", &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	})
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	reader := NewReader(fs, &http.Client{Transport: roundTripper})

	_, err := reader.ReadContent("https://example.com/my-file.txt")

	assert.EqualError(t, err, fmt.Sprintf("Received status code 500 from https://example.com/my-file.txt"))
}

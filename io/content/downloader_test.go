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

func TestDownloadFile(t *testing.T) {
	testClient := fixtures.NewTestClient()
	testClient.StubBody("https://example.com/my-file.txt", "my file")
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	writer := NewWriter(fs, container.Logger())
	downloader := NewDownloader(fs, writer, testClient.Client)

	err := downloader.DownloadFile("https://example.com/my-file.txt", "/my/file")

	content, _ := fs.ReadFile("/my/file")
	assert.Equal(t, "my file", string(content))
	assert.Nil(t, err)
}

func TestHttpError(t *testing.T) {
	testClient := fixtures.NewTestClient()
	testClient.StubResponse("https://example.com/my-file.txt", &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	})
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	writer := NewWriter(fs, container.Logger())
	downloader := NewDownloader(fs, writer, testClient.Client)

	err := downloader.DownloadFile("https://example.com/my-file.txt", "/my/file")

	assert.EqualError(t, err, fmt.Sprintf("Received status code 500 from https://example.com/my-file.txt"))
}

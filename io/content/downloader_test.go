package content

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "my file")
	}))
	defer server.Close()
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	writer := NewWriter(fs, container.Logger())
	downloader := NewDownloader(fs, writer)

	err := downloader.DownloadFile(server.URL, "/my/file")

	content, _ := fs.ReadFile("/my/file")
	assert.Equal(t, "my file\n", string(content))
	assert.Nil(t, err)
}

func TestHttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	container, _, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	writer := NewWriter(fs, container.Logger())
	downloader := NewDownloader(fs, writer)

	err := downloader.DownloadFile(server.URL, "/my/file")

	assert.EqualError(t, err, fmt.Sprintf("Received status code 500 from %s", server.URL))
}

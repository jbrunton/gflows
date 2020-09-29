package content

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jbrunton/gflows/io/pkg"
	"github.com/spf13/afero"
)

// Reader - reads files that are either local or remote
type Reader struct {
	fs         *afero.Afero
	httpClient *http.Client
}

func NewReader(fs *afero.Afero, httpClient *http.Client) *Reader {
	return &Reader{
		fs:         fs,
		httpClient: httpClient,
	}
}

func (reader *Reader) ReadContent(path string) (string, error) {
	if !pkg.IsRemotePath(path) {
		data, err := reader.fs.ReadFile(path)
		return string(data), err
	}

	resp, err := reader.httpClient.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Received status code %d from %s", resp.StatusCode, path)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

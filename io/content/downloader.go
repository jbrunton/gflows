package content

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/afero"
)

type Downloader struct {
	fs         *afero.Afero
	writer     *Writer
	httpClient *http.Client
}

func NewDownloader(fs *afero.Afero, writer *Writer, httpClient *http.Client) *Downloader {
	return &Downloader{
		fs:         fs,
		writer:     writer,
		httpClient: httpClient,
	}
}

func (downloader *Downloader) DownloadFile(url string, path string) error {
	resp, err := downloader.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Received status code %d from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	downloader.writer.SafelyWriteFile(path, string(body))
	fmt.Println("  Downloaded", url)

	return nil
}

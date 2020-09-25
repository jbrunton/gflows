package content

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jbrunton/gflows/io"
	"github.com/spf13/afero"
)

type Downloader struct {
	fs         *afero.Afero
	writer     *Writer
	httpClient *http.Client
	logger     *io.Logger
}

func NewDownloader(fs *afero.Afero, writer *Writer, httpClient *http.Client, logger *io.Logger) *Downloader {
	return &Downloader{
		fs:         fs,
		writer:     writer,
		httpClient: httpClient,
		logger:     logger,
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
	downloader.logger.Debug("Downloaded", url)

	return nil
}

func (downloader *Downloader) CopyFile(sourcePath string, destPath string) error {
	data, err := downloader.fs.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return downloader.writer.SafelyWriteFile(destPath, string(data))
}

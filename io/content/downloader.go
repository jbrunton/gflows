package content

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/afero"
)

type Downloader struct {
	fs     *afero.Afero
	writer *Writer
}

func NewDownloader(fs *afero.Afero, writer *Writer) *Downloader {
	return &Downloader{
		fs:     fs,
		writer: writer,
	}
}

func (downloader *Downloader) DownloadFile(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	downloader.writer.SafelyWriteFile(path, string(body))
	fmt.Println("  Downloaded", url)

	return nil
}

package content

import (
	"net/http"

	"github.com/jbrunton/gflows/io"
)

type Container struct {
	*io.Container
	httpClient *http.Client
}

func (container *Container) HttpClient() *http.Client {
	return container.httpClient
}

func (container *Container) ContentWriter() *Writer {
	return NewWriter(container.FileSystem(), container.Logger())
}

func (container *Container) ContentReader() *Reader {
	return NewReader(container.FileSystem(), container.HttpClient())
}

func NewContainer(parentContainer *io.Container, httpClient *http.Client) *Container {
	return &Container{parentContainer, httpClient}
}

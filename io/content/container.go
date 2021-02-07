package content

import (
	"net/http"

	"github.com/jbrunton/gflows/io"
)

type Container struct {
	*io.Container
	httpClient  *http.Client
	repoManager *RepoManager
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

func (container *Container) RepoManager() *RepoManager {
	return container.repoManager
}

func NewContainer(parentContainer *io.Container, httpClient *http.Client) *Container {
	repoManager := NewRepoManager(parentContainer.GitAdapter(), parentContainer.FileSystem(), parentContainer.Logger())
	return &Container{parentContainer, httpClient, repoManager}
}

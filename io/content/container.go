package content

import "github.com/jbrunton/gflows/io"

type Container struct {
	*io.Container
}

func (container *Container) ContentWriter() *Writer {
	return NewWriter(container.FileSystem(), container.Logger())
}

func (container *Container) Downloader() *Downloader {
	return NewDownloader(container.FileSystem(), container.ContentWriter())
}

func NewContainer(parentContainer *io.Container) *Container {
	return &Container{parentContainer}
}

// func CreateContainer(enableColors bool) *Container {
// 	return NewContainer(io.CreateContainer(enableColors))
// }

package content

import "github.com/jbrunton/gflows/adapters"

type Container struct {
	*adapters.Container
}

func (container *Container) ContentWriter() *Writer {
	return NewWriter(container.FileSystem(), container.Logger())
}

func CreateContainer() *Container {
	return &Container{
		adapters.CreateContainer(),
	}
}

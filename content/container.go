package content

import "github.com/jbrunton/gflows/adapters"

type Container struct {
	*adapters.Container
}

func (container *Container) ContentWriter() *Writer {
	return NewWriter(container.FileSystem(), container.Logger())
}

func NewContainer(parentContainer *adapters.Container) *Container {
	return &Container{parentContainer}
}

func CreateContainer() *Container {
	return NewContainer(adapters.CreateContainer())
}

package adapters

import (
	"os"

	"github.com/spf13/afero"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *Logger
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *Logger {
	return container.logger
}

func NewContainer(fs *afero.Afero, logger *Logger) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
	}
}

func CreateContainer() *Container {
	return NewContainer(
		CreateOsFs(),
		NewLogger(os.Stdout),
	)
}

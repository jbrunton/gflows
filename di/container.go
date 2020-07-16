package di

import (
	"os"

	"github.com/jbrunton/gflows/fs"
	"github.com/jbrunton/gflows/logs"
	"github.com/spf13/afero"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *logs.Logger
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *logs.Logger {
	return container.logger
}

func NewContainer() *Container {
	return BuildContainer(
		fs.CreateOsFs(),
		logs.NewLogger(os.Stdout),
	)
}

func BuildContainer(fs *afero.Afero, logger *logs.Logger) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
	}
}

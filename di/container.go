package di

import (
	"github.com/jbrunton/gflows/logs"
	"github.com/spf13/afero"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *logs.Logger
}

func NewContainer(fileSystem *afero.Afero, logger *logs.Logger) *Container {
	return &Container{
		fileSystem: fileSystem,
		logger:     logger,
	}
}

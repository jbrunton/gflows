package adapters

import (
	"github.com/jbrunton/gflows/styles"
	"github.com/spf13/afero"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *Logger
	styles     *styles.Styles
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *Logger {
	return container.logger
}

func (container *Container) Styles() *styles.Styles {
	return container.styles
}

func NewContainer(fs *afero.Afero, logger *Logger, styles *styles.Styles) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
		styles:     styles,
	}
}

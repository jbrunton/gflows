package io

import (
	"github.com/jbrunton/gflows/io/styles"
	"github.com/spf13/afero"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *Logger
	styles     *styles.Styles
	gitAdapter GitAdapter
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

func (container *Container) GitAdapter() GitAdapter {
	return container.gitAdapter
}

func NewContainer(fs *afero.Afero, logger *Logger, styles *styles.Styles, gitAdapter GitAdapter) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
		styles:     styles,
		gitAdapter: gitAdapter,
	}
}

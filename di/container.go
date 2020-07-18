package di

import (
	"os"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Container struct {
	fileSystem *afero.Afero
	logger     *adapters.Logger
	context    *config.GFlowsContext
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *adapters.Logger {
	return container.logger
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func NewContainer(cmd *cobra.Command) (*Container, error) {
	fs := adapters.CreateOsFs()
	context, err := config.GetContext(fs, cmd)
	if err != nil {
		return nil, err
	}
	return BuildContainer(
		fs,
		adapters.NewLogger(os.Stdout),
		context,
	), nil
}

func BuildContainer(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
		context:    context,
	}
}

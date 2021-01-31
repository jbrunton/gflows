package di

import (
	"os"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/fs"
	"github.com/jbrunton/gflows/logs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Factory interface {
	FileSystem() *afero.Afero
	Logger() *logs.Logger
	Context() *config.GFlowsContext
}

type Container struct {
	fileSystem *afero.Afero
	logger     *logs.Logger
	context    *config.GFlowsContext
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *logs.Logger {
	return container.logger
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func NewContainer(cmd *cobra.Command) (*Container, error) {
	fs := fs.CreateOsFs()
	context, err := config.GetContext(fs, cmd)
	if err != nil {
		return nil, err
	}
	return BuildContainer(
		fs,
		logs.NewLogger(os.Stdout),
		context,
	), nil
}

func BuildContainer(fs *afero.Afero, logger *logs.Logger, context *config.GFlowsContext) *Container {
	return &Container{
		fileSystem: fs,
		logger:     logger,
		context:    context,
	}
}

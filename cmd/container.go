package cmd

import (
	"os"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/styles"
	"github.com/jbrunton/gflows/workflow/action"
	"github.com/spf13/cobra"
)

// ContainerBuilderFunc - factory function to create a new container for the given command
type ContainerBuilderFunc func(cmd *cobra.Command) (*action.Container, error)

func buildContainer(cmd *cobra.Command) (*action.Container, error) {
	fs := adapters.CreateOsFs()
	logger := adapters.NewLogger(os.Stdout)
	context, err := config.GetContext(fs, logger, cmd)
	if err != nil {
		return nil, err
	}
	adaptersContainer := adapters.NewContainer(fs, logger, styles.NewStyles(context.EnableColors))
	contentContainer := content.NewContainer(adaptersContainer)
	return action.NewContainer(contentContainer, context), nil
}

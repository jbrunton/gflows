package cmd

import (
	"os"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/styles"
	"github.com/jbrunton/gflows/workflows"
	"github.com/spf13/cobra"
)

// ContainerBuilderFunc - factory function to create a new container for the given command
type ContainerBuilderFunc func(cmd *cobra.Command) (*workflows.Container, error)

func buildContainer(cmd *cobra.Command) (*workflows.Container, error) {
	fs := adapters.CreateOsFs()
	context, err := config.GetContext(fs, cmd)
	if err != nil {
		return nil, err
	}
	adaptersContainer := adapters.NewContainer(fs, adapters.NewLogger(os.Stdout), styles.NewStyles(context.EnableColors))
	contentContainer := content.NewContainer(adaptersContainer)
	return workflows.NewContainer(contentContainer, context), nil
}

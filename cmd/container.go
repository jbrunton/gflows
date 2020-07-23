package cmd

import (
	"os"

	"github.com/jbrunton/gflows/actions"
	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/styles"
	"github.com/spf13/cobra"
)

// ContainerBuilderFunc - factory function to create a new container for the given command
type ContainerBuilderFunc func(cmd *cobra.Command) (*actions.Container, error)

func buildContainer(cmd *cobra.Command) (*actions.Container, error) {
	fs := adapters.CreateOsFs()
	context, err := config.GetContext(fs, cmd)
	if err != nil {
		return nil, err
	}
	adaptersContainer := adapters.NewContainer(fs, adapters.NewLogger(os.Stdout), styles.NewStyles(context.EnableColors))
	contentContainer := content.NewContainer(adaptersContainer)
	return actions.NewContainer(contentContainer, context), nil
}

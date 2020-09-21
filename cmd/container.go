package cmd

import (
	"net/http"
	"os"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/styles"
	"github.com/jbrunton/gflows/workflow/action"
	"github.com/spf13/cobra"
)

// ContainerBuilderFunc - factory function to create a new container for the given command
type ContainerBuilderFunc func(cmd *cobra.Command) (*action.Container, error)

func buildContainer(cmd *cobra.Command) (*action.Container, error) {
	fs := io.CreateOsFs()
	opts := config.CreateContextOpts(cmd)
	logger := io.NewLogger(os.Stdout, opts.EnableColors, opts.Debug)
	context, err := config.NewContext(fs, logger, opts)
	if err != nil {
		return nil, err
	}
	ioContainer := io.NewContainer(fs, logger, styles.NewStyles(context.EnableColors))
	contentContainer := content.NewContainer(ioContainer, http.DefaultClient)
	return action.NewContainer(contentContainer, context), nil
}

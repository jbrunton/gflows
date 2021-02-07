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

var containers map[*cobra.Command]*action.Container

func buildContainer(cmd *cobra.Command) (*action.Container, error) {
	if containers[cmd] != nil {
		return containers[cmd], nil
	}

	fs := io.CreateOsFs()
	opts := config.CreateContextOpts(cmd)
	logger := io.NewLogger(os.Stdout, opts.EnableColors, opts.Debug)
	gitAdapter := io.NewGoGitAdapter()
	context, err := config.NewContext(fs, logger, opts)
	if err != nil {
		return nil, err
	}
	ioContainer := io.NewContainer(fs, logger, styles.NewStyles(context.EnableColors), gitAdapter)
	contentContainer := content.NewContainer(
		ioContainer,
		http.DefaultClient,
	)
	container, err := action.NewContainer(contentContainer, context), nil
	if err == nil {
		containers[cmd.Root()] = container
	}
	return container, err
}

func init() {
	containers = make(map[*cobra.Command]*action.Container)
}

func CleanUp(cmd *cobra.Command) {
	container := containers[cmd.Root()]
	if container != nil {
		container.Environment().CleanUp()
		delete(containers, cmd)
	}
}

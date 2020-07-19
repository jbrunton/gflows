package cmd

import (
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflows"
	"github.com/spf13/cobra"
)

func buildContainer(cmd *cobra.Command) (*workflows.Container, error) {
	contentContainer := content.CreateContainer()
	context, err := config.GetContext(contentContainer.FileSystem(), cmd)
	if err != nil {
		return nil, err
	}
	return workflows.NewContainer(contentContainer, context), nil
}

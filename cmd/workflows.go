package cmd

import (
	"fmt"

	"github.com/jbrunton/gflows/styles"
	"github.com/jbrunton/gflows/workflows"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

func newListWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}

			workflowManager := container.WorkflowManager()
			definitions, err := workflowManager.GetWorkflowDefinitions()
			if err != nil {
				return err
			}
			validator := container.WorkflowValidator()

			table := tablewriter.NewWriter(container.Logger())
			table.SetHeader([]string{"Name", "Source", "Target", "Status"})
			for _, definition := range definitions {
				colors := []tablewriter.Colors{
					tablewriter.Colors{tablewriter.FgGreenColor},
					tablewriter.Colors{tablewriter.FgYellowColor},
					tablewriter.Colors{tablewriter.FgYellowColor},
					tablewriter.Colors{},
				}
				var status string
				if !validator.ValidateSchema(definition).Valid {
					status = "INVALID SCHEMA"
					colors[3] = tablewriter.Colors{tablewriter.FgRedColor}
				} else if !validator.ValidateContent(definition).Valid {
					status = "OUT OF DATE"
					colors[3] = tablewriter.Colors{tablewriter.FgRedColor}
				} else {
					status = "UP TO DATE"
					colors[3] = tablewriter.Colors{tablewriter.FgGreenColor}
				}

				row := []string{definition.Name, definition.Source, definition.Destination, status}
				table.Rich(row, colors)
			}
			table.Render()
			return nil
		},
	}
}

func newUpdateWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates workflow files",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}
			workflowManager := container.WorkflowManager()
			err = workflowManager.UpdateWorkflows()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func newInitCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Setup config and templates for first time use",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}
			err = workflows.InitWorkflows(container.FileSystem(), container.Logger(), container.Context())
			return err
		},
	}
}

func checkWorkflows(workflowManager *workflows.WorkflowManager, styles *styles.Styles, watch bool, showDiff bool) error {
	err := workflowManager.ValidateWorkflows(showDiff)
	if err != nil {
		return err
	}
	fmt.Println(styles.StyleCommand("Workflows up to date"))
	return nil
}

func newCheckWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check workflow files are up to date",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}

			watch, err := cmd.Flags().GetBool("watch")
			if err != nil {
				return err
			}

			showDiff, err := cmd.Flags().GetBool("show-diffs")
			if err != nil {
				return err
			}

			workflowManager := container.WorkflowManager()
			if watch {
				watcher := container.Watcher()
				watcher.WatchWorkflows(func() {
					checkWorkflows(workflowManager, container.Styles(), watch, showDiff)
				})
			} else {
				err = checkWorkflows(workflowManager, container.Styles(), watch, showDiff)
			}
			return err
		},
	}
	cmd.Flags().BoolP("watch", "w", false, "watch workflow templates for changes")
	cmd.Flags().Bool("show-diffs", false, "show diff with generated workflow (useful when refactoring)")
	return cmd
}

func newWatchWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Alias for check --watch --show-diffs",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}

			workflowManager := container.WorkflowManager()
			watcher := container.Watcher()
			watcher.WatchWorkflows(func() {
				checkWorkflows(workflowManager, container.Styles(), true, true)
			})
			return nil
		},
	}
	return cmd
}

func newImportWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import workflow files",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}
			manager := container.WorkflowManager()
			manager.ImportWorkflows()
			return nil
		},
	}
}

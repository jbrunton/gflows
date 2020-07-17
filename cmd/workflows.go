package cmd

import (
	"fmt"
	"os"

	"github.com/jbrunton/gflows/di"
	"github.com/jbrunton/gflows/styles"
	"github.com/jbrunton/gflows/workflows"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

func newListWorkflowsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List workflows",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}

			workflowManager := workflows.NewWorkflowManager(container)
			definitions, err := workflowManager.GetWorkflowDefinitions()
			if err != nil {
				panic(err)
			}
			validator := workflows.NewWorkflowValidator(container)

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
		},
	}
}

func newUpdateWorkflowsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates workflow files",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			workflowManager := workflows.NewWorkflowManager(container)
			err = workflowManager.UpdateWorkflows()
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Setup config and templates for first time use",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			workflows.InitWorkflows(container)
		},
	}
}

func checkWorkflows(container *di.Container, watch bool, showDiff bool) {
	workflowManager := workflows.NewWorkflowManager(container)
	err := workflowManager.ValidateWorkflows(showDiff)
	if err != nil {
		fmt.Println(styles.StyleError(err.Error()))
		if !watch {
			os.Exit(1)
		}
	} else {
		fmt.Println(styles.StyleCommand("Workflows up to date"))
	}
}

func newCheckWorkflowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check workflow files are up to date",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}

			watch, err := cmd.Flags().GetBool("watch")
			if err != nil {
				panic(err)
			}

			showDiff, err := cmd.Flags().GetBool("show-diffs")
			if err != nil {
				panic(err)
			}

			if watch {
				workflows.WatchWorkflows(container, func() {
					checkWorkflows(container, watch, showDiff)
				})
			} else {
				checkWorkflows(container, watch, showDiff)
			}
		},
	}
	cmd.Flags().BoolP("watch", "w", false, "watch workflow templates for changes")
	cmd.Flags().Bool("show-diffs", false, "show diff with generated workflow (useful when refactoring)")
	return cmd
}

func newWatchWorkflowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Alias for check --watch --show-diffs",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}

			workflows.WatchWorkflows(container, func() {
				checkWorkflows(container, true, true)
			})
		},
	}
	return cmd
}

func newImportWorkflowsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import workflow files",
		Run: func(cmd *cobra.Command, args []string) {
			container, err := di.NewContainer(cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			workflows.ImportWorkflows(container)
		},
	}
}

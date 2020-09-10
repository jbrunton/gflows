package cmd

import (
	"errors"
	"fmt"

	"github.com/jbrunton/gflows/workflow/action"
	"github.com/olekukonko/tablewriter"
	"github.com/thoas/go-funk"

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
			validator := container.Validator()

			table := tablewriter.NewWriter(container.Logger())
			table.SetHeader([]string{"Name", "Source", "Target", "Status"})
			context := container.Context()
			for _, definition := range definitions {
				colors := []tablewriter.Colors{
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
				}
				if context.EnableColors {
					colors[0] = tablewriter.Colors{tablewriter.FgGreenColor}
					colors[1] = tablewriter.Colors{tablewriter.FgYellowColor}
					colors[2] = tablewriter.Colors{tablewriter.FgYellowColor}
				}
				var status string
				if !definition.Status.Valid {
					status = "TEMPLATE ERROR"
					if context.EnableColors {
						colors[3] = tablewriter.Colors{tablewriter.FgRedColor}
					}
				} else if !validator.ValidateSchema(definition).Valid {
					status = "INVALID SCHEMA"
					if context.EnableColors {
						colors[3] = tablewriter.Colors{tablewriter.FgRedColor}
					}
				} else if !validator.ValidateContent(definition).Valid {
					status = "OUT OF DATE"
					if context.EnableColors {
						colors[3] = tablewriter.Colors{tablewriter.FgRedColor}
					}
				} else {
					status = "UP TO DATE"
					if context.EnableColors {
						colors[3] = tablewriter.Colors{tablewriter.FgGreenColor}
					}
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
	cmd := &cobra.Command{
		Use:   "init --engine <ytt|jsonnet>",
		Short: "Setup config and templates for first time use using the given template engine",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := cmd.Flags().GetString("engine")
			if err != nil {
				panic(err)
			}
			if engine == "" {
				return errors.New("--engine flag required")
			}
			if !funk.ContainsString([]string{"jsonnet", "ytt"}, engine) {
				return fmt.Errorf("Unexpected engine name: %q, valid options are ytt or jsonnet", engine)
			}
			workflowName, err := cmd.Flags().GetString("name")
			if err != nil {
				panic(err)
			}

			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}

			container.WorkflowManager().InitWorkflows(workflowName)
			return nil
		},
	}
	cmd.Flags().String("engine", "", "the template engine to use (either jsonnet or ytt)")
	cmd.Flags().String("name", "gflows", "the name of the workflow to generate")
	return cmd
}

func checkWorkflows(workflowManager *action.WorkflowManager, container *action.Container, watch bool, showDiff bool) error {
	err := workflowManager.ValidateWorkflows(showDiff)
	if err != nil {
		return err
	}
	container.Logger().Println(container.Styles().StyleCommand("Workflows up to date"))
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
					checkWorkflows(workflowManager, container, watch, showDiff)
				})
			} else {
				err = checkWorkflows(workflowManager, container, watch, showDiff)
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
				checkWorkflows(workflowManager, container, true, true)
			})
			return nil
		},
	}
	return cmd
}

func newImportWorkflowsCmd(containerFunc ContainerBuilderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import existing workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := containerFunc(cmd)
			if err != nil {
				return err
			}
			manager := container.WorkflowManager()
			err = manager.ImportWorkflows()
			return err
		},
	}
}

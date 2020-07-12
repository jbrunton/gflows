package cmd

import (
	"fmt"
	"os"

	"github.com/jbrunton/gflows/styles"
	"github.com/olekukonko/tablewriter"

	"github.com/jbrunton/gflows/lib"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newListWorkflowsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List workflows",
		Run: func(cmd *cobra.Command, args []string) {
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}

			definitions, err := lib.GetWorkflowDefinitions(fs, context)
			if err != nil {
				panic(err)
			}
			validator := lib.NewWorkflowValidator(fs, context)

			table := tablewriter.NewWriter(os.Stdout)
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
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			lib.UpdateWorkflows(fs, context)
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Setup config and templates for first time use",
		Run: func(cmd *cobra.Command, args []string) {
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			lib.InitWorkflows(fs, context)
		},
	}
}

func checkWorkflows(fs *afero.Afero, context *lib.GFlowsContext, watch bool, showDiff bool) {
	err := lib.ValidateWorkflows(fs, context, showDiff)
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
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
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
				lib.WatchWorkflows(fs, context, func() {
					checkWorkflows(fs, context, watch, showDiff)
				})
			} else {
				checkWorkflows(fs, context, watch, showDiff)
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
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}

			lib.WatchWorkflows(fs, context, func() {
				checkWorkflows(fs, context, true, true)
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
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			lib.ImportWorkflows(fs, context)
		},
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/jbrunton/jflows/styles"
	"github.com/olekukonko/tablewriter"

	"github.com/jbrunton/jflows/lib"
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

			definitions := lib.GetWorkflowDefinitions(fs, context)
			validator := lib.NewWorkflowValidator(fs)

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

func newCheckWorkflowsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check workflow files are up to date",
		Run: func(cmd *cobra.Command, args []string) {
			fs := lib.CreateOsFs()
			context, err := lib.GetContext(fs, cmd)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			}
			err = lib.ValidateWorkflows(fs, context)
			if err != nil {
				fmt.Println(styles.StyleError(err.Error()))
				os.Exit(1)
			} else {
				fmt.Println(styles.StyleCommand("Workflows up to date"))
			}
		},
	}
}

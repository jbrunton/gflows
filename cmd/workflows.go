package cmd

import (
	"fmt"
	"os"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/jbrunton/jflows/diff"
	"github.com/jbrunton/jflows/styles"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"

	"github.com/jbrunton/jflows/lib"
	"github.com/spf13/cobra"
)

func diffWorkflows(cmd *cobra.Command) {
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

	for _, definition := range definitions {
		result := validator.ValidateContent(definition)
		fmt.Printf("Checking %s ... ", aurora.Bold(definition.Name))

		if !definition.Status.Valid {
			fmt.Println(styles.StyleError("ERROR"))
			fmt.Println("  Error parsing template:")
			for _, err := range definition.Status.Errors {
				fmt.Printf("  ► %s\n\n", err)
			}
			fmt.Println()
			continue
		}

		if result.Valid {
			fmt.Println(styles.StyleOK("UP TO DATE"))
		} else {
			fmt.Println(styles.StyleWarning("OUT OF DATE"))
			fpatch, err := diff.CreateFilePatch(definition.Content, result.ActualContent)
			if err != nil {
				panic(err)
			}
			message := fmt.Sprintf("--- <generated> (source: %s)\n+++ %s", definition.Source, definition.Destination)
			patch := diff.NewPatch([]fdiff.FilePatch{fpatch}, message)
			lib.PrettyPrintDiff(patch.Format())
		}
		schemaResult := validator.ValidateSchema(definition)
		if !schemaResult.Valid {
			fmt.Println(styles.StyleWarning("Warning:"), aurora.Bold(definition.Name), "failed schema validation:")
			for _, err := range schemaResult.Errors {
				fmt.Printf("  ► %s\n", err)
			}
			fmt.Println()
		}
	}
}

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff workflows",
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

			if watch {
				lib.WatchWorkflows(fs, context, func() { diffWorkflows(cmd) })
			} else {
				diffWorkflows(cmd)
			}
		},
	}
	cmd.Flags().BoolP("watch", "w", false, "watch workflow templates for changes")
	return cmd
}

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

			if watch {
				lib.WatchWorkflows(fs, context, func() {
					err := lib.ValidateWorkflows(fs, context)
					if err != nil {
						fmt.Println(styles.StyleError(err.Error()))
					} else {
						fmt.Println(styles.StyleCommand("Workflows up to date"))
					}
				})
			} else {
				err := lib.ValidateWorkflows(fs, context)
				if err != nil {
					fmt.Println(styles.StyleError(err.Error()))
					os.Exit(1)
				} else {
					fmt.Println(styles.StyleCommand("Workflows up to date"))
				}
			}

		},
	}
	cmd.Flags().BoolP("watch", "w", false, "watch workflow templates for changes")
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

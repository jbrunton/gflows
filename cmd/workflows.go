package cmd

import (
	"fmt"
	"log"
	"os"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/jbrunton/jflows/diff"
	"github.com/jbrunton/jflows/styles"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"

	"github.com/fsnotify/fsnotify"
	"github.com/inancgumus/screen"
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

	definitions := lib.GetWorkflowDefinitions(fs, context)
	validator := lib.NewWorkflowValidator(fs)

	for _, definition := range definitions {
		result := validator.ValidateContent(definition)
		fmt.Printf("Checking %s ... ", aurora.Bold(definition.Name))
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
			ue := fdiff.NewUnifiedEncoder(os.Stdout, fdiff.DefaultContextLines)
			ue.Encode(patch)
		}
		schemaResult := validator.ValidateSchema(definition)
		if !schemaResult.Valid {
			fmt.Println(styles.StyleWarning("Warning:"), definition.Name, "failed schema validation:")
			for _, err := range schemaResult.Errors {
				fmt.Printf("  ► %s\n", err)
			}
		}
	}
}

func watchWorkflows(cmd *cobra.Command, onChange func()) {
	log.Println("Watching workflows")
	fs := lib.CreateOsFs()
	context, err := lib.GetContext(fs, cmd)
	if err != nil {
		fmt.Println(styles.StyleError(err.Error()))
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	screen.Clear()
	screen.MoveTopLeft()
	log.Println("Watching workflow templates")
	onChange()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					screen.Clear()
					screen.MoveTopLeft()
					log.Println("modified file:", event.Name)
					onChange()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	definitions := lib.GetWorkflowDefinitions(fs, context)

	for _, definition := range definitions {
		err = watcher.Add(definition.Source)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	<-done
}

// func newWatchCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "watch",
// 		Short: "Watch workflows",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			watchWorkflows(cmd)
// 		},
// 	}
// }

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff workflows",
		Run: func(cmd *cobra.Command, args []string) {
			watch, err := cmd.Flags().GetBool("watch")
			if err != nil {
				panic(err)
			}
			if watch {
				watchWorkflows(cmd, func() { diffWorkflows(cmd) })
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

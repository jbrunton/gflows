package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/diff"
	"github.com/jbrunton/jflows/styles"
	"github.com/olekukonko/tablewriter"
	dmp "github.com/sergi/go-diff/diffmatchpatch"

	"github.com/fsnotify/fsnotify"
	"github.com/inancgumus/screen"
	"github.com/jbrunton/jflows/lib"
	"github.com/spf13/cobra"
)

func getPatch(wanted, got string) (fdiff.FilePatch, error) {
	diffs := diff.Do(wanted, got)
	var chunks []fdiff.Chunk
	for _, d := range diffs {

		var op fdiff.Operation
		switch d.Type {
		case dmp.DiffEqual:
			op = fdiff.Equal
		case dmp.DiffDelete:
			op = fdiff.Delete
		case dmp.DiffInsert:
			op = fdiff.Add
		}

		chunks = append(chunks, &textChunk{d.Text, op})
	}

	return &textFilePatch{
		chunks: chunks,
		from: object.ChangeEntry{
			Name: wanted,
		},
		to: object.ChangeEntry{
			Name: got,
		},
	}, nil
}

// textFilePatch is an implementation of fdiff.FilePatch interface
type textFilePatch struct {
	chunks   []fdiff.Chunk
	from, to object.ChangeEntry
}

func (tf *textFilePatch) Files() (from fdiff.File, to fdiff.File) {
	f := &changeEntryWrapper{tf.from}
	t := &changeEntryWrapper{tf.to}

	if !f.Empty() {
		from = f
	}

	if !t.Empty() {
		to = t
	}

	return
}

func (t *textFilePatch) IsBinary() bool {
	return len(t.chunks) == 0
}

func (t *textFilePatch) Chunks() []fdiff.Chunk {
	return t.chunks
}

// changeEntryWrapper is an implementation of fdiff.File interface
type changeEntryWrapper struct {
	ce object.ChangeEntry
}

func (f *changeEntryWrapper) Hash() plumbing.Hash {
	if !f.ce.TreeEntry.Mode.IsFile() {
		return plumbing.ZeroHash
	}

	return f.ce.TreeEntry.Hash
}

func (f *changeEntryWrapper) Mode() filemode.FileMode {
	return f.ce.TreeEntry.Mode
}
func (f *changeEntryWrapper) Path() string {
	if !f.ce.TreeEntry.Mode.IsFile() {
		return ""
	}

	return f.ce.Name
}

func (f *changeEntryWrapper) Empty() bool {
	return !f.ce.TreeEntry.Mode.IsFile()
}

type Patch struct {
	message     string
	filePatches []fdiff.FilePatch
}

func (t *Patch) FilePatches() []fdiff.FilePatch {
	return t.filePatches
}

func (t *Patch) Message() string {
	return t.message
}

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
		if result.Valid {
			fmt.Printf("%s is up to date\n", definition.Name)
		} else {

			fpatch, err := getPatch(definition.Content, result.ActualContent)
			if err != nil {
				panic(err)
			}
			ue := fdiff.NewUnifiedEncoder(os.Stdout, fdiff.DefaultContextLines)
			message := fmt.Sprintf("--- %s (generated)\n+++ %s", definition.Source, definition.Destination)
			patch := &Patch{
				filePatches: []fdiff.FilePatch{fpatch},
				message:     message,
			}
			ue.Encode(patch)
			//fmt.Printf("%s it out of date. Diff:\n%s\n", definition.Name, dmp.DiffPrettyText(diffs))
		}
	}
}

type textChunk struct {
	content string
	op      fdiff.Operation
}

func (t *textChunk) Content() string {
	return t.content
}

func (t *textChunk) Type() fdiff.Operation {
	return t.op
}

func watchWorkflows(cmd *cobra.Command) {
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

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Watch workflows",
		Run: func(cmd *cobra.Command, args []string) {
			watchWorkflows(cmd)
		},
	}
}

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Diff workflows",
		Run: func(cmd *cobra.Command, args []string) {
			diffWorkflows(cmd)
		},
	}
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

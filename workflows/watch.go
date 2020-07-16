package workflows

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/inancgumus/screen"
	"github.com/jbrunton/gflows/di"
)

func getWatchFiles(container *di.Container) []string {
	context := container.Context()
	workflowManager := NewWorkflowManager(container)
	files := workflowManager.getWorkflowSources(context)
	for _, workflow := range getWorkflows(container) {
		files = append(files, workflow.path)
	}
	return files
}

// WatchWorkflows - watch workflow files and invoke onChange on any changes
func WatchWorkflows(container *di.Container, onChange func()) {
	log.Println("Watching workflows")

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

	screen.Clear()
	screen.MoveTopLeft()
	log.Println("Watching workflow templates")

	sources := getWatchFiles(container)

	for _, source := range sources {
		fmt.Println("  Watching", source)
		err = watcher.Add(source)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	onChange()

	<-done
}

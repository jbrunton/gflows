package lib

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/inancgumus/screen"
	"github.com/spf13/afero"
)

// WatchWorkflows - watch workflow files and invoke onChange on any changes
func WatchWorkflows(fs *afero.Afero, context *JFlowsContext, onChange func()) {
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

	screen.Clear()
	screen.MoveTopLeft()
	log.Println("Watching workflow templates")

	sources := getWorkflowSources(fs, context)

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

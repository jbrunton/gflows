package lib

import (
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

	definitions := GetWorkflowDefinitions(fs, context)

	for _, definition := range definitions {
		err = watcher.Add(definition.Source)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	<-done
}

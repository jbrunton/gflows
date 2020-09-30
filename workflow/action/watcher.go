package action

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/inancgumus/screen"
	"github.com/jbrunton/gflows/config"
)

type Watcher struct {
	manager *WorkflowManager
	context *config.GFlowsContext
}

func NewWatcher(manager *WorkflowManager, context *config.GFlowsContext) *Watcher {
	return &Watcher{
		manager: manager,
		context: context,
	}
}

func (watcher *Watcher) getWatchFiles() ([]string, error) {
	files, err := watcher.manager.GetObservableSources()
	if err != nil {
		return nil, err
	}
	for _, workflow := range watcher.manager.GetWorkflows() {
		files = append(files, workflow.Path)
	}
	return files, nil
}

// WatchWorkflows - watch workflow files and invoke onChange on any changes
func (watcher *Watcher) WatchWorkflows(onChange func()) {
	log.Println("Watching workflows")

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer fswatcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-fswatcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					screen.Clear()
					screen.MoveTopLeft()
					log.Println("modified file:", event.Name)
					onChange()
				}
			case err, ok := <-fswatcher.Errors:
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

	sources, err := watcher.getWatchFiles()
	if err != nil {
		panic(err)
	}

	for _, source := range sources {
		fmt.Println("  Watching", source)
		err = fswatcher.Add(source)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	onChange()

	<-done
}

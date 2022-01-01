package watch

import (
	"context"
	"log"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph"
	"github.com/fsnotify/fsnotify"
)

type Watch struct {
	Dir string
}

func (w *Watch) Worker(ctx context.Context, g *graph.Graph) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(w.Dir)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-watcher.Events:
			if !ok {
				continue
			}
			b, err := backup.LoadFromDirectory(w.Dir)
			if err != nil {
				log.Print(err)
			} else {
				g.SetRoot(b)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Print("fsnotify:", err)
		}
	}
}

package watch

import (
	"context"
	"log"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/tachiql"
	"github.com/fsnotify/fsnotify"
)

type Watch struct {
	Dir string
}

func (w *Watch) Worker(ctx context.Context, t *tachiql.Tachiql) error {
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
		case _, ok := <-watcher.Events:
			if !ok {
				continue
			}
			b, err := backup.LoadFromDirectory(w.Dir)
			if err != nil {
				log.Print(err)
			} else {
				t.SetBackup(b)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Print("fsnotify:", err)
		}
	}
}

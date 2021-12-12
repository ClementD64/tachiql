package server

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph"
	"github.com/clementd64/tachiql/pkg/graph/generated"
	"github.com/clementd64/tachiql/pkg/tachiql"
	"github.com/fsnotify/fsnotify"
)

type Config struct {
	BackupDir         string
	BackupFile        string
	Bind              string
	QueryPath         string
	ThumbnailDir      string
	ThumbnailUrl      string
	ThumbnailPath     string
	ThumbnailGetReq   func(manga *backup.Manga) (*http.Request, error)
	ThumbnailDownload func(manga *backup.Manga) ([]byte, string, error)
}

type Server struct {
	config Config
}

func New(config Config) *Server {
	return &Server{
		config: config,
	}
}

func (t *Server) Run() {
	indexer := tachiql.NewIndexer(tachiql.NewThumbnail(&tachiql.Thumbnail{
		Path:     t.config.ThumbnailDir,
		GetReq:   t.config.ThumbnailGetReq,
		Download: t.config.ThumbnailDownload,
	}), t.config.ThumbnailUrl)

	if t.config.BackupFile != "" {
		b, err := backup.LoadBackup(t.config.BackupFile)
		if err != nil {
			log.Fatal(err)
		}
		indexer.IndexBackup(b)
	} else if t.config.BackupDir != "" {
		b, err := backup.LoadFromDirectory(t.config.BackupDir)
		if err != nil {
			log.Fatal(err)
		}
		indexer.IndexBackup(b)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		go func() {
			for {
				select {
				case _, ok := <-watcher.Events:
					if !ok {
						return
					}
					b, err := backup.LoadFromDirectory(t.config.BackupDir)
					if err != nil {
						log.Print(err)
					}
					indexer.IndexBackup(b)
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Print("fsnotify:", err)
				}
			}
		}()

		err = watcher.Add(t.config.BackupDir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("No backup file specified")
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			Indexer: indexer,
		},
	}))

	http.Handle(t.config.QueryPath, srv)

	if t.config.ThumbnailPath != "" {
		http.Handle(t.config.ThumbnailPath, http.StripPrefix(t.config.ThumbnailPath, http.FileServer(http.Dir(t.config.ThumbnailDir))))
	}

	log.Fatal(http.ListenAndServe(t.config.Bind, nil))
}

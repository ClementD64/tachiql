package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/clementd64/tachiql/pkg/graph"
	"github.com/graphql-go/handler"
)

type Server struct {
	Addr                    string
	Path                    string
	ShutdownTimeout         time.Duration
	ShutdownTimeoutExceeded func(err error)
	ServeMux                *http.ServeMux
}

func (s *Server) Worker(ctx context.Context, t *graph.Graph) error {
	if s.ServeMux == nil {
		s.ServeMux = http.NewServeMux()
	}

	if s.Addr == "" {
		s.Addr = ":8080"
	}

	if s.Path == "" {
		s.Path = "/"
	}

	if s.ShutdownTimeout == 0 {
		s.ShutdownTimeout = 5 * time.Second
	}

	if s.ShutdownTimeoutExceeded == nil {
		s.ShutdownTimeoutExceeded = func(err error) {
			log.Fatal(err)
		}
	}

	s.ServeMux.Handle(s.Path, handler.New(&handler.Config{
		Schema: &t.Schema,
		RootObjectFn: func(ctx context.Context, r *http.Request) map[string]interface{} {
			return graph.ToMap(t.Root)
		},
	}))

	srv := &http.Server{
		Addr:    s.Addr,
		Handler: s.ServeMux,
	}

	go func() {
		<-ctx.Done()
		timeout, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(timeout); err != nil {
			s.ShutdownTimeoutExceeded(err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

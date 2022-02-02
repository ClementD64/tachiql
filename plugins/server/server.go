package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
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
	FastCGI                 bool
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

	if s.FastCGI {
		return s.serveFcgi(ctx)
	}

	return s.serveHTTP(ctx)
}

func (s *Server) serveHTTP(ctx context.Context) error {
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

func (s *Server) serveFcgi(ctx context.Context) error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	return fcgi.Serve(l, s.ServeMux)
}

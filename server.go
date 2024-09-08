package goproj

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	phttpServer *http.Server // same pointers?
}

func (pserver *Server) Run(port string, handler http.Handler) error {
	pserver.phttpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,          // same bit manipulations?
		ReadTimeout:    10 * time.Second, // no in std::something::something::chrono::chrono_literals?
		WriteTimeout:   10 * time.Second,
	}

	return pserver.phttpServer.ListenAndServe()
}

func (pserver *Server) Shutdown(ctx context.Context) error {
	return pserver.phttpServer.Shutdown(ctx)
}

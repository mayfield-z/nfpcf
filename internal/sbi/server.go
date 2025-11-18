package sbi

import (
	"fmt"
	"net/http"

	"github.com/free5gc/nfpcf/internal/sbi/processor"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	processor  *processor.Processor
	bindAddr   string
}

func NewServer(processor *processor.Processor, bindAddr string) *Server {
	s := &Server{
		processor: processor,
		bindAddr:  bindAddr,
		mux:       http.NewServeMux(),
	}

	s.setupRoutes()

	h2s := &http2.Server{}
	handler := h2c.NewHandler(s.mux, h2s)

	s.httpServer = &http.Server{
		Addr:    bindAddr,
		Handler: handler,
	}

	return s
}

func (s *Server) Run() error {
	fmt.Printf("NFPCF server listening on %s (HTTP/2 cleartext)\n", s.bindAddr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

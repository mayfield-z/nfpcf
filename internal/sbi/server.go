package sbi

import (
	"fmt"
	"net/http"

	"github.com/free5gc/nfpcf/internal/sbi/processor"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	processor  *processor.Processor
	bindAddr   string
}

func NewServer(processor *processor.Processor, bindAddr string) *Server {
	s := &Server{
		processor: processor,
		bindAddr:  bindAddr,
	}

	s.router = s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:    bindAddr,
		Handler: s.router,
	}

	return s
}

func (s *Server) Run() error {
	fmt.Printf("NFPCF server listening on %s\n", s.bindAddr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

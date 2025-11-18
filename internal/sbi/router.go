package sbi

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/nnrf-nfm/v1/nf-instances/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[NFPCF] %s %s %s from %s\n", r.Proto, r.Method, r.URL.Path, r.RemoteAddr)

		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid path", http.StatusNotFound)
			return
		}

		nfInstanceID := pathParts[3]

		switch r.Method {
		case http.MethodPut:
			s.handleRegisterNFInstance(w, r, nfInstanceID)
		case http.MethodGet:
			s.handleGetNFInstance(w, r, nfInstanceID)
		case http.MethodDelete:
			s.handleDeregisterNFInstance(w, r, nfInstanceID)
		case http.MethodPatch:
			s.handleUpdateNFInstance(w, r, nfInstanceID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.mux.HandleFunc("/nnrf-disc/v1/nf-instances", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[NFPCF] %s %s %s from %s\n", r.Proto, r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method == http.MethodGet {
			s.handleDiscoverNFInstances(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

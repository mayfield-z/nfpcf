package sbi

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) discoverNFInstances(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	s.processor.HandleNFDiscoveryRequest(c, queryParams)
}

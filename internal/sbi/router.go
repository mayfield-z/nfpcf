package sbi

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) setupRoutes() *gin.Engine {
	router := gin.Default()

	nfmGroup := router.Group("/nnrf-nfm/v1")
	{
		nfmGroup.PUT("/nf-instances/:nfInstanceID", s.registerNFInstance)
		nfmGroup.GET("/nf-instances/:nfInstanceID", s.getNFInstance)
		nfmGroup.DELETE("/nf-instances/:nfInstanceID", s.deregisterNFInstance)
		nfmGroup.PATCH("/nf-instances/:nfInstanceID", s.updateNFInstance)
	}

	discGroup := router.Group("/nnrf-disc/v1")
	{
		discGroup.GET("/nf-instances", s.discoverNFInstances)
	}

	return router
}

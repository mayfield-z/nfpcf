package sbi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/free5gc/openapi/models"
)

func (s *Server) registerNFInstance(c *gin.Context) {
	var nfProfile models.NrfNfManagementNfProfile

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "INVALID_MSG_FORMAT",
		})
		return
	}

	if err := json.Unmarshal(body, &nfProfile); err != nil {
		c.JSON(http.StatusBadRequest, models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "INVALID_MSG_FORMAT",
		})
		return
	}

	nfInstanceID := c.Param("nfInstanceID")
	nfProfile.NfInstanceId = nfInstanceID

	s.processor.HandleNFRegisterRequest(c, &nfProfile)
}

func (s *Server) getNFInstance(c *gin.Context) {
	nfInstanceID := c.Param("nfInstanceID")
	s.processor.HandleGetNFInstanceRequest(c, nfInstanceID)
}

func (s *Server) deregisterNFInstance(c *gin.Context) {
	nfInstanceID := c.Param("nfInstanceID")
	s.processor.HandleNFDeregisterRequest(c, nfInstanceID)
}

func (s *Server) updateNFInstance(c *gin.Context) {
	nfInstanceID := c.Param("nfInstanceID")

	patchJSON, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "INVALID_MSG_FORMAT",
		})
		return
	}

	s.processor.HandleUpdateNFInstanceRequest(c, nfInstanceID, patchJSON)
}

package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/free5gc/openapi/models"
)

func (p *Processor) HandleNFRegisterRequest(c *gin.Context, nfProfile *models.NrfNfManagementNfProfile) {
	profile, problemDetails, err := p.nrfClient.RegisterNF(c.Request.Context(), nfProfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: err.Error(),
		})
		return
	}

	if problemDetails != nil {
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	if profile != nil {
		c.JSON(http.StatusCreated, profile)
		return
	}

	c.Status(http.StatusOK)
}

func (p *Processor) HandleGetNFInstanceRequest(c *gin.Context, nfInstanceID string) {
	profile, problemDetails, err := p.nrfClient.GetNFInstance(c.Request.Context(), nfInstanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: err.Error(),
		})
		return
	}

	if problemDetails != nil {
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	if profile != nil {
		c.JSON(http.StatusOK, profile)
		return
	}

	c.JSON(http.StatusNotFound, models.ProblemDetails{
		Status: http.StatusNotFound,
		Cause:  "CONTEXT_NOT_FOUND",
	})
}

func (p *Processor) HandleNFDeregisterRequest(c *gin.Context, nfInstanceID string) {
	p.cache.Delete(nfInstanceID)

	problemDetails, err := p.nrfClient.DeregisterNF(c.Request.Context(), nfInstanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: err.Error(),
		})
		return
	}

	if problemDetails != nil {
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	c.Status(http.StatusNoContent)
}

func (p *Processor) HandleUpdateNFInstanceRequest(c *gin.Context, nfInstanceID string, patchJSON []byte) {
	p.cache.Delete(nfInstanceID)

	profile, err := p.nrfClient.UpdateNFInstance(c.Request.Context(), nfInstanceID, patchJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: err.Error(),
		})
		return
	}

	if profile != nil {
		c.JSON(http.StatusOK, profile)
		return
	}

	c.Status(http.StatusNoContent)
}

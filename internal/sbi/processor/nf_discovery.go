package processor

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/free5gc/openapi/models"
)

func (p *Processor) HandleNFDiscoveryRequest(c *gin.Context, queryParams url.Values) {
	cachedResults := p.cache.Search(queryParams)
	if len(cachedResults) > 0 {
		var nfDiscProfiles []models.NrfNfDiscoveryNfProfile
		for _, profile := range cachedResults {
			nfDiscProfiles = append(nfDiscProfiles, *profile)
		}

		response := &models.SearchResult{
			NfInstances: nfDiscProfiles,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	searchResult, problemDetails, err := p.nrfClient.DiscoverNF(c.Request.Context(), queryParams)
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

	if searchResult != nil && searchResult.NfInstances != nil {
		for i := range searchResult.NfInstances {
			p.cache.Put(&searchResult.NfInstances[i])
		}

		c.JSON(http.StatusOK, searchResult)
		return
	}

	c.JSON(http.StatusOK, models.SearchResult{
		NfInstances: make([]models.NrfNfDiscoveryNfProfile, 0),
	})
}

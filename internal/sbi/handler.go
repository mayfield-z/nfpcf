package sbi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/free5gc/openapi/models"
)

func (s *Server) handleRegisterNFInstance(w http.ResponseWriter, r *http.Request, nfInstanceID string) {
	var nfProfile models.NrfNfManagementNfProfile

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendProblemDetails(w, http.StatusBadRequest, "INVALID_MSG_FORMAT", "")
		return
	}

	if err := json.Unmarshal(body, &nfProfile); err != nil {
		sendProblemDetails(w, http.StatusBadRequest, "INVALID_MSG_FORMAT", "")
		return
	}

	nfProfile.NfInstanceId = nfInstanceID

	profile, problemDetails, err := s.processor.GetNRFClient().RegisterNF(r.Context(), &nfProfile)
	if err != nil {
		sendProblemDetails(w, http.StatusInternalServerError, "SYSTEM_FAILURE", err.Error())
		return
	}

	if problemDetails != nil {
		sendJSON(w, int(problemDetails.Status), problemDetails)
		return
	}

	if profile != nil {
		// Add Location header as per TS 29.510
		w.Header().Set("Location", fmt.Sprintf("/nnrf-nfm/v1/nf-instances/%s", profile.NfInstanceId))
		sendJSON(w, http.StatusCreated, profile)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetNFInstance(w http.ResponseWriter, r *http.Request, nfInstanceID string) {
	profile, problemDetails, err := s.processor.GetNRFClient().GetNFInstance(r.Context(), nfInstanceID)
	if err != nil {
		sendProblemDetails(w, http.StatusInternalServerError, "SYSTEM_FAILURE", err.Error())
		return
	}

	if problemDetails != nil {
		sendJSON(w, int(problemDetails.Status), problemDetails)
		return
	}

	if profile != nil {
		sendJSON(w, http.StatusOK, profile)
		return
	}

	sendProblemDetails(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "")
}

func (s *Server) handleDeregisterNFInstance(w http.ResponseWriter, r *http.Request, nfInstanceID string) {
	s.processor.GetCache().Delete(nfInstanceID)

	problemDetails, err := s.processor.GetNRFClient().DeregisterNF(r.Context(), nfInstanceID)
	if err != nil {
		sendProblemDetails(w, http.StatusInternalServerError, "SYSTEM_FAILURE", err.Error())
		return
	}

	if problemDetails != nil {
		sendJSON(w, int(problemDetails.Status), problemDetails)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpdateNFInstance(w http.ResponseWriter, r *http.Request, nfInstanceID string) {
	s.processor.GetCache().Delete(nfInstanceID)

	patchJSON, err := io.ReadAll(r.Body)
	if err != nil {
		sendProblemDetails(w, http.StatusBadRequest, "INVALID_MSG_FORMAT", "")
		return
	}

	profile, err := s.processor.GetNRFClient().UpdateNFInstance(r.Context(), nfInstanceID, patchJSON)
	if err != nil {
		sendProblemDetails(w, http.StatusInternalServerError, "SYSTEM_FAILURE", err.Error())
		return
	}

	if profile != nil {
		sendJSON(w, http.StatusOK, profile)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDiscoverNFInstances(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	targetNfType := queryParams.Get("target-nf-type")
	requesterNfType := queryParams.Get("requester-nf-type")

	if targetNfType == "" || requesterNfType == "" {
		sendProblemDetails(w, http.StatusBadRequest, "Loss mandatory parameter", "")
		return
	}

	// Check cache first
	if cachedResult, found := s.processor.GetCache().GetSearchResult(queryParams); found {
		fmt.Printf("[NFPCF] Cache HIT for discovery: target=%s, requester=%s\n", targetNfType, requesterNfType)
		sendJSON(w, http.StatusOK, cachedResult)
		return
	}

	// Cache miss, query NRF
	fmt.Printf("[NFPCF] Cache MISS for discovery: target=%s, requester=%s, querying NRF\n", targetNfType, requesterNfType)
	searchResult, problemDetails, err := s.processor.GetNRFClient().DiscoverNF(r.Context(), queryParams)
	if err != nil {
		sendProblemDetails(w, http.StatusInternalServerError, "SYSTEM_FAILURE", err.Error())
		return
	}

	if problemDetails != nil {
		sendJSON(w, int(problemDetails.Status), problemDetails)
		return
	}

	if searchResult != nil {
		// Cache the result
		s.processor.GetCache().SetSearchResult(queryParams, searchResult)
		sendJSON(w, http.StatusOK, searchResult)
		return
	}

	sendProblemDetails(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "")
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sendProblemDetails(w http.ResponseWriter, status int, cause string, detail string) {
	pd := models.ProblemDetails{
		Status: int32(status),
		Cause:  cause,
	}
	if detail != "" {
		pd.Detail = detail
	}
	sendJSON(w, status, pd)
}

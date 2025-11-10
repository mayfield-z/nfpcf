package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/free5gc/openapi/models"
)

type NRFClient struct {
	nrfURL     string
	httpClient *http.Client
}

func NewNRFClient(nrfURL string) *NRFClient {
	return &NRFClient{
		nrfURL:     nrfURL,
		httpClient: &http.Client{},
	}
}

func (c *NRFClient) RegisterNF(
	ctx context.Context,
	nfProfile *models.NrfNfManagementNfProfile,
) (*models.NrfNfManagementNfProfile, *models.ProblemDetails, error) {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.nrfURL, nfProfile.NfInstanceId)

	body, err := json.Marshal(nfProfile)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal profile: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var profile models.NrfNfManagementNfProfile
		if err := json.Unmarshal(respBody, &profile); err != nil {
			return nil, nil, fmt.Errorf("unmarshal response: %w", err)
		}
		return &profile, nil, nil
	}

	var problemDetails models.ProblemDetails
	if err := json.Unmarshal(respBody, &problemDetails); err == nil {
		return nil, &problemDetails, nil
	}

	return nil, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func (c *NRFClient) GetNFInstance(
	ctx context.Context,
	nfInstanceID string,
) (*models.NrfNfManagementNfProfile, *models.ProblemDetails, error) {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.nrfURL, nfInstanceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var profile models.NrfNfManagementNfProfile
		if err := json.Unmarshal(respBody, &profile); err != nil {
			return nil, nil, fmt.Errorf("unmarshal response: %w", err)
		}
		return &profile, nil, nil
	}

	var problemDetails models.ProblemDetails
	if err := json.Unmarshal(respBody, &problemDetails); err == nil {
		return nil, &problemDetails, nil
	}

	return nil, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func (c *NRFClient) DeregisterNF(ctx context.Context, nfInstanceID string) (*models.ProblemDetails, error) {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.nrfURL, nfInstanceID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var problemDetails models.ProblemDetails
	if err := json.Unmarshal(respBody, &problemDetails); err == nil {
		return &problemDetails, nil
	}

	return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func (c *NRFClient) DiscoverNF(
	ctx context.Context,
	queryParams url.Values,
) (*models.SearchResult, *models.ProblemDetails, error) {
	url := fmt.Sprintf("%s/nnrf-disc/v1/nf-instances?%s", c.nrfURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var searchResult models.SearchResult
		if err := json.Unmarshal(respBody, &searchResult); err != nil {
			return nil, nil, fmt.Errorf("unmarshal response: %w", err)
		}
		return &searchResult, nil, nil
	}

	var problemDetails models.ProblemDetails
	if err := json.Unmarshal(respBody, &problemDetails); err == nil {
		return nil, &problemDetails, nil
	}

	return nil, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

func (c *NRFClient) UpdateNFInstance(
	ctx context.Context,
	nfInstanceID string,
	patchData []byte,
) (*models.NrfNfManagementNfProfile, error) {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.nrfURL, nfInstanceID)

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(patchData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode == http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		var profile models.NrfNfManagementNfProfile
		if err := json.Unmarshal(respBody, &profile); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
		return &profile, nil
	}

	return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

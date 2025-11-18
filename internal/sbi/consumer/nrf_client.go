package consumer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/free5gc/openapi/models"
	"golang.org/x/net/http2"
)

type NRFClient struct {
	nrfURL     string
	httpClient *http.Client
}

func NewNRFClient(nrfURL string) *NRFClient {
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	return &NRFClient{
		nrfURL:     nrfURL,
		httpClient: &http.Client{Transport: transport},
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
		fmt.Printf("[NFPCF] RegisterNF: send request error: %v\n", err)
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[NFPCF] RegisterNF: got response status %d, proto %s\n", resp.StatusCode, resp.Proto)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[NFPCF] RegisterNF: read response error: %v\n", err)
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[NFPCF] RegisterNF: response body length: %d\n", len(respBody))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var profile models.NrfNfManagementNfProfile
		if err := json.Unmarshal(respBody, &profile); err != nil {
			fmt.Printf("[NFPCF] RegisterNF: unmarshal error: %v, body: %s\n", err, string(respBody))
			return nil, nil, fmt.Errorf("unmarshal response: %w", err)
		}
		fmt.Printf("[NFPCF] RegisterNF: successfully parsed profile %s\n", profile.NfInstanceId)
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

	fmt.Printf("[NFPCF] DiscoverNF: querying %s\n", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("[NFPCF] DiscoverNF: create request error: %v\n", err)
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[NFPCF] DiscoverNF: send request error: %v\n", err)
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[NFPCF] DiscoverNF: got response status %d, proto %s\n", resp.StatusCode, resp.Proto)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[NFPCF] DiscoverNF: read response error: %v\n", err)
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[NFPCF] DiscoverNF: response body length: %d\n", len(respBody))

	if resp.StatusCode == http.StatusOK {
		var searchResult models.SearchResult
		if err := json.Unmarshal(respBody, &searchResult); err != nil {
			fmt.Printf("[NFPCF] DiscoverNF: unmarshal error: %v, body: %s\n", err, string(respBody))
			return nil, nil, fmt.Errorf("unmarshal response: %w", err)
		}
		fmt.Printf("[NFPCF] DiscoverNF: found %d NF instances\n", len(searchResult.NfInstances))
		return &searchResult, nil, nil
	}

	var problemDetails models.ProblemDetails
	if err := json.Unmarshal(respBody, &problemDetails); err == nil {
		fmt.Printf("[NFPCF] DiscoverNF: got problem details: status=%d, cause=%s\n", problemDetails.Status, problemDetails.Cause)
		return nil, &problemDetails, nil
	}

	fmt.Printf("[NFPCF] DiscoverNF: unexpected status %d, body: %s\n", resp.StatusCode, string(respBody))
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

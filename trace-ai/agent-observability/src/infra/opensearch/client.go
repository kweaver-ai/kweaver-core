package opensearch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	auth       AuthConfig
	httpClient *http.Client
}

type AuthConfig struct {
	Enabled  bool
	Username string
	Password string
}

func New(baseURL string, auth AuthConfig, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		auth:    auth,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Search(ctx context.Context, index string, query []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/_search", c.baseURL, strings.TrimLeft(index, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("create opensearch search request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.auth.Enabled {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute opensearch search request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read opensearch search response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("opensearch search failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

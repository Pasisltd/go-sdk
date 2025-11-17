package pasis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

const (
	// DefaultBaseURL is the default API base URL.
	DefaultBaseURL = "http://localhost:8080/api"
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
	// TokenRefreshBuffer is the time before expiration to refresh the token.
	TokenRefreshBuffer = 5 * time.Minute
	// DefaultRetryCount is the default number of retries for failed requests.
	DefaultRetryCount = 3
	// DefaultRetryBackoff is the initial backoff delay between retries.
	DefaultRetryBackoff = 100 * time.Millisecond
)

// Client represents the Pasis SDK client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	appKey     string
	secretKey  string
	tokenCache TokenCache
	retryCount int

	mu             sync.RWMutex
	accessToken    string
	refreshToken   string
	tokenExpiresAt time.Time
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTokenCache sets a custom token cache implementation.
func WithTokenCache(cache TokenCache) ClientOption {
	return func(c *Client) {
		c.tokenCache = cache
	}
}

// WithRetryCount sets the number of retries for failed requests.
// Default is 3. Set to 0 to disable retries.
func WithRetryCount(count int) ClientOption {
	return func(c *Client) {
		if count < 0 {
			count = 0
		}
		c.retryCount = count
	}
}

// NewClient creates a new Pasis SDK client.
func NewClient(appKey, secretKey string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		appKey:     appKey,
		secretKey:  secretKey,
		tokenCache: NewInMemoryTokenCache(),
		retryCount: DefaultRetryCount,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// HTTPClient returns the HTTP client.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// AccessToken returns the current access token.
func (c *Client) AccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

// setTokens sets the access token, refresh token, and expiration time.
func (c *Client) setTokens(token, refreshToken string, expiresAt time.Time) {
	c.mu.Lock()
	c.accessToken = token
	c.refreshToken = refreshToken
	c.tokenExpiresAt = expiresAt
	c.mu.Unlock()
	_ = c.tokenCache.Set(token, refreshToken, expiresAt)
}

// doRequest executes a request to the API and handles authentication.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any, result any) error {
	if err := c.EnsureToken(ctx); err != nil {
		return fmt.Errorf("failed to ensure token: %w", err)
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	if endpoint != "" && endpoint[0] == '/' {
		endpoint = endpoint[1:]
	}
	baseURL.Path = path.Join(baseURL.Path, endpoint)
	reqURL := baseURL.String()

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	var lastErr error
	backoff := DefaultRetryBackoff
	maxRetries := max(c.retryCount, 0)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		c.mu.RLock()
		accessToken := c.accessToken
		c.mu.RUnlock()
		if accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt < maxRetries {
				continue
			}
			return lastErr
		}

		statusCode := resp.StatusCode
		if statusCode >= http.StatusInternalServerError {
			err := parseErrorResponse(resp)
			resp.Body.Close()
			lastErr = err
			if attempt < maxRetries {
				continue
			}
			return lastErr
		}

		if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
			err := parseErrorResponse(resp)
			resp.Body.Close()
			return err
		}

		if result != nil {
			var successResp SuccessResponse
			if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to decode response: %w", err)
			}

			if successResp.Data != nil {
				dataBytes, err := json.Marshal(successResp.Data)
				if err != nil {
					resp.Body.Close()
					return fmt.Errorf("failed to marshal response data: %w", err)
				}
				if err := json.Unmarshal(dataBytes, result); err != nil {
					resp.Body.Close()
					return fmt.Errorf("failed to unmarshal response data: %w", err)
				}
			}
		}

		resp.Body.Close()
		return nil
	}

	return lastErr
}

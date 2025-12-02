package pasis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

// authenticate authenticates the application and retrieves access tokens.
func (c *Client) authenticate(ctx context.Context) error {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	baseURL.Path = path.Join(baseURL.Path, "auth", "app")
	reqURL := baseURL.String()

	reqBody := AppAuthRequest{
		AppKey:    c.appKey,
		SecretKey: c.secretKey,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &AuthError{Message: "failed to authenticate", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return parseErrorResponse(resp)
	}

	var successResp SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	authDataBytes, err := json.Marshal(successResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	var authResp AppAuthResponse
	if err := json.Unmarshal(authDataBytes, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	c.setTokens(authResp.AccessToken, authResp.RefreshToken, expiresAt)

	return nil
}

// refreshAccessToken refreshes the access token using the refresh token.
func (c *Client) refreshAccessToken(ctx context.Context) error {
	c.mu.RLock()
	refreshToken := c.refreshToken
	c.mu.RUnlock()

	if refreshToken == "" {
		return &AuthError{Message: "no refresh token available"}
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	baseURL.Path = path.Join(baseURL.Path, "auth", "refresh")
	reqURL := baseURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &AuthError{Message: "failed to refresh token", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return parseErrorResponse(resp)
	}

	var successResp SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	authDataBytes, err := json.Marshal(successResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	var authResp AppAuthResponse
	if err := json.Unmarshal(authDataBytes, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	c.setTokens(authResp.AccessToken, authResp.RefreshToken, expiresAt)

	return nil
}

// EnsureToken ensures a valid access token is available.
// It checks the cache first, then validates expiration, and refreshes if needed.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.RLock()
	expiresAt := c.tokenExpiresAt
	c.mu.RUnlock()

	if token, refreshToken, cachedExpiresAt, err := c.tokenCache.Get(); err == nil && token != "" {
		c.mu.Lock()
		c.accessToken = token
		c.refreshToken = refreshToken
		c.tokenExpiresAt = cachedExpiresAt
		c.mu.Unlock()
		expiresAt = cachedExpiresAt
	}

	now := time.Now()
	if expiresAt.IsZero() || now.Add(TokenRefreshBuffer).After(expiresAt) {
		c.mu.RLock()
		hasRefreshToken := c.refreshToken != ""
		c.mu.RUnlock()

		if hasRefreshToken {
			if err := c.refreshAccessToken(ctx); err != nil {
				return c.authenticate(ctx)
			}
		} else {
			return c.authenticate(ctx)
		}
	}

	return nil
}

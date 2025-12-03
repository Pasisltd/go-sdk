package pasis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// authenticate authenticates the application and retrieves access tokens.
func (c *Client) authenticate(ctx context.Context) error {
	urlStr := c.baseURL + "/auth/app"
	reqBody, _ := json.Marshal(AppAuthRequest{AppKey: c.appKey, SecretKey: c.secretKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &AuthError{Message: "failed to authenticate", Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	var res SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	authResp, ok := res.Data.(AppAuthResponse)
	if !ok {
		return fmt.Errorf("invalid auth response")
	}
	c.setTokens(authResp.AccessToken, authResp.RefreshToken, time.Now().Add(time.Duration(authResp.ExpiresIn)*time.Second))
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

	urlStr := c.baseURL + "/auth/refresh"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &AuthError{Message: "failed to refresh token", Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	var res SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	authResp, ok := res.Data.(AppAuthResponse)
	if !ok {
		return fmt.Errorf("invalid auth response")
	}
	c.setTokens(authResp.AccessToken, authResp.RefreshToken, time.Now().Add(time.Duration(authResp.ExpiresIn)*time.Second))
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

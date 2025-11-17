package pasis

import (
	"context"
	"fmt"
)

// GetMerchantProfile retrieves the merchant profile for the authenticated user.
func (c *Client) GetMerchantProfile(ctx context.Context) (*MerchantProfile, error) {
	var profile MerchantProfile
	if err := c.doRequest(ctx, "GET", "/user/me", nil, &profile); err != nil {
		return nil, fmt.Errorf("failed to get merchant profile: %w", err)
	}
	return &profile, nil
}


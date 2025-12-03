package pasis

import (
	"context"
	"fmt"
)

// GetMerchantProfile retrieves the merchant profile for the authenticated user.
func (c *Client) GetMerchantProfile(ctx context.Context) (*MerchantProfile, error) {
	var res SuccessResponse
	if err := c.doRequest(ctx, "GET", "/user/me", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to get merchant profile: %w", err)
	}

	profile, ok := res.Data.(MerchantProfile)
	if !ok {
		return nil, fmt.Errorf("failed to cast data to MerchantProfile")
	}

	return &profile, nil
}

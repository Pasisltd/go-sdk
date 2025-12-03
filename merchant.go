package pasis

import (
	"context"
	"fmt"
)

// GetMerchantProfile retrieves the merchant profile for the authenticated user.
func (c *Client) GetMerchantProfile(ctx context.Context) (*MerchantProfile, error) {
	var res SuccessResponse[MerchantProfile]
	if err := c.doRequest(ctx, "GET", "/user/me", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to get merchant profile: %w", err)
	}

	return &res.Data, nil
}

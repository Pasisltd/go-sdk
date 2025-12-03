package pasis

import (
	"context"
	"fmt"
)

// GetWallet retrieves wallet details for the authenticated merchant.
func (c *Client) GetWallet(ctx context.Context) (*Wallet, error) {
	var res SuccessResponse[Wallet]
	if err := c.doRequest(ctx, "GET", "/wallet", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &res.Data, nil
}

// Deposit initiates a deposit transaction to the wallet.
func (c *Client) Deposit(ctx context.Context, req *DepositRequest) (*WalletTransaction, error) {
	if req == nil {
		return nil, fmt.Errorf("deposit request cannot be nil")
	}

	var res SuccessResponse[WalletTransaction]
	if err := c.doRequest(ctx, "POST", "/wallet/deposit", req, &res); err != nil {
		return nil, fmt.Errorf("failed to deposit: %w", err)
	}

	return &res.Data, nil
}

// Withdraw initiates a withdrawal transaction from the wallet.
func (c *Client) Withdraw(ctx context.Context, req *WithdrawRequest) (*WalletTransaction, error) {
	if req == nil {
		return nil, fmt.Errorf("withdraw request cannot be nil")
	}

	var res SuccessResponse[WalletTransaction]
	if err := c.doRequest(ctx, "POST", "/wallet/withdraw", req, &res); err != nil {
		return nil, fmt.Errorf("failed to withdraw: %w", err)
	}

	return &res.Data, nil
}

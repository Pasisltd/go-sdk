package pasis

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListTransactions retrieves a paginated list of wallet transactions.
func (c *Client) ListTransactions(ctx context.Context, page, size int) (*TransactionsResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		params.Set("size", strconv.Itoa(size))
	}

	endpoint := "/wallet/transactions"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	var resp TransactionsResponse
	if err := c.doRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	return &resp, nil
}

// GetTransaction retrieves details of a specific transaction by ID.
func (c *Client) GetTransaction(ctx context.Context, id string) (*WalletTransaction, error) {
	if id == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/wallet/transactions/%s", url.PathEscape(id))
	var transaction WalletTransaction
	if err := c.doRequest(ctx, "GET", endpoint, nil, &transaction); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &transaction, nil
}

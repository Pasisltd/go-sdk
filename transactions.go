package pasis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListTransactions retrieves a paginated list of wallet transactions.
func (c *Client) ListTransactions(ctx context.Context, page, perPage int) (*TransactionsResponse, error) {
	q := make(url.Values)

	if page > 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Add("per_page", strconv.Itoa(perPage))
	}

	u := "/wallet/transactions"

	var res TransactionsResponse
	if err := c.doRequest(ctx, http.MethodGet, u, q, &res); err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	return &res, nil
}

// GetTransaction retrieves details of a specific transaction by ID.
func (c *Client) GetTransaction(ctx context.Context, id string) (*WalletTransaction, error) {
	if id == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/wallet/transactions/%s", url.PathEscape(id))
	var res SuccessResponse[WalletTransaction]
	if err := c.doRequest(ctx, "GET", endpoint, nil, &res); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &res.Data, nil
}

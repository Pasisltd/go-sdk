package pasis

import "time"

// AppAuthRequest represents the request payload for application authentication.
type AppAuthRequest struct {
	AppKey    string `json:"app_key"`
	SecretKey string `json:"secret_key"`
}

// AppAuthResponse represents the response from application authentication.
type AppAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Token expiration time in seconds
}

// DepositRequest represents a request to deposit funds into a wallet.
type DepositRequest struct {
	Amount      string            `json:"amount"`       // Required: Amount as string to preserve decimal precision
	Currency    string            `json:"currency"`    // Required: Currency code
	Provider    string            `json:"provider"`    // Required: Payment provider
	Region      string            `json:"region"`       // Required: Region code
	PhoneNumber string            `json:"phone_number,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// WithdrawRequest represents a request to withdraw funds from a wallet.
type WithdrawRequest struct {
	Amount      string            `json:"amount"`       // Required: Amount as string to preserve decimal precision
	Currency    string            `json:"currency"`    // Required: Currency code
	Provider    string            `json:"provider"`    // Required: Payment provider
	Region      string            `json:"region"`       // Required: Region code
	PhoneNumber string            `json:"phone_number,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Fees represents transaction fees breakdown.
type Fees struct {
	Provider string `json:"provider"`
	System   string `json:"system"`
	Total    string `json:"total"`
}

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
	TransactionStatusUnknown   TransactionStatus = "UNKNOWN"
)

// TransactionType represents the type of a transaction.
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeFee        TransactionType = "FEE"
)

// WalletTransaction represents a wallet transaction.
type WalletTransaction struct {
	ID               string            `json:"id"`
	WalletID         string            `json:"wallet_id"`
	Amount           string            `json:"amount"` // Stored as string to preserve decimal precision
	Currency         string            `json:"currency"`
	Type             TransactionType   `json:"type"`
	Status           TransactionStatus `json:"status"`
	Description     string            `json:"description,omitempty"`
	Provider         string            `json:"provider,omitempty"`
	ProviderReference string          `json:"provider_reference,omitempty"`
	Fees             *Fees             `json:"fees,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

// PaginationMeta represents pagination metadata.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// SuccessResponse represents a successful API response.
type SuccessResponse struct {
	Data       interface{}     `json:"data"`
	Message    string          `json:"message,omitempty"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// ErrorResponse represents an error API response.
type ErrorResponse struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// TransactionsResponse represents a paginated list of transactions.
type TransactionsResponse struct {
	Data       []WalletTransaction `json:"data"`
	Pagination *PaginationMeta    `json:"pagination,omitempty"`
}

// Wallet represents wallet details.
// Note: The exact structure is not fully defined in swagger, so this is inferred.
// The API returns SuccessResponse with wallet data in the data field.
type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Balance   string    `json:"balance,omitempty"` // Balance as string to preserve decimal precision
	Currency  string    `json:"currency,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// MerchantProfile represents merchant/user profile information.
// Note: The exact structure is not fully defined in swagger, so this is inferred.
// The API returns SuccessResponse with user data in the data field.
type MerchantProfile struct {
	ID          string    `json:"id"`
	Email       string    `json:"email,omitempty"`
	FirstName   string    `json:"first_name,omitempty"`
	LastName    string    `json:"last_name,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	CountryCode string    `json:"country_code,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}


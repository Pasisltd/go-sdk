# Pasis Go SDK

A lightweight Go SDK for integrating with the Pasis payment platform. This SDK provides a simple, intuitive interface for merchants to interact with Pasis APIs without writing HTTP requests manually.

## Installation

```bash
go get github.com/pasisltd/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/pasisltd/go-sdk"
)

func main() {
    // Create a new client with your app credentials
    client := pasis.NewClient("your-app-key", "your-secret-key")
    
    ctx := context.Background()
    
    // Get wallet details
    wallet, err := client.GetWallet(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Wallet ID: %s\n", wallet.ID)
    
    // Get merchant profile
    profile, err := client.GetMerchantProfile(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Merchant: %s %s\n", profile.FirstName, profile.LastName)
}
```

## Configuration

### Client Options

The SDK provides several options to customize the client behavior:

#### Custom Base URL

```go
client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithBaseURL("https://api.pasis.com/api"),
)
```

#### Custom HTTP Client

```go
customClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
}

client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithHTTPClient(customClient),
)
```

#### Retry Configuration

The SDK automatically retries failed requests (network errors and 5xx server errors) up to 3 times by default with exponential backoff. You can customize the retry count:

```go
// Disable retries
client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithRetryCount(0),
)

// Set custom retry count (e.g., 5 retries)
client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithRetryCount(5),
)
```

**Note:** Retries only occur for:
- Network errors (connection failures, timeouts)
- Server errors (5xx status codes)

Client errors (4xx status codes) are not retried as they indicate invalid requests.

#### Custom Token Cache

For applications that need to share tokens across multiple instances or persist tokens, you can provide a custom cache implementation:

```go
// Implement the TokenCache interface
type RedisTokenCache struct {
    client *redis.Client
}

func (r *RedisTokenCache) Get() (string, string, time.Time, error) {
    // Retrieve tokens from Redis
    // ...
}

func (r *RedisTokenCache) Set(token, refreshToken string, expiresAt time.Time) error {
    // Store tokens in Redis
    // ...
}

func (r *RedisTokenCache) Clear() error {
    // Clear tokens from Redis
    // ...
}

// Use the custom cache
redisCache := &RedisTokenCache{client: redisClient}
client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithTokenCache(redisCache),
)
```

## API Reference

### Authentication

Authentication is handled automatically by the SDK. When you create a client, it will authenticate using your app key and secret key. Tokens are automatically refreshed when they expire.

### Wallet Operations

#### Get Wallet Details

```go
wallet, err := client.GetWallet(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Balance: %s %s\n", wallet.Balance, wallet.Currency)
```

#### Deposit Funds

```go
depositReq := &pasis.DepositRequest{
    Amount:   "100.00",
    Currency: "USD",
    Provider: "mobile_money",
    Region:   "US",
    PhoneNumber: "+1234567890",
    Metadata: map[string]string{
        "reference": "order-123",
    },
}

transaction, err := client.Deposit(ctx, depositReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Transaction ID: %s\n", transaction.ID)
fmt.Printf("Status: %s\n", transaction.Status)
```

#### Withdraw Funds

```go
withdrawReq := &pasis.WithdrawRequest{
    Amount:   "50.00",
    Currency: "USD",
    Provider: "bank_transfer",
    Region:   "US",
    PhoneNumber: "+1234567890",
    Metadata: map[string]string{
        "reference": "payout-456",
    },
}

transaction, err := client.Withdraw(ctx, withdrawReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Transaction ID: %s\n", transaction.ID)
```

### Transaction Operations

#### List Transactions

```go
// List first page with 20 items
transactions, err := client.ListTransactions(ctx, 1, 20)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total transactions: %d\n", transactions.Pagination.Total)
for _, tx := range transactions.Data {
    fmt.Printf("Transaction %s: %s %s (%s)\n", 
        tx.ID, tx.Amount, tx.Currency, tx.Status)
}
```

#### Get Transaction Details

```go
transaction, err := client.GetTransaction(ctx, "txn-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Amount: %s\n", transaction.Amount)
fmt.Printf("Type: %s\n", transaction.Type)
fmt.Printf("Status: %s\n", transaction.Status)
if transaction.Fees != nil {
    fmt.Printf("Fees: %s\n", transaction.Fees.Total)
}
```

### Merchant Profile

#### Get Merchant Profile

```go
profile, err := client.GetMerchantProfile(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s %s\n", profile.FirstName, profile.LastName)
fmt.Printf("Email: %s\n", profile.Email)
fmt.Printf("Phone: %s\n", profile.PhoneNumber)
```

## Error Handling

The SDK provides specific error types to help you handle different error scenarios:

```go
transaction, err := client.Deposit(ctx, depositReq)
if err != nil {
    // Check for authentication errors
    var authErr *pasis.AuthError
    if errors.As(err, &authErr) {
        log.Printf("Authentication failed: %v", authErr)
        // Handle authentication failure
        return
    }
    
    // Check for validation errors
    var valErr *pasis.ValidationError
    if errors.As(err, &valErr) {
        log.Printf("Validation failed: %v", valErr)
        // Handle validation failure
        return
    }
    
    // Check for API errors
    var apiErr *pasis.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API error (status %d): %s", apiErr.StatusCode, apiErr.Message)
        if len(apiErr.Errors) > 0 {
            for _, e := range apiErr.Errors {
                log.Printf("  - %s", e)
            }
        }
        // Handle API error
        return
    }
    
    // Generic error
    log.Printf("Unexpected error: %v", err)
}
```

## Context Support

All SDK methods accept a `context.Context` parameter, allowing you to:

- Set timeouts for requests
- Cancel long-running operations
- Pass request-scoped values

```go
// Set a timeout for the request
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

wallet, err := client.GetWallet(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Request timed out")
    }
    log.Fatal(err)
}
```

## Best Practices

### 1. Reuse Client Instances

Create a single client instance and reuse it across your application. The client is thread-safe and handles token management efficiently.

```go
// Good: Create once, reuse everywhere
var client *pasis.Client

func init() {
    client = pasis.NewClient("app-key", "secret-key")
}

func handleRequest() {
    wallet, err := client.GetWallet(ctx)
    // ...
}
```

### 2. Use Context for Timeouts

Always use context with timeouts for production applications to prevent hanging requests.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

transaction, err := client.Deposit(ctx, req)
```

### 3. Implement Custom Token Cache for Distributed Systems

If you're running multiple instances of your application, implement a shared token cache (e.g., Redis) to avoid unnecessary authentication requests.

### 4. Handle Errors Appropriately

Use error type assertions to handle different error scenarios appropriately.

```go
if err != nil {
    var apiErr *pasis.APIError
    if errors.As(err, &apiErr) {
        // Handle API errors
    }
    // ...
}
```

### 5. Use Custom HTTP Client for Production

Configure your HTTP client with appropriate timeouts, connection pooling, and retry logic.

```go
client := pasis.NewClient(
    "app-key",
    "secret-key",
    pasis.WithHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }),
)
```

## Data Models

### Transaction Status

- `TransactionStatusPending` - Transaction is pending
- `TransactionStatusCompleted` - Transaction completed successfully
- `TransactionStatusFailed` - Transaction failed
- `TransactionStatusCancelled` - Transaction was cancelled
- `TransactionStatusUnknown` - Unknown status

### Transaction Type

- `TransactionTypeDeposit` - Deposit transaction
- `TransactionTypeWithdrawal` - Withdrawal transaction
- `TransactionTypeTransfer` - Transfer transaction
- `TransactionTypeFee` - Fee transaction

## License

This SDK is provided as-is. Please refer to your Pasis service agreement for usage terms.

## Support

For API documentation and support, please visit the [Pasis documentation](https://docs.pasis.com) or contact support.


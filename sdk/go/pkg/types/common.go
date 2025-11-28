package types

import (
	"cosmossdk.io/math"
)

// TxResponse represents a transaction response
type TxResponse struct {
	TxHash    string `json:"tx_hash"`
	Height    int64  `json:"height"`
	Code      uint32 `json:"code"`
	Data      string `json:"data"`
	RawLog    string `json:"raw_log"`
	GasWanted int64  `json:"gas_wanted"`
	GasUsed   int64  `json:"gas_used"`
}

// QueryOptions contains common query options
type QueryOptions struct {
	Height     int64
	Prove      bool
	Pagination *PageRequest
}

// PageRequest contains pagination parameters
type PageRequest struct {
	Key        []byte
	Offset     uint64
	Limit      uint64
	CountTotal bool
	Reverse    bool
}

// PageResponse contains pagination response
type PageResponse struct {
	NextKey []byte
	Total   uint64
}

// Coin represents a token amount
type Coin struct {
	Denom  string
	Amount math.Int
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// Local stub types for queries until proto is properly defined
type QueryServer interface{}

// Stub query request/response types
type (
	QueryParamsRequest struct{}

	QueryParamsResponse struct {
		Params *Params
	}

	Params struct {
		MaxMempoolSize      uint64
		MaxTxAge            uint64
		GasLimitMultiplier  string
		EnableBatching      bool
		BatchSize           uint64
		EnableSimulation    bool
		EnableCensorResist  bool
		EnablePrivacyChecks bool
		MinGasPrice         string
	}

	QueryValidateTransactionRequest struct {
		Sender    string
		Recipient string
		Amount    string
		Data      []byte
		Nonce     uint64
	}

	QueryValidateTransactionResponse struct {
		Valid             bool
		GasEstimate       uint64
		Error             string
		SufficientBalance bool
	}

	QueryMempoolRequest struct{}

	QueryMempoolResponse struct {
		Transactions []*Transaction
		Count        uint64
	}

	Transaction struct {
		Sender    string
		Recipient string
		Amount    string
		Data      []byte
		Nonce     uint64
		Signature []byte
	}

	QueryEstimateGasRequest struct {
		Sender    string
		Recipient string
		Amount    string
		Data      []byte
	}

	QueryEstimateGasResponse struct {
		GasEstimate uint64
		GasLimit    uint64
	}

	QueryGetNonceRequest struct {
		Address string
	}

	QueryGetNonceResponse struct {
		Nonce uint64
	}
)

type queryServer struct {
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) QueryServer {
	return &queryServer{keeper: keeper}
}

// Params queries module parameters
func (qs queryServer) Params(goCtx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	// Return stub params - actual params are stored in proto format
	// TODO: Map proto Params fields to stub Params when proto is finalized
	return &QueryParamsResponse{
		Params: &Params{
			MaxMempoolSize:      1000,
			MaxTxAge:            3600,
			GasLimitMultiplier:  "1.5",
			EnableBatching:      true,
			BatchSize:           100,
			EnableSimulation:    true,
			EnableCensorResist:  true,
			EnablePrivacyChecks: true,
			MinGasPrice:         "0.001",
		},
	}, nil
}

// ValidateTransaction queries transaction validation status
func (qs queryServer) ValidateTransaction(goCtx context.Context, req *QueryValidateTransactionRequest) (*QueryValidateTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx := types.Transaction{
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Amount:    req.Amount,
		Data:      req.Data,
		Nonce:     req.Nonce,
	}

	valid, err := qs.keeper.ValidateTransaction(ctx, tx)

	response := &QueryValidateTransactionResponse{
		Valid:       valid,
		GasEstimate: qs.keeper.EstimateGas(ctx, tx),
	}

	if err != nil {
		response.Error = err.Error()
	}

	// Check balance
	hasBalance := qs.keeper.CheckSufficientBalance(ctx, tx.Sender, tx.Amount)
	response.SufficientBalance = hasBalance

	return response, nil
}

// Mempool queries current mempool transactions
func (qs queryServer) Mempool(goCtx context.Context, req *QueryMempoolRequest) (*QueryMempoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	transactions := qs.keeper.GetMempoolTransactions(ctx)

	// Convert to stub format
	stubTxs := make([]*Transaction, len(transactions))
	for i, tx := range transactions {
		stubTxs[i] = &Transaction{
			Sender:    tx.Sender,
			Recipient: tx.Recipient,
			Amount:    tx.Amount,
			Data:      tx.Data,
			Nonce:     tx.Nonce,
			Signature: tx.Signature,
		}
	}

	return &QueryMempoolResponse{
		Transactions: stubTxs,
		Count:        uint64(len(transactions)),
	}, nil
}

// EstimateGas queries gas estimation for a transaction
func (qs queryServer) EstimateGas(goCtx context.Context, req *QueryEstimateGasRequest) (*QueryEstimateGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx := types.Transaction{
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Amount:    req.Amount,
		Data:      req.Data,
	}

	gasEstimate := qs.keeper.EstimateGas(ctx, tx)

	return &QueryEstimateGasResponse{
		GasEstimate: gasEstimate,
		GasLimit:    gasEstimate + (gasEstimate / 10), // Add 10% buffer
	}, nil
}

// GetNonce queries the current nonce for an address
func (qs queryServer) GetNonce(goCtx context.Context, req *QueryGetNonceRequest) (*QueryGetNonceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	nonce := qs.keeper.GetNonce(ctx, req.Address)

	return &QueryGetNonceResponse{
		Nonce: nonce,
	}, nil
}

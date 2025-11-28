package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// Local stub types for messages until proto is properly defined
type MsgServer interface{}

// Stub message types
type (
	MsgPreValidateTransaction struct {
		Sender    string
		Recipient string
		Amount    string
		Data      []byte
		Nonce     uint64
		Signature []byte
	}

	MsgPreValidateTransactionResponse struct {
		Valid       bool
		GasEstimate uint64
		TxHash      string
	}

	MsgBatchPreValidate struct {
		Transactions []*MsgPreValidateTransaction
	}

	MsgBatchPreValidateResponse struct {
		Results []*ValidationResult
	}

	MsgUpdateParams struct {
		Params *types.Params
	}

	MsgUpdateParamsResponse struct{}

	MsgSimulateTransaction struct {
		Sender    string
		Recipient string
		Amount    string
		Data      []byte
		Nonce     uint64
	}

	MsgSimulateTransactionResponse struct {
		Success  bool
		GasUsed  uint64
		GasLimit uint64
		Error    string
	}

	MsgCancelPreValidation struct {
		Sender string
		TxHash string
	}

	MsgCancelPreValidationResponse struct{}

	ValidationResult struct {
		TxHash      string
		Valid       bool
		GasEstimate uint64
		Error       string
	}
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) MsgServer {
	return &msgServer{keeper: keeper}
}

// PreValidateTransaction pre-validates a transaction
func (ms msgServer) PreValidateTransaction(goCtx context.Context, msg *MsgPreValidateTransaction) (*MsgPreValidateTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert proto message to internal transaction type
	tx := types.Transaction{
		Sender:    msg.Sender,
		Recipient: msg.Recipient,
		Amount:    msg.Amount,
		Data:      msg.Data,
		Nonce:     msg.Nonce,
		Signature: msg.Signature,
	}

	// Validate transaction
	valid, err := ms.keeper.ValidateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("transaction validation failed")
	}

	// Check sufficient balance
	if !ms.keeper.CheckSufficientBalance(ctx, tx.Sender, tx.Amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Validate signature
	if !ms.keeper.ValidateSignature(ctx, tx.Sender, []byte(tx.Sender+tx.Recipient+tx.Amount), tx.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Estimate gas
	gasEstimate := ms.keeper.EstimateGas(ctx, tx)

	// Add to mempool
	if err := ms.keeper.AddToMempool(ctx, tx); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"prevalidate_transaction",
		sdk.NewAttribute("sender", tx.Sender),
		sdk.NewAttribute("recipient", tx.Recipient),
		sdk.NewAttribute("amount", tx.Amount),
		sdk.NewAttribute("gas_estimate", fmt.Sprintf("%d", gasEstimate)),
	))

	return &MsgPreValidateTransactionResponse{
		Valid:       true,
		GasEstimate: gasEstimate,
		TxHash:      ms.keeper.GetTransactionHash(tx),
	}, nil
}

// BatchPreValidate pre-validates multiple transactions
func (ms msgServer) BatchPreValidate(goCtx context.Context, msg *MsgBatchPreValidate) (*MsgBatchPreValidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	results := make([]*ValidationResult, len(msg.Transactions))

	for i, txMsg := range msg.Transactions {
		tx := types.Transaction{
			Sender:    txMsg.Sender,
			Recipient: txMsg.Recipient,
			Amount:    txMsg.Amount,
			Data:      txMsg.Data,
			Nonce:     txMsg.Nonce,
			Signature: txMsg.Signature,
		}

		valid, err := ms.keeper.ValidateTransaction(ctx, tx)
		result := &ValidationResult{
			TxHash:      ms.keeper.GetTransactionHash(tx),
			Valid:       valid,
			GasEstimate: ms.keeper.EstimateGas(ctx, tx),
		}

		if err != nil {
			result.Valid = false
			result.Error = err.Error()
		} else if valid {
			// Add to mempool if valid
			if err := ms.keeper.AddToMempool(ctx, tx); err != nil {
				result.Error = err.Error()
			}
		}

		results[i] = result
	}

	return &MsgBatchPreValidateResponse{
		Results: results,
	}, nil
}

// UpdateParams updates module parameters
func (ms msgServer) UpdateParams(goCtx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// In production, verify authority
	params := msg.Params

	if err := ms.keeper.SetParams(ctx, params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"params_updated",
		sdk.NewAttribute("module", "prevalidation"),
	))

	return &MsgUpdateParamsResponse{}, nil
}

// SimulateTransaction simulates a transaction without executing it
func (ms msgServer) SimulateTransaction(goCtx context.Context, msg *MsgSimulateTransaction) (*MsgSimulateTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx := types.Transaction{
		Sender:    msg.Sender,
		Recipient: msg.Recipient,
		Amount:    msg.Amount,
		Data:      msg.Data,
		Nonce:     msg.Nonce,
	}

	// Simulate transaction
	gasUsed := ms.keeper.EstimateGas(ctx, tx)
	valid, err := ms.keeper.ValidateTransaction(ctx, tx)

	response := &MsgSimulateTransactionResponse{
		Success:  valid && err == nil,
		GasUsed:  gasUsed,
		GasLimit: gasUsed + (gasUsed / 10), // Add 10% buffer
	}

	if err != nil {
		response.Error = err.Error()
	}

	// Check balance
	if !ms.keeper.CheckSufficientBalance(ctx, tx.Sender, tx.Amount) {
		response.Success = false
		response.Error = "insufficient balance"
	}

	return response, nil
}

// CancelPreValidation cancels a pre-validated transaction
func (ms msgServer) CancelPreValidation(goCtx context.Context, msg *MsgCancelPreValidation) (*MsgCancelPreValidationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remove from mempool
	ms.keeper.RemoveFromMempool(ctx, msg.TxHash)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"prevalidation_cancelled",
		sdk.NewAttribute("tx_hash", msg.TxHash),
		sdk.NewAttribute("sender", msg.Sender),
	))

	return &MsgCancelPreValidationResponse{}, nil
}

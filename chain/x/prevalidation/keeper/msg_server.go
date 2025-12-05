package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

type msgServer struct {
	pb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) pb.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ pb.MsgServer = (*msgServer)(nil)

// PreValidateTransaction pre-validates a transaction
func (ms msgServer) PreValidateTransaction(goCtx context.Context, msg *pb.MsgPreValidateTransaction) (*pb.MsgPreValidateTransactionResponse, error) {
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

	// Generate transaction hash
	txHash := ms.keeper.GetTransactionHash(tx)

	// Emit event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"prevalidate_transaction",
		sdk.NewAttribute("sender", tx.Sender),
		sdk.NewAttribute("recipient", tx.Recipient),
		sdk.NewAttribute("amount", tx.Amount),
		sdk.NewAttribute("gas_estimate", fmt.Sprintf("%d", gasEstimate)),
		sdk.NewAttribute("tx_hash", txHash),
	))

	return &pb.MsgPreValidateTransactionResponse{
		Valid:            true,
		GasEstimate:      gasEstimate,
		TxHash:           txHash,
		PreValidationId:  txHash, // Use tx hash as pre-validation ID
	}, nil
}

// BatchPreValidate pre-validates multiple transactions
func (ms msgServer) BatchPreValidate(goCtx context.Context, msg *pb.MsgBatchPreValidate) (*pb.MsgBatchPreValidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	results := make([]*pb.ValidationResult, len(msg.Transactions))

	for i, txMsg := range msg.Transactions {
		tx := types.Transaction{
			Sender:    txMsg.Sender,
			Recipient: txMsg.Recipient,
			Amount:    txMsg.Amount,
			Data:      txMsg.Data,
			Nonce:     txMsg.Nonce,
			Signature: txMsg.Signature,
		}

		txHash := ms.keeper.GetTransactionHash(tx)
		valid, err := ms.keeper.ValidateTransaction(ctx, tx)
		result := &pb.ValidationResult{
			TxHash:          txHash,
			Valid:           valid,
			GasEstimate:     ms.keeper.EstimateGas(ctx, tx),
			PreValidationId: txHash,
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

	return &pb.MsgBatchPreValidateResponse{
		Results: results,
	}, nil
}

// UpdateParams updates module parameters
func (ms msgServer) UpdateParams(goCtx context.Context, msg *pb.MsgUpdateParams) (*pb.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// In production, verify authority matches governance module address
	// For now, accept the parameters
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"params_updated",
		sdk.NewAttribute("module", "prevalidation"),
		sdk.NewAttribute("authority", msg.Authority),
	))

	return &pb.MsgUpdateParamsResponse{}, nil
}

// SimulateTransaction simulates a transaction without executing it
func (ms msgServer) SimulateTransaction(goCtx context.Context, msg *pb.MsgSimulateTransaction) (*pb.MsgSimulateTransactionResponse, error) {
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

	response := &pb.MsgSimulateTransactionResponse{
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
func (ms msgServer) CancelPreValidation(goCtx context.Context, msg *pb.MsgCancelPreValidation) (*pb.MsgCancelPreValidationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remove from mempool
	ms.keeper.RemoveFromMempool(ctx, msg.TxHash)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"prevalidation_cancelled",
		sdk.NewAttribute("tx_hash", msg.TxHash),
		sdk.NewAttribute("sender", msg.Sender),
	))

	return &pb.MsgCancelPreValidationResponse{}, nil
}

// CreateTemplate creates a new validation template
func (ms msgServer) CreateTemplate(goCtx context.Context, msg *pb.MsgCreateTemplate) (*pb.MsgCreateTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Generate template ID from hash of name and creator
	templateID := fmt.Sprintf("%s-%s", msg.Name, msg.Creator[:8])

	// Create template (would store in state in production)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"template_created",
		sdk.NewAttribute("template_id", templateID),
		sdk.NewAttribute("creator", msg.Creator),
		sdk.NewAttribute("name", msg.Name),
		sdk.NewAttribute("tx_type", msg.TxType.String()),
	))

	return &pb.MsgCreateTemplateResponse{
		TemplateId: templateID,
	}, nil
}

// UpdateTemplate updates an existing validation template
func (ms msgServer) UpdateTemplate(goCtx context.Context, msg *pb.MsgUpdateTemplate) (*pb.MsgUpdateTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update template (would update state in production)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"template_updated",
		sdk.NewAttribute("template_id", msg.TemplateId),
		sdk.NewAttribute("updater", msg.Updater),
		sdk.NewAttribute("name", msg.Name),
	))

	return &pb.MsgUpdateTemplateResponse{}, nil
}

// DeactivateTemplate deactivates a validation template
func (ms msgServer) DeactivateTemplate(goCtx context.Context, msg *pb.MsgDeactivateTemplate) (*pb.MsgDeactivateTemplateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Deactivate template (would update state in production)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"template_deactivated",
		sdk.NewAttribute("template_id", msg.TemplateId),
		sdk.NewAttribute("deactivator", msg.Deactivator),
	))

	return &pb.MsgDeactivateTemplateResponse{}, nil
}

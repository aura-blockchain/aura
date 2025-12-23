package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// AIModel represents an AI model with versioning
type AIModel struct {
	ModelHash      string
	Version        string
	Name           string
	Description    string
	Provider       string
	Capabilities   []string
	MaxInputSize   uint64
	MaxOutputSize  uint64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Status         ModelStatus
	PreviousHash   string // Previous version hash
	DeprecatedAt   *time.Time
	Metadata       map[string]string
}

// ModelStatus defines model status
type ModelStatus string

const (
	ModelStatusActive     ModelStatus = "active"
	ModelStatusDeprecated ModelStatus = "deprecated"
	ModelStatusRetired    ModelStatus = "retired"
)

// RegisterModel registers a new AI model version
func (k Keeper) RegisterModel(ctx sdk.Context, model AIModel) error {
	// Validate model
	if err := validateModel(model); err != nil {
		return fmt.Errorf("error in RegisterModel for Validate: %w", err)
	}

	// Check if model already exists
	if _, exists := k.GetModel(ctx, model.ModelHash); exists {
		return sdkerrors.Wrap(types.ErrInvalidParams, "model already registered")
	}

	// If this is an update, link to previous version
	if model.PreviousHash != "" {
		prevModel, exists := k.GetModel(ctx, model.PreviousHash)
		if !exists {
			return sdkerrors.Wrap(types.ErrInvalidParams, "previous model version not found")
		}
		// Optionally deprecate previous version
		if prevModel.Status == ModelStatusActive {
			now := ctx.BlockTime()
			prevModel.DeprecatedAt = &now
			prevModel.Status = ModelStatusDeprecated
			k.setModel(ctx, prevModel)
		}
	}

	model.CreatedAt = ctx.BlockTime()
	model.UpdatedAt = ctx.BlockTime()
	model.Status = ModelStatusActive

	k.setModel(ctx, model)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ai_model_registered",
			sdk.NewAttribute("model_hash", model.ModelHash),
			sdk.NewAttribute("version", model.Version),
			sdk.NewAttribute("name", model.Name),
		),
	)

	return nil
}

// GetModel retrieves a model by hash
func (k Keeper) GetModel(ctx sdk.Context, modelHash string) (AIModel, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.ModelKey(modelHash)

	bz := store.Get(key)
	if len(bz) == 0 {
		return AIModel{}, false
	}

	var model AIModel
	if err := json.Unmarshal(bz, &model); err != nil {
		return AIModel{}, false
	}

	return model, true
}

// setModel stores a model
func (k Keeper) setModel(ctx sdk.Context, model AIModel) {
	store := ctx.KVStore(k.storeKey)
	key := types.ModelKey(model.ModelHash)

	bz, err := json.Marshal(&model)
	if err != nil {
		return
	}

	store.Set(key, bz)

	// Index by version
	k.indexModelVersion(ctx, model)
}

// indexModelVersion creates version index for model
func (k Keeper) indexModelVersion(ctx sdk.Context, model AIModel) {
	store := ctx.KVStore(k.storeKey)
	key := types.ModelVersionKey(model.Name, model.Version)
	store.Set(key, []byte(model.ModelHash))
}

// GetModelByVersion retrieves model by name and version
func (k Keeper) GetModelByVersion(ctx sdk.Context, name, version string) (AIModel, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.ModelVersionKey(name, version)

	bz := store.Get(key)
	if len(bz) == 0 {
		return AIModel{}, false
	}

	modelHash := string(bz)
	return k.GetModel(ctx, modelHash)
}

// ListModels returns all registered models
func (k Keeper) ListModels(ctx sdk.Context) []AIModel {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ModelKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var models []AIModel
	for ; iterator.Valid(); iterator.Next() {
		var model AIModel
		if err := json.Unmarshal(iterator.Value(), &model); err != nil {
			continue
		}
		models = append(models, model)
	}

	return models
}

// ListActiveModels returns only active models
func (k Keeper) ListActiveModels(ctx sdk.Context) []AIModel {
	allModels := k.ListModels(ctx)
	var activeModels []AIModel

	for _, model := range allModels {
		if model.Status == ModelStatusActive {
			activeModels = append(activeModels, model)
		}
	}

	return activeModels
}

// DeprecateModel marks a model as deprecated
func (k Keeper) DeprecateModel(ctx sdk.Context, modelHash string) error {
	model, exists := k.GetModel(ctx, modelHash)
	if !exists {
		return sdkerrors.Wrap(types.ErrInvalidParams, "model not found")
	}

	if model.Status == ModelStatusRetired {
		return sdkerrors.Wrap(types.ErrInvalidParams, "model already retired")
	}

	now := ctx.BlockTime()
	model.DeprecatedAt = &now
	model.Status = ModelStatusDeprecated
	model.UpdatedAt = now

	k.setModel(ctx, model)

	return nil
}

// RetireModel permanently retires a model
func (k Keeper) RetireModel(ctx sdk.Context, modelHash string) error {
	model, exists := k.GetModel(ctx, modelHash)
	if !exists {
		return sdkerrors.Wrap(types.ErrInvalidParams, "model not found")
	}

	model.Status = ModelStatusRetired
	model.UpdatedAt = ctx.BlockTime()

	k.setModel(ctx, model)

	return nil
}

// GetModelVersionHistory returns version history for a model
func (k Keeper) GetModelVersionHistory(ctx sdk.Context, modelName string) []AIModel {
	var history []AIModel

	// Get all models and filter by name
	allModels := k.ListModels(ctx)
	for _, model := range allModels {
		if model.Name == modelName {
			history = append(history, model)
		}
	}

	return history
}

// GetLatestModelVersion returns the latest active version of a model
func (k Keeper) GetLatestModelVersion(ctx sdk.Context, modelName string) (AIModel, bool) {
	allModels := k.ListModels(ctx)
	var latestModel *AIModel

	for i, model := range allModels {
		if model.Name == modelName && model.Status == ModelStatusActive {
			if latestModel == nil || model.CreatedAt.After(latestModel.CreatedAt) {
				latestModel = &allModels[i]
			}
		}
	}

	if latestModel == nil {
		return AIModel{}, false
	}

	return *latestModel, true
}

// validateModel validates model fields
func validateModel(model AIModel) error {
	if model.ModelHash == "" {
		return fmt.Errorf("model hash required")
	}
	if model.Version == "" {
		return fmt.Errorf("model version required")
	}
	if model.Name == "" {
		return fmt.Errorf("model name required")
	}
	if len(model.Capabilities) == 0 {
		return fmt.Errorf("at least one capability required")
	}
	if model.MaxInputSize == 0 {
		return fmt.Errorf("max input size must be positive")
	}
	if model.MaxOutputSize == 0 {
		return fmt.Errorf("max output size must be positive")
	}
	return nil
}

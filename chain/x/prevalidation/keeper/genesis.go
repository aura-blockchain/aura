package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(genesis types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Set params
	if err := k.paramsStore.SetParams(genesis.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Load pre-validated transactions
	for _, tx := range genesis.PreValidatedTransactions {
		k.preValidatedTxs[tx.Id] = tx
		k.userPreValidatedTxs[tx.Signer] = append(k.userPreValidatedTxs[tx.Signer], tx.Id)

		// Initialize cache tracking
		k.cacheOrder = append(k.cacheOrder, tx.Id)
		k.cacheAccessCount[tx.Id] = 0
		if tx.ValidatedAt != nil {
			k.cacheAccessTime[tx.Id] = tx.ValidatedAt.AsTime()
		}
	}

	// Load templates
	for _, template := range genesis.Templates {
		k.templatesById[template.Id] = template
		k.templatesByType[template.TxType] = append(k.templatesByType[template.TxType], template)
	}

	// Load metrics
	if genesis.Metrics != nil {
		k.metrics = genesis.Metrics
	} else {
		k.metrics = &types.PreValidationMetrics{
			MetricsByType: make(map[string]*types.TypeMetrics),
			Last24Hours:   []*types.HourlyMetrics{},
			ControlGroup:  &types.ControlGroupMetrics{},
		}
	}

	// Initialize type amounts
	k.initializeTypeAmounts()

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Export pre-validated transactions
	preValidatedTxs := []*types.PreValidatedTransaction{}
	for _, tx := range k.preValidatedTxs {
		preValidatedTxs = append(preValidatedTxs, tx)
	}

	// Export templates
	templates := []*types.ValidationTemplate{}
	for _, template := range k.templatesById {
		templates = append(templates, template)
	}

	return types.GenesisState{
		Params:                   k.GetParams(),
		PreValidatedTransactions: preValidatedTxs,
		Templates:                templates,
		Metrics:                  k.metrics,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(genesis types.GenesisState) error {
	// Validate params
	if err := genesis.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate pre-validated transactions
	txIDs := make(map[string]bool)
	for i, tx := range genesis.PreValidatedTransactions {
		if tx.Id == "" {
			return fmt.Errorf("pre-validated transaction %d has empty ID", i)
		}

		if txIDs[tx.Id] {
			return fmt.Errorf("duplicate transaction ID: %s", tx.Id)
		}
		txIDs[tx.Id] = true

		if tx.Signer == "" {
			return fmt.Errorf("transaction %s has empty signer", tx.Id)
		}

		if tx.TxType == types.TxTypeUnspecified {
			return fmt.Errorf("transaction %s has unspecified type", tx.Id)
		}

		if len(tx.EncryptedData) == 0 {
			return fmt.Errorf("transaction %s has no encrypted data", tx.Id)
		}
	}

	// Validate templates
	templateIDs := make(map[string]bool)
	for i, template := range genesis.Templates {
		if template.Id == "" {
			return fmt.Errorf("template %d has empty ID", i)
		}

		if templateIDs[template.Id] {
			return fmt.Errorf("duplicate template ID: %s", template.Id)
		}
		templateIDs[template.Id] = true

		if template.TxType == types.TxTypeUnspecified {
			return fmt.Errorf("template %s has unspecified type", template.Id)
		}

		if template.Name == "" {
			return fmt.Errorf("template %s has empty name", template.Id)
		}
	}

	// Validate metrics
	if genesis.Metrics != nil {
		if genesis.Metrics.OverallCacheHitRate < 0 || genesis.Metrics.OverallCacheHitRate > 1 {
			return fmt.Errorf("invalid overall cache hit rate: %f", genesis.Metrics.OverallCacheHitRate)
		}
	}

	return nil
}

// DefaultGenesisState returns the default genesis state with some sample templates
func DefaultGenesisStateWithTemplates() types.GenesisState {
	genesis := types.DefaultGenesisState()

	// Add default templates for each transaction type
	templates := []*types.ValidationTemplate{
		{
			Id:                 "ir-completion-basic",
			TxType:             types.TxTypeIRCompletion,
			Name:               "Basic IR Completion",
			Description:        "Standard inclusion routine completion template",
			ValidationRules:    `{"min_confidence_score": 100}`,
			ParameterSchema:    `{"ir_id": "string", "wallet": "string"}`,
			GasFormula:         "50000",
			PriorityWeight:     100,
			MinConfidenceScore: 100,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "dex-swap-common-pairs",
			TxType:             types.TxTypeDexSwap,
			Name:               "Common DEX Pairs Swap",
			Description:        "Template for common trading pairs (USDC-AURA, ETH-AURA, etc.)",
			ValidationRules:    `{"min_liquidity": 1000}`,
			ParameterSchema:    `{"from_token": "string", "to_token": "string", "amount": "uint64"}`,
			GasFormula:         "100000",
			PriorityWeight:     80,
			MinConfidenceScore: 50,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "lp-deposit-standard",
			TxType:             types.TxTypeLPDeposit,
			Name:               "Standard LP Deposit",
			Description:        "Template for standard liquidity pool deposits",
			ValidationRules:    `{"min_amount": 100}`,
			ParameterSchema:    `{"pool_id": "string", "token_a": "string", "token_b": "string", "amount_a": "uint64", "amount_b": "uint64"}`,
			GasFormula:         "80000",
			PriorityWeight:     60,
			MinConfidenceScore: 50,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "lp-withdrawal-standard",
			TxType:             types.TxTypeLPWithdrawal,
			Name:               "Standard LP Withdrawal",
			Description:        "Template for standard liquidity pool withdrawals",
			ValidationRules:    `{"min_lp_tokens": 10}`,
			ParameterSchema:    `{"pool_id": "string", "lp_tokens": "uint64"}`,
			GasFormula:         "80000",
			PriorityWeight:     60,
			MinConfidenceScore: 50,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "vc-mint-basic",
			TxType:             types.TxTypeVCMint,
			Name:               "Basic VC Mint",
			Description:        "Template for basic verifiable credential minting",
			ValidationRules:    `{"min_confidence_score": 500}`,
			ParameterSchema:    `{"vc_type": "string", "holder": "string", "claims": "object"}`,
			GasFormula:         "120000",
			PriorityWeight:     70,
			MinConfidenceScore: 500,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "bridge-transfer-standard",
			TxType:             types.TxTypeBridgeTransfer,
			Name:               "Standard Bridge Transfer",
			Description:        "Template for standard cross-chain bridge transfers",
			ValidationRules:    `{"min_amount": 10, "supported_chains": ["ethereum", "polygon"]}`,
			ParameterSchema:    `{"to_chain": "string", "token": "string", "amount": "uint64", "recipient": "string"}`,
			GasFormula:         "150000",
			PriorityWeight:     50,
			MinConfidenceScore: 200,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "confidence-score-update-basic",
			TxType:             types.TxTypeConfidenceScoreUpdate,
			Name:               "Basic Confidence Score Update",
			Description:        "Template for confidence score updates",
			ValidationRules:    `{}`,
			ParameterSchema:    `{"wallet": "string", "score_delta": "int64"}`,
			GasFormula:         "60000",
			PriorityWeight:     90,
			MinConfidenceScore: 0,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
		{
			Id:                 "identity-change-basic",
			TxType:             types.TxTypeIdentityChange,
			Name:               "Basic Identity Change",
			Description:        "Template for basic identity change requests",
			ValidationRules:    `{"min_confidence_score": 1000}`,
			ParameterSchema:    `{"change_type": "string", "new_value": "string"}`,
			GasFormula:         "90000",
			PriorityWeight:     40,
			MinConfidenceScore: 1000,
			Active:             true,
			Stats:              &types.TemplateStats{},
		},
	}

	genesis.Templates = templates

	return genesis
}

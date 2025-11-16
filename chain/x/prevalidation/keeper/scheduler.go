package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SchedulerState tracks the state of the scheduler
type SchedulerState struct {
	LastRun          time.Time
	NextScheduledRun time.Time
	IsRunning        bool
	TotalRuns        uint64
}

// ShouldRunScheduler checks if the scheduler should run now
func (k *Keeper) ShouldRunScheduler() bool {
	params := k.GetParams()

	if !params.Enabled || params.SchedulerConfig == nil || !params.SchedulerConfig.Enabled {
		return false
	}

	currentTime := time.Unix(k.currentTime, 0)

	// Check if enough time has passed since last run
	if !k.lastSchedulerRun.IsZero() {
		intervalDuration := time.Duration(params.SchedulerConfig.RunIntervalMinutes) * time.Minute
		if currentTime.Sub(k.lastSchedulerRun) < intervalDuration {
			return false
		}
	}

	// Check if we're in off-peak hours (or peak hours allowed)
	return params.SchedulerConfig.ShouldRunScheduler(currentTime)
}

// RunScheduler executes the pre-validation scheduler
func (k *Keeper) RunScheduler() (*types.EventSchedulerRun, error) {
	params := k.GetParams()

	if !k.ShouldRunScheduler() {
		return nil, types.ErrSchedulerDisabled
	}

	startTime := time.Unix(k.currentTime, 0)
	k.lastSchedulerRun = startTime

	event := &types.EventSchedulerRun{
		StartedAt:     timestamppb.New(startTime),
		CreatedByType: make(map[string]uint64),
		IsOffPeak:     params.SchedulerConfig.IsOffPeakHour(uint32(startTime.Hour())),
	}

	totalCreated := uint64(0)

	// Pre-validate transactions for each type
	for txType, amount := range k.typeAmounts {
		if txType == types.TxTypeUnspecified {
			continue
		}

		created := k.preValidateForType(txType, amount)
		event.CreatedByType[txType.String()] = created
		totalCreated += created

		// Stop if we hit the max per run
		if totalCreated >= params.SchedulerConfig.MaxPerRun {
			break
		}
	}

	endTime := time.Unix(k.currentTime, 0)
	event.CompletedAt = timestamppb.New(endTime)
	event.PreValidationsCreated = totalCreated

	return event, nil
}

// preValidateForType creates pre-validated transactions for a specific type
func (k *Keeper) preValidateForType(txType types.TransactionType, amount uint64) uint64 {
	// Get templates for this type
	templates := k.GetTemplatesByType(txType)
	if len(templates) == 0 {
		// Create a default template if none exists
		k.createDefaultTemplate(txType)
		templates = k.GetTemplatesByType(txType)
	}

	if len(templates) == 0 {
		return 0
	}

	created := uint64(0)

	// Distribute pre-validations across templates based on priority weights
	totalWeight := uint32(0)
	for _, template := range templates {
		if template.Active {
			totalWeight += template.PriorityWeight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	for _, template := range templates {
		if !template.Active {
			continue
		}

		// Calculate how many to create for this template
		templateAmount := uint64(float64(amount) * (float64(template.PriorityWeight) / float64(totalWeight)))

		for i := uint64(0); i < templateAmount && created < amount; i++ {
			// Generate synthetic transaction data for pre-validation
			txData := k.generateSyntheticTxData(txType, template)

			// Use a synthetic signer (in production, this would be based on predicted users)
			signer := fmt.Sprintf("synthetic-signer-%d-%d", txType, i)

			// Estimate gas
			estimatedGas := k.estimateGas(txType, template)

			// Create pre-validated transaction
			_, err := k.CreatePreValidatedTransaction(
				txType,
				template.Id,
				txData,
				signer,
				estimatedGas,
				k.generateContextFromTemplate(template),
			)

			if err != nil {
				continue
			}

			created++

			// Update template stats
			k.mu.Lock()
			template.Stats.TotalCreated++
			k.mu.Unlock()
		}
	}

	return created
}

// createDefaultTemplate creates a default template for a transaction type
func (k *Keeper) createDefaultTemplate(txType types.TransactionType) {
	template := &types.ValidationTemplate{
		Id:                 fmt.Sprintf("default-%s", txType.String()),
		TxType:             txType,
		Name:               fmt.Sprintf("Default %s Template", types.TransactionTypeName(txType)),
		Description:        fmt.Sprintf("Default template for %s transactions", types.TransactionTypeName(txType)),
		ValidationRules:    "{}",
		ParameterSchema:    "{}",
		GasFormula:         "base_gas",
		PriorityWeight:     100,
		MinConfidenceScore: 100,
		Active:             true,
		Stats:              &types.TemplateStats{},
	}

	k.RegisterTemplate(template)
}

// generateSyntheticTxData generates synthetic transaction data for pre-validation
func (k *Keeper) generateSyntheticTxData(txType types.TransactionType, template *types.ValidationTemplate) []byte {
	// In production, this would generate realistic transaction data based on the template
	// For now, return a simple placeholder
	data := fmt.Sprintf("synthetic-tx-data:%s:%s", txType.String(), template.Id)
	return []byte(data)
}

// estimateGas estimates gas for a transaction type
func (k *Keeper) estimateGas(txType types.TransactionType, template *types.ValidationTemplate) uint64 {
	// Simple gas estimation based on transaction type
	// In production, this would use the template's gas formula
	baseGas := map[types.TransactionType]uint64{
		types.TxTypeIRCompletion:          50000,
		types.TxTypeDexSwap:               100000,
		types.TxTypeLPDeposit:             80000,
		types.TxTypeLPWithdrawal:          80000,
		types.TxTypeVCMint:                120000,
		types.TxTypeBridgeTransfer:        150000,
		types.TxTypeConfidenceScoreUpdate: 60000,
		types.TxTypeIdentityChange:        90000,
	}

	if gas, ok := baseGas[txType]; ok {
		return gas
	}

	return 100000 // Default
}

// generateContextFromTemplate generates context data from a template
func (k *Keeper) generateContextFromTemplate(template *types.ValidationTemplate) map[string]string {
	// In production, parse the template's parameter schema and generate context
	return map[string]string{
		"template_id":   template.Id,
		"template_name": template.Name,
	}
}

// GetSchedulerState returns the current scheduler state
func (k *Keeper) GetSchedulerState() SchedulerState {
	params := k.GetParams()

	state := SchedulerState{
		LastRun:   k.lastSchedulerRun,
		IsRunning: false,
		TotalRuns: 0,
	}

	// Calculate next scheduled run
	if params.SchedulerConfig != nil && params.SchedulerConfig.Enabled {
		if k.lastSchedulerRun.IsZero() {
			state.NextScheduledRun = k.calculateNextScheduledRun(time.Unix(k.currentTime, 0))
		} else {
			state.NextScheduledRun = k.calculateNextScheduledRun(k.lastSchedulerRun)
		}
	}

	return state
}

// calculateNextScheduledRun calculates when the scheduler should run next
func (k *Keeper) calculateNextScheduledRun(fromTime time.Time) time.Time {
	params := k.GetParams()

	if params.SchedulerConfig == nil || !params.SchedulerConfig.Enabled {
		return time.Time{}
	}

	// Add interval to get next possible run time
	nextTime := fromTime.Add(time.Duration(params.SchedulerConfig.RunIntervalMinutes) * time.Minute)

	// If we're not in off-peak hours and peak hours not allowed, advance to next off-peak hour
	if !params.SchedulerConfig.AllowPeakHours {
		for !params.SchedulerConfig.IsOffPeakHour(uint32(nextTime.Hour())) {
			nextTime = nextTime.Add(1 * time.Hour)
		}
	}

	return nextTime
}

// ForceSchedulerRun forces the scheduler to run immediately (for testing/admin)
func (k *Keeper) ForceSchedulerRun() (*types.EventSchedulerRun, error) {
	params := k.GetParams()

	if !params.Enabled {
		return nil, types.ErrSchedulerDisabled
	}

	// Temporarily allow peak hours
	originalAllowPeakHours := params.SchedulerConfig.AllowPeakHours
	params.SchedulerConfig.AllowPeakHours = true
	k.SetParams(params)

	// Run scheduler
	event, err := k.RunScheduler()

	// Restore original setting
	params.SchedulerConfig.AllowPeakHours = originalAllowPeakHours
	k.SetParams(params)

	return event, err
}

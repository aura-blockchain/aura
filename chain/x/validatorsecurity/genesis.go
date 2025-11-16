package validatorsecurity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set params
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Set validator security info
	for _, val := range genState.Validators {
		k.SetValidatorSecurityInfo(ctx, val)
	}

	// Set double sign evidences
	for _, evidence := range genState.DoubleSignEvidences {
		k.SetDoubleSignEvidence(ctx, evidence)
	}

	// Set downtime infractions
	for _, infraction := range genState.DowntimeInfractions {
		k.SetDowntimeInfraction(ctx, infraction)
	}

	// Set alerts
	for _, alert := range genState.Alerts {
		k.CreateAlert(ctx, alert)
	}

	// Set sentry nodes
	for _, node := range genState.SentryNodes {
		k.SetSentryNodeInfo(ctx, node)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesisState()

	genesis.Params = k.GetParams(ctx)
	genesis.Validators = k.GetAllValidators(ctx)
	genesis.DoubleSignEvidences = k.GetAllDoubleSignEvidences(ctx)
	genesis.Alerts = k.GetAllAlerts(ctx)

	// Export downtime infractions
	for _, val := range genesis.Validators {
		if infraction, found := k.GetDowntimeInfraction(ctx, val.ValidatorAddress); found {
			genesis.DowntimeInfractions = append(genesis.DowntimeInfractions, infraction)
		}

		// Export sentry nodes
		sentryNodes := k.GetValidatorSentryNodes(ctx, val.ValidatorAddress)
		genesis.SentryNodes = append(genesis.SentryNodes, sentryNodes...)
	}

	return genesis
}

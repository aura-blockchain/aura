package types

import (
	"unsafe"

	v1beta1 "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// ParamsFromProto converts v1beta1.Params to types.Params
// Uses unsafe pointer conversion since the types are structurally identical
func ParamsFromProto(p *v1beta1.Params) Params {
	if p == nil {
		return Params{}
	}

	return Params{
		Tokenomics:              (*TokenomicsConfig)(unsafe.Pointer(p.Tokenomics)),
		WhaleProtection:         (*WhaleProtection)(unsafe.Pointer(p.WhaleProtection)),
		TransferTax:             (*TransferTaxConfig)(unsafe.Pointer(p.TransferTax)),
		LiquidityMining:         (*LiquidityMiningConfig)(unsafe.Pointer(p.LiquidityMining)),
		Governance:              (*GovernanceConfig)(unsafe.Pointer(p.Governance)),
		TreasuryMultisig:        (*TreasuryMultisig)(unsafe.Pointer(p.TreasuryMultisig)),
		DynamicFees:             (*DynamicFeeConfig)(unsafe.Pointer(p.DynamicFees)),
		Mev:                     (*MEVConfig)(unsafe.Pointer(p.Mev)),
		InflationAlertThreshold: p.InflationAlertThreshold,
		InflationCheckInterval:  p.InflationCheckInterval,
	}
}

// ParamsToProto converts types.Params to v1beta1.Params
// Uses unsafe pointer conversion since the types are structurally identical
func ParamsToProto(p Params) *v1beta1.Params {
	return &v1beta1.Params{
		Tokenomics:              (*v1beta1.TokenomicsConfig)(unsafe.Pointer(p.Tokenomics)),
		WhaleProtection:         (*v1beta1.WhaleProtection)(unsafe.Pointer(p.WhaleProtection)),
		TransferTax:             (*v1beta1.TransferTaxConfig)(unsafe.Pointer(p.TransferTax)),
		LiquidityMining:         (*v1beta1.LiquidityMiningConfig)(unsafe.Pointer(p.LiquidityMining)),
		Governance:              (*v1beta1.GovernanceConfig)(unsafe.Pointer(p.Governance)),
		TreasuryMultisig:        (*v1beta1.TreasuryMultisig)(unsafe.Pointer(p.TreasuryMultisig)),
		DynamicFees:             (*v1beta1.DynamicFeeConfig)(unsafe.Pointer(p.DynamicFees)),
		Mev:                     (*v1beta1.MEVConfig)(unsafe.Pointer(p.Mev)),
		InflationAlertThreshold: p.InflationAlertThreshold,
		InflationCheckInterval:  p.InflationCheckInterval,
	}
}

// VestingScheduleToProto converts types.VestingSchedule to v1beta1.VestingSchedule
func VestingScheduleToProto(v *VestingSchedule) *v1beta1.VestingSchedule {
	if v == nil {
		return nil
	}
	return (*v1beta1.VestingSchedule)(unsafe.Pointer(v))
}

// VestingSchedulesSliceToProto converts slice of types.VestingSchedule to v1beta1.VestingSchedule
func VestingSchedulesSliceToProto(schedules []*VestingSchedule) []*v1beta1.VestingSchedule {
	result := make([]*v1beta1.VestingSchedule, len(schedules))
	for i, s := range schedules {
		result[i] = VestingScheduleToProto(s)
	}
	return result
}

// VoteLockToProto converts types.VoteLock to v1beta1.VoteLock
func VoteLockToProto(l *VoteLock) *v1beta1.VoteLock {
	if l == nil {
		return nil
	}
	return (*v1beta1.VoteLock)(unsafe.Pointer(l))
}

// VoteLocksSliceToProto converts slice of types.VoteLock to v1beta1.VoteLock
func VoteLocksSliceToProto(locks []*VoteLock) []*v1beta1.VoteLock {
	result := make([]*v1beta1.VoteLock, len(locks))
	for i, l := range locks {
		result[i] = VoteLockToProto(l)
	}
	return result
}

// PendingTreasuryTxToProto converts types.PendingTreasuryTx to v1beta1.PendingTreasuryTx
func PendingTreasuryTxToProto(tx *PendingTreasuryTx) *v1beta1.PendingTreasuryTx {
	if tx == nil {
		return nil
	}
	return (*v1beta1.PendingTreasuryTx)(unsafe.Pointer(tx))
}

// PendingTreasuryTxsSliceToProto converts slice of types.PendingTreasuryTx to v1beta1.PendingTreasuryTx
func PendingTreasuryTxsSliceToProto(txs []*PendingTreasuryTx) []*v1beta1.PendingTreasuryTx {
	result := make([]*v1beta1.PendingTreasuryTx, len(txs))
	for i, tx := range txs {
		result[i] = PendingTreasuryTxToProto(tx)
	}
	return result
}

// InflationAlertToProto converts types.InflationAlert to v1beta1.InflationAlert
func InflationAlertToProto(a *InflationAlert) *v1beta1.InflationAlert {
	if a == nil {
		return nil
	}
	return (*v1beta1.InflationAlert)(unsafe.Pointer(a))
}

// InflationAlertsSliceToProto converts slice of types.InflationAlert to v1beta1.InflationAlert
func InflationAlertsSliceToProto(alerts []*InflationAlert) []*v1beta1.InflationAlert {
	result := make([]*v1beta1.InflationAlert, len(alerts))
	for i, a := range alerts {
		result[i] = InflationAlertToProto(a)
	}
	return result
}

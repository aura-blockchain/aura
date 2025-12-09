package types

import (
    "fmt"

    pb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// GenesisState re-exports the protobuf genesis structure.
type GenesisState = pb.GenesisState

// DefaultGenesis returns a baseline compliance genesis state.
func DefaultGenesis() *GenesisState {
    params := DefaultParams()
    return &GenesisState{
        Params:               params,
        KycRecords:           []*pb.KYCRecord{},
        AmlProfiles:          []*pb.AMLProfile{},
        SuspiciousActivities: []*pb.SuspiciousActivity{},
        MonitoringRules:      []*pb.TransactionMonitoringRule{},
        TransactionAlerts:    []*pb.TransactionAlert{},
        SanctionsResults:     []*pb.SanctionsScreeningResult{},
        GdprConsents:         []*pb.GDPRConsent{},
        GdprRequests:         []*pb.GDPRDataRequest{},
        TaxReports:           []*pb.TaxReport{},
    }
}

// ValidateGenesis performs minimal checks on the compliance genesis state.
func ValidateGenesis(gen *GenesisState) error {
    if gen == nil {
        return fmt.Errorf("genesis state cannot be nil")
    }
    if err := ValidateParams(gen.Params); err != nil {
        return err
    }
    return nil
}

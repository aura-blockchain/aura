package types

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params == nil {
		t.Fatal("DefaultParams should not return nil")
	}

	if params.MinDeposit != "10000000stake" {
		t.Errorf("expected MinDeposit to be 10000000stake, got %s", params.MinDeposit)
	}

	if params.Quorum != "0.334" {
		t.Errorf("expected Quorum to be 0.334, got %s", params.Quorum)
	}

	if params.Threshold != "0.50" {
		t.Errorf("expected Threshold to be 0.50, got %s", params.Threshold)
	}

	if params.VetoThreshold != "0.334" {
		t.Errorf("expected VetoThreshold to be 0.334, got %s", params.VetoThreshold)
	}

	if params.VetoCosignersRequired != 3 {
		t.Errorf("expected VetoCosignersRequired to be 3, got %d", params.VetoCosignersRequired)
	}
}

func TestDefaultGovernanceParams(t *testing.T) {
	params := DefaultGovernanceParams()

	if params.CategoryParams == nil {
		t.Fatal("expected CategoryParams to be initialized")
	}

	// Check category-specific params
	textParams, ok := params.CategoryParams[CategoryText.String()]
	if !ok {
		t.Error("expected TEXT category params to be set")
	} else {
		if textParams.MinDeposit != "10000000stake" {
			t.Errorf("expected TEXT MinDeposit to be 10000000stake, got %s", textParams.MinDeposit)
		}
	}

	emergencyParams, ok := params.CategoryParams[CategoryEmergency.String()]
	if !ok {
		t.Error("expected EMERGENCY category params to be set")
	} else {
		if emergencyParams.VotingPeriod.AsDuration() != 24*time.Hour {
			t.Errorf("expected EMERGENCY VotingPeriod to be 24h, got %v", emergencyParams.VotingPeriod.AsDuration())
		}
		if emergencyParams.Quorum != "0.600" {
			t.Errorf("expected EMERGENCY Quorum to be 0.600, got %s", emergencyParams.Quorum)
		}
	}
}

func TestCategoryAliases(t *testing.T) {
	if CategoryText != ProposalCategory_PROPOSAL_CATEGORY_TEXT {
		t.Error("CategoryText alias mismatch")
	}

	if CategoryParameterChange != ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE {
		t.Error("CategoryParameterChange alias mismatch")
	}

	if CategoryEmergency != ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY {
		t.Error("CategoryEmergency alias mismatch")
	}

	if CategoryConstitution != ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION {
		t.Error("CategoryConstitution alias mismatch")
	}
}

func TestStatusAliases(t *testing.T) {
	if StatusDepositPeriod != ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
		t.Error("StatusDepositPeriod alias mismatch")
	}

	if StatusVotingPeriod != ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		t.Error("StatusVotingPeriod alias mismatch")
	}

	if StatusPassed != ProposalStatus_PROPOSAL_STATUS_PASSED {
		t.Error("StatusPassed alias mismatch")
	}

	if StatusExecuted != ProposalStatus_PROPOSAL_STATUS_EXECUTED {
		t.Error("StatusExecuted alias mismatch")
	}
}

func TestVoteOptionAliases(t *testing.T) {
	if OptionYes != VoteOption_VOTE_OPTION_YES {
		t.Error("OptionYes alias mismatch")
	}

	if OptionNo != VoteOption_VOTE_OPTION_NO {
		t.Error("OptionNo alias mismatch")
	}

	if OptionAbstain != VoteOption_VOTE_OPTION_ABSTAIN {
		t.Error("OptionAbstain alias mismatch")
	}

	if OptionNoWithVeto != VoteOption_VOTE_OPTION_NO_WITH_VETO {
		t.Error("OptionNoWithVeto alias mismatch")
	}
}

func TestTimeParameters(t *testing.T) {
	params := DefaultGovernanceParams()

	expectedVotingPeriod := 7 * 24 * time.Hour
	if params.VotingPeriod.AsDuration() != expectedVotingPeriod {
		t.Errorf("expected VotingPeriod to be %v, got %v", expectedVotingPeriod, params.VotingPeriod.AsDuration())
	}

	expectedExecutionDelay := 48 * time.Hour
	if params.ExecutionDelay.AsDuration() != expectedExecutionDelay {
		t.Errorf("expected ExecutionDelay to be %v, got %v", expectedExecutionDelay, params.ExecutionDelay.AsDuration())
	}

	expectedEmergencyVoting := 24 * time.Hour
	if params.EmergencyVotingPeriod.AsDuration() != expectedEmergencyVoting {
		t.Errorf("expected EmergencyVotingPeriod to be %v, got %v", expectedEmergencyVoting, params.EmergencyVotingPeriod.AsDuration())
	}
}

func TestCategoryParamsCount(t *testing.T) {
	params := DefaultGovernanceParams()

	expectedCategories := 6 // Text, ParamChange, SoftwareUpgrade, Spending, Emergency, Constitution
	if len(params.CategoryParams) != expectedCategories {
		t.Errorf("expected %d category params, got %d", expectedCategories, len(params.CategoryParams))
	}
}

func TestDurationPbConversion(t *testing.T) {
	duration := 7 * 24 * time.Hour
	dpb := durationpb.New(duration)

	if dpb.AsDuration() != duration {
		t.Errorf("duration conversion failed: expected %v, got %v", duration, dpb.AsDuration())
	}
}

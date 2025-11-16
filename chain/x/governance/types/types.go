package types

import (
	"fmt"
	"time"
)

// ProposalStatus defines the status of a governance proposal
type ProposalStatus int32

const (
	StatusUnspecified ProposalStatus = iota
	StatusDepositPeriod
	StatusVotingPeriod
	StatusPassed
	StatusRejected
	StatusFailed
	StatusVetoed
	StatusExecutionDelay
	StatusReadyForExecution
	StatusExecuted
)

func (s ProposalStatus) String() string {
	switch s {
	case StatusDepositPeriod:
		return "DEPOSIT_PERIOD"
	case StatusVotingPeriod:
		return "VOTING_PERIOD"
	case StatusPassed:
		return "PASSED"
	case StatusRejected:
		return "REJECTED"
	case StatusFailed:
		return "FAILED"
	case StatusVetoed:
		return "VETOED"
	case StatusExecutionDelay:
		return "EXECUTION_DELAY"
	case StatusReadyForExecution:
		return "READY_FOR_EXECUTION"
	case StatusExecuted:
		return "EXECUTED"
	default:
		return "UNSPECIFIED"
	}
}

// VoteOption defines a vote option
type VoteOption int32

const (
	OptionUnspecified VoteOption = iota
	OptionYes
	OptionAbstain
	OptionNo
	OptionNoWithVeto
)

func (o VoteOption) String() string {
	switch o {
	case OptionYes:
		return "YES"
	case OptionAbstain:
		return "ABSTAIN"
	case OptionNo:
		return "NO"
	case OptionNoWithVeto:
		return "NO_WITH_VETO"
	default:
		return "UNSPECIFIED"
	}
}

// ProposalCategory defines the category of a proposal
type ProposalCategory int32

const (
	CategoryUnspecified ProposalCategory = iota
	CategoryText
	CategoryParameterChange
	CategorySoftwareUpgrade
	CategorySpending
	CategoryEmergency
	CategoryConstitution
)

func (c ProposalCategory) String() string {
	switch c {
	case CategoryText:
		return "TEXT"
	case CategoryParameterChange:
		return "PARAMETER_CHANGE"
	case CategorySoftwareUpgrade:
		return "SOFTWARE_UPGRADE"
	case CategorySpending:
		return "SPENDING"
	case CategoryEmergency:
		return "EMERGENCY"
	case CategoryConstitution:
		return "CONSTITUTION"
	default:
		return "UNSPECIFIED"
	}
}

// Proposal defines a governance proposal
type Proposal struct {
	ID          uint64
	Title       string
	Description string
	Category    ProposalCategory
	Status      ProposalStatus

	// Voting details
	FinalTallyResult TallyResult
	SubmitTime       time.Time
	DepositEndTime   time.Time
	VotingStartTime  time.Time
	VotingEndTime    time.Time
	ExecutionTime    time.Time

	// Deposit tracking
	TotalDeposit string
	Depositors   []string

	// Proposer
	Proposer string

	// Execution delay (time-lock)
	ExecutionDelay time.Duration

	// Emergency flag
	IsEmergency bool

	// Snapshot block height
	SnapshotHeight uint64
}

// ValidateBasic performs basic validation on a proposal
func (p *Proposal) ValidateBasic() error {
	if p.Title == "" {
		return fmt.Errorf("proposal title cannot be empty")
	}
	if p.Description == "" {
		return fmt.Errorf("proposal description cannot be empty")
	}
	if p.Proposer == "" {
		return fmt.Errorf("proposer cannot be empty")
	}
	if p.Category == CategoryUnspecified {
		return fmt.Errorf("proposal category must be specified")
	}
	return nil
}

// Proto message interface methods for Proposal
func (p *Proposal) Reset() {
	*p = Proposal{}
}

func (p *Proposal) String() string {
	return fmt.Sprintf(`Proposal{
  ID: %d,
  Title: %s,
  Description: %s,
  Category: %s,
  Status: %s,
  Proposer: %s,
  TotalDeposit: %s,
  SubmitTime: %s,
  VotingStartTime: %s,
  VotingEndTime: %s,
  ExecutionTime: %s,
  ExecutionDelay: %s,
  IsEmergency: %v,
  SnapshotHeight: %d,
}`,
		p.ID,
		p.Title,
		p.Description,
		p.Category.String(),
		p.Status.String(),
		p.Proposer,
		p.TotalDeposit,
		p.SubmitTime,
		p.VotingStartTime,
		p.VotingEndTime,
		p.ExecutionTime,
		p.ExecutionDelay,
		p.IsEmergency,
		p.SnapshotHeight,
	)
}

func (p *Proposal) ProtoMessage() {}

// Deposit represents a deposit made on a proposal
type Deposit struct {
	ProposalID uint64
	Depositor  string
	Amount     string
	Timestamp  time.Time
}

// Vote represents a vote on a proposal
type Vote struct {
	ProposalID    uint64
	Voter         string
	Option        VoteOption
	Timestamp     time.Time
	IsSecret      bool
	EncryptedVote string
	Commitment    string
	VotingPower   string
}

// ValidateBasic performs basic validation on a vote
func (v *Vote) ValidateBasic() error {
	if v.Voter == "" {
		return fmt.Errorf("voter cannot be empty")
	}
	// Secret votes have unspecified option until revealed
	if !v.IsSecret && v.Option == OptionUnspecified {
		return fmt.Errorf("vote option must be specified")
	}
	if v.IsSecret && v.Commitment == "" {
		return fmt.Errorf("secret vote must have commitment")
	}
	return nil
}

// TallyResult defines the tally of a proposal vote
type TallyResult struct {
	Yes              string
	Abstain          string
	No               string
	NoWithVeto       string
	TotalVotingPower string
	TurnoutPercent   string
}

// VoteDelegation represents delegation of voting power
type VoteDelegation struct {
	Delegator      string
	Delegate       string
	DelegationTime time.Time
	DelegatedPower string
	Categories     []ProposalCategory
}

// TokenLock represents locked tokens during voting
type TokenLock struct {
	Owner        string
	ProposalID   uint64
	LockedAmount string
	LockTime     time.Time
	UnlockTime   time.Time
}

// VetoRequest represents a veto request for emergency situations
type VetoRequest struct {
	ProposalID uint64
	Vetoer     string
	Reason     string
	Timestamp  time.Time
	Cosigners  []string
}

// SnapshotVote represents an off-chain signaling vote
type SnapshotVote struct {
	ProposalID            uint64
	SnapshotHeight        uint64
	Voter                 string
	Option                VoteOption
	VotingPowerAtSnapshot string
	Signature             string
	Timestamp             time.Time
}

// WeightedVoteOption defines a vote option with weight
type WeightedVoteOption struct {
	Option VoteOption
	Weight string
}

// Proto message interface methods for Vote
func (v *Vote) Reset() {
	*v = Vote{}
}

func (v *Vote) String() string {
	return fmt.Sprintf(`Vote{
  ProposalID: %d,
  Voter: %s,
  Option: %s,
  Timestamp: %s,
  IsSecret: %v,
  VotingPower: %s,
}`,
		v.ProposalID,
		v.Voter,
		v.Option.String(),
		v.Timestamp,
		v.IsSecret,
		v.VotingPower,
	)
}

func (v *Vote) ProtoMessage() {}

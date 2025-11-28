package e2e

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil/keeper"
)

// Chain represents a test blockchain
type Chain struct {
	T           *testing.T
	ChainID     string
	Validators  []*Validator
	FullNodes   []*FullNode
	Genesis     *GenesisState
	App         *keeper.TestApp
	Ctx         sdk.Context
}

// Validator represents a validator node
type Validator struct {
	Address    sdk.ValAddress
	Moniker    string
	Power      int64
	Commission math.LegacyDec
}

// FullNode represents a full node (non-validator)
type FullNode struct {
	Address sdk.AccAddress
	Moniker string
}

// GenesisState represents the genesis state
type GenesisState struct {
	ChainID       string
	GenesisTime   time.Time
	Validators    []Validator
	Accounts      []sdk.AccAddress
	InitialSupply map[string]int64
}

// NewChain creates a new test chain with validators
func NewChain(t *testing.T, chainID string, validatorCount int) *Chain {
	t.Helper()

	// Create validators
	validators := make([]*Validator, validatorCount)
	for i := 0; i < validatorCount; i++ {
		valAddr := keeper.GenTestValidatorAddress()
		validators[i] = &Validator{
			Address:    valAddr,
			Moniker:    "validator-" + string(rune(i)),
			Power:      100,
			Commission: math.LegacyNewDecWithPrec(5, 2), // 5%
		}
	}

	// Create genesis state
	genesis := &GenesisState{
		ChainID:       chainID,
		GenesisTime:   time.Now().UTC(),
		Validators:    make([]Validator, validatorCount),
		Accounts:      make([]sdk.AccAddress, 0),
		InitialSupply: make(map[string]int64),
	}

	for i, val := range validators {
		genesis.Validators[i] = *val
	}

	// Setup test app
	app, ctx := keeper.SetupTestApp(t)

	chain := &Chain{
		T:          t,
		ChainID:    chainID,
		Validators: validators,
		FullNodes:  make([]*FullNode, 0),
		Genesis:    genesis,
		App:        app,
		Ctx:        ctx,
	}

	return chain
}

// Start starts the test chain
func (c *Chain) Start() error {
	// Initialize chain state
	c.T.Log("Starting chain:", c.ChainID)
	return nil
}

// Stop stops the test chain
func (c *Chain) Stop() error {
	// Cleanup chain resources
	c.T.Log("Stopping chain:", c.ChainID)
	return nil
}

// AddFullNode adds a full node to the chain
func (c *Chain) AddFullNode(moniker string) *FullNode {
	addr := keeper.GenTestAddress()
	node := &FullNode{
		Address: addr,
		Moniker: moniker,
	}
	c.FullNodes = append(c.FullNodes, node)
	return node
}

// GetValidator returns a validator by index
func (c *Chain) GetValidator(index int) *Validator {
	require.Less(c.T, index, len(c.Validators), "validator index out of range")
	return c.Validators[index]
}

// GetFullNode returns a full node by index
func (c *Chain) GetFullNode(index int) *FullNode {
	require.Less(c.T, index, len(c.FullNodes), "full node index out of range")
	return c.FullNodes[index]
}

// WaitForHeight waits for the chain to reach a specific height
func (c *Chain) WaitForHeight(height int64) error {
	// Simulate waiting for block height
	header := c.Ctx.BlockHeader()
	header.Height = height
	c.Ctx = c.Ctx.WithBlockHeader(header)
	return nil
}

// WaitForNextBlock waits for the next block
func (c *Chain) WaitForNextBlock() error {
	header := c.Ctx.BlockHeader()
	header.Height++
	header.Time = header.Time.Add(6 * time.Second) // ~6s block time
	c.Ctx = c.Ctx.WithBlockHeader(header)
	return nil
}

// SendTransaction simulates sending a transaction
func (c *Chain) SendTransaction(from sdk.AccAddress, msg sdk.Msg) error {
	// Simulate transaction execution
	c.T.Log("Sending transaction from:", from.String())
	return nil
}

// Query simulates querying chain state
func (c *Chain) Query(path string, data []byte) ([]byte, error) {
	// Simulate query
	c.T.Log("Querying:", path)
	return []byte{}, nil
}

// CreateAccount creates a funded account on the chain
func (c *Chain) CreateAccount(balance sdk.Coins) sdk.AccAddress {
	addr := keeper.GenTestAddress()
	c.Genesis.Accounts = append(c.Genesis.Accounts, addr)
	return addr
}

// GetHeight returns the current block height
func (c *Chain) GetHeight() int64 {
	return c.Ctx.BlockHeight()
}

// GetTime returns the current block time
func (c *Chain) GetTime() time.Time {
	return c.Ctx.BlockTime()
}

// MultiChainTestSuite manages multiple test chains for IBC-like testing
type MultiChainTestSuite struct {
	T      *testing.T
	Chains []*Chain
}

// NewMultiChainTest creates a multi-chain test environment
func NewMultiChainTest(t *testing.T, numChains int) *MultiChainTestSuite {
	t.Helper()

	chains := make([]*Chain, numChains)
	for i := 0; i < numChains; i++ {
		chainID := "test-chain-" + string(rune('A'+i))
		chains[i] = NewChain(t, chainID, 4)
	}

	return &MultiChainTestSuite{
		T:      t,
		Chains: chains,
	}
}

// StartAll starts all chains
func (m *MultiChainTestSuite) StartAll() error {
	for _, chain := range m.Chains {
		if err := chain.Start(); err != nil {
			return err
		}
	}
	return nil
}

// StopAll stops all chains
func (m *MultiChainTestSuite) StopAll() error {
	for _, chain := range m.Chains {
		if err := chain.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// GetChain returns a chain by index
func (m *MultiChainTestSuite) GetChain(index int) *Chain {
	require.Less(m.T, index, len(m.Chains), "chain index out of range")
	return m.Chains[index]
}

// SimulateIBCTransfer simulates an IBC transfer between chains
func (m *MultiChainTestSuite) SimulateIBCTransfer(
	sourceChain, destChain int,
	from sdk.AccAddress,
	to sdk.AccAddress,
	amount sdk.Coins,
) error {
	// Simulate IBC transfer
	source := m.GetChain(sourceChain)
	dest := m.GetChain(destChain)

	m.T.Logf("IBC transfer from %s to %s", source.ChainID, dest.ChainID)
	return nil
}

// WaitForRelayer simulates relayer operations
func (m *MultiChainTestSuite) WaitForRelayer() error {
	// Simulate relayer packet relay
	m.T.Log("Relaying packets...")
	time.Sleep(100 * time.Millisecond)
	return nil
}

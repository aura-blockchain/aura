package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// MockStakingKeeperWithStake allows setting stake amounts for testing
type MockStakingKeeperWithStake struct {
	delegatorBonded map[string]sdkmath.Int
}

func NewMockStakingKeeperWithStake() *MockStakingKeeperWithStake {
	return &MockStakingKeeperWithStake{
		delegatorBonded: make(map[string]sdkmath.Int),
	}
}

func (m *MockStakingKeeperWithStake) GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdkmath.Int {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount
	}
	return sdkmath.ZeroInt()
}

func (m *MockStakingKeeperWithStake) SetDelegatorBonded(delegator sdk.AccAddress, amount sdkmath.Int) {
	m.delegatorBonded[delegator.String()] = amount
}

// TestVotingPowerBasedOnStake verifies voting power is derived from staked tokens
func TestVotingPowerBasedOnStake(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := NewMockStakingKeeperWithStake()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking)

	// Create test address
	addr := keepertest.GenTestAddr()

	// Set staked amount
	stakedAmount := sdkmath.NewInt(1000000)
	mockStaking.SetDelegatorBonded(addr, stakedAmount)

	// Get voting power
	power, err := k.GetVotingPower(input.Ctx, addr.String())
	require.NoError(t, err)

	// Voting power should equal staked amount (no delegations)
	assert.Equal(t, stakedAmount, power)
}

// TestVotingPowerWithDelegations verifies voting power calculations with delegations
func TestVotingPowerWithDelegations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := NewMockStakingKeeperWithStake()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking)

	// Create addresses
	delegatee := keepertest.GenTestAddr() // Person receiving delegation
	delegator := keepertest.GenTestAddr() // Person delegating

	// Set stakes
	delegateeStake := sdkmath.NewInt(500000)
	delegatorStake := sdkmath.NewInt(300000)
	mockStaking.SetDelegatorBonded(delegatee, delegateeStake)
	mockStaking.SetDelegatorBonded(delegator, delegatorStake)

	// Create vote delegation
	delegation := &types.VoteDelegation{
		Delegator:      delegator.String(),
		Delegate:       delegatee.String(),
		DelegatedPower: delegatorStake.String(),
	}
	err := k.SetVoteDelegation(input.Ctx, delegation)
	require.NoError(t, err)

	// Delegatee's voting power should be: their stake + delegated stake
	delegateePower, err := k.GetVotingPower(input.Ctx, delegatee.String())
	require.NoError(t, err)
	expectedPower := delegateeStake.Add(delegatorStake)
	assert.Equal(t, expectedPower, delegateePower)

	// Delegator's voting power should be zero (delegated away)
	delegatorPower, err := k.GetVotingPower(input.Ctx, delegator.String())
	require.NoError(t, err)
	assert.Equal(t, sdkmath.ZeroInt(), delegatorPower)
}

// TestSybilResistance verifies addresses without stake have no voting power
func TestSybilResistance(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := NewMockStakingKeeperWithStake()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking)

	// Create addresses with no stake
	addrs := []sdk.AccAddress{
		keepertest.GenTestAddr(),
		keepertest.GenTestAddr(),
		keepertest.GenTestAddr(),
	}

	// All should have zero voting power
	for _, addr := range addrs {
		power, err := k.GetVotingPower(input.Ctx, addr.String())
		require.NoError(t, err)
		assert.Equal(t, sdkmath.ZeroInt(), power, "Address without stake should have zero voting power")
	}
}

// TestWhaleVotingPower verifies large stakeholders have proportional power
func TestWhaleVotingPower(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := NewMockStakingKeeperWithStake()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking)

	whale := keepertest.GenTestAddr()
	regular := keepertest.GenTestAddr()

	// Whale has 1000x more stake
	whaleStake := sdkmath.NewInt(1000000000) // 1B tokens
	regularStake := sdkmath.NewInt(1000000)   // 1M tokens

	mockStaking.SetDelegatorBonded(whale, whaleStake)
	mockStaking.SetDelegatorBonded(regular, regularStake)

	whalePower, err := k.GetVotingPower(input.Ctx, whale.String())
	require.NoError(t, err)
	regularPower, err := k.GetVotingPower(input.Ctx, regular.String())
	require.NoError(t, err)

	// Whale should have 1000x more voting power
	assert.Equal(t, whaleStake, whalePower)
	assert.Equal(t, regularStake, regularPower)
	assert.True(t, whalePower.GT(regularPower))
}

// TestMultipleDelegations verifies handling of multiple delegations to same address
func TestMultipleDelegations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockStaking := NewMockStakingKeeperWithStake()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockStaking)

	delegatee := keepertest.GenTestAddr()
	delegator1 := keepertest.GenTestAddr()
	delegator2 := keepertest.GenTestAddr()
	delegator3 := keepertest.GenTestAddr()

	delegateeStake := sdkmath.NewInt(100000)
	del1Stake := sdkmath.NewInt(50000)
	del2Stake := sdkmath.NewInt(75000)
	del3Stake := sdkmath.NewInt(25000)

	mockStaking.SetDelegatorBonded(delegatee, delegateeStake)
	mockStaking.SetDelegatorBonded(delegator1, del1Stake)
	mockStaking.SetDelegatorBonded(delegator2, del2Stake)
	mockStaking.SetDelegatorBonded(delegator3, del3Stake)

	// Create delegations
	delegations := []*types.VoteDelegation{
		{Delegator: delegator1.String(), Delegate: delegatee.String(), DelegatedPower: del1Stake.String()},
		{Delegator: delegator2.String(), Delegate: delegatee.String(), DelegatedPower: del2Stake.String()},
		{Delegator: delegator3.String(), Delegate: delegatee.String(), DelegatedPower: del3Stake.String()},
	}

	for _, delegation := range delegations {
		err := k.SetVoteDelegation(input.Ctx, delegation)
		require.NoError(t, err)
	}

	// Delegatee should have their stake + all delegated stakes
	delegateePower, err := k.GetVotingPower(input.Ctx, delegatee.String())
	require.NoError(t, err)

	expectedPower := delegateeStake.Add(del1Stake).Add(del2Stake).Add(del3Stake)
	assert.Equal(t, expectedPower, delegateePower)
}

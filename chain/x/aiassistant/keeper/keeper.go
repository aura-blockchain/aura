package keeper

import (
	"fmt"
	"sort"
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	metrics "github.com/hashicorp/go-metrics"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// Keeper maintains the AI assistant registry state.
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string

	bankKeeper types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, authority string, bankKeeper types.BankKeeper) Keeper {
	if cdc == nil {
		panic("aiassistant keeper requires codec")
	}
	if key == nil {
		panic("aiassistant keeper requires store key")
	}
	if authority == "" {
		authority = authtypes.NewModuleAddress(types.ModuleName).String()
	}
	if bankKeeper == nil {
		panic("aiassistant keeper requires bank keeper")
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   key,
		authority:  authority,
		bankKeeper: bankKeeper,
	}
}

func (k Keeper) RegisterAssistant(ctx sdk.Context, msg *types.MsgRegisterAssistant) (*types.Assistant, error) {
	if msg == nil {
		return nil, sdkerrors.Wrap(types.ErrAssistantExists, "message cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.AssistantAddress); err != nil {
		return nil, err
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return nil, err
	}
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.AssistantKey(msg.AssistantAddress)) {
		return nil, types.ErrAssistantExists
	}

	params := k.GetParams(ctx)
	if err := types.ValidateParams(params); err != nil {
		return nil, types.ErrInvalidParams.Wrap(err.Error())
	}

	stakeCoin, err := balanceToCoin(msg.Stake)
	if err != nil {
		return nil, err
	}
	if stakeCoin.Amount.IsZero() {
		return nil, types.ErrInsufficientStake
	}
	minStake, err := balanceToCoin(params.MinStake)
	if err != nil {
		return nil, err
	}
	if stakeCoin.Denom != minStake.Denom {
		return nil, sdkerrors.Wrapf(types.ErrInsufficientStake, "expected denom %s", minStake.Denom)
	}
	if stakeCoin.Amount.LT(minStake.Amount) {
		return nil, sdkerrors.Wrapf(types.ErrInsufficientStake, "minimum %s%s required", minStake.Amount, minStake.Denom)
	}

	sponsorshipCoin, err := balanceToCoin(msg.Sponsorship)
	if err != nil {
		return nil, err
	}

	normalizedLocales, err := normalizeLocales(msg.Locales, int(params.MaxLocales))
	if err != nil {
		return nil, err
	}

	toTransfer := sdk.NewCoins()
	if !stakeCoin.Amount.IsZero() {
		toTransfer = toTransfer.Add(stakeCoin)
	}
	if sponsorshipCoin.Amount.IsPositive() {
		if sponsorshipCoin.Denom != stakeCoin.Denom {
			return nil, sdkerrors.Wrapf(types.ErrInvalidLocale, "sponsorship denom mismatch %s", sponsorshipCoin.Denom)
		}
		toTransfer = toTransfer.Add(sponsorshipCoin)
	}

	ownerAddr, _ := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if !toTransfer.IsZero() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddr, types.ModuleName, toTransfer); err != nil {
			return nil, err
		}
	}

	assistant := &types.Assistant{
		AssistantAddress:   msg.AssistantAddress,
		OwnerAddress:       msg.OwnerAddress,
		Stake:              coinToBalance(stakeCoin),
		SponsorshipBalance: coinToBalance(sponsorshipCoin),
		Locales:            normalizedLocales,
		ModelHash:          strings.ToLower(msg.ModelHash),
		ApiKeyFingerprint:  msg.ApiKeyFingerprint,
		LastHeartbeat:      timestamppb.New(ctx.BlockTime()),
		Status:             types.AssistantStatus_ACTIVE,
	}

	k.setAssistant(ctx, assistant)
	telemetry.IncrCounterWithLabels(
		[]string{"aiassistant", "msg", "register"},
		float32(1),
		[]metrics.Label{
			telemetry.NewLabel("sponsor", msg.OwnerAddress),
			telemetry.NewLabel("assistant", msg.AssistantAddress),
		},
	)
	return assistant, nil
}

func (k Keeper) UpdateLocales(ctx sdk.Context, msg *types.MsgUpdateLocales) (*types.Assistant, error) {
	assistant, err := k.mustGetAssistant(ctx, msg.AssistantAddress)
	if err != nil {
		return nil, err
	}
	if assistant.OwnerAddress != msg.OwnerAddress {
		return nil, types.ErrUnauthorizedOperator
	}
	params := k.GetParams(ctx)
	newLocales, err := normalizeLocales(msg.Locales, int(params.MaxLocales))
	if err != nil {
		return nil, err
	}
	k.clearLocaleIndex(ctx, assistant)
	assistant.Locales = newLocales
	k.setAssistant(ctx, assistant)
	telemetry.SetGaugeWithLabels(
		[]string{"aiassistant", "locales", "configured"},
		float32(len(newLocales)),
		[]metrics.Label{telemetry.NewLabel("assistant", assistant.AssistantAddress)},
	)
	return assistant, nil
}

func (k Keeper) Heartbeat(ctx sdk.Context, msg *types.MsgHeartbeat) (uint64, error) {
	assistant, err := k.mustGetAssistant(ctx, msg.AssistantAddress)
	if err != nil {
		return 0, err
	}
	if assistant.Status != types.AssistantStatus_ACTIVE {
		return 0, sdkerrors.Wrapf(types.ErrAssistantNotFound, "assistant is not active: %s", assistant.Status.String())
	}

	if msg.OperatorAddress != assistant.OwnerAddress && msg.OperatorAddress != assistant.AssistantAddress {
		return 0, types.ErrUnauthorizedOperator
	}

	params := k.GetParams(ctx)
	window := time.Duration(params.HeartbeatWindowSeconds) * time.Second
	grace := time.Duration(params.HeartbeatGraceSeconds) * time.Second

	last := assistant.LastHeartbeat.AsTime()
	if ctx.BlockTime().After(last.Add(window + grace)) {
		assistant.HeartbeatFailures++
		if _, err := k.slashAssistant(ctx, assistant, params.SlashFractionDowntime); err != nil {
			return 0, err
		}
	}

	latency := ctx.BlockTime().Sub(last).Seconds()
	assistant.LastHeartbeat = timestamppb.New(ctx.BlockTime())
	k.setAssistant(ctx, assistant)

	telemetry.IncrCounterWithLabels(
		[]string{"aiassistant", "heartbeat", "success"},
		float32(1),
		[]metrics.Label{telemetry.NewLabel("assistant", assistant.AssistantAddress)},
	)
	if latency >= 0 {
		telemetry.SetGaugeWithLabels(
			[]string{"aiassistant", "heartbeat", "age_seconds"},
			float32(latency),
			[]metrics.Label{telemetry.NewLabel("assistant", assistant.AssistantAddress)},
		)
	}

	nextSlash := ctx.BlockTime().Add(window + grace).Unix()
	return uint64(nextSlash), nil
}

func (k Keeper) ReportMisbehavior(ctx sdk.Context, msg *types.MsgReportMisbehavior) (*types.Assistant, sdk.Coin, error) {
	assistant, err := k.mustGetAssistant(ctx, msg.AssistantAddress)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	assistant.MisbehaviorReports++
	slashed, err := k.slashAssistant(ctx, assistant, k.GetParams(ctx).SlashFractionMisbehavior)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	currentStake, err := balanceToCoin(assistant.Stake)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	if currentStake.Amount.IsZero() {
		assistant.Status = types.AssistantStatus_TOMBSTONED
	} else {
		assistant.Status = types.AssistantStatus_JAILED
	}
	k.setAssistant(ctx, assistant)
	telemetry.IncrCounterWithLabels(
		[]string{"aiassistant", "msg", "misbehavior"},
		float32(1),
		[]metrics.Label{telemetry.NewLabel("assistant", assistant.AssistantAddress)},
	)
	return assistant, slashed, nil
}

func (k Keeper) mustGetAssistant(ctx sdk.Context, addr string) (*types.Assistant, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AssistantKey(addr))
	if len(bz) == 0 {
		return nil, types.ErrAssistantNotFound
	}
	var assistant types.Assistant
	k.cdc.MustUnmarshal(bz, &assistant)
	return &assistant, nil
}

func (k Keeper) GetAssistant(ctx sdk.Context, addr string) (*types.Assistant, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AssistantKey(addr))
	if len(bz) == 0 {
		return nil, false
	}
	var assistant types.Assistant
	k.cdc.MustUnmarshal(bz, &assistant)
	return &assistant, true
}

func (k Keeper) ListAssistants(ctx sdk.Context, pageReq *query.PageRequest) ([]*types.Assistant, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

	var assistants []*types.Assistant
	pageRes, err := query.Paginate(prefixStore, pageReq, func(_, value []byte) error {
		var assistant types.Assistant
		if err := k.cdc.Unmarshal(value, &assistant); err != nil {
			return err
		}
		assistants = append(assistants, &assistant)
		return nil
	})
	return assistants, pageRes, err
}

func (k Keeper) AssistantsByLocale(ctx sdk.Context, locale string) ([]*types.Assistant, error) {
	normalized := normalizeLocaleString(locale)
	if normalized == "" {
		return nil, types.ErrInvalidLocale
	}

	store := storeprefix.NewStore(ctx.KVStore(k.storeKey), localePrefix(normalized))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	assistants := make([]*types.Assistant, 0)
	for ; iter.Valid(); iter.Next() {
		addr := string(iter.Value())
		if asst, ok := k.GetAssistant(ctx, addr); ok {
			assistants = append(assistants, asst)
		}
	}
	return assistants, nil
}

func (k Keeper) setAssistant(ctx sdk.Context, assistant *types.Assistant) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(assistant)
	store.Set(types.AssistantKey(assistant.AssistantAddress), bz)
	k.indexLocales(ctx, assistant)
}

func (k Keeper) indexLocales(ctx sdk.Context, assistant *types.Assistant) {
	store := ctx.KVStore(k.storeKey)
	for _, loc := range assistant.Locales {
		key := types.LocaleAssistantKey(loc, assistant.AssistantAddress)
		store.Set(key, []byte(assistant.AssistantAddress))
	}
}

func (k Keeper) clearLocaleIndex(ctx sdk.Context, assistant *types.Assistant) {
	store := ctx.KVStore(k.storeKey)
	for _, loc := range assistant.Locales {
		store.Delete(types.LocaleAssistantKey(loc, assistant.AssistantAddress))
	}
}

func (k Keeper) slashAssistant(ctx sdk.Context, assistant *types.Assistant, fractionStr string) (sdk.Coin, error) {
	currentStake, err := balanceToCoin(assistant.Stake)
	if err != nil {
		return sdk.Coin{}, err
	}
	if currentStake.Amount.IsZero() {
		return sdk.Coin{}, nil
	}
	fraction, err := sdkmath.LegacyNewDecFromStr(fractionStr)
	if err != nil {
		return sdk.Coin{}, err
	}
	if fraction.IsNegative() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidParams, "slash fraction negative")
	}
	if fraction.IsZero() {
		return sdk.Coin{}, nil
	}
	slashAmtDec := fraction.MulInt(currentStake.Amount)
	slashAmt := slashAmtDec.RoundInt()
	if slashAmt.IsZero() {
		return sdk.Coin{}, nil
	}
	slashCoin := sdk.NewCoin(currentStake.Denom, slashAmt)
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(slashCoin)); err != nil {
		return sdk.Coin{}, err
	}
	newStake := currentStake.Sub(slashCoin)
	assistant.Stake = coinToBalance(newStake)
	assistant.SlashCount++
	return slashCoin, nil
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.ParamsKey) {
		def := types.DefaultParams()
		_ = k.SetParams(ctx, def) // Default params should always be valid
		return def
	}
	var params types.Params
	k.cdc.MustUnmarshal(store.Get(types.ParamsKey), &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
	return nil
}

func normalizeLocales(locales []string, max int) ([]string, error) {
	if len(locales) == 0 {
		return nil, types.ErrInvalidLocale
	}
	if len(locales) > max {
		return nil, types.ErrLocaleCapacity
	}
	normalizedSet := make(map[string]struct{})
	for _, loc := range locales {
		normalized := normalizeLocaleString(loc)
		if normalized == "" {
			return nil, sdkerrors.Wrap(types.ErrInvalidLocale, loc)
		}
		normalizedSet[normalized] = struct{}{}
	}
	normalizedLocales := make([]string, 0, len(normalizedSet))
	for loc := range normalizedSet {
		normalizedLocales = append(normalizedLocales, loc)
	}
	sort.Strings(normalizedLocales)
	return normalizedLocales, nil
}

func normalizeLocaleString(loc string) string {
	loc = strings.TrimSpace(strings.ToLower(loc))
	if len(loc) < 2 || len(loc) > 8 {
		return ""
	}
	return loc
}

func localePrefix(locale string) []byte {
	buf := append(types.LocaleKeyPrefix, []byte(locale)...)
	buf = append(buf, byte(0x00))
	return buf
}

func balanceToCoin(balance *types.Balance) (sdk.Coin, error) {
	if balance == nil {
		return sdk.NewCoin(types.DefaultStakeDenom, sdkmath.ZeroInt()), nil
	}
	if balance.Denom == "" {
		return sdk.Coin{}, fmt.Errorf("balance denom required")
	}
	if balance.Amount == "" {
		return sdk.Coin{}, fmt.Errorf("balance amount required")
	}
	amount, ok := sdkmath.NewIntFromString(balance.Amount)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("invalid amount %s", balance.Amount)
	}
	return sdk.NewCoin(balance.Denom, amount), nil
}

func coinToBalance(coin sdk.Coin) *types.Balance {
	return &types.Balance{
		Denom:  coin.Denom,
		Amount: coin.Amount.String(),
	}
}

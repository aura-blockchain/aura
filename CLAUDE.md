# Aura Project Guidelines

**Read `../CLAUDE.md` first** - contains all general instructions.

## Project-Specific

**Node:** `~/.aura/` (not in repo)
**Binary:** `cd chain && go build -o aurad ./cmd/aurad`
**Init:** `./aurad init <moniker> --chain-id aura-testnet-1`
**Proto:** `make proto-gen` after modifying `.proto` files

## CRITICAL: Gogoproto Types

**Read `chain/docs/GOGOPROTO_TYPES.md` before writing tests.**

Common mistakes:
- `timestamppb.New()` → use `time.Now()`
- `"1000"` strings → use `sdkmath.NewInt(1000)`
- `&Type{}` pointers → use `Type{}` when `nullable=false`

Use helpers from `chain/testutil/proto_helpers.go`.
- NO LONG SUMMARIES FOR ANY WORK. KEEP ALL SUMMARIES UNDER 50 LINES. NO EXHAUSTIVE SUMMARIES. BE BRIEF AND TO THE POINT!!!!
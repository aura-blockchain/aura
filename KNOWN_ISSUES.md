# Known Issues

## SDK 0.53.4 Keyring Migration Bug

**Status:** BLOCKING - Prevents transaction signing on local testnet
**Severity:** Critical
**Affected Components:** All CLI commands requiring transaction signing

### Description

Cosmos SDK 0.53.x has a bug in the keyring migration system that prevents keys from being read after creation. The SDK stores keys in JWE (JSON Web Encryption) format in `.info` files, but the migration code attempts to parse them using proto/amino unmarshal functions.

### Error Message

```
migrate err for key <keyname>.info: unable to unmarshal item.Data: Bytes left over in UnmarshalBinaryLengthPrefixed, should read 10 more bytes but have 155
```

### Root Cause

The SDK's keyring implementation in v0.53.x:
1. Creates keys and stores them in JWE encrypted format
2. When listing/reading keys, attempts to migrate from legacy amino/proto format
3. JWE-encrypted data fails the proto unmarshal, causing the error

This affects all keyring backends: `test`, `file`, `os`, and `pass`.

### Impact

- `aurad keys list` - Fails
- `aurad keys show <key>` - Fails
- `aurad tx ... --from <key>` - Fails (cannot sign transactions)
- `aurad keys add <key>` - Works (shows address at creation time only)

### Workarounds Attempted

1. **Removing corrupted keyring and recreating keys** - Same error occurs
2. **Switching backends (test → file)** - Same error occurs
3. **Using SDK's standard `keys.Commands()`** - Same error occurs
4. **Renaming .info files** - SDK requires .info suffix

### Potential Workarounds

1. **Memory backend** - ❌ DOES NOT WORK - Each CLI command runs in a separate process with its own empty memory keyring
2. **Downgrade SDK** - Possible but would require significant work (SDK 0.50.x)
3. **Wait for SDK patch** - Monitor SDK releases for fix
4. **Use external signer** - Sign transactions with external tool
5. **Custom signing code** - Implement custom signing in Go that loads key from mnemonic directly (requires code changes)
6. **Use test backend with fresh keyring** - Clear existing keyring files and only read keys immediately after creation

### Implemented Solution

The `aurad mnemonic-signer` command bypasses the keyring entirely:

```bash
# Derive address from mnemonic (verify it's correct)
aurad mnemonic-signer derive-address --mnemonic-file ./deployer.mnemonic

# Store WASM contract (foundation implemented, full tx signing TBD)
aurad mnemonic-signer wasm-store ./contract.wasm --mnemonic-file ./deployer.mnemonic
```

**How it works:**
1. Reads mnemonic from file (or stdin)
2. Derives private key using HD path (m/44'/118'/0'/0/0)
3. Signs transactions in-memory without touching the keyring
4. Broadcasts via RPC

**Files:**
- `/chain/cmd/aurad/cmd/mnemonic_signer.go` - Mnemonic signing tool implementation

**Security Notes:**
- Store mnemonic file with restricted permissions (chmod 600)
- Do not pass mnemonic on command line (may be logged)
- This is for development/testing only

### References

- Cosmos SDK Issue Tracker: Check for issues related to keyring migration
- SDK Version: github.com/cosmos/cosmos-sdk v0.53.4
- Related: JWE format change in SDK keyring system

### Date Identified

2025-11-28

### Affected AURA Roadmap Items

- [ ] WASM contract deployment (Phase 1)
- [ ] DEX pool creation and swaps (Phase 1)
- [ ] Bridge cross-chain transfers (Phase 1)
- [ ] Any transaction requiring signing

---

## Future Issues

(None documented yet)

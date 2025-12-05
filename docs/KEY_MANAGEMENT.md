# AURA Key Management Guide

## Overview

This guide covers secure key management practices for AURA blockchain validators, operators, and developers. Key management is critical for blockchain security - compromised keys can lead to loss of funds, validator slashing, or unauthorized transactions.

## Table of Contents

- [Keyring Backends](#keyring-backends)
- [Key Generation](#key-generation)
- [Key Validation](#key-validation)
- [Environment Variables](#environment-variables)
- [Security Best Practices](#security-best-practices)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)

---

## Keyring Backends

AURA supports multiple keyring backends for different use cases:

### 1. OS Backend (Production/Mainnet)

**Description:** Uses the operating system's native keyring
- **macOS:** Keychain
- **Windows:** Credential Vault
- **Linux:** Secret Service API (libsecret)

**Security:** ✅ High - keys encrypted by OS

**Use Cases:**
- Mainnet validators
- Production deployments
- Long-term key storage

**Usage:**
```bash
export AURA_KEYRING_BACKEND=os
aurad keys add my-key --keyring-backend os
```

**Requirements:**
- Properly configured OS keyring
- User authentication
- May require additional setup on Linux

### 2. File Backend (Testnet)

**Description:** Encrypted file-based keyring with passphrase

**Security:** ✅ Medium-High - passphrase protected

**Use Cases:**
- Testnet validators
- Development with persistent keys
- Backup key storage

**Usage:**
```bash
export AURA_KEYRING_BACKEND=file
aurad keys add my-key --keyring-backend file
# You will be prompted for a passphrase
```

**Best Practices:**
- Use strong passphrases (20+ characters)
- Store passphrases in password manager
- Never commit encrypted keyring files to git

### 3. Test Backend (Local Development)

**Description:** Unencrypted file-based keyring

**Security:** ⚠️ Low - NO encryption

**Use Cases:**
- Local development
- Testing
- CI/CD pipelines
- Scripted operations

**Usage:**
```bash
export AURA_KEYRING_BACKEND=test
aurad keys add my-key --keyring-backend test
```

**⚠️ WARNING:** Never use test backend for mainnet or real funds!

### 4. Memory Backend (CI/CD)

**Description:** In-memory keyring (session only)

**Security:** ✅ High - ephemeral, no persistence

**Use Cases:**
- CI/CD pipelines
- Automated testing
- Single-use operations
- Scripts with environment-provided mnemonics

**Usage:**
```bash
export AURA_KEYRING_BACKEND=memory
echo "your mnemonic here" | aurad keys add my-key --recover --keyring-backend memory
```

**Benefits:**
- Keys never touch disk
- Automatic cleanup on exit
- Ideal for secrets from environment

---

## Key Generation

### Random Key Generation

Generate a new key with a random mnemonic:

```bash
aurad keys add validator-key --keyring-backend test
```

**Output:**
- Key name
- Address (aura1...)
- Public key
- 24-word mnemonic

**⚠️ CRITICAL:** Save the mnemonic securely! This is the ONLY way to recover the key.

### Deterministic Key Generation (Testing)

For reproducible testing environments, use the key management library:

```bash
source scripts/lib-key-management.sh
setup_keyring_backend test
generate_key_deterministic "validator1" "$HOME/.aura" "test-seed-1"
```

This generates the same key every time from the same seed - useful for:
- Automated testing
- CI/CD pipelines
- Reproducible testnet deployments

### Key Recovery

Recover a key from an existing mnemonic:

```bash
echo "your 24 word mnemonic here" | \
  aurad keys add recovered-key --recover --keyring-backend test
```

---

## Key Validation

### Validate Single Key

Check if a key exists in the keyring:

```bash
source scripts/lib-key-management.sh
validate_key_exists "validator-key" "$HOME/.aura" "test"
```

### Validate Multiple Keys

Validate a batch of keys:

```bash
source scripts/lib-key-management.sh
validate_keys_batch "$HOME/.aura" "test" validator-1 validator-2 validator-3
```

**Output:**
```
=== Validating Keys ===
Home: /home/user/.aura
Backend: test

✓ Key exists: validator-1
✓ Key exists: validator-2
✓ Key exists: validator-3

✓ All keys validated successfully
```

### Get Key Address

Retrieve the address for a key:

```bash
source scripts/lib-key-management.sh
address=$(get_key_address "validator-key" "$HOME/.aura" "test")
echo "Address: $address"
```

---

## Environment Variables

### Recommended Pattern

Load secrets from environment variables (never hardcode):

```bash
# Set environment variables
export AURA_KEYRING_BACKEND=test
export AURA_VALIDATOR_MNEMONIC="your 24 word mnemonic here"
export AURA_OPERATOR_MNEMONIC="your 24 word mnemonic here"

# Load secrets
source scripts/lib-key-management.sh
load_secrets_from_env
```

### Environment Variable Reference

| Variable | Description | Required |
|----------|-------------|----------|
| `AURA_KEYRING_BACKEND` | Keyring backend (os/file/test/memory) | Yes |
| `AURA_VALIDATOR_MNEMONIC` | Validator key mnemonic | For validators |
| `AURA_OPERATOR_MNEMONIC` | Operator key mnemonic | For operators |
| `AURA_HOME` | Node data directory | No (defaults to ~/.aura) |

### Using .env Files (Development Only)

For local development, you can use a `.env` file:

```bash
# .env (DO NOT COMMIT TO GIT!)
AURA_KEYRING_BACKEND=test
AURA_VALIDATOR_MNEMONIC="word1 word2 word3 ... word24"
```

Load it:
```bash
set -a  # Auto-export variables
source .env
set +a
```

**⚠️ CRITICAL:** Add `.env` to `.gitignore`!

---

## Security Best Practices

### 1. Key Storage

**✅ DO:**
- Use OS keyring for production
- Use hardware wallets (Ledger) for large stakes
- Store backup mnemonics offline (encrypted USB, paper wallet)
- Use multi-signature schemes for critical operations

**❌ DON'T:**
- Store keys in plain text
- Commit keys to git repositories
- Share keys via email/chat
- Use the same key across environments

### 2. Mnemonic Security

**✅ DO:**
- Write down mnemonics on paper
- Store in a safe or safety deposit box
- Use metal backups for long-term storage
- Verify backups after creation

**❌ DON'T:**
- Store mnemonics in cloud storage
- Take photos of mnemonics
- Store on internet-connected devices
- Share mnemonics with anyone

### 3. Access Control

**✅ DO:**
- Use separate keys for different roles (validator, operator, admin)
- Implement principle of least privilege
- Rotate keys periodically
- Audit key access logs

**❌ DON'T:**
- Share keys between multiple operators
- Use personal keys for production
- Grant unnecessary permissions

### 4. Key Rotation

Rotate keys periodically to limit exposure:

```bash
# Generate new key
aurad keys add validator-key-2 --keyring-backend os

# Update validator operator
aurad tx staking edit-validator \
  --new-operator-address $(aurad keys show validator-key-2 -a) \
  --from validator-key \
  --chain-id aura-mainnet-1

# After confirmation period, securely delete old key
source scripts/lib-key-management.sh
secure_delete ~/.aura/keyring-os/validator-key
```

### 5. Hardware Security Modules (HSM)

For maximum security, use HSMs:

**Supported:**
- YubiHSM2
- Ledger Nano S/X
- AWS CloudHSM
- Azure Key Vault

**Setup:** See `/docs/security/HSM_INTEGRATION.md`

---

## Common Operations

### List All Keys

```bash
aurad keys list --keyring-backend test
```

### Show Key Details

```bash
aurad keys show validator-key --keyring-backend test
```

### Export Key (Encrypted)

```bash
aurad keys export validator-key --keyring-backend test > key-backup.json
```

**⚠️ This exports an encrypted version. Still treat as sensitive!**

### Import Key

```bash
aurad keys import validator-key key-backup.json --keyring-backend test
```

### Delete Key

```bash
# Regular delete
aurad keys delete validator-key --keyring-backend test

# Secure delete (overwrite file)
source scripts/lib-key-management.sh
secure_delete ~/.aura/keyring-test/validator-key
```

### Sign Transaction

```bash
aurad tx bank send \
  validator-key \
  aura1recipientaddress \
  1000000uaura \
  --keyring-backend test \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-adjustment 1.3
```

---

## Troubleshooting

### Key Not Found

**Problem:** `aurad keys show my-key` returns "key not found"

**Solutions:**
1. Check keyring backend:
   ```bash
   aurad keys list --keyring-backend test
   aurad keys list --keyring-backend os
   aurad keys list --keyring-backend file
   ```

2. Check home directory:
   ```bash
   aurad keys list --home ~/.aura
   ```

3. Validate key existence:
   ```bash
   source scripts/lib-key-management.sh
   validate_key_exists "my-key" "$HOME/.aura" "test"
   ```

### Wrong Keyring Backend

**Problem:** Script can't find keys created with different backend

**Solution:** Always specify backend consistently:
```bash
export AURA_KEYRING_BACKEND=test
# Now all commands use test backend
```

### Permission Denied (OS Backend)

**Problem:** Can't access OS keyring

**Solutions:**
- **macOS:** Unlock Keychain, grant terminal access
- **Linux:** Install and configure libsecret:
  ```bash
  sudo apt-get install libsecret-1-0 libsecret-1-dev
  ```
- **Windows:** Run as administrator if needed

### Corrupted Keyring

**Problem:** Keyring file corrupted

**Solutions:**
1. Restore from backup:
   ```bash
   cp ~/.aura/keyring-test.backup ~/.aura/keyring-test
   ```

2. Recover from mnemonic:
   ```bash
   echo "your mnemonic" | \
     aurad keys add my-key --recover --keyring-backend test
   ```

### Memory Backend Not Persisting

**Problem:** Keys disappear after script exit

**Expected Behavior:** Memory backend is ephemeral by design

**Solution:** Use test/file/os backend for persistence

---

## Integration with Scripts

### Testnet Initialization

The testnet initialization script uses key management:

```bash
# Set backend
export AURA_KEYRING_BACKEND=test

# Run initialization
./scripts/testnet-init.sh

# Keys are validated automatically
```

### Transaction Scripts

Use the key management library in your scripts:

```bash
#!/bin/bash
source scripts/lib-key-management.sh

# Setup
setup_keyring_backend "test"

# Validate key exists
if ! validate_key_exists "validator-key" "$HOME/.aura" "test"; then
    echo "Error: Key not found"
    exit 1
fi

# Get address
address=$(get_key_address "validator-key" "$HOME/.aura" "test")

# Use in transaction
aurad tx bank send \
  validator-key \
  aura1recipient \
  1000uaura \
  --from "$address" \
  --keyring-backend test
```

---

## HSM Integration

For production validators, integrate with Hardware Security Modules:

### YubiHSM2

```bash
# Setup YubiHSM connector
yubihsm-connector -d

# Generate key on HSM
aurad keys add validator-key \
  --ledger \
  --keyring-backend os
```

### Ledger Hardware Wallet

```bash
# Connect Ledger device
# Open Cosmos app on device

# Add Ledger key
aurad keys add validator-key \
  --ledger \
  --keyring-backend os

# Sign transactions via Ledger
aurad tx staking create-validator \
  --from validator-key \
  --ledger \
  --keyring-backend os \
  ...
```

See `/docs/security/HSM_INTEGRATION.md` for complete setup.

---

## Key Management Checklist

### Development
- [ ] Use `test` backend for local development
- [ ] Keys stored in `~/.aura/keyring-test/`
- [ ] Keys validated in initialization scripts
- [ ] Mnemonics saved for recovery (testing only)

### Testnet
- [ ] Use `file` backend with strong passphrase
- [ ] Backup encrypted keyring files
- [ ] Mnemonics stored offline (encrypted)
- [ ] Key rotation plan documented
- [ ] Access control policies defined

### Production/Mainnet
- [ ] Use `os` backend or HSM
- [ ] Multi-signature for critical operations
- [ ] Hardware wallet for operator keys
- [ ] Mnemonics stored in safe/vault
- [ ] Key rotation schedule (quarterly)
- [ ] Incident response plan for key compromise
- [ ] Insurance/bonding for validator stake

---

## References

- [Cosmos SDK Keyring Documentation](https://docs.cosmos.network/main/run-node/keyring)
- [Key Management Library](/scripts/lib-key-management.sh)
- [HSM Integration Guide](/docs/security/HSM_INTEGRATION.md)
- [Testnet Initialization Script](/scripts/testnet-init.sh)
- [OWASP Key Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Key_Management_Cheat_Sheet.html)

---

## Support

For questions or issues:
- Review this documentation
- Check troubleshooting section
- Review script source code in `/scripts/lib-key-management.sh`
- Consult Cosmos SDK documentation

**⚠️ NEVER share private keys or mnemonics when seeking help!**

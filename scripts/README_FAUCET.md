# Faucet Setup Scripts

This directory contains scripts for deploying and managing the AURA testnet faucet.

## Scripts

### faucet-setup.sh

**Purpose**: Creates and funds a faucet wallet for the testnet.

**Prerequisites**:
- Testnet must be running (`docker-compose -f docker-compose.testnet.yml up -d`)
- `aurad` binary must be built
- `jq` command-line tool installed (for JSON parsing)

**What it does**:
1. Verifies prerequisites (binary, testnet running)
2. Creates a new faucet account with keyring backend
3. Funds the faucet with 100,000,000 AURA from validator-1
4. Generates `.env.faucet` configuration file
5. Displays setup instructions and next steps

**Usage**:
```bash
./scripts/faucet-setup.sh
```

**Output**:
- Faucet wallet created in `testnet-data/faucet/`
- Mnemonic saved to `testnet-data/faucet/faucet.mnemonic`
- Configuration saved to `.env.faucet`
- Funding transaction saved to `testnet-data/faucet/funding-tx.json`

**Security Note**:
- The faucet mnemonic is stored unencrypted in `testnet-data/faucet/faucet.mnemonic`
- This is acceptable for local testnet, but **NEVER** use this approach for mainnet
- Add `testnet-data/` to `.gitignore` to prevent committing secrets

## Environment Variables

The `faucet-setup.sh` script respects these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `AURA_KEYRING_BACKEND` | `test` | Keyring backend (test/os/file) |

**Example with custom keyring**:
```bash
AURA_KEYRING_BACKEND=file ./scripts/faucet-setup.sh
```

## Troubleshooting

### Script fails with "testnet not initialized"

**Solution**: Initialize and start the testnet first:
```bash
./scripts/testnet-init.sh
docker build -t aurad:latest -f docker/Dockerfile.testnet .
docker-compose -f docker-compose.testnet.yml up -d
```

### Script fails with "jq: command not found"

**Solution**: Install jq:
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq

# Alpine (in Docker)
apk add jq
```

### Faucet key already exists

The script will detect existing keys and prompt you to:
- Use the existing key (preserves existing balance)
- Delete and create new key (starts fresh)

If you choose to use the existing key, make sure the mnemonic file exists at `testnet-data/faucet/faucet.mnemonic`.

### Funding transaction fails

**Possible causes**:
1. Validator-1 doesn't have sufficient balance
2. Chain is not synced yet
3. Gas prices are incorrect

**Solution**:
```bash
# Check validator-1 balance
docker exec aura-validator-1 aurad query bank balances \
  $(docker exec aura-validator-1 aurad keys show validator-1 --keyring-backend test --address) \
  --chain-id aura-local-4

# Check chain status
docker exec aura-validator-1 aurad status

# View detailed logs
docker logs aura-validator-1
```

## Manual Faucet Creation (Alternative)

If the automated script doesn't work, you can create a faucet wallet manually:

```bash
# 1. Create faucet key
cd chain
mkdir -p ../testnet-data/faucet
echo "password123" | ./aurad keys add faucet \
  --home ../testnet-data/faucet \
  --keyring-backend test \
  --output json > ../testnet-data/faucet/faucet-key.json

# 2. Extract address and mnemonic
FAUCET_ADDRESS=$(jq -r '.address' ../testnet-data/faucet/faucet-key.json)
FAUCET_MNEMONIC=$(jq -r '.mnemonic' ../testnet-data/faucet/faucet-key.json)

# 3. Save mnemonic
echo "${FAUCET_MNEMONIC}" > ../testnet-data/faucet/faucet.mnemonic
chmod 600 ../testnet-data/faucet/faucet.mnemonic

# 4. Get validator address
VALIDATOR_ADDRESS=$(docker exec aura-validator-1 \
  aurad keys show validator-1 --keyring-backend test --address)

# 5. Fund faucet
docker exec aura-validator-1 bash -c "
  echo 'password123' | aurad tx bank send ${VALIDATOR_ADDRESS} ${FAUCET_ADDRESS} \
    100000000000000uaura \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --gas 200000 \
    --gas-prices 0.025uaura \
    --yes
"

# 6. Wait and verify
sleep 6
docker exec aura-validator-1 aurad query bank balances ${FAUCET_ADDRESS} \
  --chain-id aura-local-4

# 7. Create .env.faucet (manually populate FAUCET_MNEMONIC and FAUCET_ADDRESS)
cat > ../.env.faucet <<EOF
FAUCET_MNEMONIC="${FAUCET_MNEMONIC}"
FAUCET_ADDRESS="${FAUCET_ADDRESS}"
FAUCET_AMOUNT_PER_REQUEST=100000000
FAUCET_RATE_LIMIT_IP=20
FAUCET_RATE_LIMIT_ADDR=3
FAUCET_RATE_WINDOW=24
FAUCET_DB_PASSWORD=faucet_secure_password
EOF

chmod 600 ../.env.faucet
```

## Integration with Testnet

The faucet connects to the testnet via:

1. **Docker Network**: `aura_aura-testnet` (external network)
2. **RPC Endpoint**: `http://aura-validator-1:26657`
3. **Chain ID**: `aura-local-4`
4. **Denom**: `uaura`

The faucet backend container joins the `aura_aura-testnet` network, allowing it to communicate with validator nodes directly.

## Next Steps After Setup

Once `faucet-setup.sh` completes successfully:

1. **Review configuration**:
   ```bash
   cat .env.faucet
   ```

2. **Start faucet service**:
   ```bash
   docker-compose -f docker-compose.faucet.yml --env-file .env.faucet up -d
   ```

3. **Verify services**:
   ```bash
   docker-compose -f docker-compose.faucet.yml ps
   ```

4. **Test faucet**:
   ```bash
   curl http://localhost:8081/api/v1/health
   curl http://localhost:8081/api/v1/faucet/info
   ```

5. **Access Web UI**:
   ```
   http://localhost:8081
   ```

## Security Best Practices

1. **Never commit secrets**:
   - Add `.env.faucet` to `.gitignore`
   - Add `testnet-data/faucet/` to `.gitignore`

2. **Use secure passwords in production**:
   - Generate strong database passwords
   - Use `file` or `os` keyring backend (not `test`)

3. **Limit access**:
   - Use firewall rules to restrict access to faucet services
   - Enable CAPTCHA for public deployments

4. **Monitor activity**:
   - Review faucet logs regularly
   - Set up alerts for unusual activity
   - Monitor faucet balance

5. **Testnet only**:
   - These scripts are for TESTNET use only
   - Never use testnet mnemonics on mainnet
   - Use different security measures for mainnet

## Support

For detailed documentation:
- [FAUCET_DEPLOYMENT.md](../FAUCET_DEPLOYMENT.md) - Comprehensive deployment guide
- [FAUCET_QUICK_START.md](../FAUCET_QUICK_START.md) - Quick start guide
- [faucet-service/README.md](../faucet-service/README.md) - Backend documentation

For issues:
- Check testnet is running: `docker-compose -f docker-compose.testnet.yml ps`
- Check logs: `docker-compose -f docker-compose.faucet.yml logs -f`
- Verify network: `docker network inspect aura_aura-testnet`

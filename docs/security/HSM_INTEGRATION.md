# HSM Integration Guide for AURA Validators

This guide covers Hardware Security Module (HSM) integration for AURA blockchain validators to securely manage validator signing keys.

## Overview

HSMs provide hardware-based key protection for validator consensus keys, preventing key extraction even if the server is compromised. This is critical for production validators managing significant stake.

## Supported HSM Solutions

### 1. YubiHSM 2 (Recommended for Small-Medium Validators)

**Cost:** ~$650 USD
**Interface:** USB, supports PKCS#11

#### Installation

```bash
# Install YubiHSM SDK (Ubuntu/Debian)
wget https://developers.yubico.com/YubiHSM2/Releases/yubihsm2-sdk-2023.01-ubuntu2204-amd64.tar.gz
tar -xzf yubihsm2-sdk-*.tar.gz
sudo dpkg -i yubihsm2-sdk/*.deb

# Start the connector service
sudo systemctl enable yubihsm-connector
sudo systemctl start yubihsm-connector

# Verify connection
yubihsm-shell
> connect
> session open 1 password
> list objects 0
```

#### Configure with AURA

```bash
# Generate validator key in HSM
yubihsm-shell -a generate-asymmetric-key \
  -i 0x0001 \
  -l "aura-validator-key" \
  -d "1" \
  -c "sign-eddsa" \
  -A "ed25519"

# Configure aurad to use HSM
cat >> ~/.aura/config/config.toml << EOF

[priv_validator]
# Use external signer via PKCS#11
laddr = "tcp://127.0.0.1:26658"
EOF

# Start tmkms (Tendermint KMS) with YubiHSM
tmkms init /etc/tmkms
```

#### tmkms Configuration

```toml
# /etc/tmkms/tmkms.toml
[[chain]]
id = "aura-mainnet-1"
key_format = { type = "bech32", account_key_prefix = "aurapub", consensus_key_prefix = "auravalconspub" }
state_file = "/var/lib/tmkms/aura-mainnet-1-state.json"

[[validator]]
chain_id = "aura-mainnet-1"
addr = "tcp://127.0.0.1:26658"
secret_key = "/etc/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"

[[providers.yubihsm]]
adapter = { type = "usb" }
auth = { key = 1, password_file = "/etc/tmkms/secrets/yubihsm-password" }
keys = [{ chain_ids = ["aura-mainnet-1"], key = 1 }]
serial_number = "YOUR_YUBIHSM_SERIAL"
```

### 2. AWS CloudHSM (Recommended for Enterprise/Cloud Deployments)

**Cost:** ~$1.60/hour per HSM + network costs
**Compliance:** FIPS 140-2 Level 3

#### Setup

```bash
# Install CloudHSM client
wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/EL7/cloudhsm-client-latest.el7.x86_64.rpm
sudo yum install -y ./cloudhsm-client-latest.el7.x86_64.rpm

# Configure client
sudo /opt/cloudhsm/bin/configure -a <HSM_IP_ADDRESS>

# Initialize HSM (first time only)
/opt/cloudhsm/bin/cloudhsm_mgmt_util /opt/cloudhsm/etc/cloudhsm_mgmt_util.cfg
> loginHSM PRECO admin password
> changePswd PRECO admin <new_password>
> createUser CU validator_user <password>
```

#### Generate Validator Key

```bash
# Using PKCS#11 interface
/opt/cloudhsm/bin/key_mgmt_util
> loginHSM -u CU -s validator_user -p <password>
> genECCKeyPair -i 25519 -l aura_validator_key
> getAttribute -o <key_handle> -a 512
```

### 3. HashiCorp Vault (Software-Based Alternative)

**Cost:** Open source (Enterprise features paid)
**Use Case:** Development, staging, or when HSM hardware unavailable

#### Setup

```bash
# Install Vault
wget https://releases.hashicorp.com/vault/1.15.0/vault_1.15.0_linux_amd64.zip
unzip vault_*.zip
sudo mv vault /usr/local/bin/

# Initialize and unseal
vault operator init -key-shares=5 -key-threshold=3
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>

# Enable transit secrets engine
vault secrets enable transit
vault write transit/keys/aura-validator type=ed25519
```

#### Integration with aurad

```bash
# Store validator key in Vault
vault write transit/keys/aura-validator/config min_decryption_version=1 min_encryption_version=1

# Configure remote signer
export VAULT_ADDR="https://vault.example.com:8200"
export VAULT_TOKEN="<token>"
```

## Tendermint KMS (tmkms) Setup

tmkms is the recommended solution for connecting HSMs to CometBFT/Tendermint validators.

### Installation

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Build tmkms with appropriate features
cargo install tmkms --features=yubihsm  # For YubiHSM
# OR
cargo install tmkms --features=softsign  # For software signing (testing only)
```

### Configuration

```toml
# /etc/tmkms/tmkms.toml

# Global settings
log_level = "info"
log_format = "json"

# Chain configuration
[[chain]]
id = "aura-mainnet-1"
key_format = { type = "bech32", account_key_prefix = "aurapub", consensus_key_prefix = "auravalconspub" }
state_file = "/var/lib/tmkms/state/aura-mainnet-1-consensus.json"

# Validator connection
[[validator]]
chain_id = "aura-mainnet-1"
addr = "tcp://validator.internal:26658"
secret_key = "/etc/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true

# HSM provider (choose one)
[[providers.yubihsm]]
adapter = { type = "usb" }
auth = { key = 1, password_file = "/etc/tmkms/secrets/yubihsm-password" }
keys = [
  { chain_ids = ["aura-mainnet-1"], key = 1, type = "consensus" }
]
```

### Running tmkms

```bash
# Create systemd service
sudo cat > /etc/systemd/system/tmkms.service << EOF
[Unit]
Description=Tendermint KMS
After=network.target

[Service]
Type=simple
User=tmkms
ExecStart=/usr/local/bin/tmkms start -c /etc/tmkms/tmkms.toml
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable tmkms
sudo systemctl start tmkms
```

## Validator Node Configuration

Configure your AURA validator to use the remote signer:

```toml
# ~/.aura/config/config.toml

[priv_validator]
# Disable local file-based signing
key = ""
state = ""

# Enable remote signer
laddr = "tcp://0.0.0.0:26658"

# Bind to internal network only
# Use firewall to restrict access to tmkms host only
```

## Key Backup and Recovery

### YubiHSM Backup

```bash
# Export wrapped backup key
yubihsm-shell
> session open 1 <password>
> get-wrapped --wrap-id 0x0002 --object-id 0x0001 --object-type asymmetric-key -o validator-backup.enc

# Store backup securely (offline, multiple locations)
# Never store unwrapped keys!
```

### Recovery Procedure

1. Initialize new HSM with same wrap key
2. Import wrapped backup: `put-wrapped -i 0x0001 -w validator-backup.enc`
3. Verify key: `get-public-key -i 0x0001`
4. Update tmkms configuration with new HSM serial
5. Restart tmkms service

## Security Best Practices

### Physical Security

- [ ] HSM stored in locked server cabinet
- [ ] Access logging enabled on physical access
- [ ] Tamper-evident seals on HSM ports
- [ ] Separate HSM from validator in different physical hosts

### Network Security

- [ ] tmkms and validator communicate over private network
- [ ] Firewall rules restrict HSM access to tmkms only
- [ ] TLS encryption for remote HSM connections
- [ ] No HSM management interface exposed to internet

### Operational Security

- [ ] Multi-party authentication for HSM admin operations
- [ ] Regular key rotation schedule (annual recommended)
- [ ] Audit logs exported to SIEM system
- [ ] Incident response plan documented

### Monitoring

```bash
# Monitor tmkms logs for signing events
journalctl -u tmkms -f | grep -E "(signed|error|warning)"

# Alert on signing failures
# Add to monitoring system (Prometheus/Grafana)
```

## Troubleshooting

### tmkms Connection Issues

```bash
# Check validator is listening
ss -tlnp | grep 26658

# Test connectivity from tmkms host
nc -zv validator.internal 26658

# Check tmkms logs
journalctl -u tmkms -n 100 --no-pager
```

### HSM Not Responding

```bash
# YubiHSM: Check USB connection
lsusb | grep Yubico

# Restart connector
sudo systemctl restart yubihsm-connector

# CloudHSM: Check network and credentials
/opt/cloudhsm/bin/cloudhsm_mgmt_util /opt/cloudhsm/etc/cloudhsm_mgmt_util.cfg
> getHSMInfo
```

### Double Signing Prevention

The state file prevents double signing. **NEVER**:
- Run multiple tmkms instances with same state file
- Copy state file between environments
- Delete state file without proper procedure

```bash
# Verify state file integrity
cat /var/lib/tmkms/state/aura-mainnet-1-consensus.json
# Should show last signed height/round/step
```

## Cost Comparison

| Solution | Initial Cost | Monthly Cost | Compliance | Recommended For |
|----------|-------------|--------------|------------|-----------------|
| YubiHSM 2 | $650 | $0 | FIPS 140-2 L2 | Independent validators |
| AWS CloudHSM | $5,000 setup | $1,200+ | FIPS 140-2 L3 | Enterprise, exchanges |
| Azure Dedicated HSM | $4,500 setup | $4,500+ | FIPS 140-2 L3 | Azure-native deployments |
| Vault (Enterprise) | $0 | Variable | SOC 2 | Development, cost-sensitive |

## References

- [Tendermint KMS Documentation](https://github.com/iqlusioninc/tmkms)
- [YubiHSM 2 Documentation](https://developers.yubico.com/YubiHSM2/)
- [AWS CloudHSM User Guide](https://docs.aws.amazon.com/cloudhsm/)
- [CometBFT Remote Signer](https://docs.cometbft.com/v0.38/core/validators)
- [AURA Validator Security Guidelines](/docs/ops/SECURITY_HARDENING.md)

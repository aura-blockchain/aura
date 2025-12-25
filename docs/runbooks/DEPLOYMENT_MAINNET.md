# Mainnet Deployment Runbook

## Prerequisites
- Enterprise server (32GB RAM, 8 CPU, 1TB NVMe SSD)
- Redundant network connectivity
- HSM for validator keys (recommended)
- 24/7 monitoring and alerting

## Pre-Deployment Checklist
- [ ] Security audit completed
- [ ] Load testing passed
- [ ] Backup procedures tested
- [ ] Incident response plan documented
- [ ] Team on-call schedule established

## Deployment Steps

### 1. Secure Server Setup
```bash
# Harden SSH
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Configure firewall
sudo ufw default deny incoming
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 26656/tcp # P2P
sudo ufw allow 26657/tcp # RPC (restrict to known IPs)
sudo ufw enable
```

### 2. Install with Cosmovisor
```bash
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

export DAEMON_NAME=aurad
export DAEMON_HOME=$HOME/.aura

cosmovisor init /usr/local/bin/aurad
```

### 3. Configure for Production
```bash
# config.toml optimizations
sed -i 's/max_num_inbound_peers = 40/max_num_inbound_peers = 100/' ~/.aura/config/config.toml
sed -i 's/max_num_outbound_peers = 10/max_num_outbound_peers = 50/' ~/.aura/config/config.toml
```

### 4. Systemd with Cosmovisor
```bash
sudo tee /etc/systemd/system/aurad.service << EOF
[Unit]
Description=Aura Mainnet
After=network.target

[Service]
User=$USER
ExecStart=$(which cosmovisor) run start
Restart=always
Environment="DAEMON_NAME=aurad"
Environment="DAEMON_HOME=$HOME/.aura"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```

## Post-Deployment
- Configure monitoring (see TESTNET_MONITORING.md)
- Set up log aggregation
- Test failover procedures
- Document recovery procedures

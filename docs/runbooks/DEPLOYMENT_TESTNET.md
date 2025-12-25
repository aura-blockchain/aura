# Testnet Deployment Runbook

## Prerequisites
- Production-grade server (16GB RAM, 4 CPU, 500GB SSD)
- Domain with SSL certificate
- Firewall configured (ports 26656, 26657, 9090, 1317)

## Deployment Steps

### 1. Server Setup
```bash
# Install dependencies
sudo apt update && sudo apt install -y build-essential git

# Install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. Build Binary
```bash
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain
make build
sudo cp build/aurad /usr/local/bin/
```

### 3. Initialize Node
```bash
aurad init <moniker> --chain-id aura-testnet-1
curl -o ~/.aura/config/genesis.json https://testnet.aura.network/genesis.json
```

### 4. Configure Systemd
```bash
sudo tee /etc/systemd/system/aurad.service << EOF
[Unit]
Description=Aura Daemon
After=network.target

[Service]
User=$USER
ExecStart=/usr/local/bin/aurad start
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable aurad
sudo systemctl start aurad
```

### 5. Verify
```bash
aurad status
curl localhost:26657/status
```

## Monitoring
- Prometheus metrics: port 26660
- See TESTNET_MONITORING.md for Grafana setup

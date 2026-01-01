# AURA Network Infrastructure

This document provides transparency into the infrastructure supporting the AURA blockchain network. We believe in open communication with our community about how the network is operated and secured.

## Philosophy

AURA follows blockchain community best practices for infrastructure:

- **Dedicated bare-metal servers** - No shared virtualization or cloud instances for validator nodes
- **Geographic distribution** - Nodes distributed across multiple data centers and regions
- **Independent operation** - Each blockchain in our ecosystem runs on its own dedicated infrastructure
- **Transparency** - Open documentation of our infrastructure choices and security practices

## Current Testnet Infrastructure

### Validator Node

| Specification | Details |
|--------------|---------|
| **Server Type** | Dedicated bare-metal server |
| **Provider** | OVH (SoYouStart) |
| **Location** | Beauharnois, Quebec, Canada (BHS) |
| **CPU** | Intel Xeon E3-1270 v6 @ 3.80GHz (4 cores / 8 threads) |
| **Memory** | 64 GB DDR4 ECC |
| **Storage** | 2x 480GB SSD (RAID 1) |
| **Network** | 500 Mbps dedicated bandwidth |
| **IPv4** | Dedicated static IP |

### Why Bare-Metal?

We chose dedicated bare-metal servers over cloud instances for several reasons:

1. **Predictable Performance** - No noisy neighbor issues or resource contention
2. **Security** - Full control over the hardware stack with no hypervisor layer
3. **Cost Efficiency** - Better price-to-performance for long-running blockchain nodes
4. **Community Trust** - Bare-metal is the accepted standard for serious blockchain infrastructure

### Network Architecture

```
                    ┌─────────────────┐
                    │   Cloudflare    │
                    │   (DDoS/CDN)    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │     Nginx       │
                    │  (Reverse Proxy)│
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐   ┌───────▼───────┐   ┌───────▼───────┐
│   RPC (26657) │   │  REST (1317)  │   │  gRPC (9090)  │
└───────────────┘   └───────────────┘   └───────────────┘
```

### Public Endpoints

| Service | URL | Port |
|---------|-----|------|
| RPC | https://testnet-rpc.aurablockchain.org | 443 |
| REST API | https://testnet-api.aurablockchain.org | 443 |
| gRPC | testnet-rpc.aurablockchain.org | 9090 |
| Explorer | https://testnet-explorer.aurablockchain.org | 443 |
| Faucet | https://testnet-faucet.aurablockchain.org | 443 |
| Monitoring | https://monitoring.aurablockchain.org | 443 |

### Security Measures

- **Firewall** - UFW with strict ingress rules; only required ports exposed
- **SSH** - Key-based authentication only, no password login
- **TLS** - All public endpoints secured with Let's Encrypt certificates
- **Updates** - Automated security updates via unattended-upgrades
- **Monitoring** - 24/7 monitoring via Grafana with alerting

## Mainnet Plans

For mainnet launch, we plan to expand infrastructure with:

- **Multiple validator nodes** across different geographic regions (NA, EU, Asia)
- **Sentry node architecture** to protect validator nodes from direct exposure
- **Multiple hosting providers** to avoid single provider dependency
- **Hardware Security Modules (HSM)** for validator key protection
- **Independent snapshot providers** for quick node bootstrapping

## Running Your Own Node

We encourage community members to run their own nodes. See our documentation:

- [Getting Started Guide](docs/GETTING_STARTED.md)
- [Validator Hardware Requirements](docs/VALIDATOR_HARDWARE_REQUIREMENTS.md)
- [Network Parameters](docs/NETWORK_PARAMETERS.md)

### Minimum Requirements for Validators

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores | 8+ cores |
| Memory | 16 GB | 32+ GB |
| Storage | 500 GB SSD | 1 TB NVMe |
| Network | 100 Mbps | 1 Gbps |

## Contact

For infrastructure-related inquiries:
- Email: info@aurablockchain.org
- Discord: [Join our community](https://discord.gg/aura)
- GitHub: [aura-blockchain/aura](https://github.com/aura-blockchain/aura)

## Changelog

| Date | Change |
|------|--------|
| 2025-01-01 | Initial testnet infrastructure deployed |

---

*Last updated: January 2025*

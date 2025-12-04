# Aura 4-Node Testnet - Quick Start Guide

## Status: Docker Daemon Required

**Current Issue**: Docker daemon is not running on this WSL2 system.

## Step 1: Start Docker

Run this command:

```bash
sudo systemctl start docker
```

Verify Docker is running:

```bash
docker ps
```

Expected output: Container listing (empty is OK).

## Step 2: Launch Testnet (Automated)

```bash
cd /home/decri/blockchain-projects/aura
./launch-testnet.sh
```

This script will automatically:
1. Verify Docker is running
2. Populate validator volumes
3. Start all 4 nodes
4. Wait for initialization
5. Verify health
6. Display status

**Total time**: ~2 minutes

## Alternative: Manual Launch

```bash
# 1. Start Docker
sudo systemctl start docker

# 2. Navigate to project
cd /home/decri/blockchain-projects/aura

# 3. Populate volumes
cd testnet-data
./populate-volumes.sh
cd ..

# 4. Start testnet
./scripts/testnet-manage.sh start

# 5. Wait for initialization
sleep 90

# 6. Check status
docker ps
./scripts/testnet-monitor.sh quick
```

## Testnet Configuration

| Node   | RPC Port | P2P Port | Container     |
|--------|----------|----------|---------------|
| Node 1 | 27657    | 27656    | aura-node1    |
| Node 2 | 27757    | 27756    | aura-node2    |
| Node 3 | 27857    | 27856    | aura-node3    |
| Node 4 | 27957    | 27956    | aura-node4    |

**Chain ID**: aura-testnet-1  
**Consensus**: Tendermint BFT (2/3+ voting power required)

## RPC Endpoints

```
http://localhost:27657  # Node 1
http://localhost:27757  # Node 2
http://localhost:27857  # Node 3
http://localhost:27957  # Node 4
```

## Management Commands

```bash
# Monitor testnet
./scripts/testnet-monitor.sh
./scripts/testnet-monitor.sh quick

# Stop testnet
./scripts/testnet-manage.sh stop

# Restart testnet
./scripts/testnet-manage.sh restart

# View logs
docker logs -f aura-node1
docker logs -f aura-node2

# Container status
docker ps
docker stats
```

## Health Checks

```bash
# Quick health check
for port in 27657 27757 27857 27957; do
  curl -s "http://localhost:$port/health"
  echo ""
done

# Node status
curl -s http://localhost:27657/status | jq .

# Net info (peers)
curl -s http://localhost:27657/net_info | jq .result.peers
```

## Troubleshooting

### Docker not starting
```bash
sudo systemctl status docker
journalctl -u docker -n 50
```

### Containers not starting
```bash
docker logs aura-node1
docker-compose -f docker-compose.yml ps
```

### Port conflicts
```bash
netstat -tulpn | grep -E "276[5-9]7"
```

### Consensus stalled
```bash
./scripts/testnet-monitor.sh
docker logs aura-node1 | grep -i error
```

## System Requirements

**Minimum**:
- CPU: 4 cores
- RAM: 2 GB
- Disk: 1 GB

**Current System**: ✓ Exceeds requirements
- CPU: 12 cores (AMD Ryzen AI 5 340)
- RAM: 7.3 GB (4.4 GB available)
- Disk: 780 GB available

## Resource Usage (Expected)

- **CPU**: 20-40% total (~5-10% per node)
- **Memory**: ~1 GB total (~200-300 MB per node)
- **Disk**: ~200 MB total (~50 MB per node)
- **Network**: Minimal (local only)

## Performance Metrics

- **Block Time**: 5-6 seconds
- **TPS**: 100-500 tx/sec (per node)
- **Startup**: 90-120 seconds
- **Latency**: <100ms (local)

## Comparison: Single vs Multi-Node

| Feature             | Single Node | 4-Node Testnet |
|---------------------|-------------|----------------|
| Consensus Testing   | No          | Yes            |
| Fault Tolerance     | None        | 1 node failure |
| Network Simulation  | No          | Yes            |
| Production-like     | 20%         | 90%            |
| Startup Time        | 10 sec      | 90 sec         |

## Next Steps After Launch

1. **Verify consensus**: All nodes should show same block height
2. **Test transactions**: Submit test txs via RPC
3. **Monitor performance**: Watch block production
4. **Test scenarios**: Network partitions, node failures
5. **Load testing**: Stress test with high tx volume

## Files Created

- `/home/decri/blockchain-projects/aura/launch-testnet.sh` - Automated launcher
- `/tmp/testnet-launch-report.md` - Detailed instructions
- `/tmp/TESTNET_LAUNCH_SUMMARY.txt` - Full summary
- `/tmp/resource-assessment.txt` - Resource analysis

## Support

For issues or questions:
1. Check Docker logs: `docker logs aura-node1`
2. Check testnet monitor: `./scripts/testnet-monitor.sh`
3. Review system resources: `docker stats`
4. Check documentation in `/home/decri/blockchain-projects/aura/docs/`

---

**Ready to start?**

```bash
sudo systemctl start docker && ./launch-testnet.sh
```

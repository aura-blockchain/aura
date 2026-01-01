# Upgrade Procedures

## Manual upgrade (basic)

1. Stop the node.
2. Build or install the new `aurad` binary.
3. Replace the running binary.
4. Restart the node and monitor logs.

```bash
cd chain
make build
sudo systemctl restart aurad
```

## Notes

- Follow the on-chain upgrade proposal timing if the upgrade uses governance.
- If you use Cosmovisor, follow your existing service configuration and replace the binary in the upgrade directory.

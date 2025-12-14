# Quick Start: Building and Installing Aura Wallet .deb Package

## Build the Package

```bash
# Build React frontend
npm run build:react

# Build .deb package
npm run build:linux
```

**Output**: `dist/aura-desktop-wallet_1.0.0_amd64.deb` (~279 MB)

## Verify the Package

```bash
./verify-deb.sh
```

## Install

```bash
sudo dpkg -i dist/aura-desktop-wallet_1.0.0_amd64.deb
sudo apt-get install -f  # If dependencies are missing
```

## Launch

Choose any method:

1. **Application Menu**: Search for "Aura Wallet"
2. **Terminal**: Run `aura-wallet`
3. **Direct**: `/opt/Aura\ Wallet/aura-desktop-wallet`

## Uninstall

```bash
sudo apt remove aura-desktop-wallet
```

## Documentation

- `DEB_PACKAGING.md` - Complete packaging guide
- `PACKAGING_SUMMARY.md` - Implementation details
- `build/README.md` - Build assets information

## Support

For issues: https://github.com/aequitas/aura/issues

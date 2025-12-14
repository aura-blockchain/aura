# Aura Wallet - Debian Package (.deb) Guide

## Overview

The Aura Desktop Wallet now supports Debian/Ubuntu package installation via `.deb` format, making it easy to install and manage on Debian-based Linux distributions.

## Package Information

- **Package Name**: `aura-desktop-wallet`
- **Version**: 1.0.0
- **Architecture**: amd64
- **Category**: Finance/Network
- **Maintainer**: Aura Blockchain Team <support@aura.network>

## Building the .deb Package

### Prerequisites

1. Node.js (v18 or higher)
2. npm dependencies installed
3. All build assets in place (icons, scripts, desktop file)

### Build Commands

```bash
# Build React application
npm run build:react

# Build .deb package only
npm run build:linux -- --targets deb

# Or build all Linux formats (AppImage, deb, rpm)
npm run build:linux
```

The built package will be located at:
```
dist/aura-desktop-wallet_1.0.0_amd64.deb
```

### Package Size

- Compressed: ~141 MB
- Installed size: ~535 MB

## Installation

### Standard Installation

```bash
sudo dpkg -i aura-desktop-wallet_1.0.0_amd64.deb
```

### Install with Dependency Resolution

```bash
sudo apt install ./aura-desktop-wallet_1.0.0_amd64.deb
```

### Fix Missing Dependencies

If installation fails due to missing dependencies:

```bash
sudo apt-get install -f
```

## Package Contents

### Installed Files

```
/opt/Aura Wallet/
├── aura-desktop-wallet              # Main executable
├── resources/                        # Application resources
├── locales/                          # Internationalization
└── [Electron runtime files]

/usr/bin/
└── aura-wallet                       # Symlink to main executable

/usr/share/applications/
└── aura-desktop-wallet.desktop       # Desktop entry

/usr/share/icons/hicolor/
├── 16x16/apps/aura-desktop-wallet.png
├── 32x32/apps/aura-desktop-wallet.png
├── 48x48/apps/aura-desktop-wallet.png
├── 64x64/apps/aura-desktop-wallet.png
├── 128x128/apps/aura-desktop-wallet.png
├── 256x256/apps/aura-desktop-wallet.png
└── 512x512/apps/aura-desktop-wallet.png
```

## Usage

After installation, you can launch Aura Wallet in three ways:

1. **Application Menu**: Search for "Aura Wallet" in your desktop environment's application launcher
2. **Terminal Command**: Run `aura-wallet`
3. **Desktop File**: Click the desktop entry if you've added it to your desktop

## Dependencies

The package requires the following system libraries:

- libgtk-3-0
- libnotify4
- libnss3
- libxss1
- libxtst6
- xdg-utils
- libatspi2.0-0
- libdrm2
- libgbm1
- libxcb-dri3-0

These are automatically installed when using `apt install`.

## Uninstallation

### Remove Package

```bash
sudo apt remove aura-desktop-wallet
```

### Purge Package and Configuration

```bash
sudo apt purge aura-desktop-wallet
```

**Note**: User wallet data is stored in `~/.config/aura-wallet` and is NOT automatically removed during uninstallation for security reasons. To completely remove all data:

```bash
rm -rf ~/.config/aura-wallet
```

## Post-Installation Tasks

The post-install script automatically:

1. Creates symlink `/usr/bin/aura-wallet` → `/opt/Aura Wallet/aura-desktop-wallet`
2. Updates desktop database for application menu integration
3. Updates MIME database for protocol handlers (aura://)
4. Updates icon cache for proper icon display
5. Sets up Chrome sandbox permissions

## Post-Removal Tasks

The post-remove script automatically:

1. Removes the `/usr/bin/aura-wallet` symlink
2. Updates desktop database
3. Updates MIME database
4. Updates icon cache
5. Displays warning about user data location (on purge)

## Build Assets

### Icons

Icons are generated automatically by the Python script:

```bash
build/generate-icons.py
```

This creates icons in the following sizes:
- 16x16, 32x32, 48x48, 64x64, 128x128, 256x256, 512x512

Icons follow the Aura brand colors (purple gradient with white "A" letter).

### Desktop Entry

The desktop entry file (`build/aura-wallet.desktop`) provides:
- Application metadata
- MIME type associations (x-scheme-handler/aura)
- Category placement (Finance, Network)
- Keywords for search

### Scripts

**Post-Install** (`build/deb-postinstall.sh`):
- Runs after package installation
- Sets up system integration
- Creates necessary symlinks

**Post-Remove** (`build/deb-postremove.sh`):
- Runs after package removal
- Cleans up system integration
- Warns about user data

## Package Configuration

The `.deb` configuration is defined in `package.json`:

```json
{
  "linux": {
    "target": ["AppImage", "deb", "rpm"],
    "icon": "build/icons",
    "category": "Finance",
    "maintainer": "Aura Blockchain Team <support@aura.network>",
    "vendor": "Aura Blockchain",
    "synopsis": "Secure desktop wallet for Aura blockchain",
    "description": "...",
    "desktop": { ... }
  },
  "deb": {
    "packageCategory": "net",
    "priority": "optional",
    "depends": [...],
    "compression": "xz",
    "afterInstall": "build/deb-postinstall.sh",
    "afterRemove": "build/deb-postremove.sh"
  }
}
```

## Verification

After building, verify the package:

```bash
# View package information
dpkg --info dist/aura-desktop-wallet_1.0.0_amd64.deb

# List package contents
dpkg --contents dist/aura-desktop-wallet_1.0.0_amd64.deb

# Check package dependencies
dpkg-deb -f dist/aura-desktop-wallet_1.0.0_amd64.deb Depends
```

## Troubleshooting

### Issue: Missing Dependencies

**Solution**: Run `sudo apt-get install -f` to auto-install missing dependencies

### Issue: Desktop Entry Not Appearing

**Solution**:
```bash
sudo update-desktop-database /usr/share/applications
```

### Issue: Icons Not Displaying

**Solution**:
```bash
sudo gtk-update-icon-cache -f -t /usr/share/icons/hicolor
```

### Issue: Command Not Found

**Solution**: Verify symlink exists:
```bash
ls -la /usr/bin/aura-wallet
```

If missing, reinstall the package.

## Protocol Handler

The package registers the `aura://` URL scheme handler. When users click aura:// links in browsers or other applications, Aura Wallet will automatically open.

## Security

- Application files are installed in `/opt/Aura Wallet/` with root ownership
- User data is stored in `~/.config/aura-wallet` with user ownership
- Chrome sandbox is properly configured with SUID permissions
- Wallet private keys are never included in package files

## Distribution

### Package Hosting

The `.deb` package can be:

1. **Direct Download**: Host on GitHub Releases
2. **Custom Repository**: Set up an APT repository
3. **PPA**: Publish to a Personal Package Archive

### Adding to APT Repository

Example for custom repository:

```bash
# Add repository key
wget -qO - https://repo.aura.network/key.gpg | sudo apt-key add -

# Add repository
echo "deb https://repo.aura.network/apt stable main" | sudo tee /etc/apt/sources.list.d/aura.list

# Install
sudo apt update
sudo apt install aura-desktop-wallet
```

## Continuous Integration

For automated builds:

```yaml
# GitHub Actions example
- name: Build .deb package
  run: |
    npm ci
    npm run build:react
    npx electron-builder --linux deb

- name: Upload .deb artifact
  uses: actions/upload-artifact@v3
  with:
    name: aura-wallet-deb
    path: dist/*.deb
```

## License

MIT License - See LICENSE file for details

## Support

For issues related to the .deb package:
- GitHub Issues: https://github.com/aequitas/aura/issues
- Email: support@aura.network

---

**Last Updated**: December 14, 2025
**Package Version**: 1.0.0

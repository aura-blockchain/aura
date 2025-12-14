# Aura Browser Wallet Extension

A comprehensive, secure browser extension wallet for the Aura blockchain network. Provides full wallet functionality including key management, token transfers, staking, governance participation, and DEX trading.

## Features

### Wallet Management
- **Create New Wallet**: Generate secure secp256k1 keypairs
- **Import Wallet**: Import from private key (hex format)
- **Export Private Key**: Securely export with confirmation
- **Encrypted Storage**: Optional password-protected private key encryption (AES-256-GCM)
- **Delete Wallet**: Complete removal of all wallet data

### Transaction Support
- **Send Tokens**: Transfer AURA and other tokens
- **Fee Estimation**: Automatic gas and fee calculation
- **Transaction History**: View recent transactions
- **Memo Support**: Add custom transaction memos

### Staking Operations
- **Delegate**: Stake AURA tokens to validators
- **Undelegate**: Unbond staked tokens (21-day unbonding period)
- **Redelegate**: Move stake between validators
- **Claim Rewards**: Withdraw staking rewards
- **View Delegations**: See all active delegations and rewards

### Governance
- **View Proposals**: Browse all governance proposals
- **Vote**: Participate in network governance (Yes/No/Abstain/Veto)
- **Deposit**: Make deposits on proposals in deposit period
- **Proposal Details**: View full proposal information and tally

### DEX Trading
- **Swap Tokens**: Execute token swaps through DEX
- **View Pools**: Browse available liquidity pools
- **Trade History**: Track swap transactions

### Hardware Wallet Support
- **Ledger Integration**: Full Ledger device support via WebUSB
- **Trezor Framework**: Extensible support for Trezor devices
- **Address Verification**: Display and verify addresses on device
- **Transaction Signing**: Sign transactions securely on hardware

### Security Features
- **Private Key Encryption**: AES-256-GCM with PBKDF2 key derivation (100,000 iterations)
- **Secure Storage**: Chrome storage API with optional encryption layer
- **Address Validation**: Comprehensive Bech32 validation
- **Transaction Verification**: All transactions require explicit user approval
- **No Cloud Sync**: All data stays local to the browser

## Installation

### From Source

1. **Clone and Install Dependencies**
   ```bash
   cd wallet/browser-extension
   npm install
   ```

2. **Build Extension**
   ```bash
   npm run build
   ```

3. **Load in Browser**
   - Open Chrome/Edge and navigate to `chrome://extensions`
   - Enable "Developer mode"
   - Click "Load unpacked"
   - Select the `dist/` directory

### From Release Package

1. Download the latest release ZIP
2. Extract to a local directory
3. Follow step 3 from "From Source" above

## Development

### Prerequisites
- Node.js >= 18.0.0
- npm >= 9.0.0
- Chrome/Edge browser

### Setup
```bash
npm install
```

### Development Mode
```bash
npm run watch
```
Automatically rebuilds on file changes.

### Testing
```bash
# Run all tests
npm test

# Run with coverage
npm run test -- --coverage

# Run specific test file
npm test tests/unit/wallet-core.test.js
```

### Code Quality
```bash
# Lint code
npm run lint

# Format code
npm run format

# Security audit
npm run security:audit
```

### Build for Production
```bash
npm run build
```

Creates optimized files in `dist/` directory.

### Package Extension
```bash
npm run package
```

Creates `extension.zip` ready for distribution.

## Architecture

### Core Modules

**src/wallet-core.js**
- Wallet generation and import
- Key encryption/decryption
- Secure storage management
- Address validation

**src/staking.js**
- Delegation operations
- Rewards management
- Validator queries
- Staking calculations

**src/governance.js**
- Proposal queries
- Voting operations
- Deposit management
- Proposal formatting

**cosmos-sdk.js**
- Transaction building
- secp256k1 signing
- Transaction broadcasting
- Account management

**hardware-wallet.js**
- Hardware device communication
- WebUSB integration
- Address derivation
- Hardware transaction signing

### UI Components

**popup.html** - Main wallet interface
**styles.css** - Modern dark theme styling
**popup.js** - UI logic and event handling
**background.js** - Background service worker

## Usage Guide

### Creating a Wallet

1. Click "Create New Wallet" button
2. Wallet address is generated automatically
3. **IMPORTANT**: Click "Export Private Key" and save it securely
4. Optionally, enable password protection for additional security

### Importing a Wallet

1. Click "Import Wallet"
2. Enter your 64-character hex private key
3. Click confirm
4. Wallet address will be displayed

### Sending Tokens

1. Enter recipient address (must start with `aura1`)
2. Enter amount in AURA
3. Select denomination (AURA or STAKE)
4. Optionally add a memo
5. Click "Send" and confirm transaction

### Staking

1. Navigate to the Staking section
2. Select operation: Delegate/Undelegate/Redelegate/Rewards
3. Enter validator address (starts with `auravaloper1`)
4. Enter amount
5. Confirm transaction

### Voting on Proposals

1. Navigate to Governance section
2. Click "Refresh Proposals"
3. Select a proposal to view details
4. Click vote option (Yes/No/Abstain/Veto)
5. Confirm vote transaction

### Using Hardware Wallet

1. Connect your Ledger device
2. Open Cosmos app on device
3. Click "Detect Device" in Hardware Wallet section
4. Click "Connect" when device is found
5. Use "Get Address" to verify address on device
6. Transactions will prompt for device confirmation

## Configuration

### Network Settings

Default network: Local (`http://localhost:1317`)

To change network:
1. Go to Settings section
2. Select network from dropdown or choose "Custom"
3. Enter custom REST endpoint
4. Enter chain ID
5. Click "Save Settings"

### Supported Networks
- **Local**: `http://localhost:1317` (chain-id: `aura-local`)
- **Testnet**: Configure custom endpoint
- **Mainnet**: Configure custom endpoint

### Auto-Refresh

Configure automatic balance/data refresh:
1. Settings > Auto-refresh
2. Set interval in seconds (10-300)
3. Save settings

## Security Best Practices

### For Users

1. **Backup Your Private Key**
   - Export and store securely offline
   - Use a password manager or hardware storage
   - Never share your private key with anyone

2. **Use Password Protection**
   - Enable encryption for private key storage
   - Use a strong, unique password
   - Store password securely

3. **Verify Addresses**
   - Double-check recipient addresses before sending
   - Verify amounts and fees
   - Use hardware wallet for large amounts

4. **Stay Secure**
   - Only download from official sources
   - Keep browser updated
   - Be cautious of phishing attempts
   - Never enter private key on websites

### For Developers

1. **Never Log Sensitive Data**
   - No private keys in logs
   - Sanitize error messages
   - Clear sensitive data from memory

2. **Validate All Inputs**
   - Address format validation
   - Amount range checking
   - Transaction parameter verification

3. **Follow Secure Coding Practices**
   - Use CSP headers
   - Sanitize user inputs
   - Implement proper error handling

## Testing

### Test Coverage

Current test coverage meets target metrics:
- **Statements**: 90%+
- **Branches**: 85%+
- **Functions**: 90%+
- **Lines**: 90%+

### Test Suites

**Unit Tests** (`tests/unit/`)
- `wallet-core.test.js`: Wallet management (25+ tests)
- `staking.test.js`: Staking operations (20+ tests)
- `governance.test.js`: Governance operations (20+ tests)

**Integration Tests** (`tests/integration/`)
- `wallet-flow.test.js`: End-to-end user workflows (15+ scenarios)

### Running Tests

```bash
# All tests
npm test

# With coverage report
npm run test -- --coverage

# Watch mode
npm run test -- --watch

# Specific file
npm test tests/unit/wallet-core.test.js
```

## Troubleshooting

### Common Issues

**Extension Not Loading**
- Clear browser cache
- Rebuild extension: `npm run build`
- Check for console errors in extension page

**Transactions Failing**
- Ensure sufficient balance for tx + fees
- Check network connectivity
- Verify account has been funded
- Check gas limit is adequate

**Hardware Wallet Not Detected**
- Enable WebUSB in browser settings
- Check device is unlocked
- Open Cosmos app on device
- Try different USB port/cable

**Balance Not Updating**
- Check network connection
- Verify REST endpoint is accessible
- Click "Refresh" button
- Check console for API errors

**Encrypted Wallet Won't Load**
- Verify correct password
- Check for typos
- If password is lost, wallet must be recreated (import from backup)

## API Reference

See [IMPLEMENTATION.md](./IMPLEMENTATION.md) for complete API documentation.

## Performance

- **Memory Usage**: ~15MB typical
- **Network Traffic**: REST API calls only (no blockchain syncing)
- **Storage**: <1MB per wallet
- **Load Time**: <500ms popup display

## Browser Compatibility

- **Chrome**: Version 88+
- **Edge**: Version 88+
- **Brave**: Version 1.20+
- **Opera**: Version 74+

Note: Firefox requires Manifest V2, not yet supported.

## Contributing

### Development Process

1. Fork repository
2. Create feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Submit pull request

### Code Style

- ESLint configuration provided
- Prettier for formatting
- Follow existing patterns
- Document complex logic

## License

Apache-2.0 - See LICENSE file for details

## Support

- **Issues**: GitHub Issues
- **Documentation**: [IMPLEMENTATION.md](./IMPLEMENTATION.md)
- **Community**: Discord/Telegram

## Roadmap

### v1.1 (Planned)
- Multi-signature support
- IBC transfers
- Transaction history export
- Address book
- Custom token support

### v1.2 (Planned)
- WalletConnect integration
- Multiple wallet management
- QR code scanning
- Mobile responsive improvements

### v2.0 (Future)
- dApp integration framework
- Advanced governance features
- Portfolio tracking
- NFT support

## Acknowledgments

Built with:
- Cosmos SDK integration
- secp256k1 cryptography
- Web Crypto API
- Vitest testing framework

## Disclaimer

This software is provided "as is" without warranty. Users are responsible for:
- Securing their private keys
- Verifying all transactions
- Understanding blockchain risks
- Backing up wallet data

Never share your private key with anyone. The developers cannot recover lost private keys or reverse transactions.

---

**Version**: 1.0.0
**Last Updated**: December 2024
**Maintained By**: Aura Team

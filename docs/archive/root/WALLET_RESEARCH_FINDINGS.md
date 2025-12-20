# Blockchain Wallet Implementation - Research Findings & Best Practices

**Date**: 2025-12-14
**Purpose**: Comprehensive research findings on wallet testing, security, and configuration best practices for Aura project wallets

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [React Native Mobile Wallet Best Practices](#react-native-mobile-wallet-best-practices)
3. [Electron Desktop Wallet Security](#electron-desktop-wallet-security)
4. [Cosmos SDK Wallet Patterns](#cosmos-sdk-wallet-patterns)
5. [Seed Phrase & Private Key Security](#seed-phrase--private-key-security)
6. [Secure Storage Implementation](#secure-storage-implementation)
7. [Testing Strategies](#testing-strategies)
8. [Implementation Checklist](#implementation-checklist)

---

## Executive Summary

This document synthesizes industry best practices for blockchain wallet development in 2025, specifically for:
- **Mobile**: React Native cryptocurrency wallets (iOS/Android)
- **Desktop**: Electron-based desktop wallets
- **Blockchain**: Cosmos SDK integration patterns

**Key Findings:**
- ✅ Never store private keys/mnemonics digitally without encryption
- ✅ Use OS-level secure storage (Keychain/Keystore)
- ✅ Implement comprehensive testing with mocked blockchain providers
- ✅ Follow Cosmos SDK testing patterns for consistency
- ✅ Apply defense-in-depth security architecture

---

## React Native Mobile Wallet Best Practices

### Testing Frameworks

**Recommended Stack:**
- **Jest**: Unit testing for component and logic verification
- **React Testing Library**: Component testing with user-centric approach
- **Mock Providers**: Simulate blockchain responses without live connections

Sources:
- [How to build a crypto wallet app with React Native](https://touchlane.com/how-to-build-a-crypto-wallet-app-with-react-native/)
- [Building a Web3 Wallet Example App](https://www.ianhafkenschiel.com/blog/building-a-web3-wallet-example-app-with-react-and-react-native/)

### Testnet Usage

**Safe Development Environment:**
- Ethereum Sepolia testnet (replaces deprecated Rinkeby/Ropsten)
- Solana Devnet
- Cosmos testnets (project-specific)

**Benefits:**
- Zero cost for blockchain interactions
- Safe experimentation without risking real assets
- Realistic transaction testing

Sources:
- [Building a Crypto Wallet App With React Native](https://pagepro.co/blog/crypto-wallet-app-with-react-native/)

### Testing Focus Areas

**Critical Testing Domains:**
1. **User Journey Testing**: End-to-end wallet workflows
2. **Performance Testing**: App responsiveness and efficiency
3. **Security Testing**: Sensitive data handling and transaction safety
4. **Continuous Testing**: Automated deployment and quality assurance

Sources:
- [Best DX for React Native Web3 dApps](https://www.callstack.com/blog/best-dx-for-react-native-web3-dapps-with-web3modal-and-wagmi)

### Development Tools

**Testing Infrastructure:**
- Expo development tools (note: Expo Go not supported for Web3 due to wrapper conflicts)
- Continuous integration/deployment for automated testing
- Mock blockchain providers for offline testing

---

## Electron Desktop Wallet Security

### Critical Configuration Settings

**Security-First Configuration:**

```javascript
// webPreferences settings
{
  nodeIntegration: false,        // NEVER enable with remote code
  contextIsolation: true,        // Isolate renderer context
  sandbox: true,                 // Enable sandbox for renderer
  webSecurity: true,             // Never disable in production
  enableRemoteModule: false      // Deprecated and dangerous
}
```

**Severity:** Critical security issues arise from:
- Enabling `nodeIntegration` with remote code
- Disabling `webSecurity`
- Using deprecated `remote` module

Sources:
- [Security | Electron](https://www.electronjs.org/docs/latest/tutorial/security)
- [Vulnerability in Electron-based Application](https://www.certik.com/resources/blog/vulnerability-electron-based-application-malicious-code-execution)

### Prevent Navigation & XSS

**Implementation:**

```javascript
// Disable navigation (single-page app pattern)
webContents.on('will-navigate', (event) => {
  event.preventDefault();
});

// Prevent middle-click bypass
webPreferences: {
  disableBlinkFeatures: 'Auxclick'
}
```

**Rationale:** Wallet applications are typically single-page React apps with no need for navigation. XSS vulnerabilities have higher impact in Electron than web browsers.

Sources:
- [Let's Create a Secure HD Bitcoin Wallet in Electron + React.js](https://medium.com/coinmonks/lets-create-a-secure-hd-bitcoin-wallet-in-electron-react-js-575032c42bf3)

### Permission Management

**Principle of Least Privilege:**

```javascript
// Implement permission request handler
session.setPermissionRequestHandler((webContents, permission, callback) => {
  // Deny all by default
  callback(false);

  // Only approve explicitly needed permissions
  // if (permission === 'clipboard-read') callback(true);
});
```

Without a handler, Electron silently approves all permission requests (microphone, video, clipboard, etc.).

Sources:
- [Security, Native Capabilities, and Your Responsibility](https://docs.w3cub.com/electron/tutorial/security)

### Keep Software Updated

**Critical Maintenance:**
- Use latest Electron version always
- Update immediately when new major versions release
- Older versions = easier attack targets (outdated Chromium/Node.js)

Sources:
- [Electrolint and security of electron applications](https://www.sciencedirect.com/science/article/pii/S2667295221000222)

### Automated Security Auditing

**Tool: Electronegativity**
- Automates detection of misconfigurations
- Identifies insecure patterns
- Integrates into CI/CD pipeline

---

## Cosmos SDK Wallet Patterns

### Testing Vision

**Cosmos SDK Testing Pyramid:**

1. **End-to-End (E2E) Tests**: Test full application with user-realistic flows
2. **Integration Tests**: Test module interactions with minimum-viable app
3. **Simulation Tests**: Randomize parameters to discover edge cases

**Philosophy:** Follow Cosmos SDK patterns to speak the same language as the ecosystem.

Sources:
- [Testing | Developer Portal](https://tutorials.cosmos.network/academy/2-cosmos-concepts/12-testing.html)
- [cosmos-sdk/CODING_GUIDELINES.md](https://github.com/cosmos/cosmos-sdk/blob/main/CODING_GUIDELINES.md)

### Wallet Implementation Requirements

**Prerequisite Knowledge:**
1. Hierarchical Deterministic (HD) Wallet structure (BIP32/BIP39/BIP44)
2. Tendermint key types (5+ different types)
3. Cosmos SDK message types and encoding

**Offline Signing:** Must support transaction signing without network connectivity.

Sources:
- [Building an application specific blockchain using Cosmos SDK](https://medium.com/coinmonks/building-an-application-specific-blockchain-using-cosmos-sdk-part-6-a62d02819229)

### Testing Frameworks

**Cosmos-Specific Tools:**

- **Agoric/Synpress**: E2E testing for Keplr wallet integration
  - Based on Cypress JavaScript framework
  - Seamless Keplr interaction: transaction signing, account management, network switching

- **Interchaintest**: E2E testing for IBC chains

- **Atomkraft**: E2E testing for Cosmos SDK blockchains

Sources:
- [Agoric/Synpress: A Custom E2E Testing Framework for Cosmos SDK](https://agoric.com/blog/technology/agoric-synpress-a-custom-e2e-testing-framework-for-cosmos-sdk/)
- [GitHub - cosmos/awesome-cosmos](https://github.com/cosmos/awesome-cosmos)

---

## Seed Phrase & Private Key Security

### Critical Security Rules

**NEVER:**
- ❌ Store seed phrases digitally (screenshots, cloud, clipboard)
- ❌ Type mnemonics into computers without encryption
- ❌ Share seed phrases with anyone (no legitimate support asks for this)
- ❌ Store in single location (geographic redundancy required)

Sources:
- [Understanding Secure Crypto Seed Phrases for Wallet Protection](https://web3.gate.com/en/crypto-wiki/article/understanding-secure-crypto-seed-phrases-for-wallet-protection-20251206)
- [Protecting Your Seed Phrase: Essential Security Tips](https://web3.gate.com/en/crypto-wiki/article/protecting-your-seed-phrase-essential-security-tips-20251201)

### Physical Storage Methods

**Paper Backups (Minimum Standard):**
- Write seed phrase on paper (not typed)
- Store in fireproof/waterproof lockbox or safe
- Create multiple copies for redundancy
- Verify word-by-word accuracy

Sources:
- [4 Best Ways To Securely Store Seed Phrase In 2025](https://coinsutra.com/keep-recovery-seed-safe-secure/)

**Metal Storage (Recommended for 2025):**

**Advantages:**
- Fireproof, waterproof, shockproof, acid-resistant
- Some are bulletproof
- Far superior to paper durability

**Top Solutions:**
- **Billfodl**: Premier solution when combined with Ledger Nano X
- **Cryptosteel Capsule**: Sold by Ledger, leading wallet manufacturer

Sources:
- [16 Best Metal Crypto Wallets for Seed Phrase Storage in 2025](https://coincodex.com/article/23147/best-metal-crypto-wallets-for-seed-phrase-storage/)
- [Best Practices for Storing a Wallet's Mnemonic](https://builtoncardano.com/blog/best-practices-for-storing-a-wallets-mnemonic-recovery-seed)

### Advanced Security Features

**25th Word Passphrase:**
- Adds extra layer beyond 12/24-word seed
- Seed phrase alone cannot access funds without passphrase
- Supported by advanced wallets
- Dramatically increases security

Sources:
- [Mastering the Essentials: An In-Depth Guide to Mnemonic Phrases](https://web3.gate.com/en/crypto-wiki/article/mastering-the-essentials-an-in-depth-guide-to-mnemonic-phrases-20251130)

### Geographic Redundancy

**Best Practice:**
- Multiple backups in different physical locations
- Protects against local disasters (fire, flood, theft)
- Minimum 2 locations recommended
- Consider bank safe deposit boxes

Sources:
- [How to Keep Your Seed Phrase Secure](https://onekey.so/blog/ecosystem/how-to-keep-your-seed-phrase-secure/)

### Verification & Maintenance

**After Writing Mnemonic:**
1. Perform word-by-word comparison against original
2. Verify spelling of each word
3. Confirm exact order (cryptographically significant)
4. Regular integrity checks (ensure legibility)

Sources:
- [Securing Your Mnemonic Phrase: Essential Tips for Optimal Safety](https://web3.gate.com/en/crypto-wiki/article/securing-your-mnemonic-phrase-essential-tips-for-optimal-safety-20251202)

---

## Secure Storage Implementation

### React Native: Keychain/Keystore

**Library: react-native-keychain**

**Platform Security:**
- **iOS**: Keychain Services (OS-level encryption)
- **Android**: KeyStore (hardware-backed when available)

**Usage:**

```javascript
import * as Keychain from 'react-native-keychain';

// Store private key
await Keychain.setGenericPassword('privateKey', encryptedPrivateKey, {
  service: 'com.aura.wallet',
  accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED
});

// Retrieve private key
const credentials = await Keychain.getGenericPassword({
  service: 'com.aura.wallet'
});
```

**Best Practices:**
- Generate private keys using `react-native-crypto`
- Never store unencrypted in AsyncStorage
- Combine with biometric authentication (react-native-fingerprint-scanner)
- Use password + biometric dual protection

Sources:
- [react-native-keychain - npm](https://www.npmjs.com/package/react-native-keychain)
- [GitHub - oblador/react-native-keychain](https://github.com/oblador/react-native-keychain)
- [How to build a crypto wallet app with React Native](https://touchlane.com/how-to-build-a-crypto-wallet-app-with-react-native/)

**Android Note:**
- SharedPreferences not encrypted by default
- Use Encrypted Shared Preferences wrapper
- Automatically encrypts keys and values

Sources:
- [Security · React Native](https://reactnative.dev/docs/security)

### Electron: safeStorage / Keytar

**Built-in: Electron safeStorage API (Electron 15+)**

**Platform Security:**
- **macOS**: Keychain
- **Windows**: Credentials Manager
- **Linux**: Gnome Keyring

**Usage:**

```javascript
const { safeStorage } = require('electron');

// Encrypt and store
const encrypted = safeStorage.encryptString(privateKey);
// Store encrypted buffer to disk or settings

// Decrypt when needed
const decrypted = safeStorage.decryptString(encrypted);
```

Sources:
- [safeStorage | Electron](https://www.electronjs.org/docs/latest/api/safe-storage)
- [Replacing Keytar with Electron's safeStorage](https://freek.dev/2103-replacing-keytar-with-electrons-safestorage-in-ray)

**Alternative: node-keytar**

**Usage:**

```javascript
const keytar = require('keytar');

// Store from main process only
await keytar.setPassword('aura-wallet', 'account1', privateKey);

// Retrieve
const privateKey = await keytar.getPassword('aura-wallet', 'account1');

// Delete
await keytar.deletePassword('aura-wallet', 'account1');
```

**Best Practices:**
- Only call from main process (avoids macOS permission dialogs)
- Handle missing entries gracefully (user may delete from OS store)
- Decrypt private keys only when signing transactions
- Minimize time unencrypted in memory

Sources:
- [How to securely store sensitive information in Electron with node-keytar](https://cameronnokes.com/blog/how-to-securely-store-sensitive-information-in-electron-with-node-keytar/)
- [How to securely store sensitive information in Electron with node-keytar | Medium](https://medium.com/cameron-nokes/how-to-securely-store-sensitive-information-in-electron-with-node-keytar-51af99f1cfc4)

### BIP38 Encryption for Cold Storage

**Use Case:** Exporting/printing private keys for backup

**Security:**
- Encrypts private keys with password before export
- Safe even if backup is discovered
- Decrypt only briefly when signing

Sources:
- [How to Store and Manage Private Keys in Cryptocurrency Wallets](https://www.coinspeaker.com/guides/import-export-manage-private-keys-cryptocurrency-wallet/)

---

## Testing Strategies

### React Native Testing Stack

**Unit Testing:**
```bash
# Jest configuration
jest tests/ --coverage --verbose

# With coverage reporting
npm test -- --coverage --coverageReporters=text --coverageReporters=html
```

**Component Testing:**
```javascript
import { render, fireEvent } from '@testing-library/react-native';

test('sends transaction when button pressed', () => {
  const { getByText } = render(<SendScreen />);
  const sendButton = getByText('Send');

  fireEvent.press(sendButton);

  // Assert transaction was signed
});
```

**Mock Blockchain Providers:**
```javascript
// Mock Cosmos client
jest.mock('@cosmjs/stargate', () => ({
  SigningStargateClient: {
    connectWithSigner: jest.fn().mockResolvedValue({
      sendTokens: jest.fn(),
      getBalance: jest.fn().mockResolvedValue({ amount: '1000000', denom: 'uaura' })
    })
  }
}));
```

### Electron Testing Stack

**Spectron (deprecated) → Playwright**
```javascript
const { _electron: electron } = require('playwright');

test('wallet launches successfully', async () => {
  const app = await electron.launch({ args: ['main.js'] });
  const window = await app.firstWindow();

  await expect(window.title()).resolves.toBe('Aura Wallet');

  await app.close();
});
```

**Mock IPC Communication:**
```javascript
// Renderer process test
ipcRenderer.send('sign-transaction', txData);

// Mock main process response
ipcRenderer.on('transaction-signed', (event, signedTx) => {
  expect(signedTx).toHaveProperty('signature');
});
```

### Cosmos SDK Testing Patterns

**Integration Test Example:**
```go
func TestWalletModule(t *testing.T) {
    app := simapp.Setup(t, false)
    ctx := app.BaseApp.NewContext(false, tmproto.Header{})

    // Test wallet creation
    wallet := types.NewWallet(address, publicKey)
    app.WalletKeeper.SetWallet(ctx, wallet)

    // Verify retrieval
    retrieved := app.WalletKeeper.GetWallet(ctx, address)
    require.Equal(t, wallet, retrieved)
}
```

**Simulation Test Example:**
```go
func TestRandomWalletOperations(t *testing.T) {
    app := simapp.Setup(t, false)

    // Random operations
    simtypes.SimulateMsgCreateWallet(...)
    simtypes.SimulateMsgUpdateWallet(...)

    // Invariant checks
    require.NoError(t, WalletInvariant(app))
}
```

---

## Implementation Checklist

### Mobile Wallet (React Native)

**Security:**
- [ ] Install and configure `react-native-keychain`
- [ ] Implement private key generation with `react-native-crypto`
- [ ] Add biometric authentication (react-native-fingerprint-scanner)
- [ ] Never use AsyncStorage for private keys/mnemonics
- [ ] Implement BIP39 mnemonic generation/validation
- [ ] Add BIP32 HD wallet derivation
- [ ] Implement BIP44 account discovery

**Testing:**
- [ ] Set up Jest with @testing-library/react-native
- [ ] Create mock Cosmos SDK client
- [ ] Write unit tests for all components
- [ ] Implement E2E tests for user journeys
- [ ] Test on Cosmos testnet (not mainnet)
- [ ] Add continuous integration testing
- [ ] Achieve >90% test coverage

**Features:**
- [ ] Create wallet from mnemonic
- [ ] Import existing wallet
- [ ] Display balances
- [ ] Send transactions
- [ ] Receive transactions (QR code)
- [ ] Transaction history
- [ ] Multiple account support
- [ ] Testnet/mainnet switching

### Desktop Wallet (Electron)

**Security:**
- [ ] Configure webPreferences correctly (nodeIntegration: false, etc.)
- [ ] Implement safeStorage for private key encryption
- [ ] Disable navigation with will-navigate handler
- [ ] Set up permission request handler (deny-by-default)
- [ ] Keep Electron updated to latest version
- [ ] Run electronegativity security audit
- [ ] Implement IPC security (validated channels only)

**Testing:**
- [ ] Set up Playwright for E2E testing
- [ ] Mock IPC communication
- [ ] Test all user workflows
- [ ] Security audit with electronegativity
- [ ] Test on all platforms (Windows/macOS/Linux)
- [ ] Verify code signing works

**Features:**
- [ ] Create/import wallet
- [ ] Secure password protection
- [ ] Transaction signing
- [ ] Balance display
- [ ] Transaction history
- [ ] Settings management
- [ ] Auto-update mechanism

### Cosmos SDK Integration

**Implementation:**
- [ ] Use @cosmjs/stargate for chain interaction
- [ ] Implement offline transaction signing
- [ ] Support all Cosmos message types
- [ ] Add gas estimation
- [ ] Implement fee payment in multiple denoms
- [ ] Support IBC transfers
- [ ] Add staking/delegation support

**Testing:**
- [ ] Follow Cosmos SDK testing patterns
- [ ] Write integration tests with simapp
- [ ] Add simulation tests with random ops
- [ ] Test against local testnet
- [ ] Verify with Agoric/Synpress if using Keplr
- [ ] Test IBC functionality

---

## References & Sources

### React Native Mobile Wallet
- [How to build a crypto wallet app with React Native](https://touchlane.com/how-to-build-a-crypto-wallet-app-with-react-native/)
- [Building a Web3 Wallet Example App](https://www.ianhafkenschiel.com/blog/building-a-web3-wallet-example-app-with-react-and-react-native/)
- [Building a Crypto Wallet App With React Native](https://pagepro.co/blog/crypto-wallet-app-with-react-native/)
- [Best DX for React Native Web3 dApps](https://www.callstack.com/blog/best-dx-for-react-native-web3-dapps-with-web3modal-and-wagmi)
- [Security · React Native](https://reactnative.dev/docs/security)
- [react-native-keychain - npm](https://www.npmjs.com/package/react-native-keychain)
- [GitHub - oblador/react-native-keychain](https://github.com/oblador/react-native-keychain)

### Electron Desktop Wallet
- [Security | Electron](https://www.electronjs.org/docs/latest/tutorial/security)
- [Let's Create a Secure HD Bitcoin Wallet in Electron + React.js](https://medium.com/coinmonks/lets-create-a-secure-hd-bitcoin-wallet-in-electron-react-js-575032c42bf3)
- [Vulnerability in Electron-based Application](https://www.certik.com/resources/blog/vulnerability-electron-based-application-malicious-code-execution)
- [Electrolint and security of electron applications](https://www.sciencedirect.com/science/article/pii/S2667295221000222)
- [safeStorage | Electron](https://www.electronjs.org/docs/latest/api/safe-storage)
- [How to securely store sensitive information in Electron with node-keytar](https://cameronnokes.com/blog/how-to-securely-store-sensitive-information-in-electron-with-node-keytar/)
- [Replacing Keytar with Electron's safeStorage](https://freek.dev/2103-replacing-keytar-with-electrons-safestorage-in-ray)

### Cosmos SDK Wallet Patterns
- [Testing | Developer Portal](https://tutorials.cosmos.network/academy/2-cosmos-concepts/12-testing.html)
- [cosmos-sdk/CODING_GUIDELINES.md](https://github.com/cosmos/cosmos-sdk/blob/main/CODING_GUIDELINES.md)
- [Agoric/Synpress: A Custom E2E Testing Framework for Cosmos SDK](https://agoric.com/blog/technology/agoric-synpress-a-custom-e2e-testing-framework-for-cosmos-sdk/)
- [GitHub - cosmos/awesome-cosmos](https://github.com/cosmos/awesome-cosmos)
- [Building an application specific blockchain using Cosmos SDK](https://medium.com/coinmonks/building-an-application-specific-blockchain-using-cosmos-sdk-part-6-a62d02819229)

### Seed Phrase & Private Key Security
- [Understanding Secure Crypto Seed Phrases for Wallet Protection](https://web3.gate.com/en/crypto-wiki/article/understanding-secure-crypto-seed-phrases-for-wallet-protection-20251206)
- [Protecting Your Seed Phrase: Essential Security Tips](https://web3.gate.com/en/crypto-wiki/article/protecting-your-seed-phrase-essential-security-tips-20251201)
- [4 Best Ways To Securely Store Seed Phrase In 2025](https://coinsutra.com/keep-recovery-seed-safe-secure/)
- [16 Best Metal Crypto Wallets for Seed Phrase Storage in 2025](https://coincodex.com/article/23147/best-metal-crypto-wallets-for-seed-phrase-storage/)
- [How to Keep Your Seed Phrase Secure](https://onekey.so/blog/ecosystem/how-to-keep-your-seed-phrase-secure/)
- [Securing Your Mnemonic Phrase: Essential Tips for Optimal Safety](https://web3.gate.com/en/crypto-wiki/article/securing-your-mnemonic-phrase-essential-tips-for-optimal-safety-20251202)
- [Mastering the Essentials: An In-Depth Guide to Mnemonic Phrases](https://web3.gate.com/en/crypto-wiki/article/mastering-the-essentials-an-in-depth-guide-to-mnemonic-phrases-20251130)
- [Best Practices for Storing a Wallet's Mnemonic](https://builtoncardano.com/blog/best-practices-for-storing-a-wallets-mnemonic-recovery-seed)
- [How to Store and Manage Private Keys in Cryptocurrency Wallets](https://www.coinspeaker.com/guides/import-export-manage-private-keys-cryptocurrency-wallet/)

---

**Document Version**: 1.0
**Last Updated**: 2025-12-14
**Next Review**: Before wallet implementation begins

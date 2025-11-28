# AURA Developer Playground - Quick Start Guide

## Getting Started in 5 Minutes

### 1. Install and Run

```bash
cd C:\Users\decri\GitClones\aura\developer-playground

# Install dependencies
npm install

# Start the playground
npm run dev

# Or use Docker
docker-compose up -d
```

### 2. Open in Browser
Navigate to: `http://localhost:8080`

### 3. Try Your First Example

1. Click **"Hello World"** in the left sidebar
2. Click **"Run Code"** button (or press Ctrl+Enter)
3. View the results in the Console tab

---

## Module Quick Reference

### 🎫 VCRegistry (Most Important)

#### Query a Verifiable Credential
```javascript
const vcId = '1';
const vc = await api.getVC(vcId);
console.log('VC:', vc);
```

#### Mint a New VC
```javascript
const msg = {
    type: 'aura/vcregistry/MsgMintVC',
    value: {
        issuer: wallet.address,
        subject: 'aura1...',
        vc_type: 'IdentityCredential',
        claims: JSON.stringify({ name: 'John Doe' }),
        expiration_date: '2026-01-01T00:00:00Z'
    }
};
```

#### Create VC Presentation
```javascript
const msg = {
    type: 'aura/vcregistry/MsgCreatePresentation',
    value: {
        holder: wallet.address,
        vc_ids: ['1', '2'],
        purpose: 'Identity Verification',
        verifier: 'aura1...'
    }
};
```

---

### 🔐 Auth Module

```javascript
// Query account
const account = await api.getAccount('aura1...');
console.log('Account:', account);
```

---

### 🌉 Bridge Module

```javascript
// Query bridge transfers
const transfers = await api.getBridgeTransfers();

// Initiate transfer
const msg = {
    type: 'aura/bridge/MsgInitiateTransfer',
    value: {
        sender: wallet.address,
        recipient: 'cosmos1...',
        amount: [{ denom: 'uaura', amount: '1000000' }],
        destination_chain: 'cosmoshub-4'
    }
};
```

---

### ✅ Compliance Module

```javascript
// Check compliance status
const status = await api.getComplianceStatus('aura1...');
console.log('KYC Verified:', status.kyc_verified);
console.log('AML Cleared:', status.aml_cleared);
```

---

### ⭐ ConfidenceScore Module

```javascript
// Query confidence score
const score = await api.getConfidenceScore('aura1...');
console.log('Score:', score.score);
console.log('Reputation:', score.reputation);

// Complete routine
const msg = {
    type: 'aura/confidencescore/MsgCompleteRoutine',
    value: {
        address: wallet.address,
        routine_id: '1',
        proof: 'completion_proof_hash'
    }
};
```

---

### 🔑 Cryptography Module

```javascript
// Query public key
const keyInfo = await api.getPublicKey('aura1...');

// Rotate key
const msg = {
    type: 'aura/cryptography/MsgRotateKey',
    value: {
        address: wallet.address,
        new_public_key: 'new_key_base64',
        algorithm: 'ed25519'
    }
};
```

---

### 📊 DataRegistry Module

```javascript
// Query data items
const items = await api.getDataItems();

// Register data
const msg = {
    type: 'aura/dataregistry/MsgRegisterData',
    value: {
        owner: wallet.address,
        data_hash: 'ipfs://QmHash...',
        metadata: JSON.stringify({ name: 'My Data' })
    }
};
```

---

### 💱 DEX Module

```javascript
// Query pools
const pools = await api.getPools();

// Swap tokens
const estimate = await api.estimateSwap(1, 'uaura', '1000000');
const msg = {
    type: 'aura/dex/MsgSwap',
    value: {
        sender: wallet.address,
        pool_id: 1,
        token_in: { denom: 'uaura', amount: '1000000' },
        min_token_out: estimate.amount_out
    }
};

// Add liquidity
const msg = {
    type: 'aura/dex/MsgAddLiquidity',
    value: {
        sender: wallet.address,
        pool_id: 1,
        tokens: [
            { denom: 'uaura', amount: '1000000' },
            { denom: 'uusdc', amount: '1000000' }
        ]
    }
};
```

---

### 💵 EconomicSecurity Module

```javascript
// Query dynamic fees
const fees = await api.getDynamicFees();
const mevProtection = await api.getMevProtection();
```

---

### 👤 IdentityChange Module

```javascript
// Query requests
const requests = await api.getIdentityChangeRequests();

// Request change
const msg = {
    type: 'aura/identitychange/MsgRequestChange',
    value: {
        requester: wallet.address,
        old_identity: 'old_did:aura:...',
        new_identity: 'new_did:aura:...',
        reason: 'Identity update required'
    }
};
```

---

### 📋 InclusionRoutines Module

```javascript
// Query routines
const routines = await api.getInclusionRoutines();

// Register routine
const msg = {
    type: 'aura/inclusionroutines/MsgRegisterRoutine',
    value: {
        creator: wallet.address,
        name: 'Basic Verification',
        steps: ['email', 'phone', 'document'],
        reward: '100'
    }
};
```

---

### 📈 Monitoring Module

```javascript
// Query metrics
const metrics = await api.getMonitoringMetrics();
const alerts = await api.getMonitoringAlerts();
```

---

### 🛡️ NetworkSecurity Module

```javascript
// Query status
const status = await api.getNetworkSecurityStatus();

// Report peer
const msg = {
    type: 'aura/networksecurity/MsgReportPeer',
    value: {
        reporter: wallet.address,
        peer_id: 'peer_id_...',
        reason: 'Suspicious behavior'
    }
};
```

---

### ✔️ Prevalidation Module

```javascript
// Check transaction
const status = await api.getPrevalidationStatus('tx_hash');
console.log('Valid:', status.valid);
```

---

### 🔒 Privacy Module

```javascript
// Query parameters
const params = await api.getPrivacyParams();
console.log('Mixing Enabled:', params.mixing_enabled);
```

---

### 🛡️ ValidatorSecurity Module

```javascript
// Query validator security
const status = await api.getValidatorSecurityStatus('auravaloper1...');
const slashing = await api.getValidatorSlashingEvents('auravaloper1...');
```

---

### 🔐 WalletSecurity Module

```javascript
// Query status
const status = await api.getWalletSecurityStatus('aura1...');
const sessions = await api.getWalletSessions('aura1...');

// Enable 2FA
const msg = {
    type: 'aura/walletsecurity/MsgEnable2FA',
    value: {
        address: wallet.address,
        totp_secret: 'base32_secret',
        recovery_codes: ['code1', 'code2']
    }
};
```

---

## Standard Cosmos Modules

### 💰 Bank Module

```javascript
// Query balance
const balance = await api.getBalance('aura1...', 'uaura');

// Send tokens
const msg = {
    type: 'cosmos-sdk/MsgSend',
    value: {
        from_address: wallet.address,
        to_address: 'aura1...',
        amount: [{ denom: 'uaura', amount: '1000000' }]
    }
};
```

---

### 🔒 Staking Module

```javascript
// Query validators
const validators = await api.getValidators('BOND_STATUS_BONDED');

// Delegate
const msg = {
    type: 'cosmos-sdk/MsgDelegate',
    value: {
        delegator_address: wallet.address,
        validator_address: 'auravaloper1...',
        amount: { denom: 'uaura', amount: '1000000' }
    }
};

// Claim rewards
const msg = {
    type: 'cosmos-sdk/MsgWithdrawDelegationReward',
    value: {
        delegator_address: wallet.address,
        validator_address: 'auravaloper1...'
    }
};
```

---

### 🗳️ Governance Module

```javascript
// Query proposals
const proposals = await api.getProposals('PROPOSAL_STATUS_VOTING_PERIOD');

// Submit proposal
const msg = {
    type: 'cosmos-sdk/MsgSubmitProposal',
    value: {
        content: {
            type: 'cosmos-sdk/TextProposal',
            value: { title: 'Title', description: 'Description' }
        },
        initial_deposit: [{ denom: 'uaura', amount: '10000000' }],
        proposer: wallet.address
    }
};

// Vote
const msg = {
    type: 'cosmos-sdk/MsgVote',
    value: {
        proposal_id: '1',
        voter: wallet.address,
        option: 'VOTE_OPTION_YES'
    }
};
```

---

## Network Configuration

### Endpoints

```javascript
// Local
REST: http://localhost:1317
RPC:  http://localhost:26657

// Testnet
REST: https://testnet-api.aura.zone
RPC:  https://testnet-rpc.aura.zone

// Mainnet
REST: https://api.aura.zone
RPC:  https://rpc.aura.zone
```

### Chain IDs

```javascript
local:   'aura-local'
testnet: 'aura-testnet-1'
mainnet: 'aura-1'
```

---

## Keyboard Shortcuts

- `Ctrl+Enter` / `Cmd+Enter` - Run code
- `Ctrl+S` / `Cmd+S` - Save snippet
- `Ctrl+/` / `Cmd+/` - Toggle comment
- `Alt+Shift+F` - Format code

---

## Tips & Tricks

### 1. Connect Wallet First
Most transaction examples require wallet connection:
```javascript
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}
```

### 2. Replace Placeholder Addresses
Examples use placeholders like `aura1...` - replace with actual addresses.

### 3. Check API Availability
Ensure the API endpoint is accessible before running queries.

### 4. Use Console Output
All examples log to the console for easy debugging.

### 5. Save Useful Snippets
Click "Save Snippet" to save your modified code.

### 6. Share Code
Click "Share Code" to generate a shareable URL.

### 7. Switch Languages
Click language tabs to see the same example in different languages.

### 8. Search Examples
Use the search box to quickly find examples.

---

## Troubleshooting

### Wallet Won't Connect
1. Install Keplr extension
2. Ensure correct network selected
3. Check browser console for errors

### API Request Fails
1. Verify endpoint is accessible
2. Check network selection
3. Verify address format

### Code Won't Run
1. Check for syntax errors
2. Ensure wallet is connected (if required)
3. Verify API is online

---

## Support

- Documentation: https://docs.aura.zone
- Discord: https://discord.gg/aura
- GitHub: https://github.com/aura/aura

---

**Happy Coding!** 🚀

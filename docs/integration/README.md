# Aura Blockchain Integration Guide for Wallet Developers

## Overview

This guide provides wallet developers with comprehensive information for integrating with the Aura blockchain. It covers module APIs, transaction construction, query patterns, event subscriptions, and best practices for building secure, user-friendly wallet applications.

## Quick Start

### Connection Setup

```typescript
import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";

const RPC_ENDPOINT = "https://rpc.aura.network"; // or localhost:26657 for local
const client = await SigningCosmWasmClient.connect(RPC_ENDPOINT);
```

### Account Management

```typescript
// Get account balance
const balance = await client.getBalance(address, "uaura");

// Get all balances
const allBalances = await client.getAllBalances(address);

// Query account info
const account = await client.getAccount(address);
```

## Module Integration

### Identity Module

#### Session Management

```typescript
// Create session on login
const createSession = {
  typeUrl: "/aura.identity.v1beta1.MsgCreateSession",
  value: {
    address: userAddress,
    deviceFingerprint: generateDeviceFingerprint(),
    ipAddress: getUserIP(),
    metadata: {
      userAgent: navigator.userAgent,
      platform: navigator.platform
    }
  }
};

const result = await client.signAndBroadcast(userAddress, [createSession], fee);
const sessionId = result.events.find(e => e.type === 'session_created')
  ?.attributes.find(a => a.key === 'session_id')?.value;

// Store session ID and expiration
localStorage.setItem('aura_session_id', sessionId);
localStorage.setItem('aura_session_expires', expiresAt);
```

#### Multisig Wallet Operations

```typescript
// Create multisig wallet
const createMultisig = {
  typeUrl: "/aura.identity.v1beta1.MsgCreateMultisigWallet",
  value: {
    creator: userAddress,
    signers: [signer1, signer2, signer3],
    threshold: 2,
    walletType: "WALLET_TYPE_INDIVIDUAL"
  }
};

// Sign multisig proposal
const signProposal = {
  typeUrl: "/aura.identity.v1beta1.MsgSignMultisigProposal",
  value: {
    signer: userAddress,
    proposalId: "msprop_..."
  }
};
```

#### DID Resolution

```typescript
// Resolve DID to identity record
const queryClient = await client.forceGetQueryClient();
const identityQuery = {
  identity_record: { did: "did:aura:mainnet:abc123" }
};

const response = await queryClient.wasm.queryContractSmart(
  identityModuleAddress,
  identityQuery
);
```

### DEX Module

#### Liquidity Pool Operations

```typescript
// Add liquidity to pool
const addLiquidity = {
  typeUrl: "/aura.dex.v1beta1.MsgAddLiquidity",
  value: {
    provider: userAddress,
    poolId: "pool_1",
    amountA: { denom: "uaura", amount: "1000000" },
    amountB: { denom: "ubtc", amount: "100000" }
  }
};

// Get swap quote before execution
const quoteQuery = {
  get_quote: {
    poolId: "pool_1",
    denomIn: "uaura",
    amountIn: "10000"
  }
};

const quote = await queryClient.wasm.queryContractSmart(
  dexModuleAddress,
  quoteQuery
);

// Display quote to user
console.log(`Estimated output: ${quote.estimated_output}`);
console.log(`Price impact: ${quote.price_impact}%`);
console.log(`Fee: ${quote.fee}`);

// Execute swap with slippage protection
const swap = {
  typeUrl: "/aura.dex.v1beta1.MsgSwapExactIn",
  value: {
    sender: userAddress,
    poolId: "pool_1",
    coinIn: { denom: "uaura", amount: "10000" },
    minAmountOut: calculateMinOutput(quote.estimated_output, slippageBps),
    maxSlippageBps: 500 // 5% max slippage
  }
};
```

#### Commit-Reveal Trading (Front-Running Protection)

```typescript
// Phase 1: Commit order hash
const orderData = {
  orderType: "BUY_AURA",
  auraAmount: "1000000",
  otherCoin: "ubtc",
  otherAmount: "100000",
  salt: crypto.randomBytes(32)
};

const commitHash = sha256(JSON.stringify(orderData));

const commit = {
  typeUrl: "/aura.dex.v1beta1.MsgCommitOrder",
  value: {
    sender: userAddress,
    commitHash: commitHash
  }
};

const commitResult = await client.signAndBroadcast(userAddress, [commit], fee);
const commitId = extractCommitId(commitResult);
const revealDeadline = extractRevealDeadline(commitResult);

// Phase 2: Reveal order (within 5 minutes)
setTimeout(async () => {
  const reveal = {
    typeUrl: "/aura.dex.v1beta1.MsgRevealOrder",
    value: {
      sender: userAddress,
      commitId: commitId,
      orderType: orderData.orderType,
      auraAmount: orderData.auraAmount,
      otherCoin: orderData.otherCoin,
      otherAmount: orderData.otherAmount,
      salt: orderData.salt
    }
  };
  
  await client.signAndBroadcast(userAddress, [reveal], fee);
}, 60000); // Reveal after 1 minute
```

### Bridge Module

#### Cross-Chain Transfers

```typescript
// Lock tokens for transfer to PAW
const lockTokens = {
  typeUrl: "/aura.bridge.v1beta1.MsgLockTokens",
  value: {
    sender: userAddress,
    targetChain: "paw",
    recipient: pawAddress,
    amount: { denom: "uaura", amount: "1000000" }
  }
};

const lockResult = await client.signAndBroadcast(userAddress, [lockTokens], fee);
const transferId = extractTransferId(lockResult);

// Track transfer status
const trackTransfer = async (transferId: string) => {
  const query = { transfer: { transfer_id: transferId } };
  const status = await queryClient.wasm.queryContractSmart(
    bridgeModuleAddress,
    query
  );
  
  return {
    status: status.status,
    estimatedCompletion: status.estimated_completion,
    fraudProofWindowRemaining: calculateRemainingWindow(status)
  };
};

// Display progress to user
const interval = setInterval(async () => {
  const status = await trackTransfer(transferId);
  updateUI(status);
  
  if (status.status === 'completed') {
    clearInterval(interval);
    notifyUser('Transfer completed!');
  }
}, 10000); // Poll every 10 seconds
```

#### Address Linking

```typescript
// Link addresses across chains
const linkAddress = {
  typeUrl: "/aura.bridge.v1beta1.MsgLinkAddress",
  value: {
    auraAddress: userAddress,
    pawAddress: pawAddress,
    xaiAddress: xaiAddress,
    pawSignature: await signWithPAW(linkageData),
    xaiSignature: await signWithXAI(linkageData),
    signer: userAddress
  }
};
```

### Compliance Module

#### KYC Status Display

```typescript
// Query KYC record
const kycQuery = { kyc_record: { address: userAddress } };
const kycRecord = await queryClient.wasm.queryContractSmart(
  complianceModuleAddress,
  kycQuery
);

// Display KYC status in UI
const kycBadge = (
  <div className={`kyc-badge kyc-${kycRecord.kyc_level.toLowerCase()}`}>
    <span className="level">{formatKYCLevel(kycRecord.kyc_level)}</span>
    <span className="expires">
      Expires: {formatDate(kycRecord.expires_at)}
    </span>
    {isExpiringSoon(kycRecord.expires_at) && (
      <span className="warning">Renewal required</span>
    )}
  </div>
);
```

#### Sanctions Screening

```typescript
// Screen address before high-value transaction
const screenSanctions = async (address: string, amount: string) => {
  const query = {
    sanctions_screening: {
      address: address,
      force_refresh: parseFloat(amount) > 100000 // Force refresh for large amounts
    }
  };
  
  const result = await queryClient.wasm.queryContractSmart(
    complianceModuleAddress,
    query
  );
  
  if (result.status === 'SANCTIONS_MATCH' || result.status === 'SANCTIONS_CONFIRMED') {
    throw new Error('Address flagged in sanctions screening. Transaction blocked.');
  }
  
  if (result.requires_review) {
    return showWarning('Address requires manual review. Proceed with caution.');
  }
  
  return true;
};
```

#### Tax Report Generation

```typescript
// Generate tax report for user
const generateTaxReport = {
  typeUrl: "/aura.compliance.v1beta1.MsgGenerateTaxReport",
  value: {
    address: userAddress,
    taxYear: "2025",
    jurisdiction: "US",
    reportType: "1099-MISC",
    filePath: "" // Use default
  }
};

const result = await client.signAndBroadcast(userAddress, [generateTaxReport], fee);
const reportId = extractReportId(result);

// Download generated report
const downloadReport = async (reportId: string) => {
  const query = { tax_report: { address: userAddress, tax_year: "2025", jurisdiction: "US" } };
  const report = await queryClient.wasm.queryContractSmart(
    complianceModuleAddress,
    query
  );
  
  // Fetch PDF from IPFS or local storage
  const pdfUrl = report.file_path;
  window.open(pdfUrl, '_blank');
};
```

### VC Registry Module

#### Credential Display

```typescript
// Query user's verifiable credentials
const queryVCs = {
  list_user_vcs: {
    holder_address: userAddress,
    status_filter: "VC_STATUS_ACTIVE"
  }
};

const vcs = await queryClient.wasm.queryContractSmart(
  vcRegistryModuleAddress,
  queryVCs
);

// Display credentials in wallet
const credentialList = vcs.vcs.map(vc => (
  <div key={vc.vc_id} className="credential-card">
    <div className="credential-type">{formatVCType(vc.vc_type)}</div>
    <div className="credential-status">
      <StatusBadge status={vc.status} />
    </div>
    <div className="credential-expiry">
      {vc.expires_at ? `Expires: ${formatDate(vc.expires_at)}` : 'No expiration'}
    </div>
    {isExpiringSoon(vc.expires_at) && (
      <button onClick={() => renewCredential(vc.vc_id)}>Renew</button>
    )}
  </div>
));
```

#### Eligibility Check Before Minting

```typescript
// Check eligibility before showing mint button
const checkEligibility = async (vcType: string) => {
  const query = {
    validate_mint_eligibility: {
      holder_address: userAddress,
      vc_type: vcType
    }
  };
  
  const eligibility = await queryClient.wasm.queryContractSmart(
    vcRegistryModuleAddress,
    query
  );
  
  if (!eligibility.eligible) {
    return {
      canMint: false,
      reasons: eligibility.missing_requirements,
      currentCS: eligibility.current_cs,
      requiredCS: eligibility.required_cs
    };
  }
  
  return { canMint: true };
};

// Display eligibility UI
const renderMintButton = (vcType: string) => {
  const eligibility = await checkEligibility(vcType);
  
  if (!eligibility.canMint) {
    return (
      <div className="eligibility-warning">
        <p>Requirements not met:</p>
        <ul>
          {eligibility.reasons.map(reason => <li key={reason}>{reason}</li>)}
        </ul>
        <p>Confidence Score: {eligibility.currentCS} / {eligibility.requiredCS}</p>
      </div>
    );
  }
  
  return <button onClick={() => mintVC(vcType)}>Mint Credential</button>;
};
```

## Transaction Construction Best Practices

### Fee Estimation

```typescript
// Estimate gas for transaction
const estimateGas = async (messages: any[]) => {
  try {
    const gasEstimate = await client.simulate(userAddress, messages, "");
    const gasNeeded = Math.ceil(gasEstimate * 1.3); // Add 30% buffer
    
    const fee = {
      amount: [{ denom: "uaura", amount: String(gasNeeded * gasPrice) }],
      gas: String(gasNeeded)
    };
    
    return fee;
  } catch (error) {
    // Fallback to default fee
    return {
      amount: [{ denom: "uaura", amount: "50000" }],
      gas: "200000"
    };
  }
};
```

### Transaction Signing

```typescript
// Sign transaction with Keplr
const signWithKeplr = async (messages: any[]) => {
  await window.keplr.enable(chainId);
  const offlineSigner = window.keplr.getOfflineSigner(chainId);
  
  const client = await SigningCosmWasmClient.connectWithSigner(
    RPC_ENDPOINT,
    offlineSigner
  );
  
  const fee = await estimateGas(messages);
  return await client.signAndBroadcast(userAddress, messages, fee);
};
```

### Error Handling

```typescript
const executeTransaction = async (messages: any[]) => {
  try {
    const result = await client.signAndBroadcast(userAddress, messages, fee);
    
    if (result.code !== 0) {
      throw new Error(`Transaction failed: ${result.rawLog}`);
    }
    
    return {
      success: true,
      txHash: result.transactionHash,
      events: result.events
    };
  } catch (error) {
    // Parse error for user-friendly message
    const userMessage = parseErrorMessage(error);
    
    return {
      success: false,
      error: userMessage
    };
  }
};

const parseErrorMessage = (error: any): string => {
  const rawLog = error.message || error.rawLog || '';
  
  // Map common errors to user-friendly messages
  if (rawLog.includes('insufficient funds')) {
    return 'Insufficient balance to complete transaction';
  }
  if (rawLog.includes('slippage exceeded')) {
    return 'Price moved too much. Try increasing slippage tolerance.';
  }
  if (rawLog.includes('pool not found')) {
    return 'Trading pair not available';
  }
  if (rawLog.includes('sanctions match')) {
    return 'Transaction blocked: compliance issue detected';
  }
  
  return 'Transaction failed. Please try again.';
};
```

## Event Subscriptions

### WebSocket Connection

```typescript
import { Tendermint34Client } from "@cosmjs/tendermint-rpc";

const tendermintClient = await Tendermint34Client.connect(WS_ENDPOINT);

// Subscribe to new blocks
const blockSubscription = tendermintClient.subscribeNewBlockHeader();

for await (const header of blockSubscription) {
  console.log(`New block: ${header.height}`);
  // Process block events
}

// Subscribe to specific events
const eventQuery = "message.module='dex' AND message.action='swap'";
const eventSubscription = tendermintClient.subscribeTx(eventQuery);

for await (const event of eventSubscription) {
  console.log('DEX swap detected:', event);
  // Update UI with new swap
}
```

### Event Processing

```typescript
const processEvents = (events: any[]) => {
  events.forEach(event => {
    switch (event.type) {
      case 'pool_created':
        handlePoolCreated(event.attributes);
        break;
      case 'swap_executed':
        handleSwapExecuted(event.attributes);
        break;
      case 'session_created':
        handleSessionCreated(event.attributes);
        break;
      case 'vc_minted':
        handleVCMinted(event.attributes);
        break;
      // ... handle other events
    }
  });
};
```

## Security Best Practices

### Input Validation

```typescript
// Validate addresses
const isValidAddress = (address: string): boolean => {
  return address.startsWith('aura1') && address.length === 43;
};

// Validate amounts
const isValidAmount = (amount: string): boolean => {
  const parsed = parseFloat(amount);
  return !isNaN(parsed) && parsed > 0 && parsed <= Number.MAX_SAFE_INTEGER;
};

// Sanitize user input
const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[^\w\s.-]/gi, '');
};
```

### Private Key Management

```typescript
// NEVER store private keys in wallet code
// Use browser wallet extensions (Keplr, Cosmostation)
// Or hardware wallets (Ledger)

// Connect to Keplr
const connectKeplr = async () => {
  if (!window.keplr) {
    throw new Error('Keplr extension not installed');
  }
  
  await window.keplr.enable(chainId);
  const offlineSigner = window.keplr.getOfflineSigner(chainId);
  const accounts = await offlineSigner.getAccounts();
  
  return accounts[0].address;
};
```

### Rate Limiting

```typescript
// Client-side rate limiting for expensive operations
class RateLimiter {
  private timestamps: Map<string, number[]> = new Map();
  
  canExecute(operation: string, limitPerMinute: number): boolean {
    const now = Date.now();
    const timestamps = this.timestamps.get(operation) || [];
    
    // Remove timestamps older than 1 minute
    const recentTimestamps = timestamps.filter(ts => now - ts < 60000);
    
    if (recentTimestamps.length >= limitPerMinute) {
      return false;
    }
    
    recentTimestamps.push(now);
    this.timestamps.set(operation, recentTimestamps);
    return true;
  }
}

const rateLimiter = new RateLimiter();

// Use before expensive operations
if (!rateLimiter.canExecute('sanctions_screening', 10)) {
  throw new Error('Too many requests. Please wait before trying again.');
}
```

## UI/UX Guidelines

### Transaction Confirmation

Always display transaction details before signing:

```typescript
const TransactionConfirmation = ({ messages, fee }: Props) => (
  <div className="tx-confirmation">
    <h3>Confirm Transaction</h3>
    
    <div className="tx-details">
      {messages.map((msg, i) => (
        <div key={i} className="message">
          <div className="message-type">{formatMessageType(msg.typeUrl)}</div>
          <div className="message-data">{formatMessageData(msg.value)}</div>
        </div>
      ))}
    </div>
    
    <div className="tx-fee">
      <span>Network Fee:</span>
      <span>{formatFee(fee)}</span>
    </div>
    
    <div className="tx-total">
      <span>Total:</span>
      <span>{calculateTotal(messages, fee)}</span>
    </div>
    
    <div className="tx-actions">
      <button onClick={onReject}>Cancel</button>
      <button onClick={onConfirm} className="primary">Confirm</button>
    </div>
  </div>
);
```

### Loading States

```typescript
const TransactionStatus = ({ status }: Props) => {
  switch (status) {
    case 'signing':
      return <Spinner text="Waiting for signature..." />;
    case 'broadcasting':
      return <Spinner text="Broadcasting transaction..." />;
    case 'confirming':
      return <Spinner text="Confirming on blockchain..." />;
    case 'success':
      return <SuccessMessage />;
    case 'error':
      return <ErrorMessage />;
  }
};
```

### Error Messages

Provide clear, actionable error messages:

```typescript
const ErrorDisplay = ({ error }: Props) => {
  const { message, action } = parseError(error);
  
  return (
    <div className="error-message">
      <div className="error-icon">⚠️</div>
      <div className="error-text">{message}</div>
      {action && (
        <button onClick={action.handler} className="error-action">
          {action.label}
        </button>
      )}
    </div>
  );
};
```

## Testing

### Unit Tests

```typescript
describe('DEX Module Integration', () => {
  it('should create liquidity pool', async () => {
    const result = await createPool(testAddress, 'uaura', 'ubtc', amount1, amount2);
    expect(result.success).toBe(true);
    expect(result.poolId).toBeDefined();
  });
  
  it('should handle slippage protection', async () => {
    const quote = await getQuote('pool_1', 'uaura', '10000');
    const minOutput = calculateMinOutput(quote.estimated_output, 500);
    
    const result = await executeSwap('pool_1', coinIn, minOutput, 500);
    expect(parseInt(result.amountOut)).toBeGreaterThanOrEqual(parseInt(minOutput));
  });
});
```

### Integration Tests

```typescript
describe('Cross-Chain Transfer Flow', () => {
  it('should complete full transfer lifecycle', async () => {
    // Lock tokens
    const lockResult = await lockTokens(userAddress, 'paw', pawAddress, amount);
    expect(lockResult.transferId).toBeDefined();
    
    // Wait for validators
    await sleep(60000);
    
    // Check transfer status
    const status = await getTransferStatus(lockResult.transferId);
    expect(status.status).toBe('pending');
    
    // Wait for fraud proof window
    await sleep(86400000); // 24 hours
    
    // Finalize transfer
    const finalizeResult = await finalizeTransfer(lockResult.transferId);
    expect(finalizeResult.success).toBe(true);
  });
});
```

## Resources

- **API Documentation**: https://docs.aura.network/api
- **RPC Endpoints**: https://docs.aura.network/rpc
- **Chain Registry**: https://github.com/cosmos/chain-registry
- **Example Wallet**: https://github.com/aura-network/example-wallet
- **Support**: https://discord.gg/aura-network

## Changelog

### v1.0.0 (2025-12-09)
- Initial integration guide
- Coverage for Identity, DEX, Bridge, Compliance, VC Registry modules
- Transaction construction examples
- Event subscription patterns
- Security best practices

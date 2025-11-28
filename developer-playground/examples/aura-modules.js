// All AURA Module Examples - Comprehensive Collection

export const auraModuleExamples = {
    // ============ BRIDGE MODULE ============
    'bridge-query-transfers-js': {
        title: 'Query Bridge Transfers',
        description: 'Query cross-chain bridge transfers',
        category: 'bridge',
        language: 'javascript',
        code: `// Query bridge transfers
try {
    const transfers = await api.getBridgeTransfers();
    console.log('Bridge Transfers:', transfers);

    const params = await api.getBridgeParams();
    console.log('Bridge Parameters:', params);

    return { transfers, params };
} catch (error) {
    console.error('Error:', error);
}`
    },

    'bridge-initiate-transfer-js': {
        title: 'Initiate Bridge Transfer',
        description: 'Initiate a cross-chain transfer',
        category: 'bridge',
        language: 'javascript',
        code: `// Initiate bridge transfer
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/bridge/MsgInitiateTransfer',
    value: {
        sender: wallet.address,
        recipient: 'cosmos1...', // Destination address
        amount: [{ denom: 'uaura', amount: '1000000' }],
        destination_chain: 'cosmoshub-4'
    }
};

console.log('Bridge Transfer:', msg);
return { transaction: msg };`
    },

    // ============ COMPLIANCE MODULE ============
    'compliance-check-status-js': {
        title: 'Check Compliance Status',
        description: 'Check address compliance status',
        category: 'compliance',
        language: 'javascript',
        code: `// Check compliance status
const address = 'aura1...';

try {
    const status = await api.getComplianceStatus(address);
    console.log('Compliance Status:', status);
    console.log('KYC Verified:', status.kyc_verified);
    console.log('AML Cleared:', status.aml_cleared);
    console.log('Risk Level:', status.risk_level);

    return status;
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ CONFIDENCE SCORE MODULE ============
    'confidencescore-query-score-js': {
        title: 'Query Confidence Score',
        description: 'Query user confidence score and reputation',
        category: 'confidencescore',
        language: 'javascript',
        code: `// Query confidence score
const address = 'aura1...';

try {
    const score = await api.getConfidenceScore(address);
    console.log('Confidence Score:', score);
    console.log('Score Value:', score.score);
    console.log('Completed Routines:', score.completed_routines);
    console.log('Failed Routines:', score.failed_routines);
    console.log('Reputation:', score.reputation);

    return score;
} catch (error) {
    console.error('Error:', error);
}`
    },

    'confidencescore-complete-routine-js': {
        title: 'Complete Inclusion Routine',
        description: 'Mark an inclusion routine as complete',
        category: 'confidencescore',
        language: 'javascript',
        code: `// Complete inclusion routine
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/confidencescore/MsgCompleteRoutine',
    value: {
        address: wallet.address,
        routine_id: '1',
        proof: 'completion_proof_hash'
    }
};

console.log('Complete Routine:', msg);
return { transaction: msg };`
    },

    // ============ CRYPTOGRAPHY MODULE ============
    'cryptography-query-key-js': {
        title: 'Query Public Key',
        description: 'Query cryptographic public key',
        category: 'cryptography',
        language: 'javascript',
        code: `// Query public key
const address = 'aura1...';

try {
    const keyInfo = await api.getPublicKey(address);
    console.log('Public Key:', keyInfo);
    console.log('Key Type:', keyInfo.type);
    console.log('Algorithm:', keyInfo.algorithm);

    const params = await api.getCryptographyParams();
    console.log('Cryptography Parameters:', params);

    return { keyInfo, params };
} catch (error) {
    console.error('Error:', error);
}`
    },

    'cryptography-rotate-key-js': {
        title: 'Rotate Encryption Key',
        description: 'Rotate cryptographic keys',
        category: 'cryptography',
        language: 'javascript',
        code: `// Rotate encryption key
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/cryptography/MsgRotateKey',
    value: {
        address: wallet.address,
        new_public_key: 'new_key_base64',
        algorithm: 'ed25519'
    }
};

console.log('Rotate Key:', msg);
return { transaction: msg };`
    },

    // ============ DATA REGISTRY MODULE ============
    'dataregistry-query-items-js': {
        title: 'Query Data Items',
        description: 'Query registered data items',
        category: 'dataregistry',
        language: 'javascript',
        code: `// Query data items
try {
    const items = await api.getDataItems();
    console.log('Data Items:', items);

    if (items.data && items.data.length > 0) {
        const firstItem = await api.getDataItem(items.data[0].id);
        console.log('Item Details:', firstItem);
    }

    return items;
} catch (error) {
    console.error('Error:', error);
}`
    },

    'dataregistry-register-data-js': {
        title: 'Register Data Item',
        description: 'Register a new data item',
        category: 'dataregistry',
        language: 'javascript',
        code: `// Register data item
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/dataregistry/MsgRegisterData',
    value: {
        owner: wallet.address,
        data_hash: 'ipfs://QmHash...',
        metadata: JSON.stringify({
            name: 'My Data',
            type: 'document',
            size: 1024
        })
    }
};

console.log('Register Data:', msg);
return { transaction: msg };`
    },

    // ============ ECONOMIC SECURITY MODULE ============
    'economicsecurity-query-fees-js': {
        title: 'Query Dynamic Fees',
        description: 'Query current dynamic fee structure',
        category: 'economicsecurity',
        language: 'javascript',
        code: `// Query dynamic fees
try {
    const fees = await api.getDynamicFees();
    console.log('Dynamic Fees:', fees);
    console.log('Base Fee:', fees.base_fee);
    console.log('Priority Fee:', fees.priority_fee);

    const mevProtection = await api.getMevProtection();
    console.log('MEV Protection:', mevProtection);

    const params = await api.getEconomicSecurityParams();
    console.log('Parameters:', params);

    return { fees, mevProtection, params };
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ IDENTITY CHANGE MODULE ============
    'identitychange-query-requests-js': {
        title: 'Query Identity Change Requests',
        description: 'Query pending identity change requests',
        category: 'identitychange',
        language: 'javascript',
        code: `// Query identity change requests
try {
    const requests = await api.getIdentityChangeRequests();
    console.log('Identity Change Requests:', requests);

    if (requests.requests && requests.requests.length > 0) {
        const firstRequest = await api.getIdentityChangeRequest(requests.requests[0].id);
        console.log('Request Details:', firstRequest);
    }

    return requests;
} catch (error) {
    console.error('Error:', error);
}`
    },

    'identitychange-request-change-js': {
        title: 'Request Identity Change',
        description: 'Submit an identity change request',
        category: 'identitychange',
        language: 'javascript',
        code: `// Request identity change
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/identitychange/MsgRequestChange',
    value: {
        requester: wallet.address,
        old_identity: 'old_did:aura:...',
        new_identity: 'new_did:aura:...',
        reason: 'Identity update required',
        supporting_documents: ['ipfs://QmHash1...', 'ipfs://QmHash2...']
    }
};

console.log('Identity Change Request:', msg);
return { transaction: msg };`
    },

    // ============ INCLUSION ROUTINES MODULE ============
    'inclusionroutines-query-routines-js': {
        title: 'Query Inclusion Routines',
        description: 'Query available inclusion routines',
        category: 'inclusionroutines',
        language: 'javascript',
        code: `// Query inclusion routines
try {
    const routines = await api.getInclusionRoutines();
    console.log('Inclusion Routines:', routines);

    if (routines.routines && routines.routines.length > 0) {
        const routine = await api.getInclusionRoutine(routines.routines[0].id);
        console.log('Routine Details:', routine);
        console.log('Required Steps:', routine.steps);
        console.log('Reward:', routine.reward);
    }

    const params = await api.getInclusionRoutineParams();
    console.log('Parameters:', params);

    return { routines, params };
} catch (error) {
    console.error('Error:', error);
}`
    },

    'inclusionroutines-register-js': {
        title: 'Register Inclusion Routine',
        description: 'Register a new inclusion routine',
        category: 'inclusionroutines',
        language: 'javascript',
        code: `// Register inclusion routine
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/inclusionroutines/MsgRegisterRoutine',
    value: {
        creator: wallet.address,
        name: 'Basic Verification',
        description: 'Complete basic identity verification',
        steps: ['email', 'phone', 'document'],
        reward: '100',
        deadline_blocks: 1000
    }
};

console.log('Register Routine:', msg);
return { transaction: msg };`
    },

    // ============ MONITORING MODULE ============
    'monitoring-query-metrics-js': {
        title: 'Query Monitoring Metrics',
        description: 'Query system monitoring metrics',
        category: 'monitoring',
        language: 'javascript',
        code: `// Query monitoring metrics
try {
    const metrics = await api.getMonitoringMetrics();
    console.log('System Metrics:', metrics);
    console.log('Block Height:', metrics.block_height);
    console.log('Transaction Volume:', metrics.tx_volume);
    console.log('Active Validators:', metrics.active_validators);

    const alerts = await api.getMonitoringAlerts();
    console.log('Active Alerts:', alerts);

    return { metrics, alerts };
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ NETWORK SECURITY MODULE ============
    'networksecurity-query-status-js': {
        title: 'Query Network Security Status',
        description: 'Query network security status',
        category: 'networksecurity',
        language: 'javascript',
        code: `// Query network security status
try {
    const status = await api.getNetworkSecurityStatus();
    console.log('Network Security:', status);
    console.log('Threat Level:', status.threat_level);
    console.log('Active Threats:', status.active_threats);

    const params = await api.getNetworkSecurityParams();
    console.log('Parameters:', params);

    return { status, params };
} catch (error) {
    console.error('Error:', error);
}`
    },

    'networksecurity-report-peer-js': {
        title: 'Report Malicious Peer',
        description: 'Report a malicious network peer',
        category: 'networksecurity',
        language: 'javascript',
        code: `// Report malicious peer
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/networksecurity/MsgReportPeer',
    value: {
        reporter: wallet.address,
        peer_id: 'peer_id_...',
        reason: 'Suspicious behavior detected',
        evidence: 'evidence_hash'
    }
};

console.log('Report Peer:', msg);
return { transaction: msg };`
    },

    // ============ PREVALIDATION MODULE ============
    'prevalidation-check-tx-js': {
        title: 'Prevalidate Transaction',
        description: 'Check transaction before submission',
        category: 'prevalidation',
        language: 'javascript',
        code: `// Prevalidate transaction
const txHash = '0x...'; // Transaction hash

try {
    const status = await api.getPrevalidationStatus(txHash);
    console.log('Prevalidation Status:', status);
    console.log('Valid:', status.valid);
    console.log('Errors:', status.errors);
    console.log('Warnings:', status.warnings);

    return status;
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ PRIVACY MODULE ============
    'privacy-query-params-js': {
        title: 'Query Privacy Parameters',
        description: 'Query privacy module parameters',
        category: 'privacy',
        language: 'javascript',
        code: `// Query privacy parameters
try {
    const params = await api.getPrivacyParams();
    console.log('Privacy Parameters:', params);
    console.log('Mixing Enabled:', params.mixing_enabled);
    console.log('Ring Size:', params.ring_size);
    console.log('Confidential Transfers:', params.confidential_enabled);

    return params;
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ VALIDATOR SECURITY MODULE ============
    'validatorsecurity-query-status-js': {
        title: 'Query Validator Security',
        description: 'Query validator security status',
        category: 'validatorsecurity',
        language: 'javascript',
        code: `// Query validator security
const validatorAddr = 'auravaloper1...';

try {
    const status = await api.getValidatorSecurityStatus(validatorAddr);
    console.log('Validator Security:', status);
    console.log('Jail Status:', status.jailed);
    console.log('Security Score:', status.security_score);

    const slashing = await api.getValidatorSlashingEvents(validatorAddr);
    console.log('Slashing Events:', slashing);

    return { status, slashing };
} catch (error) {
    console.error('Error:', error);
}`
    },

    // ============ WALLET SECURITY MODULE ============
    'walletsecurity-query-status-js': {
        title: 'Query Wallet Security',
        description: 'Query wallet security status',
        category: 'walletsecurity',
        language: 'javascript',
        code: `// Query wallet security
const address = 'aura1...';

try {
    const status = await api.getWalletSecurityStatus(address);
    console.log('Wallet Security:', status);
    console.log('2FA Enabled:', status.two_factor_enabled);
    console.log('Biometric Enabled:', status.biometric_enabled);
    console.log('Last Activity:', status.last_activity);

    const sessions = await api.getWalletSessions(address);
    console.log('Active Sessions:', sessions);

    return { status, sessions };
} catch (error) {
    console.error('Error:', error);
}`
    },

    'walletsecurity-enable-2fa-js': {
        title: 'Enable 2FA',
        description: 'Enable two-factor authentication',
        category: 'walletsecurity',
        language: 'javascript',
        code: `// Enable 2FA for wallet
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const msg = {
    type: 'aura/walletsecurity/MsgEnable2FA',
    value: {
        address: wallet.address,
        totp_secret: 'base32_secret',
        recovery_codes: ['code1', 'code2', 'code3']
    }
};

console.log('Enable 2FA:', msg);
return { transaction: msg };`
    },

    // ============ DEX MODULE (UPDATED) ============
    'dex-query-pools-js': {
        title: 'Query DEX Pools',
        description: 'Query available liquidity pools',
        category: 'dex',
        language: 'javascript',
        code: `// Query DEX pools
try {
    const pools = await api.getPools();
    console.log('Liquidity Pools:', pools);

    if (pools.pools && pools.pools.length > 0) {
        const poolId = pools.pools[0].id;
        const pool = await api.getPool(poolId);
        console.log('Pool Details:', pool);

        const liquidity = await api.getPoolLiquidity(poolId);
        console.log('Pool Liquidity:', liquidity);
    }

    return pools;
} catch (error) {
    console.error('Error:', error);
}`
    },

    'dex-swap-tokens-js': {
        title: 'Swap Tokens on DEX',
        description: 'Execute a token swap',
        category: 'dex',
        language: 'javascript',
        code: `// Swap tokens on DEX
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const poolId = 1;
const tokenIn = 'uaura';
const amountIn = '1000000';

// Estimate swap
const estimate = await api.estimateSwap(poolId, tokenIn, amountIn);
console.log('Swap Estimate:', estimate);

const msg = {
    type: 'aura/dex/MsgSwap',
    value: {
        sender: wallet.address,
        pool_id: poolId,
        token_in: { denom: tokenIn, amount: amountIn },
        min_token_out: estimate.amount_out
    }
};

console.log('Swap Transaction:', msg);
return { transaction: msg, estimate };`
    },

    'dex-add-liquidity-js': {
        title: 'Add Liquidity to Pool',
        description: 'Add liquidity to a DEX pool',
        category: 'dex',
        language: 'javascript',
        code: `// Add liquidity to pool
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const poolId = 1;

const msg = {
    type: 'aura/dex/MsgAddLiquidity',
    value: {
        sender: wallet.address,
        pool_id: poolId,
        tokens: [
            { denom: 'uaura', amount: '1000000' },
            { denom: 'uusdc', amount: '1000000' }
        ],
        min_liquidity: '0'
    }
};

console.log('Add Liquidity:', msg);
return { transaction: msg };`
    },

    'dex-remove-liquidity-js': {
        title: 'Remove Liquidity from Pool',
        description: 'Withdraw liquidity from a DEX pool',
        category: 'dex',
        language: 'javascript',
        code: `// Remove liquidity from pool
if (!wallet.connected) {
    console.error('Please connect your wallet');
    return;
}

const poolId = 1;
const liquidityAmount = '1000000';

const msg = {
    type: 'aura/dex/MsgRemoveLiquidity',
    value: {
        sender: wallet.address,
        pool_id: poolId,
        liquidity: liquidityAmount,
        min_tokens: []
    }
};

console.log('Remove Liquidity:', msg);
return { transaction: msg };`
    }
};

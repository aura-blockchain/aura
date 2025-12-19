/**
 * Integration Tests for Wallet Flow
 * Tests complete end-to-end wallet operations
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock Chrome APIs
global.chrome = {
  storage: {
    local: {
      data: {},
      get: vi.fn((keys, callback) => {
        const result = {};
        if (Array.isArray(keys)) {
          keys.forEach(key => {
            if (global.chrome.storage.local.data[key]) {
              result[key] = global.chrome.storage.local.data[key];
            }
          });
        }
        callback(result);
      }),
      set: vi.fn((data, callback) => {
        Object.assign(global.chrome.storage.local.data, data);
        if (callback) callback();
      }),
      remove: vi.fn((keys, callback) => {
        if (Array.isArray(keys)) {
          keys.forEach(key => delete global.chrome.storage.local.data[key]);
        }
        if (callback) callback();
      }),
    },
  },
};

// Mock COSMOS_SDK
global.COSMOS_SDK = {
  config: {
    bech32Prefix: 'aura',
    coinDenom: 'uaura',
    coinDecimals: 6,
    restEndpoint: 'http://localhost:1317',
    rpcEndpoint: 'http://localhost:26657',
  },
  generatePrivateKey: () => {
    const key = new Uint8Array(32);
    crypto.getRandomValues(key);
    return key;
  },
  getPublicKey: async (privateKey) => {
    return new Uint8Array(33).fill(2);
  },
  publicKeyToAddress: (publicKey) => {
    return `aura1${'test'.repeat(8)}`;
  },
  bytesToHex: (bytes) => {
    return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('');
  },
  hexToBytes: (hex) => {
    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      bytes[i / 2] = parseInt(hex.substr(i, 2), 16);
    }
    return bytes;
  },
  getAccount: vi.fn(async (address) => ({
    address,
    accountNumber: 1,
    sequence: 0,
    pubKey: null,
  })),
  getBalance: vi.fn(async (address) => ([
    { denom: 'uaura', amount: '1000000000' },
  ])),
  buildTransferTx: vi.fn((params) => ({
    body: {
      messages: [{
        '@type': '/cosmos.bank.v1beta1.MsgSend',
        from_address: params.fromAddress,
        to_address: params.toAddress,
        amount: [{ denom: params.denom, amount: params.amount.toString() }],
      }],
      memo: params.memo || '',
    },
    auth_info: {
      signer_infos: [],
      fee: { amount: [{ denom: 'uaura', amount: '5000' }], gas: '200000' },
    },
    signatures: [],
  })),
  signTx: vi.fn(async (tx, privateKey, accountInfo, publicKey) => {
    tx.signatures = [new Uint8Array(64).fill(1)];
    tx.auth_info.signer_infos = [{
      public_key: {
        '@type': '/cosmos.crypto.secp256k1.PubKey',
        key: 'test-public-key',
      },
      mode_info: { single: { mode: 'SIGN_MODE_DIRECT' } },
      sequence: '0',
    }];
    return tx;
  }),
  broadcastTx: vi.fn(async (signedTx) => ({
    txhash: 'ABCDEF1234567890',
    code: 0,
    raw_log: 'success',
  })),
};

// Mock fetch
global.fetch = vi.fn();

describe('Wallet Integration Tests', () => {
  let WalletCore, StakingModule, GovernanceModule, DexModule;

  beforeEach(async () => {
    // Import modules
    WalletCore = (await import('../../src/wallet-core.js')).default || (await import('../../src/wallet-core.js'));
    StakingModule = (await import('../../src/staking.js')).default || (await import('../../src/staking.js'));
    GovernanceModule = (await import('../../src/governance.js')).default || (await import('../../src/governance.js'));
    DexModule = (await import('../../src/dex.js')).default || (await import('../../src/dex.js'));

    // Clear storage
    global.chrome.storage.local.data = {};
    vi.clearAllMocks();
  });

  afterEach(() => {
    global.chrome.storage.local.data = {};
  });

  describe('Wallet Creation and Management Flow', () => {
    it('should create, save, load, and delete wallet', async () => {
      // Create new wallet
      const newWallet = await WalletCore.generateWallet();

      expect(newWallet).toHaveProperty('address');
      expect(newWallet).toHaveProperty('privateKeyHex');
      expect(newWallet.address).toMatch(/^aura1/);
      expect(newWallet.privateKeyHex).toHaveLength(64);

      // Save wallet
      await WalletCore.saveWallet(newWallet.privateKeyHex, newWallet.address);

      // Verify saved
      expect(global.chrome.storage.local.data).toHaveProperty('walletPrivateKey');
      expect(global.chrome.storage.local.data).toHaveProperty('walletAddress');

      // Load wallet
      const loadedWallet = await WalletCore.loadWallet();

      expect(loadedWallet).not.toBeNull();
      expect(loadedWallet.address).toBe(newWallet.address);
      expect(loadedWallet.privateKeyHex).toBe(newWallet.privateKeyHex);

      // Delete wallet
      await WalletCore.deleteWallet();

      // Verify deleted
      expect(global.chrome.storage.local.data).not.toHaveProperty('walletPrivateKey');
      expect(global.chrome.storage.local.data).not.toHaveProperty('walletAddress');

      // Try to load deleted wallet
      const deletedWallet = await WalletCore.loadWallet();
      expect(deletedWallet).toBeNull();
    });

    it('should import wallet from private key', async () => {
      const testPrivateKey = '0'.repeat(64);

      // Import wallet
      const importedWallet = await WalletCore.importWallet(testPrivateKey);

      expect(importedWallet).toHaveProperty('address');
      expect(importedWallet.privateKeyHex).toBe(testPrivateKey);

      // Save and reload
      await WalletCore.saveWallet(importedWallet.privateKeyHex, importedWallet.address);
      const loadedWallet = await WalletCore.loadWallet();

      expect(loadedWallet.address).toBe(importedWallet.address);
    });

    it('should handle encrypted wallet storage', async () => {
      const password = 'test-password';
      const newWallet = await WalletCore.generateWallet();

      // Save with encryption
      await WalletCore.saveWallet(newWallet.privateKeyHex, newWallet.address, password);

      // Verify encrypted
      expect(global.chrome.storage.local.data).toHaveProperty('encryptedPrivateKey');
      expect(global.chrome.storage.local.data).toHaveProperty('keySalt');
      expect(global.chrome.storage.local.data.isKeyEncrypted).toBe(true);

      // Load with password
      const loadedWallet = await WalletCore.loadWallet(password);

      expect(loadedWallet.address).toBe(newWallet.address);
      expect(loadedWallet.privateKeyHex).toBe(newWallet.privateKeyHex);

      // Try to load with wrong password
      await expect(WalletCore.loadWallet('wrong-password')).rejects.toThrow();
    });
  });

  describe('Token Transfer Flow', () => {
    it('should complete full send transaction flow', async () => {
      // Setup wallet
      const wallet = await WalletCore.generateWallet();
      await WalletCore.saveWallet(wallet.privateKeyHex, wallet.address);

      const recipientAddress = 'aura1recipient1234567890123456789012';
      const amount = 100000;

      // Get account info
      const accountInfo = await COSMOS_SDK.getAccount(wallet.address);

      // Build transaction
      const tx = COSMOS_SDK.buildTransferTx({
        fromAddress: wallet.address,
        toAddress: recipientAddress,
        amount,
        denom: 'uaura',
        memo: 'Test transfer',
      });

      expect(tx.body.messages[0].to_address).toBe(recipientAddress);

      // Sign transaction
      const privateKey = COSMOS_SDK.hexToBytes(wallet.privateKeyHex);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const signedTx = await COSMOS_SDK.signTx(tx, privateKey, accountInfo, publicKey);

      expect(signedTx.signatures).toHaveLength(1);

      // Broadcast transaction
      const result = await COSMOS_SDK.broadcastTx(signedTx);

      expect(result).toHaveProperty('txhash');
      expect(result.code).toBe(0);
    });
  });

  describe('Staking Flow', () => {
    it('should query delegations and rewards', async () => {
      const wallet = await WalletCore.generateWallet();

      // Mock delegations response
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          delegation_responses: [
            {
              delegation: {
                delegator_address: wallet.address,
                validator_address: 'auravaloper1test',
              },
              balance: { denom: 'uaura', amount: '1000000' },
            },
          ],
        }),
      });

      const delegations = await StakingModule.queryDelegations(wallet.address);

      expect(delegations).toHaveLength(1);
      expect(delegations[0].balance.amount).toBe('1000000');

      // Calculate total staked
      const totalStaked = StakingModule.calculateTotalStaked(delegations);
      expect(totalStaked).toBe(1000000);

      // Mock rewards response
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          total: [{ denom: 'uaura', amount: '50000.5' }],
        }),
      });

      const rewards = await StakingModule.queryRewards(wallet.address);
      const totalRewards = StakingModule.calculateTotalRewards(rewards);

      expect(totalRewards).toBeGreaterThan(0);
    });

    it('should build delegation transaction', async () => {
      const wallet = await WalletCore.generateWallet();
      const validatorAddress = 'auravaloper1test123456789012345678901234';

      const tx = StakingModule.buildDelegateTx({
        delegatorAddress: wallet.address,
        validatorAddress,
        amount: 1000000,
      });

      expect(tx.body.messages[0]['@type']).toBe('/cosmos.staking.v1beta1.MsgDelegate');
      expect(tx.body.messages[0].validator_address).toBe(validatorAddress);
    });
  });

  describe('Governance Flow', () => {
    it('should query proposals and build vote transaction', async () => {
      // Mock proposals response
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          proposals: [
            {
              proposal_id: '1',
              status: 2,
              content: {
                title: 'Test Proposal',
                description: 'Test Description',
              },
              submit_time: '2024-01-01T00:00:00Z',
              deposit_end_time: '2024-01-08T00:00:00Z',
              voting_start_time: '2024-01-08T00:00:00Z',
              voting_end_time: '2024-01-15T00:00:00Z',
            },
          ],
        }),
      });

      const proposals = await GovernanceModule.queryProposals();

      expect(proposals).toHaveLength(1);
      expect(proposals[0].proposal_id).toBe('1');

      const formatted = GovernanceModule.formatProposal(proposals[0]);
      expect(formatted.isActive).toBe(true);

      // Build vote transaction
      const wallet = await WalletCore.generateWallet();
      const voteTx = GovernanceModule.buildVoteTx({
        proposalId: 1,
        voter: wallet.address,
        option: GovernanceModule.VOTE_OPTIONS.YES,
      });

      expect(voteTx.body.messages[0]['@type']).toBe('/cosmos.gov.v1beta1.MsgVote');
      expect(voteTx.body.messages[0].option).toBe(1);
    });
  });

  describe('DEX Flow', () => {
    it('should quote swap and build liquidity transactions', async () => {
      const wallet = await WalletCore.generateWallet();
      await WalletCore.saveWallet(wallet.privateKeyHex, wallet.address);

      const pool = {
        pool_id: '1',
        denom_a: 'uaura',
        denom_b: 'usdt',
        reserve_a: '1500000',
        reserve_b: '500000',
        total_lp_tokens: '5000000',
        fee_percentage: '0.003',
        protocol_fee_percentage: '0.0005',
      };

      const quote = DexModule.calculateSwapQuote(pool, 'uaura', 100000);

      expect(quote.denomOut).toBe('usdt');
      expect(parseInt(quote.expectedAmountOut, 10)).toBeGreaterThan(0);

      const swapTx = DexModule.buildSwapExactInTx({
        sender: wallet.address,
        poolId: pool.pool_id,
        denomIn: 'uaura',
        amountIn: 100000,
        minAmountOut: quote.minAmountOut,
        maxSlippageBps: quote.maxSlippageBps,
        memo: 'DEX swap test',
      });

      expect(swapTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgSwapExactIn');
      expect(swapTx.body.messages[0].coin_in.amount).toBe('100000');

      const privateKey = COSMOS_SDK.hexToBytes(wallet.privateKeyHex);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const accountInfo = await COSMOS_SDK.getAccount(wallet.address);

      const signedSwapTx = await COSMOS_SDK.signTx(swapTx, privateKey, accountInfo, publicKey);
      const swapResult = await COSMOS_SDK.broadcastTx(signedSwapTx);

      expect(swapResult.code).toBe(0);

      const addLiquidityTx = DexModule.buildAddLiquidityTx({
        provider: wallet.address,
        poolId: pool.pool_id,
        amountA: 500000,
        amountB: 250000,
        denomA: 'uaura',
        denomB: 'usdt',
      });

      expect(addLiquidityTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgAddLiquidity');
      expect(addLiquidityTx.body.messages[0].amount_a.amount).toBe('500000');

      const removeLiquidityTx = DexModule.buildRemoveLiquidityTx({
        provider: wallet.address,
        poolId: pool.pool_id,
        lpTokens: '250000',
      });

      expect(removeLiquidityTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgRemoveLiquidity');
      expect(removeLiquidityTx.body.messages[0].lp_tokens).toBe('250000');
    });
  });

  describe('Error Handling and Validation', () => {
    it('should validate addresses correctly', () => {
      expect(WalletCore.validateAddress('aura1test123456789012345678901234567890')).toBe(true);
      expect(WalletCore.validateAddress('invalid')).toBe(false);
      expect(WalletCore.validateAddress('')).toBe(false);
      expect(WalletCore.validateAddress(null)).toBe(false);
    });

    it('should handle API errors gracefully', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(StakingModule.queryDelegations('aura1test')).rejects.toThrow();
    });

    it('should handle invalid transaction parameters', async () => {
      const wallet = await WalletCore.generateWallet();

      expect(() =>
        GovernanceModule.buildVoteTx({
          proposalId: 1,
          voter: wallet.address,
          option: 999, // Invalid option
        })
      ).toThrow('Invalid vote option');
    });
  });

  describe('Complete User Workflow', () => {
    it('should simulate complete user journey: create wallet -> check balance -> delegate -> vote', async () => {
      // 1. Create wallet
      const wallet = await WalletCore.generateWallet();
      await WalletCore.saveWallet(wallet.privateKeyHex, wallet.address);

      expect(wallet.address).toMatch(/^aura1/);

      // 2. Check balance
      const balances = await COSMOS_SDK.getBalance(wallet.address);

      expect(balances).toHaveLength(1);
      expect(parseInt(balances[0].amount)).toBeGreaterThan(0);

      // 3. Query validators
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          validators: [
            {
              operator_address: 'auravaloper1test',
              status: 'BOND_STATUS_BONDED',
              commission: { rate: '0.05' },
            },
          ],
        }),
      });

      const validators = await StakingModule.queryValidators();
      expect(validators).toHaveLength(1);

      // 4. Delegate to validator
      const delegateTx = StakingModule.buildDelegateTx({
        delegatorAddress: wallet.address,
        validatorAddress: validators[0].operator_address,
        amount: 1000000,
      });

      const privateKey = COSMOS_SDK.hexToBytes(wallet.privateKeyHex);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const accountInfo = await COSMOS_SDK.getAccount(wallet.address);

      const signedDelegateTx = await COSMOS_SDK.signTx(delegateTx, privateKey, accountInfo, publicKey);
      const delegateResult = await COSMOS_SDK.broadcastTx(signedDelegateTx);

      expect(delegateResult.code).toBe(0);

      // 5. Query governance proposals
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          proposals: [
            {
              proposal_id: '1',
              status: 2, // VOTING_PERIOD
              content: { title: 'Community Proposal' },
            },
          ],
        }),
      });

      const proposals = await GovernanceModule.queryProposals(2);
      expect(proposals).toHaveLength(1);

      // 6. Vote on proposal
      const voteTx = GovernanceModule.buildVoteTx({
        proposalId: proposals[0].proposal_id,
        voter: wallet.address,
        option: GovernanceModule.VOTE_OPTIONS.YES,
      });

      const signedVoteTx = await COSMOS_SDK.signTx(voteTx, privateKey, accountInfo, publicKey);
      const voteResult = await COSMOS_SDK.broadcastTx(signedVoteTx);

      expect(voteResult.code).toBe(0);

      // 7. Clean up - delete wallet
      await WalletCore.deleteWallet();
      const deletedWallet = await WalletCore.loadWallet();

      expect(deletedWallet).toBeNull();
    });
  });
});

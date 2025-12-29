/**
 * AURA Wallet Connector
 * Shared Keplr wallet integration for all Aura dashboards.
 * Provides wallet detection, connection, chain suggestion, and transaction signing.
 *
 * Dependencies:
 *   - @cosmjs/stargate (for SigningStargateClient)
 *   - @cosmjs/proto-signing (for message types)
 *
 * Usage:
 *   import { AuraWalletConnector } from './wallet-connector.js';
 *   const wallet = new AuraWalletConnector();
 *   await wallet.connect();
 *   const result = await wallet.delegate(validatorAddress, amount);
 */

// Chain configuration for Aura testnet
const AURA_CHAIN_CONFIG = {
    chainId: 'aura-testnet-1',
    chainName: 'Aura Testnet',
    rpc: 'http://localhost:26657',
    rest: 'http://localhost:1317',
    bip44: {
        coinType: 118
    },
    bech32Config: {
        bech32PrefixAccAddr: 'aura',
        bech32PrefixAccPub: 'aurapub',
        bech32PrefixValAddr: 'auravaloper',
        bech32PrefixValPub: 'auravaloperpub',
        bech32PrefixConsAddr: 'auravalcons',
        bech32PrefixConsPub: 'auravalconspub'
    },
    currencies: [{
        coinDenom: 'AURA',
        coinMinimalDenom: 'uaura',
        coinDecimals: 6,
        coinGeckoId: 'aura'
    }],
    feeCurrencies: [{
        coinDenom: 'AURA',
        coinMinimalDenom: 'uaura',
        coinDecimals: 6,
        gasPriceStep: {
            low: 0.015,
            average: 0.025,
            high: 0.04
        }
    }],
    stakeCurrency: {
        coinDenom: 'AURA',
        coinMinimalDenom: 'uaura',
        coinDecimals: 6
    },
    features: ['ibc-transfer', 'ibc-go']
};

// Gas price configuration
const GAS_CONFIG = {
    denom: 'uaura',
    low: 0.015,
    average: 0.025,
    high: 0.04
};

// Default gas limits for different transaction types
const GAS_LIMITS = {
    delegate: 250000,
    undelegate: 250000,
    redelegate: 350000,
    claimRewards: 300000,
    vote: 150000,
    deposit: 200000,
    submitProposal: 300000,
    send: 100000,
    editValidator: 200000,
    unjail: 150000
};

/**
 * Event emitter mixin for wallet events
 */
class EventEmitter {
    constructor() {
        this._events = {};
    }

    on(event, listener) {
        if (!this._events[event]) {
            this._events[event] = [];
        }
        this._events[event].push(listener);
        return this;
    }

    off(event, listener) {
        if (!this._events[event]) return this;
        this._events[event] = this._events[event].filter(l => l !== listener);
        return this;
    }

    emit(event, ...args) {
        if (!this._events[event]) return false;
        this._events[event].forEach(listener => listener(...args));
        return true;
    }
}

/**
 * AuraWalletConnector - Keplr wallet integration for Aura dashboards
 */
export class AuraWalletConnector extends EventEmitter {
    constructor(options = {}) {
        super();

        this.chainId = options.chainId || AURA_CHAIN_CONFIG.chainId;
        this.rpcEndpoint = options.rpc || AURA_CHAIN_CONFIG.rpc;
        this.restEndpoint = options.rest || AURA_CHAIN_CONFIG.rest;

        this.address = null;
        this.signer = null;
        this.signingClient = null;
        this.connected = false;

        // CosmJS modules (loaded dynamically)
        this._cosmjs = null;

        // Bind methods
        this._handleAccountChange = this._handleAccountChange.bind(this);

        // Setup Keplr event listeners
        if (typeof window !== 'undefined') {
            window.addEventListener('keplr_keystorechange', this._handleAccountChange);
        }
    }

    /**
     * Check if Keplr wallet is installed
     */
    static isKeplrInstalled() {
        return typeof window !== 'undefined' && !!window.keplr;
    }

    /**
     * Check if Keplr is available (installed and accessible)
     */
    isAvailable() {
        return AuraWalletConnector.isKeplrInstalled();
    }

    /**
     * Get Keplr wallet reference
     */
    getKeplr() {
        if (!this.isAvailable()) {
            throw new WalletError('Keplr wallet is not installed', 'KEPLR_NOT_INSTALLED');
        }
        return window.keplr;
    }

    /**
     * Suggest the Aura chain to Keplr (adds it if not present)
     */
    async suggestChain() {
        const keplr = this.getKeplr();

        try {
            await keplr.experimentalSuggestChain(AURA_CHAIN_CONFIG);
        } catch (error) {
            // Chain may already exist, which is fine
            if (!error.message?.includes('already exists')) {
                console.warn('Chain suggestion warning:', error.message);
            }
        }
    }

    /**
     * Connect to Keplr wallet
     * @returns {Promise<string>} Connected wallet address
     */
    async connect() {
        if (!this.isAvailable()) {
            throw new WalletError(
                'Keplr wallet is not installed. Please install Keplr extension from https://www.keplr.app',
                'KEPLR_NOT_INSTALLED'
            );
        }

        try {
            // Suggest chain first
            await this.suggestChain();

            // Enable the chain
            const keplr = this.getKeplr();
            await keplr.enable(this.chainId);

            // Get offline signer
            this.signer = keplr.getOfflineSigner(this.chainId);

            // Get accounts
            const accounts = await this.signer.getAccounts();
            if (!accounts || accounts.length === 0) {
                throw new WalletError('No accounts found in wallet', 'NO_ACCOUNTS');
            }

            this.address = accounts[0].address;
            this.connected = true;

            // Initialize signing client
            await this._initSigningClient();

            // Emit connection event
            this.emit('connect', { address: this.address });

            return this.address;
        } catch (error) {
            this.connected = false;
            this.address = null;
            this.signer = null;

            if (error instanceof WalletError) {
                throw error;
            }

            if (error.message?.includes('rejected')) {
                throw new WalletError('Connection request was rejected', 'USER_REJECTED');
            }

            throw new WalletError(`Failed to connect wallet: ${error.message}`, 'CONNECTION_FAILED');
        }
    }

    /**
     * Disconnect wallet
     */
    disconnect() {
        this.address = null;
        this.signer = null;
        this.signingClient = null;
        this.connected = false;

        this.emit('disconnect');
    }

    /**
     * Initialize CosmJS signing client
     */
    async _initSigningClient() {
        if (!this.signer) {
            throw new WalletError('Signer not initialized', 'NO_SIGNER');
        }

        // Load CosmJS modules
        await this._loadCosmJS();

        const { SigningStargateClient, GasPrice } = this._cosmjs;

        this.signingClient = await SigningStargateClient.connectWithSigner(
            this.rpcEndpoint,
            this.signer,
            {
                gasPrice: GasPrice.fromString(`${GAS_CONFIG.average}${GAS_CONFIG.denom}`)
            }
        );
    }

    /**
     * Load CosmJS modules dynamically
     */
    async _loadCosmJS() {
        if (this._cosmjs) return;

        // Check if CosmJS is available globally (from CDN or bundled)
        if (typeof window !== 'undefined' && window.CosmJS) {
            this._cosmjs = window.CosmJS;
            return;
        }

        // Try dynamic import (works with bundlers)
        try {
            const stargate = await import('@cosmjs/stargate');
            const protoSigning = await import('@cosmjs/proto-signing');

            this._cosmjs = {
                SigningStargateClient: stargate.SigningStargateClient,
                GasPrice: stargate.GasPrice,
                coins: stargate.coins,
                Registry: protoSigning.Registry,
                DirectSecp256k1HdWallet: protoSigning.DirectSecp256k1HdWallet
            };
        } catch (error) {
            throw new WalletError(
                'CosmJS modules not available. Include @cosmjs/stargate in your build or load from CDN.',
                'COSMJS_NOT_LOADED'
            );
        }
    }

    /**
     * Handle Keplr account changes
     */
    async _handleAccountChange() {
        if (!this.connected) return;

        try {
            const keplr = this.getKeplr();
            const key = await keplr.getKey(this.chainId);

            if (key.bech32Address !== this.address) {
                this.address = key.bech32Address;
                await this._initSigningClient();
                this.emit('accountChange', { address: this.address });
            }
        } catch (error) {
            console.error('Account change handling failed:', error);
        }
    }

    /**
     * Get account balance
     * @param {string} [address] - Address to check (defaults to connected address)
     * @param {string} [denom] - Token denom (defaults to uaura)
     */
    async getBalance(address = this.address, denom = 'uaura') {
        this._requireConnection();

        const balance = await this.signingClient.getBalance(address, denom);
        return {
            denom: balance.denom,
            amount: balance.amount,
            formatted: parseFloat(balance.amount) / 1e6
        };
    }

    /**
     * Build fee object for transactions
     */
    _buildFee(gasLimit, gasPriceLevel = 'average') {
        const gasPrice = GAS_CONFIG[gasPriceLevel] || GAS_CONFIG.average;
        const feeAmount = Math.ceil(gasLimit * gasPrice);

        return {
            amount: [{ denom: GAS_CONFIG.denom, amount: String(feeAmount) }],
            gas: String(gasLimit)
        };
    }

    /**
     * Sign and broadcast a transaction
     */
    async _signAndBroadcast(messages, memo = '', gasLimit = null) {
        this._requireConnection();

        const fee = gasLimit
            ? this._buildFee(gasLimit)
            : 'auto';

        try {
            const result = await this.signingClient.signAndBroadcast(
                this.address,
                messages,
                fee,
                memo
            );

            if (result.code !== 0) {
                throw new WalletError(
                    `Transaction failed: ${result.rawLog}`,
                    'TX_FAILED',
                    { code: result.code, rawLog: result.rawLog }
                );
            }

            return {
                success: true,
                transactionHash: result.transactionHash,
                height: result.height,
                gasUsed: result.gasUsed,
                gasWanted: result.gasWanted,
                events: result.events
            };
        } catch (error) {
            if (error instanceof WalletError) throw error;

            if (error.message?.includes('rejected')) {
                throw new WalletError('Transaction was rejected by user', 'USER_REJECTED');
            }

            throw new WalletError(`Transaction failed: ${error.message}`, 'TX_FAILED');
        }
    }

    /**
     * Check if wallet is connected
     */
    _requireConnection() {
        if (!this.connected || !this.signingClient) {
            throw new WalletError('Wallet not connected', 'NOT_CONNECTED');
        }
    }

    // ==================== STAKING TRANSACTIONS ====================

    /**
     * Delegate tokens to a validator
     * @param {string} validatorAddress - Validator operator address (auravaloper...)
     * @param {string|number} amount - Amount in uaura (micro-units)
     * @param {string} [memo] - Optional transaction memo
     */
    async delegate(validatorAddress, amount, memo = 'Delegated via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
            value: {
                delegatorAddress: this.address,
                validatorAddress: validatorAddress,
                amount: {
                    denom: 'uaura',
                    amount: String(amount)
                }
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.delegate);
    }

    /**
     * Undelegate tokens from a validator
     * @param {string} validatorAddress - Validator operator address
     * @param {string|number} amount - Amount in uaura
     * @param {string} [memo] - Optional transaction memo
     */
    async undelegate(validatorAddress, amount, memo = 'Undelegated via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgUndelegate',
            value: {
                delegatorAddress: this.address,
                validatorAddress: validatorAddress,
                amount: {
                    denom: 'uaura',
                    amount: String(amount)
                }
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.undelegate);
    }

    /**
     * Redelegate tokens between validators
     * @param {string} srcValidatorAddress - Source validator address
     * @param {string} dstValidatorAddress - Destination validator address
     * @param {string|number} amount - Amount in uaura
     * @param {string} [memo] - Optional transaction memo
     */
    async redelegate(srcValidatorAddress, dstValidatorAddress, amount, memo = 'Redelegated via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
            value: {
                delegatorAddress: this.address,
                validatorSrcAddress: srcValidatorAddress,
                validatorDstAddress: dstValidatorAddress,
                amount: {
                    denom: 'uaura',
                    amount: String(amount)
                }
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.redelegate);
    }

    /**
     * Claim staking rewards from a specific validator
     * @param {string} validatorAddress - Validator operator address
     * @param {string} [memo] - Optional transaction memo
     */
    async claimRewardsFromValidator(validatorAddress, memo = 'Claimed rewards via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
            value: {
                delegatorAddress: this.address,
                validatorAddress: validatorAddress
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.claimRewards);
    }

    /**
     * Claim staking rewards from all validators
     * @param {string[]} validatorAddresses - Array of validator addresses
     * @param {string} [memo] - Optional transaction memo
     */
    async claimAllRewards(validatorAddresses, memo = 'Claimed all rewards via Aura Dashboard') {
        this._requireConnection();

        const messages = validatorAddresses.map(validatorAddress => ({
            typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
            value: {
                delegatorAddress: this.address,
                validatorAddress: validatorAddress
            }
        }));

        // Increase gas limit based on number of validators
        const gasLimit = GAS_LIMITS.claimRewards * Math.min(validatorAddresses.length, 10);

        return this._signAndBroadcast(messages, memo, gasLimit);
    }

    // ==================== GOVERNANCE TRANSACTIONS ====================

    /**
     * Vote on a governance proposal
     * @param {string|number} proposalId - Proposal ID
     * @param {string|number} option - Vote option (1=Yes, 2=Abstain, 3=No, 4=NoWithVeto)
     * @param {string} [memo] - Optional transaction memo
     */
    async vote(proposalId, option, memo = 'Voted via Aura Governance Dashboard') {
        this._requireConnection();

        // Convert string options to numeric values
        const optionMap = {
            'yes': 1, 'VOTE_OPTION_YES': 1, '1': 1,
            'abstain': 2, 'VOTE_OPTION_ABSTAIN': 2, '2': 2,
            'no': 3, 'VOTE_OPTION_NO': 3, '3': 3,
            'no_with_veto': 4, 'nowithveto': 4, 'VOTE_OPTION_NO_WITH_VETO': 4, '4': 4
        };

        const voteOption = typeof option === 'number'
            ? option
            : optionMap[String(option).toLowerCase()] || 1;

        const msg = {
            typeUrl: '/cosmos.gov.v1beta1.MsgVote',
            value: {
                proposalId: String(proposalId),
                voter: this.address,
                option: voteOption
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.vote);
    }

    /**
     * Vote on a governance proposal with weighted votes
     * @param {string|number} proposalId - Proposal ID
     * @param {Array<{option: number, weight: string}>} options - Weighted vote options
     * @param {string} [memo] - Optional transaction memo
     */
    async voteWeighted(proposalId, options, memo = 'Voted (weighted) via Aura Governance Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.gov.v1beta1.MsgVoteWeighted',
            value: {
                proposalId: String(proposalId),
                voter: this.address,
                options: options
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.vote);
    }

    /**
     * Deposit tokens to a governance proposal
     * @param {string|number} proposalId - Proposal ID
     * @param {string|number} amount - Amount in uaura
     * @param {string} [memo] - Optional transaction memo
     */
    async depositToProposal(proposalId, amount, memo = 'Deposit via Aura Governance Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.gov.v1beta1.MsgDeposit',
            value: {
                proposalId: String(proposalId),
                depositor: this.address,
                amount: [{
                    denom: 'uaura',
                    amount: String(amount)
                }]
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.deposit);
    }

    /**
     * Submit a new governance proposal
     * @param {Object} content - Proposal content
     * @param {string} content.title - Proposal title
     * @param {string} content.description - Proposal description
     * @param {string} [content.type='text'] - Proposal type (text, parameter_change, software_upgrade)
     * @param {string|number} initialDeposit - Initial deposit in uaura
     * @param {string} [memo] - Optional transaction memo
     */
    async submitProposal(content, initialDeposit, memo = 'Proposal submitted via Aura Governance Dashboard') {
        this._requireConnection();

        const typeUrls = {
            'text': '/cosmos.gov.v1beta1.TextProposal',
            'parameter_change': '/cosmos.params.v1beta1.ParameterChangeProposal',
            'software_upgrade': '/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal'
        };

        const contentTypeUrl = typeUrls[content.type || 'text'];

        const msg = {
            typeUrl: '/cosmos.gov.v1beta1.MsgSubmitProposal',
            value: {
                content: {
                    typeUrl: contentTypeUrl,
                    value: {
                        title: content.title,
                        description: content.description,
                        ...(content.changes && { changes: content.changes }),
                        ...(content.plan && { plan: content.plan })
                    }
                },
                initialDeposit: [{
                    denom: 'uaura',
                    amount: String(initialDeposit)
                }],
                proposer: this.address
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.submitProposal);
    }

    // ==================== VALIDATOR OPERATIONS ====================

    /**
     * Edit validator description/commission
     * @param {Object} changes - Validator changes
     * @param {string} [memo] - Optional transaction memo
     */
    async editValidator(changes, memo = 'Edited validator via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgEditValidator',
            value: {
                description: {
                    moniker: changes.moniker || '',
                    identity: changes.identity || '',
                    website: changes.website || '',
                    securityContact: changes.securityContact || '',
                    details: changes.details || ''
                },
                validatorAddress: changes.validatorAddress || this.address.replace('aura', 'auravaloper'),
                commissionRate: changes.commissionRate ? String(changes.commissionRate) : null,
                minSelfDelegation: changes.minSelfDelegation ? String(changes.minSelfDelegation) : null
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.editValidator);
    }

    /**
     * Unjail a jailed validator
     * @param {string} [validatorAddress] - Validator address (defaults to derived from connected account)
     * @param {string} [memo] - Optional transaction memo
     */
    async unjail(validatorAddress = null, memo = 'Unjailed via Aura Dashboard') {
        this._requireConnection();

        const valAddr = validatorAddress || this.address.replace('aura', 'auravaloper');

        const msg = {
            typeUrl: '/cosmos.slashing.v1beta1.MsgUnjail',
            value: {
                validatorAddr: valAddr
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.unjail);
    }

    /**
     * Withdraw validator commission
     * @param {string} [validatorAddress] - Validator address
     * @param {string} [memo] - Optional transaction memo
     */
    async withdrawValidatorCommission(validatorAddress = null, memo = 'Withdrew commission via Aura Dashboard') {
        this._requireConnection();

        const valAddr = validatorAddress || this.address.replace('aura', 'auravaloper');

        const msg = {
            typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission',
            value: {
                validatorAddress: valAddr
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.claimRewards);
    }

    // ==================== UTILITY METHODS ====================

    /**
     * Send tokens to another address
     * @param {string} recipientAddress - Recipient bech32 address
     * @param {string|number} amount - Amount in uaura
     * @param {string} [memo] - Optional transaction memo
     */
    async send(recipientAddress, amount, memo = 'Sent via Aura Dashboard') {
        this._requireConnection();

        const msg = {
            typeUrl: '/cosmos.bank.v1beta1.MsgSend',
            value: {
                fromAddress: this.address,
                toAddress: recipientAddress,
                amount: [{
                    denom: 'uaura',
                    amount: String(amount)
                }]
            }
        };

        return this._signAndBroadcast([msg], memo, GAS_LIMITS.send);
    }

    /**
     * Convert display amount (AURA) to micro units (uaura)
     * @param {number} amount - Amount in AURA
     * @returns {string} Amount in uaura
     */
    static toMicroUnits(amount) {
        return String(Math.floor(parseFloat(amount) * 1e6));
    }

    /**
     * Convert micro units (uaura) to display amount (AURA)
     * @param {string|number} amount - Amount in uaura
     * @returns {number} Amount in AURA
     */
    static fromMicroUnits(amount) {
        return parseFloat(amount) / 1e6;
    }

    /**
     * Format address for display (truncated)
     * @param {string} address - Full bech32 address
     * @param {number} [prefixLen=10] - Prefix length
     * @param {number} [suffixLen=6] - Suffix length
     */
    static formatAddress(address, prefixLen = 10, suffixLen = 6) {
        if (!address) return '';
        if (address.length <= prefixLen + suffixLen) return address;
        return `${address.slice(0, prefixLen)}...${address.slice(-suffixLen)}`;
    }

    /**
     * Get explorer URL for a transaction
     * @param {string} txHash - Transaction hash
     * @returns {string} Explorer URL
     */
    getExplorerUrl(txHash) {
        return `http://localhost:8088/transactions/${txHash}`;
    }
}

/**
 * Custom error class for wallet operations
 */
export class WalletError extends Error {
    constructor(message, code, details = {}) {
        super(message);
        this.name = 'WalletError';
        this.code = code;
        this.details = details;
    }
}

// Error codes
export const WalletErrorCodes = {
    KEPLR_NOT_INSTALLED: 'KEPLR_NOT_INSTALLED',
    NO_ACCOUNTS: 'NO_ACCOUNTS',
    USER_REJECTED: 'USER_REJECTED',
    CONNECTION_FAILED: 'CONNECTION_FAILED',
    NOT_CONNECTED: 'NOT_CONNECTED',
    NO_SIGNER: 'NO_SIGNER',
    COSMJS_NOT_LOADED: 'COSMJS_NOT_LOADED',
    TX_FAILED: 'TX_FAILED'
};

// Export chain config for external use
export const CHAIN_CONFIG = AURA_CHAIN_CONFIG;
export const GAS_PRICES = GAS_CONFIG;

// Create default instance for simple usage
let defaultConnector = null;

/**
 * Get or create the default wallet connector instance
 */
export function getWalletConnector(options = {}) {
    if (!defaultConnector) {
        defaultConnector = new AuraWalletConnector(options);
    }
    return defaultConnector;
}

// Export for different module systems
if (typeof window !== 'undefined') {
    window.AuraWalletConnector = AuraWalletConnector;
    window.WalletError = WalletError;
    window.getWalletConnector = getWalletConnector;
}

export default AuraWalletConnector;

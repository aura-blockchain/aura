/**
 * AURA Network Dashboard Configuration
 * Centralized configuration for all dashboards
 */

const AuraConfig = {
    // Network Configuration
    network: {
        name: 'AURA',
        chainId: 'aura-1',
        denom: 'aura',
        decimals: 6,
        bech32Prefix: {
            account: 'aura',
            validator: 'auravaloper',
            consensus: 'auravalcons'
        }
    },

    // API Endpoints
    endpoints: {
        // REST API (LCD)
        rest: process.env.AURA_REST_ENDPOINT || 'http://localhost:1317',

        // RPC Endpoint
        rpc: process.env.AURA_RPC_ENDPOINT || 'http://localhost:26657',

        // gRPC Web (if needed)
        grpc: process.env.AURA_GRPC_ENDPOINT || 'http://localhost:9090'
    },

    // Cosmos SDK Standard Endpoints
    api: {
        // Base Endpoints
        nodeInfo: '/cosmos/base/tendermint/v1beta1/node_info',

        // Staking Module
        staking: {
            pool: '/cosmos/staking/v1beta1/pool',
            params: '/cosmos/staking/v1beta1/params',
            validators: '/cosmos/staking/v1beta1/validators',
            validator: (address) => `/cosmos/staking/v1beta1/validators/${address}`,
            validatorDelegations: (address) => `/cosmos/staking/v1beta1/validators/${address}/delegations`,
            delegations: (address) => `/cosmos/staking/v1beta1/delegations/${address}`,
            unbondingDelegations: (address) => `/cosmos/staking/v1beta1/delegators/${address}/unbonding_delegations`,
            redelegations: (address) => `/cosmos/staking/v1beta1/delegators/${address}/redelegations`
        },

        // Distribution Module (Rewards)
        distribution: {
            params: '/cosmos/distribution/v1beta1/params',
            validatorOutstandingRewards: (address) => `/cosmos/distribution/v1beta1/validators/${address}/outstanding_rewards`,
            validatorCommission: (address) => `/cosmos/distribution/v1beta1/validators/${address}/commission`,
            validatorSlashes: (address) => `/cosmos/distribution/v1beta1/validators/${address}/slashes`,
            delegationRewards: (delegator, validator) => validator
                ? `/cosmos/distribution/v1beta1/delegators/${delegator}/rewards/${validator}`
                : `/cosmos/distribution/v1beta1/delegators/${delegator}/rewards`,
            delegatorWithdrawAddress: (address) => `/cosmos/distribution/v1beta1/delegators/${address}/withdraw_address`
        },

        // Governance Module
        governance: {
            params: (paramType) => `/cosmos/gov/v1beta1/params/${paramType}`, // deposit, voting, tallying
            proposals: '/cosmos/gov/v1beta1/proposals',
            proposal: (id) => `/cosmos/gov/v1beta1/proposals/${id}`,
            proposalVotes: (id) => `/cosmos/gov/v1beta1/proposals/${id}/votes`,
            proposalVote: (id, voter) => `/cosmos/gov/v1beta1/proposals/${id}/votes/${voter}`,
            proposalDeposits: (id) => `/cosmos/gov/v1beta1/proposals/${id}/deposits`,
            proposalDeposit: (id, depositor) => `/cosmos/gov/v1beta1/proposals/${id}/deposits/${depositor}`,
            proposalTally: (id) => `/cosmos/gov/v1beta1/proposals/${id}/tally`
        },

        // Slashing Module
        slashing: {
            params: '/cosmos/slashing/v1beta1/params',
            signingInfos: '/cosmos/slashing/v1beta1/signing_infos',
            signingInfo: (consAddress) => `/cosmos/slashing/v1beta1/signing_infos/${consAddress}`
        },

        // Bank Module
        bank: {
            balance: (address, denom) => denom
                ? `/cosmos/bank/v1beta1/balances/${address}/${denom}`
                : `/cosmos/bank/v1beta1/balances/${address}`,
            totalSupply: '/cosmos/bank/v1beta1/supply',
            supplyOf: (denom) => `/cosmos/bank/v1beta1/supply/${denom}`
        },

        // AURA Custom Modules
        aura: {
            // Validator Security
            validatorSecurity: {
                params: '/aura/validatorsecurity/v1beta1/params',
                jailedValidators: '/aura/validatorsecurity/v1beta1/jailed',
                slashEvents: (address) => `/aura/validatorsecurity/v1beta1/slash_events/${address}`
            },

            // DEX Module
            dex: {
                params: '/aura/dex/v1beta1/params',
                pools: '/aura/dex/v1beta1/pools',
                pool: (id) => `/aura/dex/v1beta1/pools/${id}`,
                swapHistory: (address) => `/aura/dex/v1beta1/swaps/${address}`
            },

            // Governance Extensions
            governance: {
                params: '/aura/governance/v1beta1/params',
                proposalStats: '/aura/governance/v1beta1/stats'
            },

            // Bridge Module
            bridge: {
                params: '/aura/bridge/v1beta1/params',
                transfers: '/aura/bridge/v1beta1/transfers'
            },

            // Network Security
            networkSecurity: {
                params: '/aura/networksecurity/v1beta1/params',
                reputation: (address) => `/aura/networksecurity/v1beta1/reputation/${address}`
            }
        }
    },

    // Dashboard Settings
    dashboard: {
        // Refresh intervals (milliseconds)
        refreshInterval: {
            fast: 5000,    // 5 seconds - for critical data
            normal: 15000, // 15 seconds - for general data
            slow: 60000    // 60 seconds - for static data
        },

        // Cache settings
        cache: {
            enabled: true,
            ttl: 30000 // 30 seconds
        },

        // Mock mode for development
        mockMode: process.env.AURA_MOCK_MODE === 'true' || false,

        // UI Settings
        ui: {
            itemsPerPage: 20,
            chartColors: {
                primary: '#6366f1',
                success: '#10b981',
                warning: '#f59e0b',
                danger: '#ef4444',
                info: '#3b82f6'
            }
        }
    },

    // Governance Parameters
    governance: {
        minDeposit: 10000, // AURA tokens (will be multiplied by 10^6 for micro-aura)
        votingPeriod: 14,  // days
        depositPeriod: 14, // days
        quorum: 0.334,
        threshold: 0.5,
        vetoThreshold: 0.334
    },

    // Staking Parameters
    staking: {
        unbondingTime: 21,     // days
        maxValidators: 100,
        maxEntries: 7,
        historicalEntries: 10000,
        bondDenom: 'aura'
    },

    // Slashing Parameters
    slashing: {
        signedBlocksWindow: 10000,
        minSignedPerWindow: 0.5,
        downtimeJailDuration: 600, // seconds
        slashFractionDoubleSign: 0.05,
        slashFractionDowntime: 0.0001
    }
};

// Export for different module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AuraConfig;
}
if (typeof window !== 'undefined') {
    window.AuraConfig = AuraConfig;
}

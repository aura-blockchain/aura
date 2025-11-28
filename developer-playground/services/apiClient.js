// AURA Blockchain API Client

export class APIClient {
    constructor(network = 'testnet') {
        this.network = network;
        this.customEndpoint = null;
        this.endpoints = {
            local: 'http://localhost:1317',
            testnet: 'https://testnet-api.aura.zone',
            mainnet: 'https://api.aura.zone'
        };
        this.chainIds = {
            local: 'aura-local',
            testnet: 'aura-testnet-1',
            mainnet: 'aura-1'
        };
    }

    getEndpoint() {
        return this.customEndpoint || this.endpoints[this.network];
    }

    setNetwork(network) {
        if (!this.endpoints[network]) {
            throw new Error(`Invalid network: ${network}`);
        }
        this.network = network;
        this.customEndpoint = null;
    }

    setCustomEndpoint(endpoint) {
        this.customEndpoint = endpoint;
    }

    getChainId() {
        return this.chainIds[this.network] || 'paw-1';
    }

    async request(path, options = {}) {
        const endpoint = this.getEndpoint();
        const url = `${endpoint}${path}`;

        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(`API request failed: ${response.status} ${error}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API request error:', error);
            throw error;
        }
    }

    // Bank Module
    async getBalance(address, denom = null) {
        const path = denom
            ? `/cosmos/bank/v1beta1/balances/${address}/by_denom?denom=${denom}`
            : `/cosmos/bank/v1beta1/balances/${address}`;
        return await this.request(path);
    }

    async getAllBalances(address) {
        return await this.request(`/cosmos/bank/v1beta1/balances/${address}`);
    }

    async getSupply(denom = null) {
        const path = denom
            ? `/cosmos/bank/v1beta1/supply/by_denom?denom=${denom}`
            : `/cosmos/bank/v1beta1/supply`;
        return await this.request(path);
    }

    async getDenomMetadata(denom) {
        return await this.request(`/cosmos/bank/v1beta1/denoms_metadata/${denom}`);
    }

    // Staking Module
    async getValidators(status = null) {
        const path = status
            ? `/cosmos/staking/v1beta1/validators?status=${status}`
            : `/cosmos/staking/v1beta1/validators`;
        return await this.request(path);
    }

    async getValidator(validatorAddr) {
        return await this.request(`/cosmos/staking/v1beta1/validators/${validatorAddr}`);
    }

    async getDelegations(delegatorAddr) {
        return await this.request(`/cosmos/staking/v1beta1/delegations/${delegatorAddr}`);
    }

    async getValidatorDelegations(validatorAddr) {
        return await this.request(`/cosmos/staking/v1beta1/validators/${validatorAddr}/delegations`);
    }

    async getUnbondingDelegations(delegatorAddr) {
        return await this.request(`/cosmos/staking/v1beta1/delegators/${delegatorAddr}/unbonding_delegations`);
    }

    async getStakingPool() {
        return await this.request('/cosmos/staking/v1beta1/pool');
    }

    async getStakingParams() {
        return await this.request('/cosmos/staking/v1beta1/params');
    }

    // Distribution Module
    async getDelegationRewards(delegatorAddr, validatorAddr = null) {
        const path = validatorAddr
            ? `/cosmos/distribution/v1beta1/delegators/${delegatorAddr}/rewards/${validatorAddr}`
            : `/cosmos/distribution/v1beta1/delegators/${delegatorAddr}/rewards`;
        return await this.request(path);
    }

    async getValidatorCommission(validatorAddr) {
        return await this.request(`/cosmos/distribution/v1beta1/validators/${validatorAddr}/commission`);
    }

    async getValidatorOutstandingRewards(validatorAddr) {
        return await this.request(`/cosmos/distribution/v1beta1/validators/${validatorAddr}/outstanding_rewards`);
    }

    // Governance Module
    async getProposals(status = null) {
        const path = status
            ? `/cosmos/gov/v1beta1/proposals?proposal_status=${status}`
            : `/cosmos/gov/v1beta1/proposals`;
        return await this.request(path);
    }

    async getProposal(proposalId) {
        return await this.request(`/cosmos/gov/v1beta1/proposals/${proposalId}`);
    }

    async getProposalVotes(proposalId) {
        return await this.request(`/cosmos/gov/v1beta1/proposals/${proposalId}/votes`);
    }

    async getProposalTally(proposalId) {
        return await this.request(`/cosmos/gov/v1beta1/proposals/${proposalId}/tally`);
    }

    async getGovParams(paramsType = 'voting') {
        return await this.request(`/cosmos/gov/v1beta1/params/${paramsType}`);
    }

    // DEX Module
    async getPools() {
        return await this.request('/aura/dex/v1beta1/pools');
    }

    async getPool(poolId) {
        return await this.request(`/aura/dex/v1beta1/pools/${poolId}`);
    }

    async getPoolLiquidity(poolId) {
        return await this.request(`/aura/dex/v1beta1/pools/${poolId}/liquidity`);
    }

    async estimateSwap(poolId, tokenIn, amountIn) {
        return await this.request(`/aura/dex/v1beta1/pools/${poolId}/estimate_swap`, {
            method: 'POST',
            body: JSON.stringify({
                token_in: tokenIn,
                amount_in: amountIn
            })
        });
    }

    // Auth Module
    async getAccount(address) {
        return await this.request(`/cosmos/auth/v1beta1/accounts/${address}`);
    }

    async getAuthParams() {
        return await this.request('/cosmos/auth/v1beta1/params');
    }

    // Bridge Module
    async getBridgeParams() {
        return await this.request('/aura/bridge/v1beta1/params');
    }

    async getBridgeTransfer(id) {
        return await this.request(`/aura/bridge/v1beta1/transfers/${id}`);
    }

    async getBridgeTransfers() {
        return await this.request('/aura/bridge/v1beta1/transfers');
    }

    // Compliance Module
    async getComplianceStatus(address) {
        return await this.request(`/aura/compliance/v1beta1/status/${address}`);
    }

    async getComplianceParams() {
        return await this.request('/aura/compliance/v1beta1/params');
    }

    // ConfidenceScore Module
    async getConfidenceScore(address) {
        return await this.request(`/aura/confidencescore/v1beta1/score/${address}`);
    }

    async getConfidenceScoreParams() {
        return await this.request('/aura/confidencescore/v1beta1/params');
    }

    async getInclusionRoutineStatus(id) {
        return await this.request(`/aura/confidencescore/v1beta1/routine/${id}`);
    }

    // Cryptography Module
    async getCryptographyParams() {
        return await this.request('/aura/cryptography/v1beta1/params');
    }

    async getPublicKey(address) {
        return await this.request(`/aura/cryptography/v1beta1/public_key/${address}`);
    }

    // DataRegistry Module
    async getDataItem(id) {
        return await this.request(`/aura/dataregistry/v1beta1/data/${id}`);
    }

    async getDataItems() {
        return await this.request('/aura/dataregistry/v1beta1/data');
    }

    async getDataRegistryParams() {
        return await this.request('/aura/dataregistry/v1beta1/params');
    }

    // EconomicSecurity Module
    async getEconomicSecurityParams() {
        return await this.request('/aura/economicsecurity/v1beta1/params');
    }

    async getDynamicFees() {
        return await this.request('/aura/economicsecurity/v1beta1/fees');
    }

    async getMevProtection() {
        return await this.request('/aura/economicsecurity/v1beta1/mev_protection');
    }

    // IdentityChange Module
    async getIdentityChangeRequest(id) {
        return await this.request(`/aura/identitychange/v1beta1/request/${id}`);
    }

    async getIdentityChangeRequests() {
        return await this.request('/aura/identitychange/v1beta1/requests');
    }

    async getIdentityChangeParams() {
        return await this.request('/aura/identitychange/v1beta1/params');
    }

    // InclusionRoutines Module
    async getInclusionRoutine(id) {
        return await this.request(`/aura/inclusionroutines/v1beta1/routine/${id}`);
    }

    async getInclusionRoutines() {
        return await this.request('/aura/inclusionroutines/v1beta1/routines');
    }

    async getInclusionRoutineParams() {
        return await this.request('/aura/inclusionroutines/v1beta1/params');
    }

    // Monitoring Module
    async getMonitoringMetrics() {
        return await this.request('/aura/monitoring/v1beta1/metrics');
    }

    async getMonitoringAlerts() {
        return await this.request('/aura/monitoring/v1beta1/alerts');
    }

    async getMonitoringParams() {
        return await this.request('/aura/monitoring/v1beta1/params');
    }

    // NetworkSecurity Module
    async getNetworkSecurityStatus() {
        return await this.request('/aura/networksecurity/v1beta1/status');
    }

    async getNetworkSecurityParams() {
        return await this.request('/aura/networksecurity/v1beta1/params');
    }

    async getPeerReputation(peerId) {
        return await this.request(`/aura/networksecurity/v1beta1/reputation/${peerId}`);
    }

    // Prevalidation Module
    async getPrevalidationStatus(txHash) {
        return await this.request(`/aura/prevalidation/v1beta1/status/${txHash}`);
    }

    async getPrevalidationParams() {
        return await this.request('/aura/prevalidation/v1beta1/params');
    }

    // Privacy Module
    async getPrivacyParams() {
        return await this.request('/aura/privacy/v1beta1/params');
    }

    // ValidatorSecurity Module
    async getValidatorSecurityStatus(validatorAddr) {
        return await this.request(`/aura/validatorsecurity/v1beta1/status/${validatorAddr}`);
    }

    async getValidatorSecurityParams() {
        return await this.request('/aura/validatorsecurity/v1beta1/params');
    }

    async getValidatorSlashingEvents(validatorAddr) {
        return await this.request(`/aura/validatorsecurity/v1beta1/slashing/${validatorAddr}`);
    }

    // VCRegistry Module (Most Important)
    async getVC(vcId) {
        return await this.request(`/aura/vcregistry/v1beta1/vc/${vcId}`);
    }

    async getVCs(address = null) {
        const path = address
            ? `/aura/vcregistry/v1beta1/vcs/${address}`
            : '/aura/vcregistry/v1beta1/vcs';
        return await this.request(path);
    }

    async getVCPresentation(presentationId) {
        return await this.request(`/aura/vcregistry/v1beta1/presentation/${presentationId}`);
    }

    async getVCRegistryParams() {
        return await this.request('/aura/vcregistry/v1beta1/params');
    }

    async getVCStats() {
        return await this.request('/aura/vcregistry/v1beta1/stats');
    }

    // WalletSecurity Module
    async getWalletSecurityStatus(address) {
        return await this.request(`/aura/walletsecurity/v1beta1/status/${address}`);
    }

    async getWalletSecurityParams() {
        return await this.request('/aura/walletsecurity/v1beta1/params');
    }

    async getWalletSessions(address) {
        return await this.request(`/aura/walletsecurity/v1beta1/sessions/${address}`);
    }

    // Transaction queries
    async getTx(hash) {
        return await this.request(`/cosmos/tx/v1beta1/txs/${hash}`);
    }

    async getTxsByEvents(events) {
        const params = new URLSearchParams();
        events.forEach(event => {
            params.append('events', event);
        });
        return await this.request(`/cosmos/tx/v1beta1/txs?${params.toString()}`);
    }

    // Node info
    async getNodeInfo() {
        return await this.request('/cosmos/base/tendermint/v1beta1/node_info');
    }

    async getLatestBlock() {
        return await this.request('/cosmos/base/tendermint/v1beta1/blocks/latest');
    }

    async getBlockByHeight(height) {
        return await this.request(`/cosmos/base/tendermint/v1beta1/blocks/${height}`);
    }

    async getSyncing() {
        return await this.request('/cosmos/base/tendermint/v1beta1/syncing');
    }

    // Generic query method
    async query(path, options = {}) {
        return await this.request(path, options);
    }
}

// ValidatorAPI Service - Handles all API calls to AURA blockchain
//
// Uses @cosmjs/stargate for transaction signing and broadcasting.
// Install dependencies: npm install @cosmjs/stargate @cosmjs/proto-signing

// Import CosmJS modules (available via npm or CDN)
let SigningStargateClient, GasPrice;
try {
    const cosmjs = window.CosmJS || {};
    SigningStargateClient = cosmjs.SigningStargateClient;
    GasPrice = cosmjs.GasPrice;
} catch (e) {
    console.log('CosmJS modules will be loaded on demand');
}

class ValidatorAPI {
    static baseURL = 'http://localhost:1317'; // Default LCD endpoint
    static rpcURL = 'http://localhost:26657'; // RPC endpoint
    static timeout = 10000; // 10 seconds
    static signingClient = null;
    static gasPrice = '0.025aura';

    static async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        try {
            const response = await fetch(url, {
                ...options,
                signal: controller.signal,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            clearTimeout(timeoutId);
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            throw error;
        }
    }

    // Get validator information
    static async getValidatorInfo(validatorAddress) {
        try {
            const response = await this.request(`/cosmos/staking/v1beta1/validators/${validatorAddress}`);

            return {
                address: response.validator.operator_address,
                moniker: response.validator.description.moniker,
                website: response.validator.description.website,
                details: response.validator.description.details,
                identity: response.validator.description.identity,
                tokens: response.validator.tokens,
                delegatorShares: response.validator.delegator_shares,
                commission: {
                    rate: parseFloat(response.validator.commission.commission_rates.rate),
                    maxRate: parseFloat(response.validator.commission.commission_rates.max_rate),
                    maxChangeRate: parseFloat(response.validator.commission.commission_rates.max_change_rate)
                },
                status: response.validator.status,
                jailed: response.validator.jailed,
                unbondingHeight: response.validator.unbonding_height,
                unbondingTime: response.validator.unbonding_time,
                uptime: await this.calculateUptime(validatorAddress),
                totalRewards: await this.getTotalRewards(validatorAddress),
                delegatorCount: await this.getDelegatorCount(validatorAddress)
            };
        } catch (error) {
            console.error('Error fetching validator info:', error);
            // Return mock data for development
            return this.getMockValidatorInfo(validatorAddress);
        }
    }

    // Get validator delegations
    static async getDelegations(validatorAddress) {
        try {
            const response = await this.request(
                `/cosmos/staking/v1beta1/validators/${validatorAddress}/delegations`
            );

            return response.delegation_responses.map(del => ({
                delegatorAddress: del.delegation.delegator_address,
                shares: del.delegation.shares,
                pendingRewards: 0, // Would be fetched separately
                timestamp: new Date().toISOString() // Would come from blockchain
            }));
        } catch (error) {
            console.error('Error fetching delegations:', error);
            return this.getMockDelegations();
        }
    }

    // Get validator rewards
    static async getRewards(validatorAddress) {
        try {
            const response = await this.request(
                `/cosmos/distribution/v1beta1/validators/${validatorAddress}/outstanding_rewards`
            );

            const history = await this.getRewardHistory(validatorAddress);

            return {
                totalDistributed: parseFloat(response.rewards.rewards[0]?.amount || '0'),
                pending: parseFloat(response.rewards.rewards[0]?.amount || '0'),
                commissionEarned: await this.getCommissionEarned(validatorAddress),
                history: history
            };
        } catch (error) {
            console.error('Error fetching rewards:', error);
            return this.getMockRewards();
        }
    }

    // Get validator performance metrics
    static async getPerformance(validatorAddress) {
        try {
            const validatorInfo = await this.getValidatorInfo(validatorAddress);
            const signingInfo = await this.getSigningInfo(validatorAddress);

            // Calculate voting power percentage
            const allValidators = await this.request('/cosmos/staking/v1beta1/validators');
            const totalVotingPower = allValidators.validators.reduce(
                (sum, v) => sum + parseFloat(v.tokens || 0),
                0
            );
            const votingPower = (parseFloat(validatorInfo.tokens) / totalVotingPower) * 100;

            return {
                votingPower: votingPower,
                blockProposals: signingInfo.indexOffset || 0,
                missRate: signingInfo.missedBlocksCounter / signingInfo.indexOffset || 0,
                votingPowerHistory: this.generateMockHistory(votingPower),
                proposalHistory: this.generateMockHistory(signingInfo.indexOffset),
                missRateHistory: this.generateMockHistory(0.01, 0.05)
            };
        } catch (error) {
            console.error('Error fetching performance:', error);
            return this.getMockPerformance();
        }
    }

    // Get validator uptime data
    static async getUptime(validatorAddress) {
        try {
            const signingInfo = await this.getSigningInfo(validatorAddress);
            const blocks = await this.getRecentBlocks(validatorAddress, 1000);

            const signedBlocks = blocks.filter(b => b.signed || b.proposed).length;
            const uptimePercentage = (signedBlocks / blocks.length) * 100;

            return {
                uptimePercentage: uptimePercentage,
                totalBlocks: blocks.length,
                missedBlocks: blocks.length - signedBlocks,
                blocks: blocks,
                uptime7d: uptimePercentage,
                uptime30d: uptimePercentage,
                consecutiveMisses: this.calculateConsecutiveMisses(blocks),
                longestStreak: this.calculateLongestStreak(blocks),
                alerts: this.generateUptimeAlerts(uptimePercentage, signingInfo)
            };
        } catch (error) {
            console.error('Error fetching uptime:', error);
            return this.getMockUptime();
        }
    }

    // Get signing statistics
    static async getSigningStats(validatorAddress) {
        try {
            const signingInfo = await this.getSigningInfo(validatorAddress);

            return {
                blocksSigned: signingInfo.indexOffset - signingInfo.missedBlocksCounter,
                blocksMissed: signingInfo.missedBlocksCounter,
                history: this.generateSigningHistory(1000)
            };
        } catch (error) {
            console.error('Error fetching signing stats:', error);
            return this.getMockSigningStats();
        }
    }

    // Get slash events
    static async getSlashEvents(validatorAddress) {
        try {
            // Note: This endpoint may vary depending on the blockchain implementation
            const response = await this.request(
                `/cosmos/slashing/v1beta1/signing_infos/${validatorAddress}`
            );

            // Parse slash events from signing info or query events
            return [];
        } catch (error) {
            console.error('Error fetching slash events:', error);
            return []; // Return empty array if no slash events
        }
    }

    // Initialize signing client for transaction broadcasting
    static async initSigningClient(signer) {
        if (!SigningStargateClient) {
            throw new Error('CosmJS SigningStargateClient not available. Please load @cosmjs/stargate.');
        }

        try {
            this.signingClient = await SigningStargateClient.connectWithSigner(
                this.rpcURL,
                signer,
                { gasPrice: GasPrice.fromString(this.gasPrice) }
            );
            return true;
        } catch (error) {
            console.error('Failed to initialize signing client:', error);
            throw new Error(`Failed to connect to RPC: ${error.message}`);
        }
    }

    // Disconnect signing client
    static disconnectSigningClient() {
        if (this.signingClient) {
            this.signingClient.disconnect();
            this.signingClient = null;
        }
    }

    // Update commission rate using MsgEditValidator
    static async updateCommission(validatorAddress, newRate, signer) {
        if (!this.signingClient) {
            if (!signer) {
                throw new Error('Signer required to initialize signing client');
            }
            await this.initSigningClient(signer);
        }

        // Validate commission rate (must be between 0 and 1)
        const rate = parseFloat(newRate);
        if (isNaN(rate) || rate < 0 || rate > 1) {
            throw new Error('Commission rate must be between 0 and 1');
        }

        // Get current validator info to preserve other fields
        const currentInfo = await this.getValidatorInfo(validatorAddress);

        // Build MsgEditValidator message
        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgEditValidator',
            value: {
                validatorAddress: validatorAddress,
                description: {
                    moniker: currentInfo.moniker || '',
                    identity: currentInfo.identity || '',
                    website: currentInfo.website || '',
                    securityContact: '',
                    details: currentInfo.details || ''
                },
                commissionRate: (rate * 1e18).toString(), // Convert to SDK decimal format
                minSelfDelegation: null // Don't change
            }
        };

        try {
            const accounts = await this.signingClient.getAccounts();
            const signerAddress = accounts[0]?.address;

            if (!signerAddress) {
                throw new Error('No signer address available');
            }

            const result = await this.signingClient.signAndBroadcast(
                signerAddress,
                [msg],
                'auto',
                `Update commission rate to ${(rate * 100).toFixed(2)}%`
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                transactionHash: result.transactionHash,
                height: result.height,
                gasUsed: result.gasUsed,
                newCommissionRate: rate
            };
        } catch (error) {
            console.error('Error updating commission:', error);
            throw new Error(`Failed to update commission: ${error.message}`);
        }
    }

    // Update validator information using MsgEditValidator
    static async updateValidatorInfo(validatorAddress, info, signer) {
        if (!this.signingClient) {
            if (!signer) {
                throw new Error('Signer required to initialize signing client');
            }
            await this.initSigningClient(signer);
        }

        // Validate input
        if (!info || typeof info !== 'object') {
            throw new Error('Info object required with fields: moniker, website, details, identity');
        }

        // Build MsgEditValidator message
        const msg = {
            typeUrl: '/cosmos.staking.v1beta1.MsgEditValidator',
            value: {
                validatorAddress: validatorAddress,
                description: {
                    moniker: info.moniker || '',
                    identity: info.identity || '',
                    website: info.website || '',
                    securityContact: info.securityContact || '',
                    details: info.details || ''
                },
                commissionRate: null, // Don't change commission
                minSelfDelegation: info.minSelfDelegation || null
            }
        };

        try {
            const accounts = await this.signingClient.getAccounts();
            const signerAddress = accounts[0]?.address;

            if (!signerAddress) {
                throw new Error('No signer address available');
            }

            const result = await this.signingClient.signAndBroadcast(
                signerAddress,
                [msg],
                'auto',
                `Update validator info for ${info.moniker || validatorAddress}`
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                transactionHash: result.transactionHash,
                height: result.height,
                gasUsed: result.gasUsed,
                updatedFields: Object.keys(info).filter(k => info[k])
            };
        } catch (error) {
            console.error('Error updating validator info:', error);
            throw new Error(`Failed to update validator info: ${error.message}`);
        }
    }

    // Unjail validator using MsgUnjail
    static async unjailValidator(validatorAddress, signer) {
        if (!this.signingClient) {
            if (!signer) {
                throw new Error('Signer required to initialize signing client');
            }
            await this.initSigningClient(signer);
        }

        const msg = {
            typeUrl: '/cosmos.slashing.v1beta1.MsgUnjail',
            value: {
                validatorAddr: validatorAddress
            }
        };

        try {
            const accounts = await this.signingClient.getAccounts();
            const signerAddress = accounts[0]?.address;

            if (!signerAddress) {
                throw new Error('No signer address available');
            }

            const result = await this.signingClient.signAndBroadcast(
                signerAddress,
                [msg],
                'auto',
                'Unjail validator'
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                transactionHash: result.transactionHash,
                height: result.height,
                gasUsed: result.gasUsed
            };
        } catch (error) {
            console.error('Error unjailing validator:', error);
            throw new Error(`Failed to unjail validator: ${error.message}`);
        }
    }

    // Withdraw validator commission using MsgWithdrawValidatorCommission
    static async withdrawCommission(validatorAddress, signer) {
        if (!this.signingClient) {
            if (!signer) {
                throw new Error('Signer required to initialize signing client');
            }
            await this.initSigningClient(signer);
        }

        const msg = {
            typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission',
            value: {
                validatorAddress: validatorAddress
            }
        };

        try {
            const accounts = await this.signingClient.getAccounts();
            const signerAddress = accounts[0]?.address;

            if (!signerAddress) {
                throw new Error('No signer address available');
            }

            const result = await this.signingClient.signAndBroadcast(
                signerAddress,
                [msg],
                'auto',
                'Withdraw validator commission'
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                transactionHash: result.transactionHash,
                height: result.height,
                gasUsed: result.gasUsed
            };
        } catch (error) {
            console.error('Error withdrawing commission:', error);
            throw new Error(`Failed to withdraw commission: ${error.message}`);
        }
    }

    // Helper methods

    static async calculateUptime(validatorAddress) {
        try {
            const signingInfo = await this.getSigningInfo(validatorAddress);
            const uptime = 1 - (signingInfo.missedBlocksCounter / signingInfo.indexOffset);
            return Math.max(0, Math.min(1, uptime));
        } catch (error) {
            return 0.99; // Default mock value
        }
    }

    static async getTotalRewards(validatorAddress) {
        try {
            const response = await this.request(
                `/cosmos/distribution/v1beta1/validators/${validatorAddress}/outstanding_rewards`
            );
            return parseFloat(response.rewards.rewards[0]?.amount || '0');
        } catch (error) {
            return 1000000; // Mock value
        }
    }

    static async getDelegatorCount(validatorAddress) {
        try {
            const response = await this.request(
                `/cosmos/staking/v1beta1/validators/${validatorAddress}/delegations`
            );
            return response.delegation_responses.length;
        } catch (error) {
            return 100; // Mock value
        }
    }

    static async getSigningInfo(validatorAddress) {
        try {
            // Convert validator address to consensus address
            const consAddress = this.validatorToConsensus(validatorAddress);
            const response = await this.request(
                `/cosmos/slashing/v1beta1/signing_infos/${consAddress}`
            );
            return response.val_signing_info;
        } catch (error) {
            return {
                indexOffset: 10000,
                missedBlocksCounter: 50
            };
        }
    }

    static async getRecentBlocks(validatorAddress, count) {
        // In production, query actual block data
        // For now, generate mock data
        const blocks = [];
        for (let i = 0; i < count; i++) {
            blocks.push({
                height: 1000000 - i,
                signed: Math.random() > 0.02, // 98% signing rate
                proposed: Math.random() > 0.99, // 1% proposal rate
                timestamp: new Date(Date.now() - i * 6000).toISOString()
            });
        }
        return blocks;
    }

    static async getRewardHistory(validatorAddress) {
        // Mock reward history - in production, query from events
        const history = [];
        const now = Date.now();
        for (let i = 0; i < 30; i++) {
            history.push({
                timestamp: new Date(now - i * 24 * 60 * 60 * 1000).toISOString(),
                amount: 10000 + Math.random() * 5000,
                commission: 500 + Math.random() * 250
            });
        }
        return history.reverse();
    }

    static async getCommissionEarned(validatorAddress) {
        try {
            const response = await this.request(
                `/cosmos/distribution/v1beta1/validators/${validatorAddress}/commission`
            );
            return parseFloat(response.commission.commission[0]?.amount || '0');
        } catch (error) {
            return 50000; // Mock value
        }
    }

    static calculateConsecutiveMisses(blocks) {
        let maxConsecutive = 0;
        let current = 0;

        for (const block of blocks) {
            if (!block.signed && !block.proposed) {
                current++;
                maxConsecutive = Math.max(maxConsecutive, current);
            } else {
                current = 0;
            }
        }

        return maxConsecutive;
    }

    static calculateLongestStreak(blocks) {
        let maxStreak = 0;
        let current = 0;

        for (const block of blocks) {
            if (block.signed || block.proposed) {
                current++;
                maxStreak = Math.max(maxStreak, current);
            } else {
                current = 0;
            }
        }

        return maxStreak;
    }

    static generateUptimeAlerts(uptime, signingInfo) {
        const alerts = [];

        if (uptime < 95) {
            alerts.push({
                level: 'danger',
                title: 'Low Uptime',
                message: `Uptime is ${uptime.toFixed(2)}%, below the 95% threshold`,
                timestamp: new Date().toISOString()
            });
        }

        if (signingInfo.missedBlocksCounter > 400) {
            alerts.push({
                level: 'warning',
                title: 'High Missed Blocks',
                message: `Missed ${signingInfo.missedBlocksCounter} blocks, approaching slash threshold`,
                timestamp: new Date().toISOString()
            });
        }

        return alerts;
    }

    static generateSigningHistory(count) {
        const history = [];
        for (let i = 0; i < count; i++) {
            history.push(Math.random() > 0.02); // 98% signing rate
        }
        return history;
    }

    static generateMockHistory(baseValue, variance = 0.1) {
        const history = [];
        for (let i = 0; i < 30; i++) {
            const value = baseValue * (1 + (Math.random() - 0.5) * variance);
            history.push(Math.max(0, Math.min(1, value)));
        }
        return history;
    }

    static validatorToConsensus(validatorAddress) {
        // Simplified conversion - in production, use proper bech32 conversion
        return validatorAddress.replace('auravaloper', 'auravalcons');
    }

    // Mock data generators for development

    static getMockValidatorInfo(address) {
        return {
            address: address,
            moniker: 'AURA Validator',
            website: 'https://validator.aura.network',
            details: 'Professional validator service',
            identity: '',
            tokens: '1000000000000',
            delegatorShares: '1000000000000',
            commission: {
                rate: 0.05,
                maxRate: 0.20,
                maxChangeRate: 0.01
            },
            status: 'BOND_STATUS_BONDED',
            jailed: false,
            unbondingHeight: '0',
            unbondingTime: null,
            uptime: 0.99,
            totalRewards: 1000000,
            delegatorCount: 100
        };
    }

    static getMockDelegations() {
        const delegations = [];
        for (let i = 0; i < 50; i++) {
            delegations.push({
                delegatorAddress: `aura1${Array(38).fill(0).map(() => 'abcdefghijklmnopqrstuvwxyz0123456789'[Math.floor(Math.random() * 36)]).join('')}`,
                shares: (Math.random() * 100000 + 1000) * 1000000,
                pendingRewards: Math.random() * 1000 * 1000000,
                timestamp: new Date(Date.now() - Math.random() * 90 * 24 * 60 * 60 * 1000).toISOString()
            });
        }
        return delegations.sort((a, b) => parseFloat(b.shares) - parseFloat(a.shares));
    }

    static getMockRewards() {
        const history = [];
        for (let i = 0; i < 30; i++) {
            history.push({
                timestamp: new Date(Date.now() - i * 24 * 60 * 60 * 1000).toISOString(),
                amount: 10000 + Math.random() * 5000,
                commission: 500 + Math.random() * 250
            });
        }

        return {
            totalDistributed: 500000,
            pending: 25000,
            commissionEarned: 50000,
            history: history.reverse()
        };
    }

    static getMockPerformance() {
        return {
            votingPower: 2.5,
            blockProposals: 1234,
            missRate: 0.02,
            votingPowerHistory: this.generateMockHistory(2.5, 0.1),
            proposalHistory: this.generateMockHistory(40, 0.2),
            missRateHistory: this.generateMockHistory(0.02, 0.5)
        };
    }

    static getMockUptime() {
        const blocks = this.generateSigningHistory(1000);
        const signedCount = blocks.filter(b => b).length;

        return {
            uptimePercentage: (signedCount / blocks.length) * 100,
            totalBlocks: blocks.length,
            missedBlocks: blocks.length - signedCount,
            blocks: blocks.map((signed, i) => ({
                height: 1000000 - i,
                signed: signed,
                proposed: Math.random() > 0.99
            })),
            uptime7d: 99.5,
            uptime30d: 99.2,
            consecutiveMisses: 2,
            longestStreak: 850,
            alerts: []
        };
    }

    static getMockSigningStats() {
        return {
            blocksSigned: 9950,
            blocksMissed: 50,
            history: this.generateSigningHistory(1000)
        };
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ValidatorAPI;
}

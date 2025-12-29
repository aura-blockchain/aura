/**
 * AURA Governance API Service
 * Handles all API interactions for governance operations
 *
 * Uses @cosmjs/stargate for transaction signing and broadcasting.
 * Install dependencies: npm install @cosmjs/stargate @cosmjs/proto-signing
 */

// Import CosmJS modules (available via npm or CDN)
// In production, these would be bundled via webpack/rollup
let SigningStargateClient, GasPrice, coins;
try {
    // Try importing as ES modules (bundled environment)
    const cosmjs = window.CosmJS || {};
    SigningStargateClient = cosmjs.SigningStargateClient;
    GasPrice = cosmjs.GasPrice;
    coins = cosmjs.coins;
} catch (e) {
    // Modules will be loaded dynamically if not available
    console.log('CosmJS modules will be loaded on demand');
}

class GovernanceAPI {
    constructor() {
        this.baseURL = 'http://localhost:1317'; // REST endpoint
        this.rpcURL = 'http://localhost:26657'; // RPC endpoint
        this.connected = false;
        this.mockMode = true; // Enable mock mode for development
        this.signingClient = null;
        this.chainId = 'aura-testnet-1';
        this.gasPrice = '0.025aura';
    }

    /**
     * Initialize CosmJS signing client with a signer (e.g., Keplr wallet)
     * @param {OfflineSigner} signer - The wallet signer from Keplr or similar
     */
    async initSigningClient(signer) {
        if (!SigningStargateClient) {
            throw new Error('CosmJS not loaded. Include @cosmjs/stargate in your build.');
        }

        this.signingClient = await SigningStargateClient.connectWithSigner(
            this.rpcURL,
            signer,
            { gasPrice: GasPrice.fromString(this.gasPrice) }
        );
        this.mockMode = false;
        console.log('Signing client initialized');
    }

    /**
     * Get the connected wallet address
     */
    async getWalletAddress(signer) {
        const accounts = await signer.getAccounts();
        return accounts[0]?.address;
    }

    /**
     * Check connection to the blockchain
     */
    async checkConnection() {
        try {
            if (this.mockMode) {
                this.connected = true;
                return true;
            }

            const response = await fetch(`${this.baseURL}/cosmos/base/tendermint/v1beta1/node_info`);
            this.connected = response.ok;
            return this.connected;
        } catch (error) {
            console.error('Connection check failed:', error);
            this.connected = false;
            return false;
        }
    }

    /**
     * Get all governance proposals
     */
    async getAllProposals() {
        try {
            if (this.mockMode) {
                return this.getMockProposals();
            }

            const response = await fetch(`${this.baseURL}/cosmos/gov/v1beta1/proposals`);
            if (!response.ok) throw new Error('Failed to fetch proposals');

            const data = await response.json();
            return data.proposals || [];
        } catch (error) {
            console.error('Failed to fetch proposals:', error);
            return this.getMockProposals(); // Fallback to mock data
        }
    }

    /**
     * Get a specific proposal by ID
     */
    async getProposal(proposalId) {
        try {
            if (this.mockMode) {
                const proposals = this.getMockProposals();
                return proposals.find(p => p.proposal_id === proposalId);
            }

            const response = await fetch(`${this.baseURL}/cosmos/gov/v1beta1/proposals/${proposalId}`);
            if (!response.ok) throw new Error('Failed to fetch proposal');

            const data = await response.json();
            return data.proposal;
        } catch (error) {
            console.error('Failed to fetch proposal:', error);
            throw error;
        }
    }

    /**
     * Get votes for a proposal
     */
    async getProposalVotes(proposalId) {
        try {
            if (this.mockMode) {
                return this.getMockVotes(proposalId);
            }

            const response = await fetch(`${this.baseURL}/cosmos/gov/v1beta1/proposals/${proposalId}/votes`);
            if (!response.ok) throw new Error('Failed to fetch votes');

            const data = await response.json();
            return data.votes || [];
        } catch (error) {
            console.error('Failed to fetch votes:', error);
            return [];
        }
    }

    /**
     * Get deposits for a proposal
     */
    async getProposalDeposits(proposalId) {
        try {
            if (this.mockMode) {
                return this.getMockDeposits(proposalId);
            }

            const response = await fetch(`${this.baseURL}/cosmos/gov/v1beta1/proposals/${proposalId}/deposits`);
            if (!response.ok) throw new Error('Failed to fetch deposits');

            const data = await response.json();
            return data.deposits || [];
        } catch (error) {
            console.error('Failed to fetch deposits:', error);
            return [];
        }
    }

    /**
     * Get tally results for a proposal
     */
    async getProposalTally(proposalId) {
        try {
            if (this.mockMode) {
                const proposals = this.getMockProposals();
                const proposal = proposals.find(p => p.proposal_id === proposalId);
                return proposal?.final_tally_result || this.getEmptyTally();
            }

            const response = await fetch(`${this.baseURL}/cosmos/gov/v1beta1/proposals/${proposalId}/tally`);
            if (!response.ok) throw new Error('Failed to fetch tally');

            const data = await response.json();
            return data.tally;
        } catch (error) {
            console.error('Failed to fetch tally:', error);
            return this.getEmptyTally();
        }
    }

    /**
     * Get governance parameters
     */
    async getGovernanceParameters() {
        try {
            if (this.mockMode) {
                return this.getMockParameters();
            }

            const [depositParams, votingParams, tallyParams] = await Promise.all([
                fetch(`${this.baseURL}/cosmos/gov/v1beta1/params/deposit`).then(r => r.json()),
                fetch(`${this.baseURL}/cosmos/gov/v1beta1/params/voting`).then(r => r.json()),
                fetch(`${this.baseURL}/cosmos/gov/v1beta1/params/tallying`).then(r => r.json())
            ]);

            return {
                deposit: depositParams.deposit_params,
                voting: votingParams.voting_params,
                tally: tallyParams.tally_params
            };
        } catch (error) {
            console.error('Failed to fetch parameters:', error);
            return this.getMockParameters();
        }
    }

    /**
     * Submit a new proposal
     * @param {Object} proposalData - The proposal content (title, description, type)
     * @param {Array} initialDeposit - Initial deposit coins [{denom: 'aura', amount: '10000000'}]
     * @param {string} proposerAddress - The proposer's bech32 address
     */
    async submitProposal(proposalData, initialDeposit, proposerAddress) {
        try {
            if (this.mockMode) {
                console.log('Mock: Submitting proposal', proposalData);
                return {
                    success: true,
                    proposal_id: Math.floor(Math.random() * 1000) + 1,
                    txhash: 'mock_' + Math.random().toString(36).substring(7)
                };
            }

            if (!this.signingClient) {
                throw new Error('Signing client not initialized. Call initSigningClient first.');
            }

            // Format the proposal message based on type
            let typeUrl, content;
            switch (proposalData.type) {
                case 'text':
                    typeUrl = '/cosmos.gov.v1beta1.TextProposal';
                    content = {
                        title: proposalData.title,
                        description: proposalData.description
                    };
                    break;
                case 'parameter_change':
                    typeUrl = '/cosmos.params.v1beta1.ParameterChangeProposal';
                    content = {
                        title: proposalData.title,
                        description: proposalData.description,
                        changes: proposalData.changes || []
                    };
                    break;
                case 'software_upgrade':
                    typeUrl = '/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal';
                    content = {
                        title: proposalData.title,
                        description: proposalData.description,
                        plan: proposalData.plan
                    };
                    break;
                default:
                    typeUrl = '/cosmos.gov.v1beta1.TextProposal';
                    content = {
                        title: proposalData.title,
                        description: proposalData.description
                    };
            }

            const msg = {
                typeUrl: '/cosmos.gov.v1beta1.MsgSubmitProposal',
                value: {
                    content: {
                        typeUrl: typeUrl,
                        value: content
                    },
                    initialDeposit: initialDeposit,
                    proposer: proposerAddress
                }
            };

            const result = await this.signingClient.signAndBroadcast(
                proposerAddress,
                [msg],
                'auto',
                'Submitted via Aura Governance Dashboard'
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            // Extract proposal_id from events
            let proposalId = null;
            for (const event of result.events || []) {
                if (event.type === 'submit_proposal') {
                    const attr = event.attributes.find(a => a.key === 'proposal_id');
                    if (attr) {
                        proposalId = attr.value;
                        break;
                    }
                }
            }

            return {
                success: true,
                proposal_id: proposalId,
                txhash: result.transactionHash,
                height: result.height
            };
        } catch (error) {
            console.error('Failed to submit proposal:', error);
            throw error;
        }
    }

    /**
     * Vote on a proposal
     * @param {string} proposalId - The proposal ID
     * @param {string|number} option - Vote option: 1=Yes, 2=Abstain, 3=No, 4=NoWithVeto or string
     * @param {string} voterAddress - The voter's bech32 address
     */
    async vote(proposalId, option, voterAddress) {
        try {
            if (this.mockMode) {
                console.log(`Mock: Voting ${option} on proposal ${proposalId}`);
                return {
                    success: true,
                    txhash: 'mock_' + Math.random().toString(36).substring(7)
                };
            }

            if (!this.signingClient) {
                throw new Error('Signing client not initialized. Call initSigningClient first.');
            }

            // Convert string options to numeric values
            let voteOption;
            if (typeof option === 'string') {
                const optionMap = {
                    'VOTE_OPTION_YES': 1,
                    'VOTE_OPTION_ABSTAIN': 2,
                    'VOTE_OPTION_NO': 3,
                    'VOTE_OPTION_NO_WITH_VETO': 4,
                    'yes': 1,
                    'abstain': 2,
                    'no': 3,
                    'no_with_veto': 4
                };
                voteOption = optionMap[option] || optionMap[option.toLowerCase()] || 1;
            } else {
                voteOption = option;
            }

            const msg = {
                typeUrl: '/cosmos.gov.v1beta1.MsgVote',
                value: {
                    proposalId: Long.fromString(proposalId.toString()),
                    voter: voterAddress,
                    option: voteOption
                }
            };

            const result = await this.signingClient.signAndBroadcast(
                voterAddress,
                [msg],
                'auto',
                'Voted via Aura Governance Dashboard'
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                txhash: result.transactionHash,
                height: result.height
            };
        } catch (error) {
            console.error('Failed to vote:', error);
            throw error;
        }
    }

    /**
     * Deposit to a proposal
     * @param {string} proposalId - The proposal ID
     * @param {Array} amount - Deposit coins [{denom: 'aura', amount: '1000000'}]
     * @param {string} depositorAddress - The depositor's bech32 address
     */
    async deposit(proposalId, amount, depositorAddress) {
        try {
            if (this.mockMode) {
                console.log(`Mock: Depositing ${JSON.stringify(amount)} to proposal ${proposalId}`);
                return {
                    success: true,
                    txhash: 'mock_' + Math.random().toString(36).substring(7)
                };
            }

            if (!this.signingClient) {
                throw new Error('Signing client not initialized. Call initSigningClient first.');
            }

            // Normalize amount format
            const depositAmount = Array.isArray(amount) ? amount : [{ denom: 'aura', amount: amount.toString() }];

            const msg = {
                typeUrl: '/cosmos.gov.v1beta1.MsgDeposit',
                value: {
                    proposalId: Long.fromString(proposalId.toString()),
                    depositor: depositorAddress,
                    amount: depositAmount
                }
            };

            const result = await this.signingClient.signAndBroadcast(
                depositorAddress,
                [msg],
                'auto',
                'Deposit via Aura Governance Dashboard'
            );

            if (result.code !== 0) {
                throw new Error(`Transaction failed: ${result.rawLog}`);
            }

            return {
                success: true,
                txhash: result.transactionHash,
                height: result.height
            };
        } catch (error) {
            console.error('Failed to deposit:', error);
            throw error;
        }
    }

    /**
     * Get user's votes
     */
    async getUserVotes(address) {
        try {
            if (this.mockMode) {
                return this.getMockUserVotes(address);
            }

            const proposals = await this.getAllProposals();
            const votes = [];

            for (const proposal of proposals) {
                try {
                    const response = await fetch(
                        `${this.baseURL}/cosmos/gov/v1beta1/proposals/${proposal.proposal_id}/votes/${address}`
                    );
                    if (response.ok) {
                        const data = await response.json();
                        votes.push({
                            proposal_id: proposal.proposal_id,
                            option: data.vote.option,
                            timestamp: proposal.voting_start_time
                        });
                    }
                } catch (error) {
                    // Vote not found for this proposal
                    continue;
                }
            }

            return votes;
        } catch (error) {
            console.error('Failed to fetch user votes:', error);
            return [];
        }
    }

    // Mock data generators
    getMockProposals() {
        return [
            {
                proposal_id: '1',
                content: {
                    '@type': '/cosmos.gov.v1beta1.TextProposal',
                    title: 'Increase Block Size Limit',
                    description: 'This proposal aims to increase the block size limit from 22KB to 50KB to improve network throughput and reduce transaction costs during peak usage periods.'
                },
                status: 'VOTING_PERIOD',
                final_tally_result: {
                    yes: '45000000',
                    abstain: '5000000',
                    no: '8000000',
                    no_with_veto: '2000000'
                },
                submit_time: '2024-01-15T10:00:00Z',
                deposit_end_time: '2024-01-29T10:00:00Z',
                total_deposit: [{ denom: 'aura', amount: '10000000' }],
                voting_start_time: '2024-01-20T10:00:00Z',
                voting_end_time: '2024-02-05T10:00:00Z'
            },
            {
                proposal_id: '2',
                content: {
                    '@type': '/cosmos.params.v1beta1.ParameterChangeProposal',
                    title: 'Update Governance Voting Period',
                    description: 'Reduce the voting period from 14 days to 7 days to accelerate governance decisions while maintaining adequate time for community participation.'
                },
                status: 'VOTING_PERIOD',
                final_tally_result: {
                    yes: '38000000',
                    abstain: '12000000',
                    no: '15000000',
                    no_with_veto: '5000000'
                },
                submit_time: '2024-01-18T14:30:00Z',
                deposit_end_time: '2024-02-01T14:30:00Z',
                total_deposit: [{ denom: 'aura', amount: '10000000' }],
                voting_start_time: '2024-01-22T14:30:00Z',
                voting_end_time: '2024-02-08T14:30:00Z'
            },
            {
                proposal_id: '3',
                content: {
                    '@type': '/cosmos.gov.v1beta1.TextProposal',
                    title: 'Community Pool Fund Allocation',
                    description: 'Allocate 500,000 AURA from the community pool to fund development of a mobile wallet application and related infrastructure improvements.'
                },
                status: 'PASSED',
                final_tally_result: {
                    yes: '75000000',
                    abstain: '10000000',
                    no: '8000000',
                    no_with_veto: '1000000'
                },
                submit_time: '2024-01-05T09:00:00Z',
                deposit_end_time: '2024-01-19T09:00:00Z',
                total_deposit: [{ denom: 'aura', amount: '10000000' }],
                voting_start_time: '2024-01-10T09:00:00Z',
                voting_end_time: '2024-01-26T09:00:00Z'
            },
            {
                proposal_id: '4',
                content: {
                    '@type': '/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal',
                    title: 'Network Upgrade v2.0',
                    description: 'Upgrade to version 2.0 including new DEX features, improved oracle integration, and enhanced security measures. Scheduled for block height 1,000,000.'
                },
                status: 'DEPOSIT_PERIOD',
                final_tally_result: {
                    yes: '0',
                    abstain: '0',
                    no: '0',
                    no_with_veto: '0'
                },
                submit_time: '2024-01-25T16:00:00Z',
                deposit_end_time: '2024-02-08T16:00:00Z',
                total_deposit: [{ denom: 'aura', amount: '5000000' }],
                voting_start_time: '0001-01-01T00:00:00Z',
                voting_end_time: '0001-01-01T00:00:00Z'
            },
            {
                proposal_id: '5',
                content: {
                    '@type': '/cosmos.gov.v1beta1.TextProposal',
                    title: 'Adjust Minimum Deposit Requirement',
                    description: 'Lower the minimum deposit requirement for proposals from 10,000 AURA to 5,000 AURA to encourage more community participation in governance.'
                },
                status: 'REJECTED',
                final_tally_result: {
                    yes: '25000000',
                    abstain: '15000000',
                    no: '45000000',
                    no_with_veto: '8000000'
                },
                submit_time: '2023-12-20T11:00:00Z',
                deposit_end_time: '2024-01-03T11:00:00Z',
                total_deposit: [{ denom: 'aura', amount: '10000000' }],
                voting_start_time: '2023-12-28T11:00:00Z',
                voting_end_time: '2024-01-13T11:00:00Z'
            }
        ];
    }

    getMockVotes(proposalId) {
        const voteOptions = ['VOTE_OPTION_YES', 'VOTE_OPTION_NO', 'VOTE_OPTION_ABSTAIN', 'VOTE_OPTION_NO_WITH_VETO'];
        const votes = [];

        for (let i = 0; i < 50; i++) {
            votes.push({
                proposal_id: proposalId,
                voter: `aura1${Math.random().toString(36).substring(2, 42)}`,
                option: voteOptions[Math.floor(Math.random() * voteOptions.length)],
                timestamp: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
            });
        }

        return votes;
    }

    getMockDeposits(proposalId) {
        const deposits = [];

        for (let i = 0; i < 5; i++) {
            deposits.push({
                proposal_id: proposalId,
                depositor: `aura1${Math.random().toString(36).substring(2, 42)}`,
                amount: [{ denom: 'aura', amount: String(Math.floor(Math.random() * 5000000) + 1000000) }],
                timestamp: new Date(Date.now() - Math.random() * 14 * 24 * 60 * 60 * 1000).toISOString()
            });
        }

        return deposits;
    }

    getMockParameters() {
        return {
            deposit: {
                min_deposit: [{ denom: 'aura', amount: '10000000' }],
                max_deposit_period: '1209600s' // 14 days
            },
            voting: {
                voting_period: '1209600s' // 14 days
            },
            tally: {
                quorum: '0.334000000000000000',
                threshold: '0.500000000000000000',
                veto_threshold: '0.334000000000000000'
            }
        };
    }

    getMockUserVotes(address) {
        return [
            {
                proposal_id: '1',
                option: 'VOTE_OPTION_YES',
                timestamp: '2024-01-22T15:30:00Z'
            },
            {
                proposal_id: '3',
                option: 'VOTE_OPTION_YES',
                timestamp: '2024-01-12T10:15:00Z'
            },
            {
                proposal_id: '5',
                option: 'VOTE_OPTION_NO',
                timestamp: '2023-12-30T14:20:00Z'
            }
        ];
    }

    getEmptyTally() {
        return {
            yes: '0',
            abstain: '0',
            no: '0',
            no_with_veto: '0'
        };
    }
}

// AURA Blockchain Example Code Repository
// Comprehensive examples for all 20 AURA modules

// Import module-specific examples
import { vcregistryExamples } from './vcregistry.js';
import { authExamples } from './auth.js';
import { auraModuleExamples } from './aura-modules.js';

export const examples = {
    // ============ GETTING STARTED ============
    'hello-world': {
        title: 'Hello World',
        description: 'Simple example to get started with AURA',
        category: 'getting-started',
        language: 'javascript',
        code: `// Welcome to AURA Blockchain Playground!
// This is a simple example to get you started

console.log('Hello, AURA Blockchain!');

// Query node information
const nodeInfo = await api.getNodeInfo();
console.log('Node Info:', nodeInfo);
console.log('Chain ID:', nodeInfo.default_node_info.network);
console.log('Node Version:', nodeInfo.application_version.version);

// Return the result
return nodeInfo;`
    },

    'query-balance': {
        title: 'Query Balance',
        description: 'Check account balance',
        category: 'getting-started',
        language: 'javascript',
        code: `// Query account balance
const address = 'aura1...'; // Replace with actual address

// Get all balances
const balances = await api.getAllBalances(address);
console.log('All Balances:', balances);

// Get specific denom balance
const auraBalance = await api.getBalance(address, 'uaura');
console.log('AURA Balance:', auraBalance);

return balances;`
    },

    // ============ BANK MODULE ============
    'bank-transfer': {
        title: 'Bank Transfer',
        description: 'Send tokens to another address',
        category: 'bank',
        language: 'javascript',
        code: `// Send tokens (requires wallet connection)
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Transaction message
const msg = {
    type: 'cosmos-sdk/MsgSend',
    value: {
        from_address: wallet.address,
        to_address: 'aura1...', // Recipient address
        amount: [{
            denom: 'uaura',
            amount: '1000000' // 1 AURA (1,000,000 uaura)
        }]
    }
};

console.log('Transaction message:', msg);
console.log('Note: This is a preview. Use Keplr to sign and broadcast.');

return { transaction: msg };`
    },

    'multi-send': {
        title: 'Multi Send',
        description: 'Send tokens to multiple addresses',
        category: 'bank',
        language: 'javascript',
        code: `// Multi-send tokens
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

const msg = {
    type: 'cosmos-sdk/MsgMultiSend',
    value: {
        inputs: [{
            address: wallet.address,
            coins: [{
                denom: 'uaura',
                amount: '3000000'
            }]
        }],
        outputs: [
            {
                address: 'aura1...', // Recipient 1
                coins: [{
                    denom: 'uaura',
                    amount: '1000000'
                }]
            },
            {
                address: 'aura1...', // Recipient 2
                coins: [{
                    denom: 'uaura',
                    amount: '1000000'
                }]
            },
            {
                address: 'aura1...', // Recipient 3
                coins: [{
                    denom: 'uaura',
                    amount: '1000000'
                }]
            }
        ]
    }
};

console.log('Multi-send transaction:', msg);
return { transaction: msg };`
    },

    // ============ STAKING MODULE ============
    'staking': {
        title: 'Delegate Tokens',
        description: 'Stake tokens with a validator',
        category: 'staking',
        language: 'javascript',
        code: `// Delegate tokens to a validator
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get validators
const validators = await api.getValidators('BOND_STATUS_BONDED');
console.log('Active validators:', validators);

// Select a validator (use first one for example)
const validator = validators.validators[0];
console.log('Delegating to:', validator.description.moniker);

// Create delegation message
const msg = {
    type: 'cosmos-sdk/MsgDelegate',
    value: {
        delegator_address: wallet.address,
        validator_address: validator.operator_address,
        amount: {
            denom: 'uaura',
            amount: '1000000' // 1 AURA
        }
    }
};

console.log('Delegation transaction:', msg);
return { transaction: msg, validator };`
    },

    'unstaking': {
        title: 'Undelegate Tokens',
        description: 'Unstake tokens from a validator',
        category: 'staking',
        language: 'javascript',
        code: `// Undelegate tokens from a validator
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get current delegations
const delegations = await api.getDelegations(wallet.address);
console.log('Current delegations:', delegations);

// Create undelegation message (use first delegation for example)
if (delegations.delegation_responses.length === 0) {
    console.error('No active delegations found');
    return;
}

const delegation = delegations.delegation_responses[0];

const msg = {
    type: 'cosmos-sdk/MsgUndelegate',
    value: {
        delegator_address: wallet.address,
        validator_address: delegation.delegation.validator_address,
        amount: {
            denom: 'uaura',
            amount: '1000000' // 1 AURA
        }
    }
};

console.log('Undelegation transaction:', msg);
return { transaction: msg };`
    },

    'claim-rewards': {
        title: 'Claim Rewards',
        description: 'Claim staking rewards',
        category: 'staking',
        language: 'javascript',
        code: `// Claim staking rewards
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get pending rewards
const rewards = await api.getDelegationRewards(wallet.address);
console.log('Pending rewards:', rewards);

// Get delegations to claim from
const delegations = await api.getDelegations(wallet.address);

// Create claim rewards messages for each validator
const messages = delegations.delegation_responses.map(delegation => ({
    type: 'cosmos-sdk/MsgWithdrawDelegationReward',
    value: {
        delegator_address: wallet.address,
        validator_address: delegation.delegation.validator_address
    }
}));

console.log('Claim rewards transactions:', messages);
return { transaction: messages, rewards };`
    },

    // ============ GOVERNANCE MODULE ============
    'governance': {
        title: 'Submit Proposal',
        description: 'Create a governance proposal',
        category: 'governance',
        language: 'javascript',
        code: `// Submit a governance proposal
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get governance parameters
const params = await api.getGovParams('deposit');
console.log('Governance params:', params);

// Create proposal message
const msg = {
    type: 'cosmos-sdk/MsgSubmitProposal',
    value: {
        content: {
            type: 'cosmos-sdk/TextProposal',
            value: {
                title: 'Example Proposal',
                description: 'This is an example governance proposal for AURA'
            }
        },
        initial_deposit: [{
            denom: 'uaura',
            amount: '10000000' // 10 AURA minimum deposit
        }],
        proposer: wallet.address
    }
};

console.log('Proposal transaction:', msg);
return { transaction: msg, params };`
    },

    'vote': {
        title: 'Vote on Proposal',
        description: 'Cast a vote on a governance proposal',
        category: 'governance',
        language: 'javascript',
        code: `// Vote on a governance proposal
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get active proposals
const proposals = await api.getProposals('PROPOSAL_STATUS_VOTING_PERIOD');
console.log('Active proposals:', proposals);

if (proposals.proposals.length === 0) {
    console.log('No active proposals to vote on');
    return;
}

// Vote on first proposal
const proposalId = proposals.proposals[0].proposal_id;

// Create vote message
// Vote options: VOTE_OPTION_YES, VOTE_OPTION_NO, VOTE_OPTION_ABSTAIN, VOTE_OPTION_NO_WITH_VETO
const msg = {
    type: 'cosmos-sdk/MsgVote',
    value: {
        proposal_id: proposalId,
        voter: wallet.address,
        option: 'VOTE_OPTION_YES'
    }
};

console.log('Vote transaction:', msg);
return { transaction: msg, proposal: proposals.proposals[0] };`
    },

    // ============ MERGE ALL MODULE EXAMPLES ============
    ...vcregistryExamples,
    ...authExamples,
    ...auraModuleExamples
};

// Export module categories for UI organization
export const categories = {
    'getting-started': {
        name: 'Getting Started',
        description: 'Introduction to AURA blockchain',
        icon: '🚀'
    },
    'bank': {
        name: 'Bank Module',
        description: 'Token transfers and balance queries',
        icon: '💰'
    },
    'staking': {
        name: 'Staking',
        description: 'Delegate, undelegate, and claim rewards',
        icon: '🔒'
    },
    'governance': {
        name: 'Governance',
        description: 'Proposals and voting',
        icon: '🗳️'
    },
    'vcregistry': {
        name: 'VC Registry',
        description: 'Verifiable Credentials (Most Important)',
        icon: '🎫'
    },
    'auth': {
        name: 'Auth Module',
        description: 'Authentication and account management',
        icon: '🔐'
    },
    'bridge': {
        name: 'Bridge Module',
        description: 'Cross-chain transfers',
        icon: '🌉'
    },
    'compliance': {
        name: 'Compliance',
        description: 'KYC/AML and compliance checks',
        icon: '✅'
    },
    'confidencescore': {
        name: 'Confidence Score',
        description: 'User reputation and scoring',
        icon: '⭐'
    },
    'cryptography': {
        name: 'Cryptography',
        description: 'Encryption and key management',
        icon: '🔑'
    },
    'dataregistry': {
        name: 'Data Registry',
        description: 'Data registration and management',
        icon: '📊'
    },
    'dex': {
        name: 'DEX',
        description: 'Decentralized exchange',
        icon: '💱'
    },
    'economicsecurity': {
        name: 'Economic Security',
        description: 'Dynamic fees and MEV protection',
        icon: '💵'
    },
    'identitychange': {
        name: 'Identity Change',
        description: 'Identity management',
        icon: '👤'
    },
    'inclusionroutines': {
        name: 'Inclusion Routines',
        description: 'Verification routines',
        icon: '📋'
    },
    'monitoring': {
        name: 'Monitoring',
        description: 'System monitoring and alerts',
        icon: '📈'
    },
    'networksecurity': {
        name: 'Network Security',
        description: 'Network protection and peer management',
        icon: '🛡️'
    },
    'prevalidation': {
        name: 'Prevalidation',
        description: 'Transaction prevalidation',
        icon: '✔️'
    },
    'privacy': {
        name: 'Privacy',
        description: 'Privacy-preserving transactions',
        icon: '🔒'
    },
    'validatorsecurity': {
        name: 'Validator Security',
        description: 'Validator monitoring and security',
        icon: '🛡️'
    },
    'walletsecurity': {
        name: 'Wallet Security',
        description: 'Wallet protection and 2FA',
        icon: '🔐'
    }
};

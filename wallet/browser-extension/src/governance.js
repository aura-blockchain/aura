/**
 * Governance Module
 * Handles proposal querying, voting, and deposits
 */

const GovernanceModule = {
  /**
   * Vote options
   */
  VOTE_OPTIONS: {
    YES: 1,
    ABSTAIN: 2,
    NO: 3,
    NO_WITH_VETO: 4,
  },

  /**
   * Proposal status
   */
  PROPOSAL_STATUS: {
    UNSPECIFIED: 0,
    DEPOSIT_PERIOD: 1,
    VOTING_PERIOD: 2,
    PASSED: 3,
    REJECTED: 4,
    FAILED: 5,
  },

  /**
   * Build vote transaction
   * @param {Object} params - Vote parameters
   * @returns {Object} Unsigned transaction
   */
  buildVoteTx(params) {
    const {
      proposalId,
      voter,
      option,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '2000' }], gas: '150000' },
    } = params;

    // Validate vote option
    if (!Object.values(this.VOTE_OPTIONS).includes(option)) {
      throw new Error('Invalid vote option');
    }

    return {
      body: {
        messages: [{
          '@type': '/cosmos.gov.v1beta1.MsgVote',
          proposal_id: proposalId.toString(),
          voter,
          option,
        }],
        memo,
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      },
      auth_info: {
        signer_infos: [],
        fee,
      },
      signatures: [],
    };
  },

  /**
   * Build weighted vote transaction
   * @param {Object} params - Weighted vote parameters
   * @returns {Object} Unsigned transaction
   */
  buildWeightedVoteTx(params) {
    const {
      proposalId,
      voter,
      options, // Array of {option, weight}
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '2000' }], gas: '150000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.gov.v1beta1.MsgVoteWeighted',
          proposal_id: proposalId.toString(),
          voter,
          options,
        }],
        memo,
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      },
      auth_info: {
        signer_infos: [],
        fee,
      },
      signatures: [],
    };
  },

  /**
   * Build deposit transaction
   * @param {Object} params - Deposit parameters
   * @returns {Object} Unsigned transaction
   */
  buildDepositTx(params) {
    const {
      proposalId,
      depositor,
      amount,
      denom = COSMOS_SDK.config.coinDenom,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '5000' }], gas: '200000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.gov.v1beta1.MsgDeposit',
          proposal_id: proposalId.toString(),
          depositor,
          amount: [{ denom, amount: amount.toString() }],
        }],
        memo,
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      },
      auth_info: {
        signer_infos: [],
        fee,
      },
      signatures: [],
    };
  },

  /**
   * Build submit proposal transaction (text proposal)
   * @param {Object} params - Proposal parameters
   * @returns {Object} Unsigned transaction
   */
  buildSubmitProposalTx(params) {
    const {
      proposer,
      title,
      description,
      initialDeposit,
      denom = COSMOS_SDK.config.coinDenom,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '10000' }], gas: '300000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.gov.v1beta1.MsgSubmitProposal',
          content: {
            '@type': '/cosmos.gov.v1beta1.TextProposal',
            title,
            description,
          },
          initial_deposit: [{ denom, amount: initialDeposit.toString() }],
          proposer,
        }],
        memo,
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      },
      auth_info: {
        signer_infos: [],
        fee,
      },
      signatures: [],
    };
  },

  /**
   * Query all proposals
   * @param {string} status - Optional status filter
   * @returns {Promise<Array>} Proposals
   */
  async queryProposals(status = null) {
    try {
      let url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals`;

      if (status !== null) {
        url += `?proposal_status=${status}`;
      }

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.proposals || [];
    } catch (error) {
      console.error('Query proposals error:', error);
      throw new Error(`Failed to query proposals: ${error.message}`);
    }
  },

  /**
   * Query specific proposal
   * @param {number} proposalId - Proposal ID
   * @returns {Promise<Object>} Proposal
   */
  async queryProposal(proposalId) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.proposal;
    } catch (error) {
      console.error('Query proposal error:', error);
      throw new Error(`Failed to query proposal: ${error.message}`);
    }
  },

  /**
   * Query proposal tally
   * @param {number} proposalId - Proposal ID
   * @returns {Promise<Object>} Tally result
   */
  async queryProposalTally(proposalId) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/tally`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.tally;
    } catch (error) {
      console.error('Query tally error:', error);
      throw new Error(`Failed to query tally: ${error.message}`);
    }
  },

  /**
   * Query vote for a proposal
   * @param {number} proposalId - Proposal ID
   * @param {string} voter - Voter address
   * @returns {Promise<Object>} Vote
   */
  async queryVote(proposalId, voter) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/votes/${voter}`;
      const response = await fetch(url);

      if (!response.ok) {
        // Vote not found
        if (response.status === 404) {
          return null;
        }
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.vote;
    } catch (error) {
      console.error('Query vote error:', error);
      return null;
    }
  },

  /**
   * Query all votes for a proposal
   * @param {number} proposalId - Proposal ID
   * @returns {Promise<Array>} Votes
   */
  async queryVotes(proposalId) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/votes`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.votes || [];
    } catch (error) {
      console.error('Query votes error:', error);
      throw new Error(`Failed to query votes: ${error.message}`);
    }
  },

  /**
   * Query deposits for a proposal
   * @param {number} proposalId - Proposal ID
   * @returns {Promise<Array>} Deposits
   */
  async queryDeposits(proposalId) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/deposits`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.deposits || [];
    } catch (error) {
      console.error('Query deposits error:', error);
      throw new Error(`Failed to query deposits: ${error.message}`);
    }
  },

  /**
   * Query governance parameters
   * @param {string} paramsType - deposit, voting, or tallying
   * @returns {Promise<Object>} Parameters
   */
  async queryParams(paramsType = 'voting') {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/gov/v1beta1/params/${paramsType}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data[`${paramsType}_params`];
    } catch (error) {
      console.error('Query params error:', error);
      throw new Error(`Failed to query params: ${error.message}`);
    }
  },

  /**
   * Get vote option name
   * @param {number} option - Vote option number
   * @returns {string} Option name
   */
  getVoteOptionName(option) {
    const names = {
      [this.VOTE_OPTIONS.YES]: 'Yes',
      [this.VOTE_OPTIONS.ABSTAIN]: 'Abstain',
      [this.VOTE_OPTIONS.NO]: 'No',
      [this.VOTE_OPTIONS.NO_WITH_VETO]: 'No With Veto',
    };
    return names[option] || 'Unknown';
  },

  /**
   * Get proposal status name
   * @param {number} status - Proposal status number
   * @returns {string} Status name
   */
  getProposalStatusName(status) {
    const names = {
      [this.PROPOSAL_STATUS.UNSPECIFIED]: 'Unspecified',
      [this.PROPOSAL_STATUS.DEPOSIT_PERIOD]: 'Deposit Period',
      [this.PROPOSAL_STATUS.VOTING_PERIOD]: 'Voting Period',
      [this.PROPOSAL_STATUS.PASSED]: 'Passed',
      [this.PROPOSAL_STATUS.REJECTED]: 'Rejected',
      [this.PROPOSAL_STATUS.FAILED]: 'Failed',
    };
    return names[status] || 'Unknown';
  },

  /**
   * Check if proposal is active (can be voted on)
   * @param {Object} proposal - Proposal object
   * @returns {boolean} Is active
   */
  isProposalActive(proposal) {
    return proposal.status === this.PROPOSAL_STATUS.VOTING_PERIOD;
  },

  /**
   * Check if proposal is in deposit period
   * @param {Object} proposal - Proposal object
   * @returns {boolean} Is in deposit period
   */
  isProposalInDepositPeriod(proposal) {
    return proposal.status === this.PROPOSAL_STATUS.DEPOSIT_PERIOD;
  },

  /**
   * Format proposal for display
   * @param {Object} proposal - Proposal object
   * @returns {Object} Formatted proposal
   */
  formatProposal(proposal) {
    return {
      id: proposal.proposal_id,
      title: proposal.content?.title || 'Untitled',
      description: proposal.content?.description || 'No description',
      status: this.getProposalStatusName(proposal.status),
      statusCode: proposal.status,
      submitTime: new Date(proposal.submit_time),
      depositEndTime: new Date(proposal.deposit_end_time),
      votingStartTime: proposal.voting_start_time ? new Date(proposal.voting_start_time) : null,
      votingEndTime: proposal.voting_end_time ? new Date(proposal.voting_end_time) : null,
      totalDeposit: proposal.total_deposit,
      isActive: this.isProposalActive(proposal),
      isInDepositPeriod: this.isProposalInDepositPeriod(proposal),
    };
  },
};

// Export for use in extension
if (typeof module !== 'undefined' && module.exports) {
  module.exports = GovernanceModule;
}

/**
 * Staking Module
 * Handles delegation, undelegation, redelegation, and reward claiming
 */

const StakingModule = {
  /**
   * Build delegate transaction
   * @param {Object} params - Delegation parameters
   * @returns {Object} Unsigned transaction
   */
  buildDelegateTx(params) {
    const {
      delegatorAddress,
      validatorAddress,
      amount,
      denom = COSMOS_SDK.config.coinDenom,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '5000' }], gas: '250000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.staking.v1beta1.MsgDelegate',
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: { denom, amount: amount.toString() },
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
   * Build undelegate transaction
   * @param {Object} params - Undelegation parameters
   * @returns {Object} Unsigned transaction
   */
  buildUndelegateTx(params) {
    const {
      delegatorAddress,
      validatorAddress,
      amount,
      denom = COSMOS_SDK.config.coinDenom,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '5000' }], gas: '250000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.staking.v1beta1.MsgUndelegate',
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: { denom, amount: amount.toString() },
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
   * Build redelegate transaction
   * @param {Object} params - Redelegation parameters
   * @returns {Object} Unsigned transaction
   */
  buildRedelegateTx(params) {
    const {
      delegatorAddress,
      validatorSrcAddress,
      validatorDstAddress,
      amount,
      denom = COSMOS_SDK.config.coinDenom,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '5000' }], gas: '250000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.staking.v1beta1.MsgBeginRedelegate',
          delegator_address: delegatorAddress,
          validator_src_address: validatorSrcAddress,
          validator_dst_address: validatorDstAddress,
          amount: { denom, amount: amount.toString() },
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
   * Build withdraw rewards transaction
   * @param {Object} params - Withdraw parameters
   * @returns {Object} Unsigned transaction
   */
  buildWithdrawRewardsTx(params) {
    const {
      delegatorAddress,
      validatorAddress,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '5000' }], gas: '200000' },
    } = params;

    return {
      body: {
        messages: [{
          '@type': '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
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
   * Build withdraw all rewards transaction
   * @param {Object} params - Withdraw parameters
   * @returns {Object} Unsigned transaction
   */
  buildWithdrawAllRewardsTx(params) {
    const {
      delegatorAddress,
      validatorAddresses,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '10000' }], gas: '500000' },
    } = params;

    const messages = validatorAddresses.map(validatorAddress => ({
      '@type': '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
      delegator_address: delegatorAddress,
      validator_address: validatorAddress,
    }));

    return {
      body: {
        messages,
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
   * Query delegations for an address
   * @param {string} delegatorAddress - Delegator address
   * @returns {Promise<Array>} Delegations
   */
  async queryDelegations(delegatorAddress) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/staking/v1beta1/delegations/${delegatorAddress}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.delegation_responses || [];
    } catch (error) {
      console.error('Query delegations error:', error);
      throw new Error(`Failed to query delegations: ${error.message}`);
    }
  },

  /**
   * Query unbonding delegations for an address
   * @param {string} delegatorAddress - Delegator address
   * @returns {Promise<Array>} Unbonding delegations
   */
  async queryUnbondingDelegations(delegatorAddress) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/staking/v1beta1/delegators/${delegatorAddress}/unbonding_delegations`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.unbonding_responses || [];
    } catch (error) {
      console.error('Query unbonding delegations error:', error);
      throw new Error(`Failed to query unbonding delegations: ${error.message}`);
    }
  },

  /**
   * Query delegation rewards for an address
   * @param {string} delegatorAddress - Delegator address
   * @param {string} validatorAddress - Optional validator address
   * @returns {Promise<Object>} Rewards
   */
  async queryRewards(delegatorAddress, validatorAddress = null) {
    try {
      let url;
      if (validatorAddress) {
        url = `${COSMOS_SDK.config.restEndpoint}/cosmos/distribution/v1beta1/delegators/${delegatorAddress}/rewards/${validatorAddress}`;
      } else {
        url = `${COSMOS_SDK.config.restEndpoint}/cosmos/distribution/v1beta1/delegators/${delegatorAddress}/rewards`;
      }

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error('Query rewards error:', error);
      throw new Error(`Failed to query rewards: ${error.message}`);
    }
  },

  /**
   * Query all validators
   * @param {string} status - Validator status (BOND_STATUS_BONDED, BOND_STATUS_UNBONDED, etc.)
   * @returns {Promise<Array>} Validators
   */
  async queryValidators(status = 'BOND_STATUS_BONDED') {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/staking/v1beta1/validators?status=${status}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.validators || [];
    } catch (error) {
      console.error('Query validators error:', error);
      throw new Error(`Failed to query validators: ${error.message}`);
    }
  },

  /**
   * Query specific validator
   * @param {string} validatorAddress - Validator address
   * @returns {Promise<Object>} Validator
   */
  async queryValidator(validatorAddress) {
    try {
      const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/staking/v1beta1/validators/${validatorAddress}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.validator;
    } catch (error) {
      console.error('Query validator error:', error);
      throw new Error(`Failed to query validator: ${error.message}`);
    }
  },

  /**
   * Calculate total staked amount
   * @param {Array} delegations - Array of delegations
   * @returns {number} Total staked
   */
  calculateTotalStaked(delegations) {
    return delegations.reduce((sum, del) => {
      const amount = parseInt(del.balance?.amount || 0);
      return sum + amount;
    }, 0);
  },

  /**
   * Calculate total rewards
   * @param {Object} rewardsData - Rewards data from query
   * @returns {number} Total rewards
   */
  calculateTotalRewards(rewardsData) {
    if (!rewardsData.total || !Array.isArray(rewardsData.total)) {
      return 0;
    }

    return rewardsData.total.reduce((sum, coin) => {
      if (coin.denom === COSMOS_SDK.config.coinDenom) {
        return sum + parseFloat(coin.amount || 0);
      }
      return sum;
    }, 0);
  },

  /**
   * Validate validator address
   * @param {string} address - Validator address
   * @returns {boolean} Is valid
   */
  validateValidatorAddress(address) {
    if (!address || typeof address !== 'string') {
      return false;
    }
    // Validator addresses have 'valoper' in them
    const prefix = COSMOS_SDK.config.bech32Prefix + 'valoper';
    return address.startsWith(prefix) && address.length >= 45;
  },
};

// Export for use in extension
if (typeof module !== 'undefined' && module.exports) {
  module.exports = StakingModule;
}

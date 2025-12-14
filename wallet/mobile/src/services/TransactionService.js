/**
 * Transaction Service
 * Handles transaction signing and broadcasting for staking, governance, and other operations
 * Uses native crypto libraries instead of CosmJS for React Native compatibility
 */

import {sha256} from 'js-sha256';
import {ec as EC} from 'elliptic';
import PawAPI from './PawAPI';

const ec = new EC('secp256k1');

class TransactionServiceClass {
  /**
   * Create a signed transaction
   * @param {Object} params - Transaction parameters
   * @returns {Promise<Object>} Signed transaction ready for broadcast
   */
  async createSignedTransaction({
    messages,
    fee,
    memo,
    accountNumber,
    sequence,
    chainId,
    privateKeyHex,
  }) {
    try {
      // Create transaction body
      const txBody = {
        messages,
        memo: memo || '',
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      };

      // Create auth info
      const authInfo = {
        signer_infos: [
          {
            public_key: {
              type_url: '/cosmos.crypto.secp256k1.PubKey',
              value: this.getPubKeyBytes(privateKeyHex),
            },
            mode_info: {
              single: {
                mode: 1, // SIGN_MODE_DIRECT
              },
            },
            sequence: sequence.toString(),
          },
        ],
        fee: {
          amount: fee.amount || [{denom: 'uaura', amount: '5000'}],
          gas_limit: fee.gas || '200000',
          payer: '',
          granter: '',
        },
      };

      // Create sign doc
      const signDoc = {
        body_bytes: this.encodeTxBody(txBody),
        auth_info_bytes: this.encodeAuthInfo(authInfo),
        chain_id: chainId,
        account_number: accountNumber.toString(),
      };

      // Sign the transaction
      const signature = this.signTransaction(signDoc, privateKeyHex);

      // Create final transaction
      const tx = {
        body: txBody,
        auth_info: authInfo,
        signatures: [signature],
      };

      return tx;
    } catch (error) {
      console.error('Failed to create signed transaction:', error);
      throw error;
    }
  }

  /**
   * Get public key bytes from private key
   */
  getPubKeyBytes(privateKeyHex) {
    const keyPair = ec.keyFromPrivate(privateKeyHex, 'hex');
    const pubKey = keyPair.getPublic();
    return Buffer.from(pubKey.encode('array', true)).toString('base64');
  }

  /**
   * Encode transaction body (simplified version)
   */
  encodeTxBody(txBody) {
    // In a real implementation, use protobuf encoding
    // For now, return a simple JSON string representation
    return Buffer.from(JSON.stringify(txBody)).toString('base64');
  }

  /**
   * Encode auth info (simplified version)
   */
  encodeAuthInfo(authInfo) {
    // In a real implementation, use protobuf encoding
    return Buffer.from(JSON.stringify(authInfo)).toString('base64');
  }

  /**
   * Sign transaction
   */
  signTransaction(signDoc, privateKeyHex) {
    const keyPair = ec.keyFromPrivate(privateKeyHex, 'hex');

    // Create sign bytes
    const signBytes = Buffer.from(
      JSON.stringify({
        chain_id: signDoc.chain_id,
        account_number: signDoc.account_number,
        sequence: signDoc.sequence,
        fee: signDoc.fee,
        msgs: signDoc.messages,
        memo: signDoc.memo || '',
      }),
    );

    // Hash and sign
    const hash = sha256.array(signBytes);
    const signature = keyPair.sign(hash);

    // Return DER encoded signature
    return Buffer.from(signature.toDER()).toString('base64');
  }

  /**
   * Undelegate tokens from a validator
   */
  async undelegate({
    delegatorAddress,
    validatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgUndelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to undelegate:', error);
      throw new Error(error.message || 'Failed to undelegate tokens');
    }
  }

  /**
   * Withdraw delegation rewards
   */
  async withdrawRewards({
    delegatorAddress,
    validatorAddress,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      const message = {
        type_url: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to withdraw rewards:', error);
      throw new Error(error.message || 'Failed to withdraw rewards');
    }
  }

  /**
   * Redelegate tokens from one validator to another
   */
  async redelegate({
    delegatorAddress,
    srcValidatorAddress,
    dstValidatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_src_address: srcValidatorAddress,
          validator_dst_address: dstValidatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to redelegate:', error);
      throw new Error(error.message || 'Failed to redelegate tokens');
    }
  }

  /**
   * Vote on a governance proposal
   */
  async vote({
    voterAddress,
    proposalId,
    option,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      // Map option string to VoteOption enum
      const voteOptionMap = {
        yes: 1,
        abstain: 2,
        no: 3,
        no_with_veto: 4,
      };

      const voteOption =
        typeof option === 'string'
          ? voteOptionMap[option.toLowerCase()]
          : option;

      if (!voteOption) {
        throw new Error(
          'Invalid vote option. Must be: yes, abstain, no, or no_with_veto',
        );
      }

      const message = {
        type_url: '/cosmos.gov.v1beta1.MsgVote',
        value: {
          proposal_id: proposalId.toString(),
          voter: voterAddress,
          option: voteOption,
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to vote:', error);
      throw new Error(error.message || 'Failed to submit vote');
    }
  }

  /**
   * Delegate tokens to a validator
   */
  async delegate({
    delegatorAddress,
    validatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgDelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to delegate:', error);
      throw new Error(error.message || 'Failed to delegate tokens');
    }
  }

  /**
   * Send tokens
   */
  async sendTokens({
    fromAddress,
    toAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      const message = {
        type_url: '/cosmos.bank.v1beta1.MsgSend',
        value: {
          from_address: fromAddress,
          to_address: toAddress,
          amount: [{denom, amount: amount.toString()}],
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: 'uaura', amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to send tokens:', error);
      throw new Error(error.message || 'Failed to send tokens');
    }
  }
}

// Export singleton instance
const TransactionService = new TransactionServiceClass();
export default TransactionService;

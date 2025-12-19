/**
 * API Service Tests
 */

import { ApiService } from '../src/services/api';
import axios from 'axios';

jest.mock('axios');

describe('ApiService', () => {
  let apiService;

  beforeEach(async () => {
    jest.clearAllMocks();
    await window.electron.store.clear();
    apiService = new ApiService();
  });

  describe('Balance', () => {
    test('should fetch account balance', async () => {
      const mockBalance = {
        balances: [
          { denom: 'uaura', amount: '1000000' }
        ]
      };

      axios.get.mockResolvedValue({ data: mockBalance });

      const balance = await apiService.getBalance('aura1test');

      expect(axios.get).toHaveBeenCalled();
      expect(balance).toEqual(mockBalance);
    });

    test('should handle balance fetch error', async () => {
      axios.get.mockRejectedValue(new Error('Network error'));

      await expect(
        apiService.getBalance('aura1test')
      ).rejects.toThrow();
    });
  });

  describe('Send Tokens Validation', () => {
    test('should reject invalid sender address', async () => {
      await expect(
        apiService.sendTokens(
          'invalid',
          'aura1recipient',
          1000,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid sender address');
    });

    test('should reject non-positive amounts', async () => {
      await expect(
        apiService.sendTokens(
          'aura1sender',
          'aura1recipient',
          0,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Amount must be greater than zero');
    });

    test('should reject missing signing key', async () => {
      await expect(
        apiService.sendTokens(
          'aura1sender',
          'aura1recipient',
          10,
          'uaura',
          '',
          ''
        )
      ).rejects.toThrow('Missing signing key');
    });
  });

  describe('Delegate Validation', () => {
    test('should reject invalid delegator address', async () => {
      await expect(
        apiService.delegate(
          'invalid',
          'auravaloper1validator',
          1000,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid delegator address');
    });

    test('should reject invalid validator address', async () => {
      await expect(
        apiService.delegate(
          'aura1delegator',
          'badval',
          1000,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid validator address');
    });

    test('should reject non-positive delegate amount', async () => {
      await expect(
        apiService.delegate(
          'aura1delegator',
          'auravaloper1validator',
          0,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Delegate amount must be greater than zero');
    });
  });

  describe('Undelegate Validation', () => {
    test('should reject invalid undelegate inputs', async () => {
      await expect(
        apiService.undelegate(
          '',
          'auravaloper1validator',
          1000,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid delegator address');

      await expect(
        apiService.undelegate(
          'aura1delegator',
          '',
          1000,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid validator address');

      await expect(
        apiService.undelegate(
          'aura1delegator',
          'auravaloper1validator',
          0,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Undelegate amount must be greater than zero');
    });
  });

  describe('Redelegate Validation', () => {
    test('should reject invalid redelegate inputs', async () => {
      await expect(
        apiService.redelegate(
          '',
          'auravaloper1src',
          'auravaloper1dst',
          10,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid delegator address');

      await expect(
        apiService.redelegate(
          'aura1delegator',
          'badval',
          'auravaloper1dst',
          10,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Invalid validator address');

      await expect(
        apiService.redelegate(
          'aura1delegator',
          'auravaloper1src',
          'auravaloper1dst',
          0,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Redelegate amount must be greater than zero');
    });
  });

  describe('Account', () => {
    test('should fetch account information', async () => {
      const mockAccount = {
        account: {
          address: 'aura1test',
          account_number: '1',
          sequence: '0'
        }
      };

      axios.get.mockResolvedValue({ data: mockAccount });

      const account = await apiService.getAccount('aura1test');

      expect(axios.get).toHaveBeenCalled();
      expect(account).toEqual(mockAccount.account);
    });
  });

  describe('Transactions', () => {
    test('should fetch transaction history', async () => {
      const mockTxs = {
        txs: [
          { txhash: 'hash1', height: '100' },
          { txhash: 'hash2', height: '101' }
        ]
      };

      axios.get.mockResolvedValue({ data: mockTxs });

      const txs = await apiService.getTransactions('aura1test');

      expect(axios.get).toHaveBeenCalled();
      expect(txs).toEqual(mockTxs.txs);
    });

    test('should return empty array on error', async () => {
      axios.get.mockRejectedValue(new Error('Not found'));

      const txs = await apiService.getTransactions('aura1test');

      expect(txs).toEqual([]);
    });
  });

  describe('Node Information', () => {
    test('should fetch node info', async () => {
      const mockNodeInfo = {
        node_info: {
          network: 'aura-testnet',
          version: '1.0.0'
        }
      };

      axios.get.mockResolvedValue({ data: mockNodeInfo });

      const nodeInfo = await apiService.getNodeInfo();

      expect(axios.get).toHaveBeenCalled();
      expect(nodeInfo).toEqual(mockNodeInfo.node_info);
    });
  });

  describe('Validators', () => {
    test('should fetch validator list', async () => {
      const mockValidators = {
        validators: [
          { operator_address: 'auravaloper1test', status: 'BOND_STATUS_BONDED' }
        ]
      };

      axios.get.mockResolvedValue({ data: mockValidators });

      const validators = await apiService.getValidators();

      expect(axios.get).toHaveBeenCalled();
      expect(validators).toEqual(mockValidators.validators);
    });
  });

  describe('DEX Pools', () => {
    test('should fetch DEX pools', async () => {
      const mockPools = {
        pools: [
          { id: '1', token_a: 'uaura', token_b: 'uatom' }
        ]
      };

      axios.get.mockResolvedValue({ data: mockPools });

      const pools = await apiService.getDexPools();

      expect(axios.get).toHaveBeenCalled();
      expect(pools).toEqual(mockPools.pools);
    });

    test('should return empty array if DEX not available', async () => {
      axios.get.mockRejectedValue(new Error('Not found'));

      const pools = await apiService.getDexPools();

      expect(pools).toEqual([]);
    });
  });

  describe('Oracle Prices', () => {
    test('should fetch oracle prices', async () => {
      const mockPrices = {
        prices: [
          { symbol: 'Aura/USD', price: '1.5' }
        ]
      };

      axios.get.mockResolvedValue({ data: mockPrices });

      const prices = await apiService.getOraclePrices();

      expect(axios.get).toHaveBeenCalled();
      expect(prices).toEqual(mockPrices.prices);
    });

    test('should return empty array if oracle not available', async () => {
      axios.get.mockRejectedValue(new Error('Not found'));

      const prices = await apiService.getOraclePrices();

      expect(prices).toEqual([]);
    });
  });

  describe('Staking Operations', () => {
    describe('getDelegations', () => {
      test('should fetch delegations for an address', async () => {
        const mockDelegations = {
          delegation_responses: [
            {
              delegation: {
                delegator_address: 'aura1test',
                validator_address: 'auravaloper1test',
                shares: '1000000'
              },
              balance: { denom: 'uaura', amount: '1000000' }
            }
          ]
        };

        axios.get.mockResolvedValue({ data: mockDelegations });

        const delegations = await apiService.getDelegations('aura1test');

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/staking/v1beta1/delegations/aura1test')
        );
        expect(delegations).toEqual(mockDelegations.delegation_responses);
      });

      test('should handle delegation fetch error', async () => {
        axios.get.mockRejectedValue(new Error('Network error'));

        await expect(
          apiService.getDelegations('aura1test')
        ).rejects.toThrow();
      });
    });

    describe('getRewards', () => {
      test('should fetch delegation rewards', async () => {
        const mockRewards = {
          rewards: [
            {
              validator_address: 'auravaloper1test',
              reward: [{ denom: 'uaura', amount: '5000' }]
            }
          ],
          total: [{ denom: 'uaura', amount: '5000' }]
        };

        axios.get.mockResolvedValue({ data: mockRewards });

        const rewards = await apiService.getRewards('aura1test');

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/distribution/v1beta1/delegators/aura1test/rewards')
        );
        expect(rewards).toEqual(mockRewards);
      });

      test('should handle rewards fetch error', async () => {
        axios.get.mockRejectedValue(new Error('Network error'));

        await expect(
          apiService.getRewards('aura1test')
        ).rejects.toThrow();
      });
    });

    describe('getUnbondingDelegations', () => {
      test('should fetch unbonding delegations', async () => {
        const mockUnbonding = {
          unbonding_responses: [
            {
              delegator_address: 'aura1test',
              validator_address: 'auravaloper1test',
              entries: [
                {
                  creation_height: '1000',
                  completion_time: '2024-12-31T00:00:00Z',
                  initial_balance: '1000000',
                  balance: '1000000'
                }
              ]
            }
          ]
        };

        axios.get.mockResolvedValue({ data: mockUnbonding });

        const unbonding = await apiService.getUnbondingDelegations('aura1test');

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/staking/v1beta1/delegators/aura1test/unbonding_delegations')
        );
        expect(unbonding).toEqual(mockUnbonding.unbonding_responses);
      });

      test('should handle unbonding fetch error', async () => {
        axios.get.mockRejectedValue(new Error('Network error'));

        await expect(
          apiService.getUnbondingDelegations('aura1test')
        ).rejects.toThrow();
      });
    });
  });

  describe('Governance Operations', () => {
    describe('getProposals', () => {
      test('should fetch all proposals', async () => {
        const mockProposals = {
          proposals: [
            {
              proposal_id: '1',
              content: {
                type: '/cosmos.gov.v1beta1.TextProposal',
                value: {
                  title: 'Test Proposal',
                  description: 'Test description'
                }
              },
              status: 'PROPOSAL_STATUS_VOTING_PERIOD'
            }
          ]
        };

        axios.get.mockResolvedValue({ data: mockProposals });

        const proposals = await apiService.getProposals();

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/gov/v1beta1/proposals'),
          expect.objectContaining({ params: {} })
        );
        expect(proposals).toEqual(mockProposals.proposals);
      });

      test('should fetch proposals with status filter', async () => {
        const mockProposals = {
          proposals: [
            {
              proposal_id: '1',
              status: 'PROPOSAL_STATUS_VOTING_PERIOD'
            }
          ]
        };

        axios.get.mockResolvedValue({ data: mockProposals });

        const proposals = await apiService.getProposals('PROPOSAL_STATUS_VOTING_PERIOD');

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/gov/v1beta1/proposals'),
          expect.objectContaining({
            params: { proposal_status: 'PROPOSAL_STATUS_VOTING_PERIOD' }
          })
        );
        expect(proposals).toEqual(mockProposals.proposals);
      });

      test('should handle proposals fetch error', async () => {
        axios.get.mockRejectedValue(new Error('Network error'));

        await expect(
          apiService.getProposals()
        ).rejects.toThrow();
      });
    });

    describe('getProposal', () => {
      test('should fetch specific proposal', async () => {
        const mockProposal = {
          proposal: {
            proposal_id: '1',
            content: {
              type: '/cosmos.gov.v1beta1.TextProposal',
              value: {
                title: 'Test Proposal',
                description: 'Test description'
              }
            },
            status: 'PROPOSAL_STATUS_VOTING_PERIOD'
          }
        };

        axios.get.mockResolvedValue({ data: mockProposal });

        const proposal = await apiService.getProposal('1');

        expect(axios.get).toHaveBeenCalledWith(
          expect.stringContaining('/cosmos/gov/v1beta1/proposals/1')
        );
        expect(proposal).toEqual(mockProposal.proposal);
      });

      test('should handle proposal fetch error', async () => {
        axios.get.mockRejectedValue(new Error('Not found'));

        await expect(
          apiService.getProposal('999')
        ).rejects.toThrow();
      });
    });
  });
});

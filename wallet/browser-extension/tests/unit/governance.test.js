/**
 * Unit Tests for Governance Module
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock COSMOS_SDK
global.COSMOS_SDK = {
  config: {
    bech32Prefix: 'aura',
    coinDenom: 'uaura',
    coinDecimals: 6,
    restEndpoint: 'http://localhost:1317',
  },
};

// Mock fetch
global.fetch = vi.fn();

describe('GovernanceModule', () => {
  let GovernanceModule;

  beforeEach(async () => {
    GovernanceModule = (await import('../../src/governance.js')).default || (await import('../../src/governance.js'));
    vi.clearAllMocks();
  });

  describe('buildVoteTx', () => {
    it('should build correct vote transaction for YES', () => {
      const params = {
        proposalId: 1,
        voter: 'aura1test',
        option: GovernanceModule.VOTE_OPTIONS.YES,
      };

      const tx = GovernanceModule.buildVoteTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.gov.v1beta1.MsgVote');
      expect(tx.body.messages[0].proposal_id).toBe('1');
      expect(tx.body.messages[0].voter).toBe(params.voter);
      expect(tx.body.messages[0].option).toBe(1);
    });

    it('should throw error for invalid vote option', () => {
      const params = {
        proposalId: 1,
        voter: 'aura1test',
        option: 999, // Invalid
      };

      expect(() => GovernanceModule.buildVoteTx(params)).toThrow('Invalid vote option');
    });
  });

  describe('buildDepositTx', () => {
    it('should build correct deposit transaction', () => {
      const params = {
        proposalId: 1,
        depositor: 'aura1test',
        amount: 1000000,
      };

      const tx = GovernanceModule.buildDepositTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.gov.v1beta1.MsgDeposit');
      expect(tx.body.messages[0].proposal_id).toBe('1');
      expect(tx.body.messages[0].depositor).toBe(params.depositor);
      expect(tx.body.messages[0].amount).toEqual([{ denom: 'uaura', amount: '1000000' }]);
    });
  });

  describe('buildSubmitProposalTx', () => {
    it('should build correct submit proposal transaction', () => {
      const params = {
        proposer: 'aura1test',
        title: 'Test Proposal',
        description: 'This is a test proposal',
        initialDeposit: 5000000,
      };

      const tx = GovernanceModule.buildSubmitProposalTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.gov.v1beta1.MsgSubmitProposal');
      expect(tx.body.messages[0].content.title).toBe(params.title);
      expect(tx.body.messages[0].content.description).toBe(params.description);
      expect(tx.body.messages[0].proposer).toBe(params.proposer);
    });
  });

  describe('queryProposals', () => {
    it('should query all proposals', async () => {
      const mockProposals = [
        { proposal_id: '1', status: 2 },
        { proposal_id: '2', status: 3 },
      ];

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ proposals: mockProposals }),
      });

      const proposals = await GovernanceModule.queryProposals();

      expect(proposals).toEqual(mockProposals);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringMatching(/\/cosmos\/gov\/v1beta1\/proposals$/)
      );
    });

    it('should query proposals with status filter', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ proposals: [] }),
      });

      await GovernanceModule.queryProposals(2);

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('proposal_status=2')
      );
    });
  });

  describe('queryProposal', () => {
    it('should query specific proposal', async () => {
      const mockProposal = { proposal_id: '1', content: { title: 'Test' } };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ proposal: mockProposal }),
      });

      const proposal = await GovernanceModule.queryProposal(1);

      expect(proposal).toEqual(mockProposal);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/proposals/1')
      );
    });
  });

  describe('queryVote', () => {
    it('should query vote for proposal', async () => {
      const mockVote = { voter: 'aura1test', option: 1 };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ vote: mockVote }),
      });

      const vote = await GovernanceModule.queryVote(1, 'aura1test');

      expect(vote).toEqual(mockVote);
    });

    it('should return null if vote not found', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

      const vote = await GovernanceModule.queryVote(1, 'aura1test');

      expect(vote).toBeNull();
    });
  });

  describe('getVoteOptionName', () => {
    it('should return correct vote option names', () => {
      expect(GovernanceModule.getVoteOptionName(GovernanceModule.VOTE_OPTIONS.YES)).toBe('Yes');
      expect(GovernanceModule.getVoteOptionName(GovernanceModule.VOTE_OPTIONS.NO)).toBe('No');
      expect(GovernanceModule.getVoteOptionName(GovernanceModule.VOTE_OPTIONS.ABSTAIN)).toBe('Abstain');
      expect(GovernanceModule.getVoteOptionName(GovernanceModule.VOTE_OPTIONS.NO_WITH_VETO)).toBe('No With Veto');
      expect(GovernanceModule.getVoteOptionName(999)).toBe('Unknown');
    });
  });

  describe('getProposalStatusName', () => {
    it('should return correct status names', () => {
      expect(GovernanceModule.getProposalStatusName(GovernanceModule.PROPOSAL_STATUS.DEPOSIT_PERIOD)).toBe('Deposit Period');
      expect(GovernanceModule.getProposalStatusName(GovernanceModule.PROPOSAL_STATUS.VOTING_PERIOD)).toBe('Voting Period');
      expect(GovernanceModule.getProposalStatusName(GovernanceModule.PROPOSAL_STATUS.PASSED)).toBe('Passed');
      expect(GovernanceModule.getProposalStatusName(GovernanceModule.PROPOSAL_STATUS.REJECTED)).toBe('Rejected');
    });
  });

  describe('isProposalActive', () => {
    it('should return true for voting period proposals', () => {
      const proposal = { status: GovernanceModule.PROPOSAL_STATUS.VOTING_PERIOD };
      expect(GovernanceModule.isProposalActive(proposal)).toBe(true);
    });

    it('should return false for non-voting period proposals', () => {
      const proposal = { status: GovernanceModule.PROPOSAL_STATUS.PASSED };
      expect(GovernanceModule.isProposalActive(proposal)).toBe(false);
    });
  });

  describe('isProposalInDepositPeriod', () => {
    it('should return true for deposit period proposals', () => {
      const proposal = { status: GovernanceModule.PROPOSAL_STATUS.DEPOSIT_PERIOD };
      expect(GovernanceModule.isProposalInDepositPeriod(proposal)).toBe(true);
    });

    it('should return false for other statuses', () => {
      const proposal = { status: GovernanceModule.PROPOSAL_STATUS.VOTING_PERIOD };
      expect(GovernanceModule.isProposalInDepositPeriod(proposal)).toBe(false);
    });
  });

  describe('formatProposal', () => {
    it('should format proposal correctly', () => {
      const proposal = {
        proposal_id: '1',
        status: 2,
        content: {
          title: 'Test Proposal',
          description: 'Test Description',
        },
        submit_time: '2024-01-01T00:00:00Z',
        deposit_end_time: '2024-01-08T00:00:00Z',
        voting_start_time: '2024-01-08T00:00:00Z',
        voting_end_time: '2024-01-15T00:00:00Z',
        total_deposit: [{ denom: 'uaura', amount: '5000000' }],
      };

      const formatted = GovernanceModule.formatProposal(proposal);

      expect(formatted.id).toBe('1');
      expect(formatted.title).toBe('Test Proposal');
      expect(formatted.description).toBe('Test Description');
      expect(formatted.status).toBe('Voting Period');
      expect(formatted.statusCode).toBe(2);
      expect(formatted.isActive).toBe(true);
      expect(formatted.submitTime).toBeInstanceOf(Date);
    });

    it('should handle missing fields gracefully', () => {
      const proposal = {
        proposal_id: '1',
        status: 2,
      };

      const formatted = GovernanceModule.formatProposal(proposal);

      expect(formatted.title).toBe('Untitled');
      expect(formatted.description).toBe('No description');
    });
  });
});

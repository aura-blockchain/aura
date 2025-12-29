import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { ApiService } from '../services/api';
import { KeystoreService } from '../services/keystore';

const STAKING_DENOM = 'uaura';
const DISPLAY_DENOM = 'AURA';
const MICRO_MULTIPLIER = 1000000;

const VOTE_OPTIONS = {
  YES: { value: 'yes', label: 'Yes', color: 'var(--success)', description: 'Vote in favor of the proposal' },
  NO: { value: 'no', label: 'No', color: 'var(--error)', description: 'Vote against the proposal' },
  ABSTAIN: { value: 'abstain', label: 'Abstain', color: 'var(--warning)', description: 'Abstain from voting but count towards quorum' },
  NO_WITH_VETO: { value: 'no_with_veto', label: 'No With Veto', color: '#e55770', description: 'Strong opposition - if >33% veto, proposal fails and deposit is burned' }
};

const PROPOSAL_STATUS = {
  PROPOSAL_STATUS_DEPOSIT_PERIOD: { label: 'Deposit Period', class: 'status-pending', filter: '1' },
  PROPOSAL_STATUS_VOTING_PERIOD: { label: 'Voting', class: 'status-success', filter: '2' },
  PROPOSAL_STATUS_PASSED: { label: 'Passed', class: 'status-success', filter: '3' },
  PROPOSAL_STATUS_REJECTED: { label: 'Rejected', class: 'status-failed', filter: '4' },
  PROPOSAL_STATUS_FAILED: { label: 'Failed', class: 'status-failed', filter: '5' }
};

const Governance = ({ walletData, onSuccess }) => {
  // Data state
  const [proposals, setProposals] = useState([]);
  const [balance, setBalance] = useState('0');

  // UI state
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('voting');
  const [selectedProposal, setSelectedProposal] = useState(null);

  // Modal state
  const [modalType, setModalType] = useState(null);
  const [selectedVote, setSelectedVote] = useState(null);
  const [depositAmount, setDepositAmount] = useState('');
  const [password, setPassword] = useState('');
  const [txLoading, setTxLoading] = useState(false);
  const [txError, setTxError] = useState('');
  const [txSuccess, setTxSuccess] = useState('');

  const apiService = new ApiService();
  const keystoreService = new KeystoreService();

  useEffect(() => {
    if (walletData?.address) {
      fetchAllData();
    }
  }, [walletData]);

  const fetchAllData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [proposalsData, balanceData] = await Promise.all([
        apiService.getProposals().catch(() => []),
        apiService.getBalance(walletData.address).catch(() => ({ balances: [] }))
      ]);

      setProposals(proposalsData);

      const auraBalance = balanceData.balances?.find(b => b.denom === STAKING_DENOM);
      setBalance(auraBalance?.amount || '0');
    } catch (err) {
      console.error('Failed to fetch governance data:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatAmount = useCallback((amount, decimals = 6) => {
    if (!amount) return '0';
    const value = parseInt(amount) / MICRO_MULTIPLIER;
    return value.toLocaleString('en-US', {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    });
  }, []);

  const formatDate = useCallback((dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }, []);

  const getStatusInfo = useCallback((status) => {
    return PROPOSAL_STATUS[status] || { label: status, class: 'status-pending' };
  }, []);

  const getProposalContent = useCallback((proposal) => {
    if (proposal.content) {
      return {
        title: proposal.content.title || proposal.content['@type']?.split('.').pop() || 'Untitled',
        description: proposal.content.description || 'No description available'
      };
    }
    if (proposal.messages && proposal.messages.length > 0) {
      const msg = proposal.messages[0];
      return {
        title: msg.content?.title || proposal.title || 'Untitled',
        description: msg.content?.description || proposal.summary || 'No description available'
      };
    }
    return {
      title: proposal.title || 'Untitled',
      description: proposal.summary || proposal.description || 'No description available'
    };
  }, []);

  const calculateVotePercentages = useCallback((tally) => {
    if (!tally) return { yes: 0, no: 0, abstain: 0, noWithVeto: 0, total: 0 };

    const yes = parseInt(tally.yes || tally.yes_count || '0');
    const no = parseInt(tally.no || tally.no_count || '0');
    const abstain = parseInt(tally.abstain || tally.abstain_count || '0');
    const noWithVeto = parseInt(tally.no_with_veto || tally.no_with_veto_count || '0');
    const total = yes + no + abstain + noWithVeto;

    if (total === 0) return { yes: 0, no: 0, abstain: 0, noWithVeto: 0, total: 0 };

    return {
      yes: (yes / total) * 100,
      no: (no / total) * 100,
      abstain: (abstain / total) * 100,
      noWithVeto: (noWithVeto / total) * 100,
      total,
      raw: { yes, no, abstain, noWithVeto }
    };
  }, []);

  const getDepositProgress = useCallback((proposal) => {
    const totalDeposit = (proposal.total_deposit || []).reduce((sum, coin) => {
      if (coin.denom === STAKING_DENOM) {
        return sum + parseInt(coin.amount || '0');
      }
      return sum;
    }, 0);

    const minDeposit = 10000000000;
    const progress = Math.min((totalDeposit / minDeposit) * 100, 100);

    return {
      current: totalDeposit,
      required: minDeposit,
      progress
    };
  }, []);

  const isVotingActive = useCallback((proposal) => {
    return proposal.status === 'PROPOSAL_STATUS_VOTING_PERIOD';
  }, []);

  const isDepositPeriod = useCallback((proposal) => {
    return proposal.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD';
  }, []);

  const getTimeRemaining = useCallback((endTime) => {
    if (!endTime) return 'N/A';
    const end = new Date(endTime);
    const now = new Date();
    const diff = end - now;

    if (diff <= 0) return 'Ended';

    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (days > 0) return `${days}d ${hours}h remaining`;
    if (hours > 0) return `${hours}h ${minutes}m remaining`;
    return `${minutes}m remaining`;
  }, []);

  const filteredProposals = useMemo(() => {
    let filtered = [...proposals];

    switch (activeTab) {
      case 'voting':
        filtered = filtered.filter(p => p.status === 'PROPOSAL_STATUS_VOTING_PERIOD');
        break;
      case 'deposit':
        filtered = filtered.filter(p => p.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD');
        break;
      case 'passed':
        filtered = filtered.filter(p => p.status === 'PROPOSAL_STATUS_PASSED');
        break;
      case 'rejected':
        filtered = filtered.filter(p =>
          p.status === 'PROPOSAL_STATUS_REJECTED' ||
          p.status === 'PROPOSAL_STATUS_FAILED'
        );
        break;
      default:
        break;
    }

    filtered.sort((a, b) => {
      const aId = parseInt(a.proposal_id || a.id || '0');
      const bId = parseInt(b.proposal_id || b.id || '0');
      return bId - aId;
    });

    return filtered;
  }, [proposals, activeTab]);

  const openModal = (type, proposal = null) => {
    setModalType(type);
    setSelectedProposal(proposal);
    setSelectedVote(null);
    setDepositAmount('');
    setPassword('');
    setTxError('');
    setTxSuccess('');
  };

  const closeModal = () => {
    setModalType(null);
    setSelectedProposal(null);
    setSelectedVote(null);
    setDepositAmount('');
    setPassword('');
    setTxError('');
    setTxSuccess('');
  };

  const handleVote = async () => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!selectedVote) {
        throw new Error('Please select a vote option');
      }
      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      const proposalId = selectedProposal.proposal_id || selectedProposal.id;

      const result = await apiService.vote(
        walletData.address,
        proposalId,
        selectedVote,
        '',
        wallet.privateKey
      );

      setTxSuccess(`Vote submitted successfully! Hash: ${result.transactionHash || result.txhash}`);
      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Vote failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const handleDeposit = async () => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!depositAmount || parseFloat(depositAmount) <= 0) {
        throw new Error('Deposit amount must be greater than 0');
      }
      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      const amountInMicro = Math.floor(parseFloat(depositAmount) * MICRO_MULTIPLIER);
      if (amountInMicro > parseInt(balance)) {
        throw new Error('Insufficient balance');
      }

      const proposalId = selectedProposal.proposal_id || selectedProposal.id;

      const result = await apiService.deposit(
        walletData.address,
        proposalId,
        amountInMicro,
        STAKING_DENOM,
        '',
        wallet.privateKey
      );

      setTxSuccess(`Deposit submitted successfully! Hash: ${result.transactionHash || result.txhash}`);
      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Deposit failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const renderVoteBar = (percentages) => {
    if (!percentages || percentages.total === 0) {
      return (
        <div style={{ height: '8px', background: 'var(--bg-tertiary)', borderRadius: '4px' }}>
          <div className="text-muted" style={{ fontSize: '11px', marginTop: '5px' }}>No votes yet</div>
        </div>
      );
    }

    return (
      <div>
        <div style={{ display: 'flex', height: '8px', borderRadius: '4px', overflow: 'hidden' }}>
          {percentages.yes > 0 && (
            <div style={{ width: `${percentages.yes}%`, background: 'var(--success)' }} title={`Yes: ${percentages.yes.toFixed(1)}%`} />
          )}
          {percentages.abstain > 0 && (
            <div style={{ width: `${percentages.abstain}%`, background: 'var(--warning)' }} title={`Abstain: ${percentages.abstain.toFixed(1)}%`} />
          )}
          {percentages.no > 0 && (
            <div style={{ width: `${percentages.no}%`, background: 'var(--error)' }} title={`No: ${percentages.no.toFixed(1)}%`} />
          )}
          {percentages.noWithVeto > 0 && (
            <div style={{ width: `${percentages.noWithVeto}%`, background: '#e55770' }} title={`No With Veto: ${percentages.noWithVeto.toFixed(1)}%`} />
          )}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '5px', fontSize: '11px' }}>
          <span style={{ color: 'var(--success)' }}>Yes: {percentages.yes.toFixed(1)}%</span>
          <span style={{ color: 'var(--warning)' }}>Abstain: {percentages.abstain.toFixed(1)}%</span>
          <span style={{ color: 'var(--error)' }}>No: {percentages.no.toFixed(1)}%</span>
          <span style={{ color: '#e55770' }}>Veto: {percentages.noWithVeto.toFixed(1)}%</span>
        </div>
      </div>
    );
  };

  const renderProposalCard = (proposal) => {
    const content = getProposalContent(proposal);
    const statusInfo = getStatusInfo(proposal.status);
    const percentages = calculateVotePercentages(proposal.final_tally_result || proposal.tally);
    const depositInfo = getDepositProgress(proposal);
    const proposalId = proposal.proposal_id || proposal.id;

    return (
      <div key={proposalId} className="card" style={{ marginBottom: '15px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '15px' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '5px' }}>
              <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>#{proposalId}</span>
              <span className={`status-badge ${statusInfo.class}`}>{statusInfo.label}</span>
            </div>
            <h4 style={{ marginBottom: '5px' }}>{content.title}</h4>
          </div>
          <div style={{ textAlign: 'right', minWidth: '120px' }}>
            {isVotingActive(proposal) && (
              <>
                <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>Voting ends</div>
                <div style={{ fontSize: '12px', color: 'var(--accent)' }}>
                  {getTimeRemaining(proposal.voting_end_time)}
                </div>
              </>
            )}
            {isDepositPeriod(proposal) && (
              <>
                <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>Deposit ends</div>
                <div style={{ fontSize: '12px', color: 'var(--warning)' }}>
                  {getTimeRemaining(proposal.deposit_end_time)}
                </div>
              </>
            )}
          </div>
        </div>

        <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '15px', lineHeight: '1.5' }}>
          {content.description.length > 300
            ? `${content.description.substring(0, 300)}...`
            : content.description}
        </p>

        {isDepositPeriod(proposal) && (
          <div style={{ marginBottom: '15px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '5px' }}>
              <span className="text-muted">Deposit Progress</span>
              <span>
                {formatAmount(depositInfo.current.toString())} / {formatAmount(depositInfo.required.toString())} {DISPLAY_DENOM}
              </span>
            </div>
            <div style={{ height: '6px', background: 'var(--bg-tertiary)', borderRadius: '3px', overflow: 'hidden' }}>
              <div style={{
                width: `${depositInfo.progress}%`,
                height: '100%',
                background: depositInfo.progress >= 100 ? 'var(--success)' : 'var(--warning)',
                transition: 'width 0.3s ease'
              }} />
            </div>
          </div>
        )}

        {(isVotingActive(proposal) || proposal.status === 'PROPOSAL_STATUS_PASSED' || proposal.status === 'PROPOSAL_STATUS_REJECTED') && (
          <div style={{ marginBottom: '15px' }}>
            {renderVoteBar(percentages)}
          </div>
        )}

        <div style={{ display: 'flex', gap: '10px' }}>
          <button
            className="btn btn-secondary"
            style={{ padding: '6px 12px', fontSize: '12px' }}
            onClick={() => openModal('details', proposal)}
          >
            View Details
          </button>

          {isVotingActive(proposal) && (
            <button
              className="btn btn-primary"
              style={{ padding: '6px 12px', fontSize: '12px' }}
              onClick={() => openModal('vote', proposal)}
            >
              Vote
            </button>
          )}

          {isDepositPeriod(proposal) && (
            <button
              className="btn btn-primary"
              style={{ padding: '6px 12px', fontSize: '12px' }}
              onClick={() => openModal('deposit', proposal)}
            >
              Deposit
            </button>
          )}
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="content text-center">
        <div className="loading-spinner"></div>
        <p className="text-muted">Loading governance data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="content">
        <div className="card">
          <div className="text-error text-center">
            <p>Failed to load governance data</p>
            <p className="text-muted mt-20">{error}</p>
            <button className="btn btn-primary mt-20" onClick={fetchAllData}>
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="content">
      {/* Governance Overview */}
      <div className="card">
        <div className="flex-between mb-20">
          <h3 className="card-header" style={{ marginBottom: 0 }}>Governance</h3>
          <button className="btn btn-secondary" onClick={fetchAllData}>
            Refresh
          </button>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '15px' }}>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px', textAlign: 'center' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Total Proposals</div>
            <div style={{ fontSize: '24px', fontWeight: '600', color: 'var(--text-primary)' }}>
              {proposals.length}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px', textAlign: 'center' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Active Voting</div>
            <div style={{ fontSize: '24px', fontWeight: '600', color: 'var(--success)' }}>
              {proposals.filter(p => p.status === 'PROPOSAL_STATUS_VOTING_PERIOD').length}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px', textAlign: 'center' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>In Deposit</div>
            <div style={{ fontSize: '24px', fontWeight: '600', color: 'var(--warning)' }}>
              {proposals.filter(p => p.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD').length}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px', textAlign: 'center' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Your Balance</div>
            <div style={{ fontSize: '18px', fontWeight: '600', color: 'var(--accent)' }}>
              {formatAmount(balance, 2)} {DISPLAY_DENOM}
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="card" style={{ marginBottom: 0 }}>
        <div style={{ display: 'flex', gap: '10px', marginBottom: '20px', borderBottom: '1px solid var(--border)', paddingBottom: '15px', flexWrap: 'wrap' }}>
          <button
            className={`btn ${activeTab === 'voting' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('voting')}
          >
            Active Voting ({proposals.filter(p => p.status === 'PROPOSAL_STATUS_VOTING_PERIOD').length})
          </button>
          <button
            className={`btn ${activeTab === 'deposit' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('deposit')}
          >
            Deposit Period ({proposals.filter(p => p.status === 'PROPOSAL_STATUS_DEPOSIT_PERIOD').length})
          </button>
          <button
            className={`btn ${activeTab === 'passed' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('passed')}
          >
            Passed
          </button>
          <button
            className={`btn ${activeTab === 'rejected' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('rejected')}
          >
            Rejected
          </button>
          <button
            className={`btn ${activeTab === 'all' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('all')}
          >
            All
          </button>
        </div>

        {/* Proposals List */}
        {filteredProposals.length > 0 ? (
          <div>
            {filteredProposals.map(renderProposalCard)}
          </div>
        ) : (
          <div className="text-center text-muted" style={{ padding: '40px' }}>
            <p>No proposals found</p>
            <p style={{ fontSize: '12px', marginTop: '10px' }}>
              {activeTab === 'voting' && 'There are no proposals currently in the voting period'}
              {activeTab === 'deposit' && 'There are no proposals currently in the deposit period'}
              {activeTab === 'passed' && 'No proposals have passed yet'}
              {activeTab === 'rejected' && 'No proposals have been rejected'}
              {activeTab === 'all' && 'No governance proposals have been submitted'}
            </p>
          </div>
        )}
      </div>

      {/* Modals */}
      {modalType && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '600px' }}>

            {/* Vote Modal */}
            {modalType === 'vote' && selectedProposal && (
              <>
                <h3 className="modal-header">Vote on Proposal #{selectedProposal.proposal_id || selectedProposal.id}</h3>

                <div style={{ marginBottom: '20px', padding: '15px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                  <div style={{ fontWeight: '500', marginBottom: '10px' }}>
                    {getProposalContent(selectedProposal).title}
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
                    <span className="text-muted">Voting ends:</span>
                    <span style={{ color: 'var(--accent)' }}>
                      {getTimeRemaining(selectedProposal.voting_end_time)}
                    </span>
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">Select your vote</label>
                  <div style={{ display: 'grid', gap: '10px' }}>
                    {Object.entries(VOTE_OPTIONS).map(([key, option]) => (
                      <div
                        key={key}
                        onClick={() => setSelectedVote(option.value)}
                        style={{
                          padding: '15px',
                          background: selectedVote === option.value ? 'rgba(122, 162, 247, 0.15)' : 'var(--bg-primary)',
                          border: `2px solid ${selectedVote === option.value ? option.color : 'var(--border)'}`,
                          borderRadius: '8px',
                          cursor: 'pointer',
                          transition: 'all 0.2s'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                          <div style={{
                            width: '16px',
                            height: '16px',
                            borderRadius: '50%',
                            border: `2px solid ${option.color}`,
                            background: selectedVote === option.value ? option.color : 'transparent'
                          }} />
                          <div>
                            <div style={{ fontWeight: '500', color: option.color }}>{option.label}</div>
                            <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{option.description}</div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">Password</label>
                  <input
                    type="password"
                    className="form-input"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>

                {txError && <div className="text-error mb-20">{txError}</div>}
                {txSuccess && <div className="text-success mb-20">{txSuccess}</div>}

                <div className="modal-footer">
                  <button className="btn btn-secondary" onClick={closeModal} disabled={txLoading}>
                    Cancel
                  </button>
                  <button className="btn btn-primary" onClick={handleVote} disabled={txLoading || !selectedVote}>
                    {txLoading ? 'Submitting...' : 'Submit Vote'}
                  </button>
                </div>
              </>
            )}

            {/* Deposit Modal */}
            {modalType === 'deposit' && selectedProposal && (
              <>
                <h3 className="modal-header">Deposit to Proposal #{selectedProposal.proposal_id || selectedProposal.id}</h3>

                <div style={{ marginBottom: '20px', padding: '15px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                  <div style={{ fontWeight: '500', marginBottom: '10px' }}>
                    {getProposalContent(selectedProposal).title}
                  </div>
                  {(() => {
                    const depositInfo = getDepositProgress(selectedProposal);
                    return (
                      <>
                        <div style={{ marginBottom: '10px' }}>
                          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '5px' }}>
                            <span className="text-muted">Current Deposit</span>
                            <span>
                              {formatAmount(depositInfo.current.toString())} / {formatAmount(depositInfo.required.toString())} {DISPLAY_DENOM}
                            </span>
                          </div>
                          <div style={{ height: '6px', background: 'var(--bg-tertiary)', borderRadius: '3px', overflow: 'hidden' }}>
                            <div style={{
                              width: `${depositInfo.progress}%`,
                              height: '100%',
                              background: depositInfo.progress >= 100 ? 'var(--success)' : 'var(--warning)'
                            }} />
                          </div>
                        </div>
                        <div style={{ fontSize: '12px', color: 'var(--warning)' }}>
                          Deposit ends: {getTimeRemaining(selectedProposal.deposit_end_time)}
                        </div>
                      </>
                    );
                  })()}
                </div>

                <div style={{ marginBottom: '15px', padding: '10px', background: 'rgba(224, 175, 104, 0.1)', borderRadius: '6px', border: '1px solid var(--warning)' }}>
                  <div style={{ color: 'var(--warning)', fontSize: '12px' }}>
                    Warning: Deposits are returned if the proposal passes or is rejected. However, if the proposal fails to reach quorum or is vetoed, your deposit may be burned.
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">
                    Deposit Amount ({DISPLAY_DENOM})
                    <span className="text-muted" style={{ fontSize: '11px', marginLeft: '10px' }}>
                      Available: {formatAmount(balance)}
                    </span>
                  </label>
                  <input
                    type="number"
                    className="form-input"
                    placeholder="Enter amount"
                    value={depositAmount}
                    onChange={(e) => setDepositAmount(e.target.value)}
                    min="0"
                    step="0.000001"
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Password</label>
                  <input
                    type="password"
                    className="form-input"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>

                {txError && <div className="text-error mb-20">{txError}</div>}
                {txSuccess && <div className="text-success mb-20">{txSuccess}</div>}

                <div className="modal-footer">
                  <button className="btn btn-secondary" onClick={closeModal} disabled={txLoading}>
                    Cancel
                  </button>
                  <button className="btn btn-primary" onClick={handleDeposit} disabled={txLoading}>
                    {txLoading ? 'Submitting...' : 'Submit Deposit'}
                  </button>
                </div>
              </>
            )}

            {/* Details Modal */}
            {modalType === 'details' && selectedProposal && (
              <>
                <h3 className="modal-header">
                  Proposal #{selectedProposal.proposal_id || selectedProposal.id}
                </h3>

                {(() => {
                  const content = getProposalContent(selectedProposal);
                  const statusInfo = getStatusInfo(selectedProposal.status);
                  const percentages = calculateVotePercentages(selectedProposal.final_tally_result || selectedProposal.tally);
                  const depositInfo = getDepositProgress(selectedProposal);

                  return (
                    <div>
                      <div style={{ marginBottom: '20px' }}>
                        <span className={`status-badge ${statusInfo.class}`}>{statusInfo.label}</span>
                      </div>

                      <h4 style={{ marginBottom: '15px' }}>{content.title}</h4>

                      <div style={{
                        maxHeight: '200px',
                        overflowY: 'auto',
                        padding: '15px',
                        background: 'var(--bg-primary)',
                        borderRadius: '6px',
                        marginBottom: '20px',
                        fontSize: '13px',
                        lineHeight: '1.6',
                        whiteSpace: 'pre-wrap'
                      }}>
                        {content.description}
                      </div>

                      <div style={{ display: 'grid', gap: '10px', marginBottom: '20px' }}>
                        <div className="flex-between" style={{ fontSize: '13px' }}>
                          <span className="text-muted">Submit Time</span>
                          <span>{formatDate(selectedProposal.submit_time)}</span>
                        </div>
                        <div className="flex-between" style={{ fontSize: '13px' }}>
                          <span className="text-muted">Deposit End Time</span>
                          <span>{formatDate(selectedProposal.deposit_end_time)}</span>
                        </div>
                        {selectedProposal.voting_start_time && (
                          <div className="flex-between" style={{ fontSize: '13px' }}>
                            <span className="text-muted">Voting Start Time</span>
                            <span>{formatDate(selectedProposal.voting_start_time)}</span>
                          </div>
                        )}
                        {selectedProposal.voting_end_time && (
                          <div className="flex-between" style={{ fontSize: '13px' }}>
                            <span className="text-muted">Voting End Time</span>
                            <span>{formatDate(selectedProposal.voting_end_time)}</span>
                          </div>
                        )}
                      </div>

                      {isDepositPeriod(selectedProposal) && (
                        <div style={{ marginBottom: '20px' }}>
                          <div style={{ fontSize: '13px', marginBottom: '10px' }}>Deposit Progress</div>
                          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '5px' }}>
                            <span className="text-muted">Current</span>
                            <span>
                              {formatAmount(depositInfo.current.toString())} / {formatAmount(depositInfo.required.toString())} {DISPLAY_DENOM}
                            </span>
                          </div>
                          <div style={{ height: '8px', background: 'var(--bg-tertiary)', borderRadius: '4px', overflow: 'hidden' }}>
                            <div style={{
                              width: `${depositInfo.progress}%`,
                              height: '100%',
                              background: depositInfo.progress >= 100 ? 'var(--success)' : 'var(--warning)'
                            }} />
                          </div>
                        </div>
                      )}

                      {(isVotingActive(selectedProposal) || selectedProposal.status === 'PROPOSAL_STATUS_PASSED' || selectedProposal.status === 'PROPOSAL_STATUS_REJECTED') && (
                        <div style={{ marginBottom: '20px' }}>
                          <div style={{ fontSize: '13px', marginBottom: '10px' }}>Vote Results</div>
                          {renderVoteBar(percentages)}

                          {percentages.raw && (
                            <div style={{ marginTop: '15px', display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '10px' }}>
                              <div style={{ padding: '10px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                                <div style={{ fontSize: '11px', color: 'var(--success)' }}>Yes</div>
                                <div style={{ fontWeight: '500' }}>{formatAmount(percentages.raw.yes.toString(), 0)}</div>
                              </div>
                              <div style={{ padding: '10px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                                <div style={{ fontSize: '11px', color: 'var(--error)' }}>No</div>
                                <div style={{ fontWeight: '500' }}>{formatAmount(percentages.raw.no.toString(), 0)}</div>
                              </div>
                              <div style={{ padding: '10px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                                <div style={{ fontSize: '11px', color: 'var(--warning)' }}>Abstain</div>
                                <div style={{ fontWeight: '500' }}>{formatAmount(percentages.raw.abstain.toString(), 0)}</div>
                              </div>
                              <div style={{ padding: '10px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                                <div style={{ fontSize: '11px', color: '#e55770' }}>No With Veto</div>
                                <div style={{ fontWeight: '500' }}>{formatAmount(percentages.raw.noWithVeto.toString(), 0)}</div>
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })()}

                <div className="modal-footer">
                  <button className="btn btn-secondary" onClick={closeModal}>
                    Close
                  </button>
                  {isVotingActive(selectedProposal) && (
                    <button className="btn btn-primary" onClick={() => {
                      closeModal();
                      setTimeout(() => openModal('vote', selectedProposal), 100);
                    }}>
                      Vote
                    </button>
                  )}
                  {isDepositPeriod(selectedProposal) && (
                    <button className="btn btn-primary" onClick={() => {
                      closeModal();
                      setTimeout(() => openModal('deposit', selectedProposal), 100);
                    }}>
                      Deposit
                    </button>
                  )}
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Governance;

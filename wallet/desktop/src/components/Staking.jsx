import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { ApiService } from '../services/api';
import { KeystoreService } from '../services/keystore';

const STAKING_DENOM = 'uaura';
const DISPLAY_DENOM = 'AURA';
const MICRO_MULTIPLIER = 1000000;
const UNBONDING_DAYS = 21;

const Staking = ({ walletData, onSuccess }) => {
  // Data state
  const [validators, setValidators] = useState([]);
  const [delegations, setDelegations] = useState([]);
  const [unbondingDelegations, setUnbondingDelegations] = useState([]);
  const [rewards, setRewards] = useState(null);
  const [balance, setBalance] = useState('0');

  // UI state
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('validators');
  const [sortField, setSortField] = useState('tokens');
  const [sortDirection, setSortDirection] = useState('desc');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('bonded');

  // Modal state
  const [modalType, setModalType] = useState(null);
  const [selectedValidator, setSelectedValidator] = useState(null);
  const [delegateAmount, setDelegateAmount] = useState('');
  const [redelegateValidator, setRedelegateValidator] = useState('');
  const [password, setPassword] = useState('');
  const [txLoading, setTxLoading] = useState(false);
  const [txError, setTxError] = useState('');
  const [txSuccess, setTxSuccess] = useState('');

  // APY Calculator state
  const [apyAmount, setApyAmount] = useState('');
  const [apyDuration, setApyDuration] = useState('365');

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

      const [validatorsData, delegationsData, unbondingData, rewardsData, balanceData] = await Promise.all([
        apiService.getValidators().catch(() => []),
        apiService.getDelegations(walletData.address).catch(() => []),
        apiService.getUnbondingDelegations(walletData.address).catch(() => []),
        apiService.getRewards(walletData.address).catch(() => null),
        apiService.getBalance(walletData.address).catch(() => ({ balances: [] }))
      ]);

      setValidators(validatorsData);
      setDelegations(delegationsData);
      setUnbondingDelegations(unbondingData);
      setRewards(rewardsData);

      const auraBalance = balanceData.balances?.find(b => b.denom === STAKING_DENOM);
      setBalance(auraBalance?.amount || '0');
    } catch (err) {
      console.error('Failed to fetch staking data:', err);
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

  const formatPercent = useCallback((value, decimals = 2) => {
    if (!value) return '0%';
    const percent = parseFloat(value) * 100;
    return `${percent.toFixed(decimals)}%`;
  }, []);

  const formatCommission = useCallback((commission) => {
    if (!commission?.commission_rates?.rate) return 'N/A';
    return formatPercent(commission.commission_rates.rate);
  }, [formatPercent]);

  const getValidatorStatus = useCallback((validator) => {
    if (validator.jailed) return 'jailed';
    if (validator.status === 'BOND_STATUS_BONDED') return 'bonded';
    if (validator.status === 'BOND_STATUS_UNBONDING') return 'unbonding';
    return 'unbonded';
  }, []);

  const getStatusClass = useCallback((status) => {
    switch (status) {
      case 'bonded': return 'status-success';
      case 'unbonding': return 'status-pending';
      case 'jailed': return 'status-failed';
      default: return 'status-pending';
    }
  }, []);

  const getDelegationForValidator = useCallback((validatorAddress) => {
    const delegation = delegations.find(d =>
      d.delegation?.validator_address === validatorAddress
    );
    return delegation?.balance?.amount || '0';
  }, [delegations]);

  const getRewardsForValidator = useCallback((validatorAddress) => {
    if (!rewards?.rewards) return '0';
    const validatorReward = rewards.rewards.find(r =>
      r.validator_address === validatorAddress
    );
    if (!validatorReward?.reward) return '0';
    const auraReward = validatorReward.reward.find(r => r.denom === STAKING_DENOM);
    return auraReward?.amount || '0';
  }, [rewards]);

  const totalDelegated = useMemo(() => {
    return delegations.reduce((sum, d) => {
      return sum + parseInt(d.balance?.amount || '0');
    }, 0);
  }, [delegations]);

  const totalRewards = useMemo(() => {
    if (!rewards?.total) return 0;
    const auraReward = rewards.total.find(r => r.denom === STAKING_DENOM);
    return parseInt(auraReward?.amount || '0');
  }, [rewards]);

  const totalUnbonding = useMemo(() => {
    return unbondingDelegations.reduce((sum, u) => {
      const entrySum = (u.entries || []).reduce((s, e) => s + parseInt(e.balance || '0'), 0);
      return sum + entrySum;
    }, 0);
  }, [unbondingDelegations]);

  const estimatedAPY = useMemo(() => {
    if (validators.length === 0) return 0;
    const bondedValidators = validators.filter(v => getValidatorStatus(v) === 'bonded');
    if (bondedValidators.length === 0) return 0;

    const avgCommission = bondedValidators.reduce((sum, v) => {
      const rate = parseFloat(v.commission?.commission_rates?.rate || '0');
      return sum + rate;
    }, 0) / bondedValidators.length;

    const baseAPY = 0.15;
    return baseAPY * (1 - avgCommission);
  }, [validators, getValidatorStatus]);

  const sortedValidators = useMemo(() => {
    let filtered = [...validators];

    if (statusFilter !== 'all') {
      filtered = filtered.filter(v => getValidatorStatus(v) === statusFilter);
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(v =>
        v.description?.moniker?.toLowerCase().includes(query) ||
        v.operator_address?.toLowerCase().includes(query)
      );
    }

    filtered.sort((a, b) => {
      let aVal, bVal;
      switch (sortField) {
        case 'moniker':
          aVal = a.description?.moniker?.toLowerCase() || '';
          bVal = b.description?.moniker?.toLowerCase() || '';
          break;
        case 'tokens':
          aVal = parseInt(a.tokens || '0');
          bVal = parseInt(b.tokens || '0');
          break;
        case 'commission':
          aVal = parseFloat(a.commission?.commission_rates?.rate || '0');
          bVal = parseFloat(b.commission?.commission_rates?.rate || '0');
          break;
        case 'delegated':
          aVal = parseInt(getDelegationForValidator(a.operator_address));
          bVal = parseInt(getDelegationForValidator(b.operator_address));
          break;
        default:
          aVal = 0;
          bVal = 0;
      }

      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1;
      }
      return aVal < bVal ? 1 : -1;
    });

    return filtered;
  }, [validators, statusFilter, searchQuery, sortField, sortDirection, getValidatorStatus, getDelegationForValidator]);

  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const openModal = (type, validator = null) => {
    setModalType(type);
    setSelectedValidator(validator);
    setDelegateAmount('');
    setRedelegateValidator('');
    setPassword('');
    setTxError('');
    setTxSuccess('');
  };

  const closeModal = () => {
    setModalType(null);
    setSelectedValidator(null);
    setDelegateAmount('');
    setRedelegateValidator('');
    setPassword('');
    setTxError('');
    setTxSuccess('');
  };

  const handleDelegate = async () => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!delegateAmount || parseFloat(delegateAmount) <= 0) {
        throw new Error('Amount must be greater than 0');
      }
      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      const amountInMicro = Math.floor(parseFloat(delegateAmount) * MICRO_MULTIPLIER);
      if (amountInMicro > parseInt(balance)) {
        throw new Error('Insufficient balance');
      }

      const result = await apiService.delegate(
        walletData.address,
        selectedValidator.operator_address,
        amountInMicro,
        STAKING_DENOM,
        '',
        wallet.privateKey
      );

      setTxSuccess(`Delegation successful! Hash: ${result.transactionHash || result.txhash}`);
      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Delegation failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const handleUndelegate = async () => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!delegateAmount || parseFloat(delegateAmount) <= 0) {
        throw new Error('Amount must be greater than 0');
      }
      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      const amountInMicro = Math.floor(parseFloat(delegateAmount) * MICRO_MULTIPLIER);
      const currentDelegation = parseInt(getDelegationForValidator(selectedValidator.operator_address));
      if (amountInMicro > currentDelegation) {
        throw new Error('Amount exceeds current delegation');
      }

      const result = await apiService.undelegate(
        walletData.address,
        selectedValidator.operator_address,
        amountInMicro,
        STAKING_DENOM,
        '',
        wallet.privateKey
      );

      setTxSuccess(`Undelegation initiated! Tokens will be available in ${UNBONDING_DAYS} days. Hash: ${result.transactionHash || result.txhash}`);
      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Undelegation failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const handleRedelegate = async () => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!delegateAmount || parseFloat(delegateAmount) <= 0) {
        throw new Error('Amount must be greater than 0');
      }
      if (!redelegateValidator) {
        throw new Error('Please select a destination validator');
      }
      if (redelegateValidator === selectedValidator.operator_address) {
        throw new Error('Source and destination validators must be different');
      }
      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      const amountInMicro = Math.floor(parseFloat(delegateAmount) * MICRO_MULTIPLIER);
      const currentDelegation = parseInt(getDelegationForValidator(selectedValidator.operator_address));
      if (amountInMicro > currentDelegation) {
        throw new Error('Amount exceeds current delegation');
      }

      const result = await apiService.redelegate(
        walletData.address,
        selectedValidator.operator_address,
        redelegateValidator,
        amountInMicro,
        STAKING_DENOM,
        '',
        wallet.privateKey
      );

      setTxSuccess(`Redelegation successful! Hash: ${result.transactionHash || result.txhash}`);
      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Redelegation failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const handleClaimRewards = async (validatorAddress = null) => {
    try {
      setTxLoading(true);
      setTxError('');
      setTxSuccess('');

      if (!password) {
        throw new Error('Password is required');
      }

      const wallet = await keystoreService.unlockWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }

      if (validatorAddress) {
        const result = await apiService.withdrawRewards(
          walletData.address,
          validatorAddress,
          '',
          wallet.privateKey
        );
        setTxSuccess(`Rewards claimed! Hash: ${result.transactionHash || result.txhash}`);
      } else {
        const validatorsWithRewards = rewards?.rewards?.filter(r => {
          const auraReward = r.reward?.find(rw => rw.denom === STAKING_DENOM);
          return auraReward && parseFloat(auraReward.amount) > 0;
        }) || [];

        if (validatorsWithRewards.length === 0) {
          throw new Error('No rewards to claim');
        }

        for (const vr of validatorsWithRewards) {
          await apiService.withdrawRewards(
            walletData.address,
            vr.validator_address,
            '',
            wallet.privateKey
          );
        }
        setTxSuccess(`All rewards claimed from ${validatorsWithRewards.length} validators!`);
      }

      setTimeout(() => {
        closeModal();
        fetchAllData();
        if (onSuccess) onSuccess();
      }, 2000);
    } catch (err) {
      console.error('Claim rewards failed:', err);
      setTxError(err.message);
    } finally {
      setTxLoading(false);
    }
  };

  const calculateAPYReturns = useMemo(() => {
    if (!apyAmount || parseFloat(apyAmount) <= 0) return null;

    const principal = parseFloat(apyAmount);
    const days = parseInt(apyDuration) || 365;
    const dailyRate = estimatedAPY / 365;
    const simpleReturn = principal * estimatedAPY * (days / 365);
    const compoundReturn = principal * (Math.pow(1 + dailyRate, days) - 1);

    return {
      simple: simpleReturn,
      compound: compoundReturn,
      days
    };
  }, [apyAmount, apyDuration, estimatedAPY]);

  if (loading) {
    return (
      <div className="content text-center">
        <div className="loading-spinner"></div>
        <p className="text-muted">Loading staking data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="content">
        <div className="card">
          <div className="text-error text-center">
            <p>Failed to load staking data</p>
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
      {/* Staking Overview */}
      <div className="card">
        <div className="flex-between mb-20">
          <h3 className="card-header" style={{ marginBottom: 0 }}>Staking Overview</h3>
          <button className="btn btn-secondary" onClick={fetchAllData}>
            Refresh
          </button>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '20px' }}>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Available Balance</div>
            <div style={{ fontSize: '20px', fontWeight: '600', color: 'var(--text-primary)' }}>
              {formatAmount(balance)} {DISPLAY_DENOM}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Total Delegated</div>
            <div style={{ fontSize: '20px', fontWeight: '600', color: 'var(--accent)' }}>
              {formatAmount(totalDelegated.toString())} {DISPLAY_DENOM}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Unbonding</div>
            <div style={{ fontSize: '20px', fontWeight: '600', color: 'var(--warning)' }}>
              {formatAmount(totalUnbonding.toString())} {DISPLAY_DENOM}
            </div>
          </div>
          <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px' }}>
            <div className="text-muted" style={{ fontSize: '12px', marginBottom: '5px' }}>Pending Rewards</div>
            <div style={{ fontSize: '20px', fontWeight: '600', color: 'var(--success)' }}>
              {formatAmount(totalRewards.toString())} {DISPLAY_DENOM}
            </div>
            {totalRewards > 0 && (
              <button
                className="btn btn-primary"
                style={{ marginTop: '10px', padding: '6px 12px', fontSize: '12px' }}
                onClick={() => openModal('claimAll')}
              >
                Claim All Rewards
              </button>
            )}
          </div>
        </div>

        <div style={{ marginTop: '15px', padding: '10px 15px', background: 'rgba(122, 162, 247, 0.1)', borderRadius: '6px' }}>
          <span className="text-muted" style={{ fontSize: '12px' }}>Estimated APY: </span>
          <span style={{ fontSize: '14px', fontWeight: '600', color: 'var(--accent)' }}>
            {formatPercent(estimatedAPY)}
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="card" style={{ marginBottom: 0 }}>
        <div style={{ display: 'flex', gap: '10px', marginBottom: '20px', borderBottom: '1px solid var(--border)', paddingBottom: '15px' }}>
          <button
            className={`btn ${activeTab === 'validators' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('validators')}
          >
            Validators
          </button>
          <button
            className={`btn ${activeTab === 'delegations' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('delegations')}
          >
            My Delegations ({delegations.length})
          </button>
          <button
            className={`btn ${activeTab === 'unbonding' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('unbonding')}
          >
            Unbonding ({unbondingDelegations.length})
          </button>
          <button
            className={`btn ${activeTab === 'calculator' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('calculator')}
          >
            APY Calculator
          </button>
        </div>

        {/* Validators Tab */}
        {activeTab === 'validators' && (
          <>
            <div style={{ display: 'flex', gap: '15px', marginBottom: '20px', flexWrap: 'wrap' }}>
              <input
                type="text"
                className="form-input"
                placeholder="Search validators..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{ flex: '1', minWidth: '200px' }}
              />
              <select
                className="form-input"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                style={{ width: '150px' }}
              >
                <option value="all">All Status</option>
                <option value="bonded">Active</option>
                <option value="unbonding">Unbonding</option>
                <option value="unbonded">Inactive</option>
                <option value="jailed">Jailed</option>
              </select>
            </div>

            <div style={{ overflowX: 'auto' }}>
              <table className="table">
                <thead>
                  <tr>
                    <th onClick={() => handleSort('moniker')} style={{ cursor: 'pointer' }}>
                      Validator {sortField === 'moniker' && (sortDirection === 'asc' ? ' ^' : ' v')}
                    </th>
                    <th onClick={() => handleSort('tokens')} style={{ cursor: 'pointer' }}>
                      Voting Power {sortField === 'tokens' && (sortDirection === 'asc' ? ' ^' : ' v')}
                    </th>
                    <th onClick={() => handleSort('commission')} style={{ cursor: 'pointer' }}>
                      Commission {sortField === 'commission' && (sortDirection === 'asc' ? ' ^' : ' v')}
                    </th>
                    <th onClick={() => handleSort('delegated')} style={{ cursor: 'pointer' }}>
                      My Delegation {sortField === 'delegated' && (sortDirection === 'asc' ? ' ^' : ' v')}
                    </th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedValidators.map((validator) => {
                    const status = getValidatorStatus(validator);
                    const myDelegation = getDelegationForValidator(validator.operator_address);
                    return (
                      <tr key={validator.operator_address}>
                        <td>
                          <div style={{ fontWeight: '500' }}>{validator.description?.moniker || 'Unknown'}</div>
                          <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                            {validator.operator_address.slice(0, 20)}...
                          </div>
                        </td>
                        <td>{formatAmount(validator.tokens, 0)}</td>
                        <td>{formatCommission(validator.commission)}</td>
                        <td>
                          {parseInt(myDelegation) > 0 ? (
                            <span style={{ color: 'var(--accent)' }}>
                              {formatAmount(myDelegation)}
                            </span>
                          ) : (
                            <span className="text-muted">-</span>
                          )}
                        </td>
                        <td>
                          <span className={`status-badge ${getStatusClass(status)}`}>
                            {status}
                          </span>
                        </td>
                        <td>
                          <div style={{ display: 'flex', gap: '5px' }}>
                            <button
                              className="btn btn-primary"
                              style={{ padding: '4px 10px', fontSize: '12px' }}
                              onClick={() => openModal('delegate', validator)}
                              disabled={status === 'jailed'}
                            >
                              Delegate
                            </button>
                            {parseInt(myDelegation) > 0 && (
                              <>
                                <button
                                  className="btn btn-secondary"
                                  style={{ padding: '4px 10px', fontSize: '12px' }}
                                  onClick={() => openModal('undelegate', validator)}
                                >
                                  Undelegate
                                </button>
                                <button
                                  className="btn btn-secondary"
                                  style={{ padding: '4px 10px', fontSize: '12px' }}
                                  onClick={() => openModal('redelegate', validator)}
                                >
                                  Redelegate
                                </button>
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {sortedValidators.length === 0 && (
              <div className="text-center text-muted" style={{ padding: '40px' }}>
                No validators found matching your criteria
              </div>
            )}
          </>
        )}

        {/* My Delegations Tab */}
        {activeTab === 'delegations' && (
          <>
            {delegations.length > 0 ? (
              <div style={{ overflowX: 'auto' }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>Validator</th>
                      <th>Delegated Amount</th>
                      <th>Rewards</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {delegations.map((delegation) => {
                      const validator = validators.find(v =>
                        v.operator_address === delegation.delegation?.validator_address
                      );
                      const rewardAmount = getRewardsForValidator(delegation.delegation?.validator_address);
                      return (
                        <tr key={delegation.delegation?.validator_address}>
                          <td>
                            <div style={{ fontWeight: '500' }}>
                              {validator?.description?.moniker || 'Unknown Validator'}
                            </div>
                            <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                              {delegation.delegation?.validator_address?.slice(0, 20)}...
                            </div>
                          </td>
                          <td style={{ color: 'var(--accent)', fontWeight: '500' }}>
                            {formatAmount(delegation.balance?.amount)} {DISPLAY_DENOM}
                          </td>
                          <td style={{ color: 'var(--success)' }}>
                            {formatAmount(rewardAmount)} {DISPLAY_DENOM}
                          </td>
                          <td>
                            <div style={{ display: 'flex', gap: '5px' }}>
                              {parseFloat(rewardAmount) > 0 && (
                                <button
                                  className="btn btn-primary"
                                  style={{ padding: '4px 10px', fontSize: '12px' }}
                                  onClick={() => {
                                    setSelectedValidator(validator);
                                    openModal('claim', validator);
                                  }}
                                >
                                  Claim
                                </button>
                              )}
                              <button
                                className="btn btn-secondary"
                                style={{ padding: '4px 10px', fontSize: '12px' }}
                                onClick={() => openModal('delegate', validator)}
                              >
                                Delegate More
                              </button>
                              <button
                                className="btn btn-secondary"
                                style={{ padding: '4px 10px', fontSize: '12px' }}
                                onClick={() => openModal('undelegate', validator)}
                              >
                                Undelegate
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center text-muted" style={{ padding: '40px' }}>
                <p>No delegations found</p>
                <p style={{ fontSize: '12px', marginTop: '10px' }}>
                  Delegate your tokens to validators to earn staking rewards
                </p>
              </div>
            )}
          </>
        )}

        {/* Unbonding Tab */}
        {activeTab === 'unbonding' && (
          <>
            {unbondingDelegations.length > 0 ? (
              <div style={{ overflowX: 'auto' }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>Validator</th>
                      <th>Amount</th>
                      <th>Completion Time</th>
                    </tr>
                  </thead>
                  <tbody>
                    {unbondingDelegations.flatMap((unbonding) => {
                      const validator = validators.find(v =>
                        v.operator_address === unbonding.validator_address
                      );
                      return (unbonding.entries || []).map((entry, idx) => (
                        <tr key={`${unbonding.validator_address}-${idx}`}>
                          <td>
                            <div style={{ fontWeight: '500' }}>
                              {validator?.description?.moniker || 'Unknown Validator'}
                            </div>
                            <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                              {unbonding.validator_address?.slice(0, 20)}...
                            </div>
                          </td>
                          <td style={{ color: 'var(--warning)', fontWeight: '500' }}>
                            {formatAmount(entry.balance)} {DISPLAY_DENOM}
                          </td>
                          <td>
                            {new Date(entry.completion_time).toLocaleString('en-US', {
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </td>
                        </tr>
                      ));
                    })}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center text-muted" style={{ padding: '40px' }}>
                <p>No unbonding delegations</p>
                <p style={{ fontSize: '12px', marginTop: '10px' }}>
                  Unbonding tokens will appear here during the {UNBONDING_DAYS}-day unbonding period
                </p>
              </div>
            )}
          </>
        )}

        {/* APY Calculator Tab */}
        {activeTab === 'calculator' && (
          <div style={{ maxWidth: '500px' }}>
            <h4 style={{ marginBottom: '20px' }}>Staking Rewards Calculator</h4>

            <div className="form-group">
              <label className="form-label">Stake Amount ({DISPLAY_DENOM})</label>
              <input
                type="number"
                className="form-input"
                placeholder="Enter amount to stake"
                value={apyAmount}
                onChange={(e) => setApyAmount(e.target.value)}
                min="0"
                step="0.000001"
              />
            </div>

            <div className="form-group">
              <label className="form-label">Duration (Days)</label>
              <select
                className="form-input"
                value={apyDuration}
                onChange={(e) => setApyDuration(e.target.value)}
              >
                <option value="30">30 Days</option>
                <option value="90">90 Days</option>
                <option value="180">180 Days</option>
                <option value="365">1 Year</option>
                <option value="730">2 Years</option>
              </select>
            </div>

            <div style={{ padding: '15px', background: 'var(--bg-primary)', borderRadius: '8px', marginTop: '20px' }}>
              <div style={{ marginBottom: '15px' }}>
                <span className="text-muted" style={{ fontSize: '12px' }}>Current Est. APY</span>
                <div style={{ fontSize: '24px', fontWeight: '600', color: 'var(--accent)' }}>
                  {formatPercent(estimatedAPY)}
                </div>
              </div>

              {calculateAPYReturns && (
                <>
                  <div style={{ borderTop: '1px solid var(--border)', paddingTop: '15px', marginTop: '15px' }}>
                    <div style={{ marginBottom: '10px' }}>
                      <span className="text-muted" style={{ fontSize: '12px' }}>
                        Simple Interest ({calculateAPYReturns.days} days)
                      </span>
                      <div style={{ fontSize: '18px', fontWeight: '500', color: 'var(--success)' }}>
                        +{calculateAPYReturns.simple.toFixed(6)} {DISPLAY_DENOM}
                      </div>
                    </div>
                    <div>
                      <span className="text-muted" style={{ fontSize: '12px' }}>
                        Compound Interest (daily compounding)
                      </span>
                      <div style={{ fontSize: '18px', fontWeight: '500', color: 'var(--success)' }}>
                        +{calculateAPYReturns.compound.toFixed(6)} {DISPLAY_DENOM}
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>

            <p className="text-muted" style={{ fontSize: '11px', marginTop: '15px' }}>
              Note: APY is estimated based on current network conditions and validator commission rates.
              Actual returns may vary. Compound interest assumes daily claiming and restaking of rewards.
            </p>
          </div>
        )}
      </div>

      {/* Modals */}
      {modalType && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '500px' }}>
            {/* Delegate Modal */}
            {modalType === 'delegate' && (
              <>
                <h3 className="modal-header">Delegate to {selectedValidator?.description?.moniker}</h3>

                <div style={{ marginBottom: '20px', padding: '10px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                  <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Commission Rate</div>
                  <div style={{ fontWeight: '500' }}>{formatCommission(selectedValidator?.commission)}</div>
                </div>

                <div className="form-group">
                  <label className="form-label">
                    Amount ({DISPLAY_DENOM})
                    <span className="text-muted" style={{ fontSize: '11px', marginLeft: '10px' }}>
                      Available: {formatAmount(balance)}
                    </span>
                  </label>
                  <input
                    type="number"
                    className="form-input"
                    placeholder="Enter amount"
                    value={delegateAmount}
                    onChange={(e) => setDelegateAmount(e.target.value)}
                    min="0"
                    step="0.000001"
                  />
                  <button
                    style={{ marginTop: '5px', padding: '4px 8px', fontSize: '11px' }}
                    className="btn btn-secondary"
                    onClick={() => setDelegateAmount((parseInt(balance) / MICRO_MULTIPLIER * 0.95).toFixed(6))}
                  >
                    Max (95%)
                  </button>
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
                  <button className="btn btn-primary" onClick={handleDelegate} disabled={txLoading}>
                    {txLoading ? 'Delegating...' : 'Delegate'}
                  </button>
                </div>
              </>
            )}

            {/* Undelegate Modal */}
            {modalType === 'undelegate' && (
              <>
                <h3 className="modal-header">Undelegate from {selectedValidator?.description?.moniker}</h3>

                <div style={{ marginBottom: '20px', padding: '10px', background: 'rgba(224, 175, 104, 0.1)', borderRadius: '6px', border: '1px solid var(--warning)' }}>
                  <div style={{ color: 'var(--warning)', fontSize: '13px' }}>
                    Undelegation takes {UNBONDING_DAYS} days. Your tokens will not earn rewards during this period.
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">
                    Amount ({DISPLAY_DENOM})
                    <span className="text-muted" style={{ fontSize: '11px', marginLeft: '10px' }}>
                      Delegated: {formatAmount(getDelegationForValidator(selectedValidator?.operator_address))}
                    </span>
                  </label>
                  <input
                    type="number"
                    className="form-input"
                    placeholder="Enter amount"
                    value={delegateAmount}
                    onChange={(e) => setDelegateAmount(e.target.value)}
                    min="0"
                    step="0.000001"
                  />
                  <button
                    style={{ marginTop: '5px', padding: '4px 8px', fontSize: '11px' }}
                    className="btn btn-secondary"
                    onClick={() => setDelegateAmount(
                      (parseInt(getDelegationForValidator(selectedValidator?.operator_address)) / MICRO_MULTIPLIER).toFixed(6)
                    )}
                  >
                    Max
                  </button>
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
                  <button className="btn btn-danger" onClick={handleUndelegate} disabled={txLoading}>
                    {txLoading ? 'Undelegating...' : 'Undelegate'}
                  </button>
                </div>
              </>
            )}

            {/* Redelegate Modal */}
            {modalType === 'redelegate' && (
              <>
                <h3 className="modal-header">Redelegate from {selectedValidator?.description?.moniker}</h3>

                <div className="form-group">
                  <label className="form-label">Destination Validator</label>
                  <select
                    className="form-input"
                    value={redelegateValidator}
                    onChange={(e) => setRedelegateValidator(e.target.value)}
                  >
                    <option value="">Select a validator</option>
                    {validators
                      .filter(v =>
                        getValidatorStatus(v) === 'bonded' &&
                        v.operator_address !== selectedValidator?.operator_address
                      )
                      .map(v => (
                        <option key={v.operator_address} value={v.operator_address}>
                          {v.description?.moniker} ({formatCommission(v.commission)} commission)
                        </option>
                      ))
                    }
                  </select>
                </div>

                <div className="form-group">
                  <label className="form-label">
                    Amount ({DISPLAY_DENOM})
                    <span className="text-muted" style={{ fontSize: '11px', marginLeft: '10px' }}>
                      Delegated: {formatAmount(getDelegationForValidator(selectedValidator?.operator_address))}
                    </span>
                  </label>
                  <input
                    type="number"
                    className="form-input"
                    placeholder="Enter amount"
                    value={delegateAmount}
                    onChange={(e) => setDelegateAmount(e.target.value)}
                    min="0"
                    step="0.000001"
                  />
                  <button
                    style={{ marginTop: '5px', padding: '4px 8px', fontSize: '11px' }}
                    className="btn btn-secondary"
                    onClick={() => setDelegateAmount(
                      (parseInt(getDelegationForValidator(selectedValidator?.operator_address)) / MICRO_MULTIPLIER).toFixed(6)
                    )}
                  >
                    Max
                  </button>
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
                  <button className="btn btn-primary" onClick={handleRedelegate} disabled={txLoading}>
                    {txLoading ? 'Redelegating...' : 'Redelegate'}
                  </button>
                </div>
              </>
            )}

            {/* Claim Rewards Modal (single validator) */}
            {modalType === 'claim' && (
              <>
                <h3 className="modal-header">Claim Rewards</h3>

                <div style={{ marginBottom: '20px', padding: '15px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                  <div className="text-muted" style={{ fontSize: '12px' }}>Validator</div>
                  <div style={{ fontWeight: '500' }}>{selectedValidator?.description?.moniker}</div>
                  <div className="text-muted" style={{ fontSize: '12px', marginTop: '10px' }}>Pending Rewards</div>
                  <div style={{ fontSize: '20px', fontWeight: '600', color: 'var(--success)' }}>
                    {formatAmount(getRewardsForValidator(selectedValidator?.operator_address))} {DISPLAY_DENOM}
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
                  <button
                    className="btn btn-primary"
                    onClick={() => handleClaimRewards(selectedValidator?.operator_address)}
                    disabled={txLoading}
                  >
                    {txLoading ? 'Claiming...' : 'Claim Rewards'}
                  </button>
                </div>
              </>
            )}

            {/* Claim All Rewards Modal */}
            {modalType === 'claimAll' && (
              <>
                <h3 className="modal-header">Claim All Rewards</h3>

                <div style={{ marginBottom: '20px', padding: '15px', background: 'var(--bg-primary)', borderRadius: '6px' }}>
                  <div className="text-muted" style={{ fontSize: '12px' }}>Total Pending Rewards</div>
                  <div style={{ fontSize: '24px', fontWeight: '600', color: 'var(--success)' }}>
                    {formatAmount(totalRewards.toString())} {DISPLAY_DENOM}
                  </div>
                  <div className="text-muted" style={{ fontSize: '11px', marginTop: '10px' }}>
                    From {rewards?.rewards?.filter(r => {
                      const auraReward = r.reward?.find(rw => rw.denom === STAKING_DENOM);
                      return auraReward && parseFloat(auraReward.amount) > 0;
                    }).length || 0} validators
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
                  <button
                    className="btn btn-primary"
                    onClick={() => handleClaimRewards(null)}
                    disabled={txLoading}
                  >
                    {txLoading ? 'Claiming...' : 'Claim All'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Staking;

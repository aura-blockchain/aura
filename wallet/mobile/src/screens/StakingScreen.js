/**
 * Staking Screen
 * Validator list, delegation, undelegation, redelegation, and rewards management
 */

import React, {useState, useEffect, useCallback} from 'react';
import {
  View,
  Text,
  TextInput,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
  RefreshControl,
  Modal,
  FlatList,
} from 'react-native';
import WalletService from '../services/WalletService';
import PawAPI from '../services/PawAPI';
import TransactionService from '../services/TransactionService';
const {COIN, CHAIN_CONFIG} = require('../../../config/chain');

const UNBONDING_DAYS = 21; // Standard unbonding period

/**
 * Format token amount from base to display units
 */
function formatAmount(amount, decimals = COIN.exponent || 6) {
  if (!amount) {
    return '0';
  }
  const value = parseInt(amount, 10) / Math.pow(10, decimals);
  return value.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 6,
  });
}

/**
 * Convert display amount to base units
 */
function toBaseUnits(amount, decimals = COIN.exponent || 6) {
  const value = parseFloat(amount);
  if (isNaN(value) || value <= 0) {
    return '0';
  }
  return Math.floor(value * Math.pow(10, decimals)).toString();
}

/**
 * Truncate validator address for display
 */
function truncateAddress(address, startChars = 12, endChars = 8) {
  if (!address || address.length <= startChars + endChars) {
    return address;
  }
  return `${address.slice(0, startChars)}...${address.slice(-endChars)}`;
}

/**
 * Format commission rate as percentage
 */
function formatCommission(rate) {
  if (!rate) {
    return '0%';
  }
  const percentage = parseFloat(rate) * 100;
  return `${percentage.toFixed(2)}%`;
}

/**
 * Format voting power percentage
 */
function formatVotingPower(tokens, totalBonded) {
  if (!tokens || !totalBonded) {
    return '0%';
  }
  const percentage = (parseInt(tokens, 10) / parseInt(totalBonded, 10)) * 100;
  return `${percentage.toFixed(2)}%`;
}

/**
 * Parse validator description safely
 */
function getValidatorDescription(validator) {
  try {
    return validator?.description?.details || 'No description available';
  } catch (e) {
    return 'No description available';
  }
}

/**
 * Get validator status display
 */
function getValidatorStatus(validator) {
  if (validator.jailed) {
    return {text: 'Jailed', color: '#ff6b6b'};
  }
  if (validator.status === 'BOND_STATUS_BONDED') {
    return {text: 'Active', color: '#51cf66'};
  }
  if (validator.status === 'BOND_STATUS_UNBONDING') {
    return {text: 'Unbonding', color: '#ffd43b'};
  }
  return {text: 'Inactive', color: '#888'};
}

function StakingScreen({navigation}) {
  // State
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [walletInfo, setWalletInfo] = useState(null);
  const [balance, setBalance] = useState(null);

  // Validators
  const [validators, setValidators] = useState([]);
  const [totalBonded, setTotalBonded] = useState('0');

  // Delegations
  const [delegations, setDelegations] = useState([]);
  const [unbonding, setUnbonding] = useState([]);
  const [rewards, setRewards] = useState({total: [], rewards: []});
  const [totalRewards, setTotalRewards] = useState('0');

  // Modal state
  const [modalVisible, setModalVisible] = useState(false);
  const [modalMode, setModalMode] = useState('delegate'); // delegate, undelegate, redelegate
  const [selectedValidator, setSelectedValidator] = useState(null);
  const [destinationValidator, setDestinationValidator] = useState(null);
  const [amount, setAmount] = useState('');
  const [password, setPassword] = useState('');

  // Tab state
  const [activeTab, setActiveTab] = useState('validators'); // validators, delegations

  useEffect(() => {
    loadData();
  }, []);

  /**
   * Load all staking data
   */
  const loadData = async () => {
    try {
      setLoading(true);

      const [info, bal, vals, pool] = await Promise.all([
        WalletService.getWalletInfo(),
        WalletService.getBalance(),
        PawAPI.getValidators('BOND_STATUS_BONDED'),
        PawAPI.getStakingPool(),
      ]);

      setWalletInfo(info);
      setBalance(bal);
      setValidators(vals.sort((a, b) => parseInt(b.tokens, 10) - parseInt(a.tokens, 10)));
      setTotalBonded(pool?.bonded_tokens || '0');

      // Load user-specific data if wallet exists
      if (info?.address) {
        const [dels, unbonds, rews] = await Promise.all([
          PawAPI.getDelegations(info.address),
          PawAPI.getUnbondingDelegations(info.address),
          PawAPI.getRewards(info.address),
        ]);

        setDelegations(dels || []);
        setUnbonding(unbonds || []);
        setRewards(rews || {total: [], rewards: []});

        // Calculate total rewards
        const total = rews?.total?.find(r => r.denom === (COIN.base || 'uaura'));
        setTotalRewards(total?.amount || '0');
      }
    } catch (error) {
      console.error('Error loading staking data:', error);
      Alert.alert('Error', 'Failed to load staking data. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
  }, []);

  /**
   * Get account info for transaction
   */
  const getAccountInfo = async (address) => {
    const account = await PawAPI.getAccount(address);
    return {
      accountNumber: parseInt(account.account_number, 10),
      sequence: parseInt(account.sequence, 10),
    };
  };

  /**
   * Open delegation modal
   */
  const openDelegateModal = (validator) => {
    setSelectedValidator(validator);
    setModalMode('delegate');
    setAmount('');
    setPassword('');
    setModalVisible(true);
  };

  /**
   * Open undelegation modal
   */
  const openUndelegateModal = (delegation) => {
    const validator = validators.find(
      v => v.operator_address === delegation.delegation.validator_address
    );
    setSelectedValidator(validator || {
      operator_address: delegation.delegation.validator_address,
      description: {moniker: 'Unknown Validator'},
    });
    setModalMode('undelegate');
    setAmount('');
    setPassword('');
    setModalVisible(true);
  };

  /**
   * Open redelegation modal
   */
  const openRedelegateModal = (delegation) => {
    const validator = validators.find(
      v => v.operator_address === delegation.delegation.validator_address
    );
    setSelectedValidator(validator || {
      operator_address: delegation.delegation.validator_address,
      description: {moniker: 'Unknown Validator'},
    });
    setDestinationValidator(null);
    setModalMode('redelegate');
    setAmount('');
    setPassword('');
    setModalVisible(true);
  };

  /**
   * Close modal
   */
  const closeModal = () => {
    setModalVisible(false);
    setSelectedValidator(null);
    setDestinationValidator(null);
    setAmount('');
    setPassword('');
  };

  /**
   * Validate delegation inputs
   */
  const validateInputs = () => {
    if (!amount || parseFloat(amount) <= 0) {
      Alert.alert('Error', 'Please enter a valid amount');
      return false;
    }

    if (!password) {
      Alert.alert('Error', 'Please enter your password');
      return false;
    }

    if (modalMode === 'delegate') {
      const availableBalance = parseFloat(balance?.formatted || '0');
      if (parseFloat(amount) > availableBalance) {
        Alert.alert('Error', 'Insufficient balance');
        return false;
      }
    }

    if (modalMode === 'undelegate' || modalMode === 'redelegate') {
      const delegation = delegations.find(
        d => d.delegation.validator_address === selectedValidator?.operator_address
      );
      const delegatedAmount = parseFloat(
        formatAmount(delegation?.balance?.amount || '0')
      );
      if (parseFloat(amount) > delegatedAmount) {
        Alert.alert('Error', 'Amount exceeds delegated balance');
        return false;
      }
    }

    if (modalMode === 'redelegate' && !destinationValidator) {
      Alert.alert('Error', 'Please select a destination validator');
      return false;
    }

    return true;
  };

  /**
   * Execute delegation
   */
  const handleDelegate = async () => {
    if (!validateInputs()) {
      return;
    }

    try {
      setSubmitting(true);

      const wallet = await WalletService.unlockWallet(password);
      const {accountNumber, sequence} = await getAccountInfo(walletInfo.address);

      await TransactionService.delegate({
        delegatorAddress: walletInfo.address,
        validatorAddress: selectedValidator.operator_address,
        amount: toBaseUnits(amount),
        denom: COIN.base || 'uaura',
        memo: '',
        privateKeyHex: wallet.privateKey,
        accountNumber,
        sequence,
        chainId: CHAIN_CONFIG.chainId,
      });

      Alert.alert(
        'Success',
        `Successfully delegated ${amount} ${COIN.symbol || 'AURA'} to ${
          selectedValidator.description?.moniker || 'validator'
        }`,
        [{text: 'OK', onPress: closeModal}]
      );

      await loadData();
    } catch (error) {
      console.error('Delegation error:', error);
      Alert.alert('Error', error.message || 'Failed to delegate tokens');
    } finally {
      setSubmitting(false);
    }
  };

  /**
   * Execute undelegation
   */
  const handleUndelegate = async () => {
    if (!validateInputs()) {
      return;
    }

    Alert.alert(
      'Confirm Undelegation',
      `Undelegating tokens will take ${UNBONDING_DAYS} days to complete. During this time, your tokens will not earn rewards and cannot be transferred. Continue?`,
      [
        {text: 'Cancel', style: 'cancel'},
        {
          text: 'Continue',
          onPress: async () => {
            try {
              setSubmitting(true);

              const wallet = await WalletService.unlockWallet(password);
              const {accountNumber, sequence} = await getAccountInfo(walletInfo.address);

              await TransactionService.undelegate({
                delegatorAddress: walletInfo.address,
                validatorAddress: selectedValidator.operator_address,
                amount: toBaseUnits(amount),
                denom: COIN.base || 'uaura',
                memo: '',
                privateKeyHex: wallet.privateKey,
                accountNumber,
                sequence,
                chainId: CHAIN_CONFIG.chainId,
              });

              Alert.alert(
                'Success',
                `Undelegation initiated. Your ${amount} ${COIN.symbol || 'AURA'} will be available in ${UNBONDING_DAYS} days.`,
                [{text: 'OK', onPress: closeModal}]
              );

              await loadData();
            } catch (error) {
              console.error('Undelegation error:', error);
              Alert.alert('Error', error.message || 'Failed to undelegate tokens');
            } finally {
              setSubmitting(false);
            }
          },
        },
      ]
    );
  };

  /**
   * Execute redelegation
   */
  const handleRedelegate = async () => {
    if (!validateInputs()) {
      return;
    }

    try {
      setSubmitting(true);

      const wallet = await WalletService.unlockWallet(password);
      const {accountNumber, sequence} = await getAccountInfo(walletInfo.address);

      await TransactionService.redelegate({
        delegatorAddress: walletInfo.address,
        srcValidatorAddress: selectedValidator.operator_address,
        dstValidatorAddress: destinationValidator.operator_address,
        amount: toBaseUnits(amount),
        denom: COIN.base || 'uaura',
        memo: '',
        privateKeyHex: wallet.privateKey,
        accountNumber,
        sequence,
        chainId: CHAIN_CONFIG.chainId,
      });

      Alert.alert(
        'Success',
        `Successfully redelegated ${amount} ${COIN.symbol || 'AURA'} from ${
          selectedValidator.description?.moniker || 'source validator'
        } to ${destinationValidator.description?.moniker || 'destination validator'}`,
        [{text: 'OK', onPress: closeModal}]
      );

      await loadData();
    } catch (error) {
      console.error('Redelegation error:', error);
      Alert.alert('Error', error.message || 'Failed to redelegate tokens');
    } finally {
      setSubmitting(false);
    }
  };

  /**
   * Claim all rewards
   */
  const handleClaimRewards = async () => {
    if (parseFloat(totalRewards) <= 0) {
      Alert.alert('No Rewards', 'You have no rewards to claim');
      return;
    }

    Alert.prompt(
      'Claim Rewards',
      'Enter your password to claim all staking rewards',
      async (inputPassword) => {
        if (!inputPassword) {
          return;
        }

        try {
          setSubmitting(true);

          const wallet = await WalletService.unlockWallet(inputPassword);
          const {accountNumber, sequence} = await getAccountInfo(walletInfo.address);

          // Claim rewards from each validator
          const validatorAddresses = delegations.map(d => d.delegation.validator_address);

          for (let i = 0; i < validatorAddresses.length; i++) {
            await TransactionService.withdrawRewards({
              delegatorAddress: walletInfo.address,
              validatorAddress: validatorAddresses[i],
              memo: '',
              privateKeyHex: wallet.privateKey,
              accountNumber: accountNumber + i,
              sequence: sequence + i,
              chainId: CHAIN_CONFIG.chainId,
            });
          }

          Alert.alert(
            'Success',
            `Successfully claimed ${formatAmount(totalRewards)} ${COIN.symbol || 'AURA'} in rewards!`
          );

          await loadData();
        } catch (error) {
          console.error('Claim rewards error:', error);
          Alert.alert('Error', error.message || 'Failed to claim rewards');
        } finally {
          setSubmitting(false);
        }
      },
      'secure-text'
    );
  };

  /**
   * Set max amount
   */
  const handleSetMax = () => {
    if (modalMode === 'delegate') {
      setAmount(balance?.formatted || '0');
    } else {
      const delegation = delegations.find(
        d => d.delegation.validator_address === selectedValidator?.operator_address
      );
      setAmount(formatAmount(delegation?.balance?.amount || '0'));
    }
  };

  /**
   * Render validator item
   */
  const renderValidatorItem = ({item: validator}) => {
    const status = getValidatorStatus(validator);
    const delegation = delegations.find(
      d => d.delegation.validator_address === validator.operator_address
    );

    return (
      <TouchableOpacity
        style={styles.validatorCard}
        onPress={() => openDelegateModal(validator)}>
        <View style={styles.validatorHeader}>
          <View style={styles.validatorInfo}>
            <Text style={styles.validatorName} numberOfLines={1}>
              {validator.description?.moniker || 'Unknown'}
            </Text>
            <Text style={styles.validatorAddress}>
              {truncateAddress(validator.operator_address)}
            </Text>
          </View>
          <View style={[styles.statusBadge, {backgroundColor: status.color + '20'}]}>
            <View style={[styles.statusDot, {backgroundColor: status.color}]} />
            <Text style={[styles.statusText, {color: status.color}]}>{status.text}</Text>
          </View>
        </View>

        <View style={styles.validatorStats}>
          <View style={styles.statItem}>
            <Text style={styles.statLabel}>Commission</Text>
            <Text style={styles.statValue}>
              {formatCommission(validator.commission?.commission_rates?.rate)}
            </Text>
          </View>
          <View style={styles.statItem}>
            <Text style={styles.statLabel}>Voting Power</Text>
            <Text style={styles.statValue}>
              {formatVotingPower(validator.tokens, totalBonded)}
            </Text>
          </View>
          <View style={styles.statItem}>
            <Text style={styles.statLabel}>Delegated</Text>
            <Text style={styles.statValue}>
              {formatAmount(validator.tokens)} {COIN.symbol || 'AURA'}
            </Text>
          </View>
        </View>

        {delegation && (
          <View style={styles.userDelegation}>
            <Text style={styles.userDelegationLabel}>Your Delegation:</Text>
            <Text style={styles.userDelegationValue}>
              {formatAmount(delegation.balance?.amount)} {COIN.symbol || 'AURA'}
            </Text>
          </View>
        )}
      </TouchableOpacity>
    );
  };

  /**
   * Render delegation item
   */
  const renderDelegationItem = ({item: delegation}) => {
    const validator = validators.find(
      v => v.operator_address === delegation.delegation.validator_address
    );
    const reward = rewards.rewards?.find(
      r => r.validator_address === delegation.delegation.validator_address
    );
    const rewardAmount = reward?.reward?.find(r => r.denom === (COIN.base || 'uaura'))?.amount || '0';

    return (
      <View style={styles.delegationCard}>
        <View style={styles.delegationHeader}>
          <Text style={styles.delegationValidatorName}>
            {validator?.description?.moniker || 'Unknown Validator'}
          </Text>
          <Text style={styles.delegationAmount}>
            {formatAmount(delegation.balance?.amount)} {COIN.symbol || 'AURA'}
          </Text>
        </View>

        <View style={styles.delegationReward}>
          <Text style={styles.delegationRewardLabel}>Pending Rewards:</Text>
          <Text style={styles.delegationRewardValue}>
            {formatAmount(rewardAmount)} {COIN.symbol || 'AURA'}
          </Text>
        </View>

        <View style={styles.delegationActions}>
          <TouchableOpacity
            style={[styles.delegationButton, styles.delegateButton]}
            onPress={() => openDelegateModal(validator)}>
            <Text style={styles.delegationButtonText}>Delegate More</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.delegationButton, styles.redelegateButton]}
            onPress={() => openRedelegateModal(delegation)}>
            <Text style={styles.delegationButtonText}>Redelegate</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.delegationButton, styles.undelegateButton]}
            onPress={() => openUndelegateModal(delegation)}>
            <Text style={[styles.delegationButtonText, {color: '#ff6b6b'}]}>
              Undelegate
            </Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  };

  /**
   * Render unbonding delegation item
   */
  const renderUnbondingItem = ({item: unbondingDelegation}) => {
    const validator = validators.find(
      v => v.operator_address === unbondingDelegation.validator_address
    );

    return (
      <View style={styles.unbondingCard}>
        <Text style={styles.unbondingValidatorName}>
          {validator?.description?.moniker || 'Unknown Validator'}
        </Text>
        {unbondingDelegation.entries?.map((entry, index) => (
          <View key={index} style={styles.unbondingEntry}>
            <View>
              <Text style={styles.unbondingAmount}>
                {formatAmount(entry.balance)} {COIN.symbol || 'AURA'}
              </Text>
              <Text style={styles.unbondingDate}>
                Available: {new Date(entry.completion_time).toLocaleDateString()}
              </Text>
            </View>
          </View>
        ))}
      </View>
    );
  };

  /**
   * Render validator selection for redelegation
   */
  const renderValidatorOption = ({item: validator}) => {
    const isSelected = destinationValidator?.operator_address === validator.operator_address;
    const isSourceValidator = validator.operator_address === selectedValidator?.operator_address;

    if (isSourceValidator) {
      return null;
    }

    return (
      <TouchableOpacity
        style={[
          styles.validatorOption,
          isSelected && styles.validatorOptionSelected,
        ]}
        onPress={() => setDestinationValidator(validator)}>
        <Text style={styles.validatorOptionName}>
          {validator.description?.moniker || 'Unknown'}
        </Text>
        <Text style={styles.validatorOptionCommission}>
          {formatCommission(validator.commission?.commission_rates?.rate)}
        </Text>
        {isSelected && <Text style={styles.checkMark}>check</Text>}
      </TouchableOpacity>
    );
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
        <Text style={styles.loadingText}>Loading staking data...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Summary Card */}
      <View style={styles.summaryCard}>
        <View style={styles.summaryRow}>
          <View style={styles.summaryItem}>
            <Text style={styles.summaryLabel}>Available</Text>
            <Text style={styles.summaryValue}>
              {balance?.formatted || '0'} {COIN.symbol || 'AURA'}
            </Text>
          </View>
          <View style={styles.summaryDivider} />
          <View style={styles.summaryItem}>
            <Text style={styles.summaryLabel}>Staked</Text>
            <Text style={styles.summaryValue}>
              {formatAmount(
                delegations.reduce(
                  (sum, d) => sum + parseInt(d.balance?.amount || '0', 10),
                  0
                )
              )}{' '}
              {COIN.symbol || 'AURA'}
            </Text>
          </View>
        </View>

        <TouchableOpacity
          style={[
            styles.claimButton,
            parseFloat(totalRewards) <= 0 && styles.claimButtonDisabled,
          ]}
          onPress={handleClaimRewards}
          disabled={parseFloat(totalRewards) <= 0 || submitting}>
          {submitting ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <>
              <Text style={styles.claimButtonText}>Claim Rewards</Text>
              <Text style={styles.claimButtonAmount}>
                {formatAmount(totalRewards)} {COIN.symbol || 'AURA'}
              </Text>
            </>
          )}
        </TouchableOpacity>
      </View>

      {/* Tab Selector */}
      <View style={styles.tabContainer}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'validators' && styles.tabActive]}
          onPress={() => setActiveTab('validators')}>
          <Text
            style={[
              styles.tabText,
              activeTab === 'validators' && styles.tabTextActive,
            ]}>
            Validators ({validators.length})
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'delegations' && styles.tabActive]}
          onPress={() => setActiveTab('delegations')}>
          <Text
            style={[
              styles.tabText,
              activeTab === 'delegations' && styles.tabTextActive,
            ]}>
            My Staking ({delegations.length})
          </Text>
        </TouchableOpacity>
      </View>

      {/* Content */}
      {activeTab === 'validators' ? (
        <FlatList
          data={validators}
          keyExtractor={item => item.operator_address}
          renderItem={renderValidatorItem}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              tintColor="#4A90E2"
            />
          }
          contentContainerStyle={styles.listContent}
          ListEmptyComponent={
            <View style={styles.emptyState}>
              <Text style={styles.emptyText}>No active validators found</Text>
            </View>
          }
        />
      ) : (
        <ScrollView
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              tintColor="#4A90E2"
            />
          }
          contentContainerStyle={styles.listContent}>
          {delegations.length === 0 ? (
            <View style={styles.emptyState}>
              <Text style={styles.emptyText}>No active delegations</Text>
              <Text style={styles.emptySubtext}>
                Select a validator to start staking
              </Text>
            </View>
          ) : (
            <>
              <Text style={styles.sectionTitle}>Active Delegations</Text>
              {delegations.map((delegation, index) => (
                <View key={index}>
                  {renderDelegationItem({item: delegation})}
                </View>
              ))}
            </>
          )}

          {unbonding.length > 0 && (
            <>
              <Text style={styles.sectionTitle}>Unbonding</Text>
              {unbonding.map((u, index) => (
                <View key={index}>{renderUnbondingItem({item: u})}</View>
              ))}
            </>
          )}
        </ScrollView>
      )}

      {/* Delegation Modal */}
      <Modal
        visible={modalVisible}
        animationType="slide"
        transparent={true}
        onRequestClose={closeModal}>
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>
                {modalMode === 'delegate' && 'Delegate Tokens'}
                {modalMode === 'undelegate' && 'Undelegate Tokens'}
                {modalMode === 'redelegate' && 'Redelegate Tokens'}
              </Text>
              <TouchableOpacity onPress={closeModal}>
                <Text style={styles.modalClose}>X</Text>
              </TouchableOpacity>
            </View>

            <ScrollView style={styles.modalBody}>
              {/* Validator Info */}
              <View style={styles.modalValidatorInfo}>
                <Text style={styles.modalLabel}>
                  {modalMode === 'redelegate' ? 'From Validator:' : 'Validator:'}
                </Text>
                <Text style={styles.modalValidatorName}>
                  {selectedValidator?.description?.moniker || 'Unknown'}
                </Text>
                <Text style={styles.modalValidatorAddress}>
                  {truncateAddress(selectedValidator?.operator_address)}
                </Text>
              </View>

              {/* Destination Validator (for redelegate) */}
              {modalMode === 'redelegate' && (
                <View style={styles.destinationSection}>
                  <Text style={styles.modalLabel}>To Validator:</Text>
                  {destinationValidator ? (
                    <View style={styles.selectedDestination}>
                      <Text style={styles.modalValidatorName}>
                        {destinationValidator.description?.moniker}
                      </Text>
                      <TouchableOpacity
                        onPress={() => setDestinationValidator(null)}>
                        <Text style={styles.changeButton}>Change</Text>
                      </TouchableOpacity>
                    </View>
                  ) : (
                    <FlatList
                      data={validators.filter(
                        v => v.operator_address !== selectedValidator?.operator_address
                      )}
                      keyExtractor={item => item.operator_address}
                      renderItem={renderValidatorOption}
                      style={styles.validatorList}
                      nestedScrollEnabled
                    />
                  )}
                </View>
              )}

              {/* Amount Input */}
              <View style={styles.inputSection}>
                <View style={styles.labelRow}>
                  <Text style={styles.modalLabel}>Amount ({COIN.symbol || 'AURA'})</Text>
                  <TouchableOpacity onPress={handleSetMax}>
                    <Text style={styles.maxButton}>MAX</Text>
                  </TouchableOpacity>
                </View>
                <TextInput
                  style={styles.input}
                  placeholder="0.00"
                  placeholderTextColor="#666"
                  value={amount}
                  onChangeText={setAmount}
                  keyboardType="decimal-pad"
                />
                <Text style={styles.availableText}>
                  Available:{' '}
                  {modalMode === 'delegate'
                    ? balance?.formatted || '0'
                    : formatAmount(
                        delegations.find(
                          d =>
                            d.delegation.validator_address ===
                            selectedValidator?.operator_address
                        )?.balance?.amount || '0'
                      )}{' '}
                  {COIN.symbol || 'AURA'}
                </Text>
              </View>

              {/* Warning for undelegate */}
              {modalMode === 'undelegate' && (
                <View style={styles.warningBox}>
                  <Text style={styles.warningText}>
                    Undelegating takes {UNBONDING_DAYS} days. During this time, your
                    tokens will not earn rewards and cannot be transferred.
                  </Text>
                </View>
              )}

              {/* Password Input */}
              <View style={styles.inputSection}>
                <Text style={styles.modalLabel}>Password</Text>
                <TextInput
                  style={styles.input}
                  placeholder="Enter your password"
                  placeholderTextColor="#666"
                  value={password}
                  onChangeText={setPassword}
                  secureTextEntry
                />
              </View>
            </ScrollView>

            {/* Modal Actions */}
            <View style={styles.modalActions}>
              <TouchableOpacity
                style={styles.cancelButton}
                onPress={closeModal}
                disabled={submitting}>
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.confirmButton, submitting && styles.buttonDisabled]}
                onPress={
                  modalMode === 'delegate'
                    ? handleDelegate
                    : modalMode === 'undelegate'
                    ? handleUndelegate
                    : handleRedelegate
                }
                disabled={submitting}>
                {submitting ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text style={styles.confirmButtonText}>
                    {modalMode === 'delegate' && 'Delegate'}
                    {modalMode === 'undelegate' && 'Undelegate'}
                    {modalMode === 'redelegate' && 'Redelegate'}
                  </Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0a0a0a',
  },
  loadingContainer: {
    flex: 1,
    backgroundColor: '#0a0a0a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: '#888',
    marginTop: 12,
    fontSize: 14,
  },

  // Summary Card
  summaryCard: {
    backgroundColor: '#1a1a1a',
    margin: 16,
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  summaryRow: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  summaryItem: {
    flex: 1,
    alignItems: 'center',
  },
  summaryDivider: {
    width: 1,
    backgroundColor: '#333',
    marginHorizontal: 16,
  },
  summaryLabel: {
    color: '#888',
    fontSize: 12,
    marginBottom: 4,
  },
  summaryValue: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  claimButton: {
    backgroundColor: '#4A90E2',
    borderRadius: 8,
    padding: 12,
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'center',
  },
  claimButtonDisabled: {
    backgroundColor: '#333',
  },
  claimButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginRight: 8,
  },
  claimButtonAmount: {
    color: '#fff',
    fontSize: 14,
    opacity: 0.9,
  },

  // Tabs
  tabContainer: {
    flexDirection: 'row',
    marginHorizontal: 16,
    marginBottom: 8,
  },
  tab: {
    flex: 1,
    padding: 12,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  tabActive: {
    borderBottomColor: '#4A90E2',
  },
  tabText: {
    color: '#888',
    fontSize: 14,
  },
  tabTextActive: {
    color: '#4A90E2',
    fontWeight: 'bold',
  },

  // List
  listContent: {
    paddingHorizontal: 16,
    paddingBottom: 20,
  },
  sectionTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginTop: 16,
    marginBottom: 12,
  },

  // Validator Card
  validatorCard: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#333',
  },
  validatorHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  validatorInfo: {
    flex: 1,
  },
  validatorName: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  validatorAddress: {
    color: '#888',
    fontSize: 12,
    fontFamily: 'monospace',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    marginRight: 4,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '500',
  },
  validatorStats: {
    flexDirection: 'row',
    borderTopWidth: 1,
    borderTopColor: '#333',
    paddingTop: 12,
  },
  statItem: {
    flex: 1,
  },
  statLabel: {
    color: '#888',
    fontSize: 11,
    marginBottom: 2,
  },
  statValue: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '500',
  },
  userDelegation: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    backgroundColor: '#4A90E2' + '20',
    padding: 8,
    borderRadius: 6,
    marginTop: 12,
  },
  userDelegationLabel: {
    color: '#4A90E2',
    fontSize: 12,
  },
  userDelegationValue: {
    color: '#4A90E2',
    fontSize: 12,
    fontWeight: 'bold',
  },

  // Delegation Card
  delegationCard: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#333',
  },
  delegationHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  delegationValidatorName: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    flex: 1,
  },
  delegationAmount: {
    color: '#4A90E2',
    fontSize: 16,
    fontWeight: 'bold',
  },
  delegationReward: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 8,
    borderTopWidth: 1,
    borderTopColor: '#333',
    marginBottom: 12,
  },
  delegationRewardLabel: {
    color: '#888',
    fontSize: 13,
  },
  delegationRewardValue: {
    color: '#51cf66',
    fontSize: 13,
    fontWeight: '500',
  },
  delegationActions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  delegationButton: {
    flex: 1,
    padding: 10,
    borderRadius: 6,
    alignItems: 'center',
    marginHorizontal: 4,
  },
  delegateButton: {
    backgroundColor: '#4A90E2' + '30',
  },
  redelegateButton: {
    backgroundColor: '#ffd43b' + '30',
  },
  undelegateButton: {
    backgroundColor: '#ff6b6b' + '20',
    borderWidth: 1,
    borderColor: '#ff6b6b' + '40',
  },
  delegationButtonText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '500',
  },

  // Unbonding Card
  unbondingCard: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#ffd43b' + '40',
  },
  unbondingValidatorName: {
    color: '#ffd43b',
    fontSize: 14,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  unbondingEntry: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 8,
    borderTopWidth: 1,
    borderTopColor: '#333',
  },
  unbondingAmount: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  unbondingDate: {
    color: '#888',
    fontSize: 12,
  },

  // Empty State
  emptyState: {
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    color: '#666',
    fontSize: 16,
    marginBottom: 8,
  },
  emptySubtext: {
    color: '#444',
    fontSize: 14,
  },

  // Modal
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.8)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#1a1a1a',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '90%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  modalTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  modalClose: {
    color: '#888',
    fontSize: 20,
    padding: 8,
  },
  modalBody: {
    padding: 16,
  },
  modalValidatorInfo: {
    backgroundColor: '#0a0a0a',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  modalLabel: {
    color: '#888',
    fontSize: 12,
    marginBottom: 4,
  },
  modalValidatorName: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 2,
  },
  modalValidatorAddress: {
    color: '#666',
    fontSize: 12,
    fontFamily: 'monospace',
  },

  // Destination Section
  destinationSection: {
    marginBottom: 16,
  },
  selectedDestination: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#0a0a0a',
    padding: 12,
    borderRadius: 8,
  },
  changeButton: {
    color: '#4A90E2',
    fontSize: 14,
  },
  validatorList: {
    maxHeight: 200,
    backgroundColor: '#0a0a0a',
    borderRadius: 8,
  },
  validatorOption: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  validatorOptionSelected: {
    backgroundColor: '#4A90E2' + '20',
  },
  validatorOptionName: {
    color: '#fff',
    fontSize: 14,
    flex: 1,
  },
  validatorOptionCommission: {
    color: '#888',
    fontSize: 12,
    marginRight: 8,
  },
  checkMark: {
    color: '#4A90E2',
    fontSize: 16,
  },

  // Input Section
  inputSection: {
    marginBottom: 16,
  },
  labelRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  maxButton: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: 'bold',
  },
  input: {
    backgroundColor: '#0a0a0a',
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 16,
  },
  availableText: {
    color: '#666',
    fontSize: 12,
    marginTop: 4,
  },

  // Warning
  warningBox: {
    backgroundColor: '#332200',
    borderLeftWidth: 4,
    borderLeftColor: '#ff9800',
    padding: 12,
    borderRadius: 4,
    marginBottom: 16,
  },
  warningText: {
    color: '#ff9800',
    fontSize: 13,
  },

  // Modal Actions
  modalActions: {
    flexDirection: 'row',
    padding: 16,
    borderTopWidth: 1,
    borderTopColor: '#333',
  },
  cancelButton: {
    flex: 1,
    padding: 14,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
    alignItems: 'center',
    marginRight: 8,
  },
  cancelButtonText: {
    color: '#888',
    fontSize: 16,
  },
  confirmButton: {
    flex: 2,
    backgroundColor: '#4A90E2',
    padding: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginLeft: 8,
  },
  confirmButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
});

export default StakingScreen;

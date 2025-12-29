/**
 * Swap Screen
 * DEX token swap interface with pool selection, slippage control, and quote preview
 */

import React, {useState, useEffect, useCallback, useMemo} from 'react';
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

// Default slippage tolerance in basis points (0.5%)
const DEFAULT_SLIPPAGE_BPS = 50;
const MIN_SLIPPAGE_BPS = 10;
const MAX_SLIPPAGE_BPS = 1000;

// Swap fee percentage (typically 0.3% for AMM pools)
const SWAP_FEE_PERCENT = 0.003;

/**
 * Format token amount from base to display units
 */
function formatAmount(amount, decimals = 6) {
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
function toBaseUnits(amount, decimals = 6) {
  const value = parseFloat(amount);
  if (isNaN(value) || value <= 0) {
    return '0';
  }
  return Math.floor(value * Math.pow(10, decimals)).toString();
}

/**
 * Get token display name
 */
function getTokenDisplayName(denom) {
  if (!denom) {
    return 'Unknown';
  }
  if (denom === (COIN.base || 'uaura')) {
    return COIN.symbol || 'AURA';
  }
  // Handle IBC denoms
  if (denom.startsWith('ibc/')) {
    return denom.slice(4, 12) + '...';
  }
  // Remove 'u' prefix for micro-denominations
  if (denom.startsWith('u')) {
    return denom.slice(1).toUpperCase();
  }
  return denom.toUpperCase();
}

/**
 * Get token decimals (default 6 for Cosmos SDK tokens)
 */
function getTokenDecimals(denom) {
  // Most Cosmos SDK tokens use 6 decimals
  return 6;
}

/**
 * Calculate output amount using constant product formula (x * y = k)
 */
function calculateSwapOutput(amountIn, reserveIn, reserveOut, feePercent = SWAP_FEE_PERCENT) {
  if (!amountIn || !reserveIn || !reserveOut) {
    return '0';
  }

  const inputAmount = BigInt(amountIn);
  const inputReserve = BigInt(reserveIn);
  const outputReserve = BigInt(reserveOut);

  if (inputAmount <= 0n || inputReserve <= 0n || outputReserve <= 0n) {
    return '0';
  }

  // Apply swap fee
  const feeMultiplier = BigInt(Math.floor((1 - feePercent) * 10000));
  const inputWithFee = inputAmount * feeMultiplier;

  // Constant product formula: dy = y * dx / (x + dx)
  const numerator = inputWithFee * outputReserve;
  const denominator = inputReserve * 10000n + inputWithFee;

  return (numerator / denominator).toString();
}

/**
 * Calculate price impact percentage
 */
function calculatePriceImpact(amountIn, reserveIn) {
  if (!amountIn || !reserveIn) {
    return 0;
  }

  const input = parseInt(amountIn, 10);
  const reserve = parseInt(reserveIn, 10);

  if (reserve === 0) {
    return 0;
  }

  // Price impact is approximately: input / (reserve + input)
  const impact = (input / (reserve + input)) * 100;
  return Math.min(impact, 100);
}

/**
 * Format price impact for display
 */
function formatPriceImpact(impact) {
  if (impact < 0.01) {
    return '<0.01%';
  }
  return `${impact.toFixed(2)}%`;
}

/**
 * Get price impact severity color
 */
function getPriceImpactColor(impact) {
  if (impact < 1) {
    return '#51cf66'; // Green - low impact
  }
  if (impact < 3) {
    return '#ffd43b'; // Yellow - moderate impact
  }
  if (impact < 5) {
    return '#ff9800'; // Orange - high impact
  }
  return '#ff6b6b'; // Red - very high impact
}

/**
 * Calculate minimum received amount with slippage
 */
function calculateMinReceived(amountOut, slippageBps) {
  if (!amountOut || parseInt(amountOut, 10) <= 0) {
    return '0';
  }

  const output = BigInt(amountOut);
  const slippageMultiplier = BigInt(10000 - slippageBps);
  const minReceived = (output * slippageMultiplier) / 10000n;

  return minReceived.toString();
}

function SwapScreen({navigation}) {
  // State
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [walletInfo, setWalletInfo] = useState(null);
  const [balances, setBalances] = useState([]);

  // Pools
  const [pools, setPools] = useState([]);
  const [selectedPool, setSelectedPool] = useState(null);
  const [poolSelectVisible, setPoolSelectVisible] = useState(false);

  // Swap state
  const [inputToken, setInputToken] = useState(null);
  const [outputToken, setOutputToken] = useState(null);
  const [inputAmount, setInputAmount] = useState('');
  const [outputAmount, setOutputAmount] = useState('');
  const [slippageBps, setSlippageBps] = useState(DEFAULT_SLIPPAGE_BPS);
  const [slippageModalVisible, setSlippageModalVisible] = useState(false);
  const [customSlippage, setCustomSlippage] = useState('');

  // Password modal
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [password, setPassword] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  // Recalculate output when input changes
  useEffect(() => {
    calculateOutput();
  }, [inputAmount, inputToken, outputToken, selectedPool]);

  /**
   * Load all DEX data
   */
  const loadData = async () => {
    try {
      setLoading(true);

      const [info, dexPools] = await Promise.all([
        WalletService.getWalletInfo(),
        PawAPI.getDexPools(),
      ]);

      setWalletInfo(info);
      setPools(dexPools || []);

      // Load balances if wallet exists
      if (info?.address) {
        const balanceData = await PawAPI.getBalance(info.address);
        setBalances(balanceData?.balances || []);
      }

      // Auto-select first pool if available
      if (dexPools && dexPools.length > 0) {
        selectPool(dexPools[0]);
      }
    } catch (error) {
      console.error('Error loading DEX data:', error);
      Alert.alert('Error', 'Failed to load DEX data. Please try again.');
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
   * Select a pool and set up tokens
   */
  const selectPool = (pool) => {
    setSelectedPool(pool);

    // Extract token denoms from pool
    const tokenA = pool.reserve_coin_denoms?.[0] || pool.denom_a;
    const tokenB = pool.reserve_coin_denoms?.[1] || pool.denom_b;

    setInputToken(tokenA);
    setOutputToken(tokenB);
    setInputAmount('');
    setOutputAmount('');
    setPoolSelectVisible(false);
  };

  /**
   * Calculate output amount based on input
   */
  const calculateOutput = () => {
    if (!selectedPool || !inputAmount || !inputToken || !outputToken) {
      setOutputAmount('');
      return;
    }

    const inputValue = toBaseUnits(inputAmount, getTokenDecimals(inputToken));
    if (inputValue === '0') {
      setOutputAmount('');
      return;
    }

    // Get pool reserves
    const reserves = selectedPool.reserve_coins || [];
    const reserveA = reserves.find(r => r.denom === inputToken);
    const reserveB = reserves.find(r => r.denom === outputToken);

    if (!reserveA || !reserveB) {
      setOutputAmount('');
      return;
    }

    const calculatedOutput = calculateSwapOutput(
      inputValue,
      reserveA.amount,
      reserveB.amount
    );

    const outputDecimals = getTokenDecimals(outputToken);
    const displayOutput = formatAmount(calculatedOutput, outputDecimals);
    setOutputAmount(displayOutput);
  };

  /**
   * Swap input and output tokens
   */
  const handleSwapTokens = () => {
    const tempToken = inputToken;
    const tempAmount = inputAmount;

    setInputToken(outputToken);
    setOutputToken(tempToken);
    setInputAmount('');
    setOutputAmount('');
  };

  /**
   * Get balance for a specific token
   */
  const getTokenBalance = (denom) => {
    const balance = balances.find(b => b.denom === denom);
    return balance?.amount || '0';
  };

  /**
   * Set max input amount
   */
  const handleSetMax = () => {
    if (!inputToken) {
      return;
    }

    const balance = getTokenBalance(inputToken);
    const decimals = getTokenDecimals(inputToken);
    const maxAmount = formatAmount(balance, decimals);

    // Leave some for gas if using native token
    if (inputToken === (COIN.base || 'uaura')) {
      const value = parseFloat(maxAmount);
      const safeAmount = Math.max(0, value - 0.1).toFixed(6);
      setInputAmount(safeAmount);
    } else {
      setInputAmount(maxAmount.replace(/,/g, ''));
    }
  };

  /**
   * Calculate swap details for display
   */
  const swapDetails = useMemo(() => {
    if (!selectedPool || !inputAmount || !inputToken || !outputToken) {
      return null;
    }

    const inputValue = toBaseUnits(inputAmount, getTokenDecimals(inputToken));
    if (inputValue === '0') {
      return null;
    }

    const reserves = selectedPool.reserve_coins || [];
    const reserveIn = reserves.find(r => r.denom === inputToken);
    const reserveOut = reserves.find(r => r.denom === outputToken);

    if (!reserveIn || !reserveOut) {
      return null;
    }

    const outputValue = calculateSwapOutput(inputValue, reserveIn.amount, reserveOut.amount);
    const priceImpact = calculatePriceImpact(inputValue, reserveIn.amount);
    const minReceived = calculateMinReceived(outputValue, slippageBps);

    // Calculate exchange rate
    const inputDecimals = getTokenDecimals(inputToken);
    const outputDecimals = getTokenDecimals(outputToken);
    const inputDisplay = parseFloat(inputAmount);
    const outputDisplay = parseInt(outputValue, 10) / Math.pow(10, outputDecimals);
    const rate = outputDisplay / inputDisplay;

    return {
      outputValue,
      priceImpact,
      minReceived,
      rate: rate.toFixed(6),
      fee: (parseFloat(inputAmount) * SWAP_FEE_PERCENT).toFixed(6),
    };
  }, [selectedPool, inputAmount, inputToken, outputToken, slippageBps]);

  /**
   * Validate swap inputs
   */
  const validateSwap = () => {
    if (!selectedPool) {
      Alert.alert('Error', 'Please select a liquidity pool');
      return false;
    }

    if (!inputAmount || parseFloat(inputAmount) <= 0) {
      Alert.alert('Error', 'Please enter an amount to swap');
      return false;
    }

    const inputBalance = getTokenBalance(inputToken);
    const inputValue = toBaseUnits(inputAmount, getTokenDecimals(inputToken));

    if (BigInt(inputValue) > BigInt(inputBalance)) {
      Alert.alert('Error', 'Insufficient balance');
      return false;
    }

    if (!swapDetails || parseInt(swapDetails.outputValue, 10) <= 0) {
      Alert.alert('Error', 'Swap output is too small');
      return false;
    }

    // Warn about high price impact
    if (swapDetails.priceImpact > 5) {
      return new Promise((resolve) => {
        Alert.alert(
          'High Price Impact',
          `This swap has a ${formatPriceImpact(swapDetails.priceImpact)} price impact. Are you sure you want to continue?`,
          [
            {text: 'Cancel', style: 'cancel', onPress: () => resolve(false)},
            {text: 'Continue', onPress: () => resolve(true)},
          ]
        );
      });
    }

    return true;
  };

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
   * Execute swap transaction
   */
  const handleSwap = async () => {
    const isValid = await validateSwap();
    if (!isValid) {
      return;
    }

    setPasswordModalVisible(true);
  };

  /**
   * Confirm and execute swap with password
   */
  const confirmSwap = async () => {
    if (!password) {
      Alert.alert('Error', 'Please enter your password');
      return;
    }

    try {
      setSubmitting(true);
      setPasswordModalVisible(false);

      const wallet = await WalletService.unlockWallet(password);
      const {accountNumber, sequence} = await getAccountInfo(walletInfo.address);

      const inputValue = toBaseUnits(inputAmount, getTokenDecimals(inputToken));

      await TransactionService.swapDexExactIn({
        senderAddress: walletInfo.address,
        poolId: selectedPool.id || selectedPool.pool_id,
        denomIn: inputToken,
        amountIn: inputValue,
        minAmountOut: swapDetails.minReceived,
        maxSlippageBps: slippageBps,
        memo: '',
        privateKeyHex: wallet.privateKey,
        accountNumber,
        sequence,
        chainId: CHAIN_CONFIG.chainId,
      });

      Alert.alert(
        'Swap Successful',
        `Swapped ${inputAmount} ${getTokenDisplayName(inputToken)} for approximately ${outputAmount} ${getTokenDisplayName(outputToken)}`,
        [{text: 'OK', onPress: () => {
          setInputAmount('');
          setOutputAmount('');
          setPassword('');
          loadData();
        }}]
      );
    } catch (error) {
      console.error('Swap error:', error);
      Alert.alert('Swap Failed', error.message || 'Failed to execute swap');
    } finally {
      setSubmitting(false);
    }
  };

  /**
   * Update slippage tolerance
   */
  const handleSlippageSelect = (bps) => {
    setSlippageBps(bps);
    setSlippageModalVisible(false);
  };

  /**
   * Apply custom slippage
   */
  const handleCustomSlippage = () => {
    const value = parseFloat(customSlippage);
    if (isNaN(value) || value <= 0) {
      Alert.alert('Error', 'Please enter a valid slippage percentage');
      return;
    }

    const bps = Math.round(value * 100);
    if (bps < MIN_SLIPPAGE_BPS) {
      Alert.alert('Warning', 'Very low slippage may cause transaction to fail');
    }
    if (bps > MAX_SLIPPAGE_BPS) {
      Alert.alert('Warning', 'High slippage may result in unfavorable prices');
    }

    setSlippageBps(Math.min(MAX_SLIPPAGE_BPS, Math.max(MIN_SLIPPAGE_BPS, bps)));
    setSlippageModalVisible(false);
    setCustomSlippage('');
  };

  /**
   * Render pool selection item
   */
  const renderPoolItem = ({item: pool}) => {
    const tokenA = pool.reserve_coin_denoms?.[0] || pool.denom_a;
    const tokenB = pool.reserve_coin_denoms?.[1] || pool.denom_b;
    const isSelected = selectedPool?.id === pool.id || selectedPool?.pool_id === pool.pool_id;

    return (
      <TouchableOpacity
        style={[styles.poolItem, isSelected && styles.poolItemSelected]}
        onPress={() => selectPool(pool)}>
        <View style={styles.poolTokens}>
          <Text style={styles.poolTokenName}>
            {getTokenDisplayName(tokenA)} / {getTokenDisplayName(tokenB)}
          </Text>
          <Text style={styles.poolId}>Pool #{pool.id || pool.pool_id}</Text>
        </View>
        {pool.reserve_coins && (
          <View style={styles.poolLiquidity}>
            <Text style={styles.poolLiquidityLabel}>Liquidity</Text>
            <Text style={styles.poolLiquidityValue}>
              {formatAmount(pool.reserve_coins[0]?.amount)} / {formatAmount(pool.reserve_coins[1]?.amount)}
            </Text>
          </View>
        )}
        {isSelected && <Text style={styles.selectedMark}>SELECTED</Text>}
      </TouchableOpacity>
    );
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
        <Text style={styles.loadingText}>Loading DEX pools...</Text>
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.container}
      refreshControl={
        <RefreshControl
          refreshing={refreshing}
          onRefresh={onRefresh}
          tintColor="#4A90E2"
        />
      }>
      <View style={styles.content}>
        {/* Pool Selection */}
        <TouchableOpacity
          style={styles.poolSelector}
          onPress={() => setPoolSelectVisible(true)}>
          {selectedPool ? (
            <>
              <View style={styles.selectedPool}>
                <Text style={styles.selectedPoolLabel}>Selected Pool</Text>
                <Text style={styles.selectedPoolName}>
                  {getTokenDisplayName(inputToken)} / {getTokenDisplayName(outputToken)}
                </Text>
              </View>
              <Text style={styles.changeText}>Change</Text>
            </>
          ) : (
            <Text style={styles.selectPoolText}>Select a pool to swap</Text>
          )}
        </TouchableOpacity>

        {/* Swap Interface */}
        {selectedPool && (
          <>
            {/* Input Token */}
            <View style={styles.swapCard}>
              <View style={styles.tokenHeader}>
                <Text style={styles.tokenLabel}>From</Text>
                <Text style={styles.balanceText}>
                  Balance: {formatAmount(getTokenBalance(inputToken), getTokenDecimals(inputToken))}
                </Text>
              </View>

              <View style={styles.tokenRow}>
                <View style={styles.tokenBadge}>
                  <Text style={styles.tokenSymbol}>{getTokenDisplayName(inputToken)}</Text>
                </View>
                <TextInput
                  style={styles.amountInput}
                  placeholder="0.00"
                  placeholderTextColor="#666"
                  value={inputAmount}
                  onChangeText={setInputAmount}
                  keyboardType="decimal-pad"
                />
              </View>

              <TouchableOpacity style={styles.maxButton} onPress={handleSetMax}>
                <Text style={styles.maxButtonText}>MAX</Text>
              </TouchableOpacity>
            </View>

            {/* Swap Direction Button */}
            <TouchableOpacity style={styles.swapDirectionButton} onPress={handleSwapTokens}>
              <View style={styles.swapDirectionIcon}>
                <Text style={styles.swapDirectionText}>SWAP</Text>
              </View>
            </TouchableOpacity>

            {/* Output Token */}
            <View style={styles.swapCard}>
              <View style={styles.tokenHeader}>
                <Text style={styles.tokenLabel}>To (estimated)</Text>
                <Text style={styles.balanceText}>
                  Balance: {formatAmount(getTokenBalance(outputToken), getTokenDecimals(outputToken))}
                </Text>
              </View>

              <View style={styles.tokenRow}>
                <View style={styles.tokenBadge}>
                  <Text style={styles.tokenSymbol}>{getTokenDisplayName(outputToken)}</Text>
                </View>
                <Text style={styles.outputAmount}>
                  {outputAmount || '0.00'}
                </Text>
              </View>
            </View>

            {/* Swap Details */}
            {swapDetails && (
              <View style={styles.detailsCard}>
                <View style={styles.detailRow}>
                  <Text style={styles.detailLabel}>Exchange Rate</Text>
                  <Text style={styles.detailValue}>
                    1 {getTokenDisplayName(inputToken)} = {swapDetails.rate} {getTokenDisplayName(outputToken)}
                  </Text>
                </View>

                <View style={styles.detailRow}>
                  <Text style={styles.detailLabel}>Price Impact</Text>
                  <Text style={[styles.detailValue, {color: getPriceImpactColor(swapDetails.priceImpact)}]}>
                    {formatPriceImpact(swapDetails.priceImpact)}
                  </Text>
                </View>

                <View style={styles.detailRow}>
                  <Text style={styles.detailLabel}>Minimum Received</Text>
                  <Text style={styles.detailValue}>
                    {formatAmount(swapDetails.minReceived, getTokenDecimals(outputToken))} {getTokenDisplayName(outputToken)}
                  </Text>
                </View>

                <View style={styles.detailRow}>
                  <Text style={styles.detailLabel}>Swap Fee ({(SWAP_FEE_PERCENT * 100).toFixed(2)}%)</Text>
                  <Text style={styles.detailValue}>
                    ~{swapDetails.fee} {getTokenDisplayName(inputToken)}
                  </Text>
                </View>

                <TouchableOpacity
                  style={styles.slippageRow}
                  onPress={() => setSlippageModalVisible(true)}>
                  <Text style={styles.detailLabel}>Slippage Tolerance</Text>
                  <View style={styles.slippageValue}>
                    <Text style={styles.detailValue}>{(slippageBps / 100).toFixed(2)}%</Text>
                    <Text style={styles.editIcon}>EDIT</Text>
                  </View>
                </TouchableOpacity>
              </View>
            )}

            {/* Swap Button */}
            <TouchableOpacity
              style={[
                styles.swapButton,
                (!inputAmount || submitting) && styles.swapButtonDisabled,
              ]}
              onPress={handleSwap}
              disabled={!inputAmount || submitting}>
              {submitting ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.swapButtonText}>
                  {!inputAmount
                    ? 'Enter an amount'
                    : `Swap ${getTokenDisplayName(inputToken)} for ${getTokenDisplayName(outputToken)}`}
                </Text>
              )}
            </TouchableOpacity>

            {/* Warning for high price impact */}
            {swapDetails && swapDetails.priceImpact > 3 && (
              <View style={styles.warningBox}>
                <Text style={styles.warningText}>
                  Warning: High price impact ({formatPriceImpact(swapDetails.priceImpact)}).
                  Consider swapping a smaller amount or using a different pool.
                </Text>
              </View>
            )}
          </>
        )}

        {/* No Pools State */}
        {pools.length === 0 && (
          <View style={styles.emptyState}>
            <Text style={styles.emptyTitle}>No Pools Available</Text>
            <Text style={styles.emptyText}>
              There are no liquidity pools available for swapping at this time.
            </Text>
          </View>
        )}
      </View>

      {/* Pool Selection Modal */}
      <Modal
        visible={poolSelectVisible}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setPoolSelectVisible(false)}>
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Select Pool</Text>
              <TouchableOpacity onPress={() => setPoolSelectVisible(false)}>
                <Text style={styles.modalClose}>X</Text>
              </TouchableOpacity>
            </View>

            <FlatList
              data={pools}
              keyExtractor={item => (item.id || item.pool_id || '').toString()}
              renderItem={renderPoolItem}
              contentContainerStyle={styles.poolList}
              ListEmptyComponent={
                <View style={styles.emptyState}>
                  <Text style={styles.emptyText}>No pools available</Text>
                </View>
              }
            />
          </View>
        </View>
      </Modal>

      {/* Slippage Settings Modal */}
      <Modal
        visible={slippageModalVisible}
        animationType="fade"
        transparent={true}
        onRequestClose={() => setSlippageModalVisible(false)}>
        <View style={styles.modalOverlay}>
          <View style={styles.slippageModal}>
            <Text style={styles.slippageTitle}>Slippage Tolerance</Text>
            <Text style={styles.slippageDescription}>
              Your transaction will revert if the price changes unfavorably by more than this percentage.
            </Text>

            <View style={styles.slippageOptions}>
              {[10, 50, 100, 200].map(bps => (
                <TouchableOpacity
                  key={bps}
                  style={[
                    styles.slippageOption,
                    slippageBps === bps && styles.slippageOptionSelected,
                  ]}
                  onPress={() => handleSlippageSelect(bps)}>
                  <Text
                    style={[
                      styles.slippageOptionText,
                      slippageBps === bps && styles.slippageOptionTextSelected,
                    ]}>
                    {(bps / 100).toFixed(1)}%
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <View style={styles.customSlippage}>
              <TextInput
                style={styles.customSlippageInput}
                placeholder="Custom"
                placeholderTextColor="#666"
                value={customSlippage}
                onChangeText={setCustomSlippage}
                keyboardType="decimal-pad"
              />
              <Text style={styles.percentSymbol}>%</Text>
              <TouchableOpacity
                style={styles.applyButton}
                onPress={handleCustomSlippage}>
                <Text style={styles.applyButtonText}>Apply</Text>
              </TouchableOpacity>
            </View>

            <TouchableOpacity
              style={styles.closeSlippageButton}
              onPress={() => setSlippageModalVisible(false)}>
              <Text style={styles.closeSlippageText}>Done</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>

      {/* Password Confirmation Modal */}
      <Modal
        visible={passwordModalVisible}
        animationType="fade"
        transparent={true}
        onRequestClose={() => setPasswordModalVisible(false)}>
        <View style={styles.modalOverlay}>
          <View style={styles.passwordModal}>
            <Text style={styles.passwordTitle}>Confirm Swap</Text>

            <View style={styles.swapSummary}>
              <Text style={styles.swapSummaryText}>
                Swap {inputAmount} {getTokenDisplayName(inputToken)}
              </Text>
              <Text style={styles.swapSummaryArrow}>FOR</Text>
              <Text style={styles.swapSummaryText}>
                ~{outputAmount} {getTokenDisplayName(outputToken)}
              </Text>
            </View>

            <View style={styles.passwordInputContainer}>
              <Text style={styles.passwordLabel}>Enter Password</Text>
              <TextInput
                style={styles.passwordInput}
                placeholder="Password"
                placeholderTextColor="#666"
                value={password}
                onChangeText={setPassword}
                secureTextEntry
                autoFocus
              />
            </View>

            <View style={styles.passwordActions}>
              <TouchableOpacity
                style={styles.cancelPasswordButton}
                onPress={() => {
                  setPasswordModalVisible(false);
                  setPassword('');
                }}>
                <Text style={styles.cancelPasswordText}>Cancel</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.confirmSwapButton, submitting && styles.buttonDisabled]}
                onPress={confirmSwap}
                disabled={submitting}>
                {submitting ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text style={styles.confirmSwapText}>Confirm Swap</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </ScrollView>
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
  content: {
    padding: 16,
  },

  // Pool Selector
  poolSelector: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  selectedPool: {
    flex: 1,
  },
  selectedPoolLabel: {
    color: '#888',
    fontSize: 12,
    marginBottom: 4,
  },
  selectedPoolName: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  changeText: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: '500',
  },
  selectPoolText: {
    color: '#4A90E2',
    fontSize: 16,
    fontWeight: '500',
  },

  // Swap Card
  swapCard: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  tokenHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 12,
  },
  tokenLabel: {
    color: '#888',
    fontSize: 12,
  },
  balanceText: {
    color: '#666',
    fontSize: 12,
  },
  tokenRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  tokenBadge: {
    backgroundColor: '#4A90E2',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 8,
    marginRight: 12,
  },
  tokenSymbol: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
  amountInput: {
    flex: 1,
    fontSize: 24,
    color: '#fff',
    fontWeight: 'bold',
    textAlign: 'right',
  },
  outputAmount: {
    flex: 1,
    fontSize: 24,
    color: '#fff',
    fontWeight: 'bold',
    textAlign: 'right',
  },
  maxButton: {
    position: 'absolute',
    right: 16,
    top: 12,
    paddingHorizontal: 8,
    paddingVertical: 4,
    backgroundColor: '#4A90E2' + '30',
    borderRadius: 4,
  },
  maxButtonText: {
    color: '#4A90E2',
    fontSize: 11,
    fontWeight: 'bold',
  },

  // Swap Direction Button
  swapDirectionButton: {
    alignItems: 'center',
    marginVertical: 4,
  },
  swapDirectionIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#333',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 3,
    borderColor: '#0a0a0a',
  },
  swapDirectionText: {
    color: '#888',
    fontSize: 10,
    fontWeight: 'bold',
  },

  // Details Card
  detailsCard: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 16,
    marginTop: 16,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  detailRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  detailLabel: {
    color: '#888',
    fontSize: 13,
  },
  detailValue: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '500',
  },
  slippageRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 0,
  },
  slippageValue: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  editIcon: {
    color: '#4A90E2',
    fontSize: 11,
    marginLeft: 8,
    fontWeight: 'bold',
  },

  // Swap Button
  swapButton: {
    backgroundColor: '#4A90E2',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    marginBottom: 16,
  },
  swapButtonDisabled: {
    backgroundColor: '#333',
  },
  swapButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },

  // Warning Box
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

  // Empty State
  emptyState: {
    padding: 40,
    alignItems: 'center',
  },
  emptyTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  emptyText: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
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
    maxHeight: '80%',
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

  // Pool List
  poolList: {
    padding: 16,
  },
  poolItem: {
    backgroundColor: '#0a0a0a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#333',
  },
  poolItemSelected: {
    borderColor: '#4A90E2',
    backgroundColor: '#4A90E2' + '10',
  },
  poolTokens: {
    marginBottom: 8,
  },
  poolTokenName: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 2,
  },
  poolId: {
    color: '#888',
    fontSize: 12,
  },
  poolLiquidity: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#333',
  },
  poolLiquidityLabel: {
    color: '#888',
    fontSize: 12,
  },
  poolLiquidityValue: {
    color: '#fff',
    fontSize: 12,
  },
  selectedMark: {
    position: 'absolute',
    top: 8,
    right: 8,
    color: '#4A90E2',
    fontSize: 10,
    fontWeight: 'bold',
  },

  // Slippage Modal
  slippageModal: {
    backgroundColor: '#1a1a1a',
    borderRadius: 16,
    margin: 20,
    padding: 20,
    marginTop: 'auto',
    marginBottom: 'auto',
  },
  slippageTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  slippageDescription: {
    color: '#888',
    fontSize: 13,
    textAlign: 'center',
    marginBottom: 20,
  },
  slippageOptions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  slippageOption: {
    flex: 1,
    padding: 12,
    borderRadius: 8,
    backgroundColor: '#333',
    marginHorizontal: 4,
    alignItems: 'center',
  },
  slippageOptionSelected: {
    backgroundColor: '#4A90E2',
  },
  slippageOptionText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  slippageOptionTextSelected: {
    fontWeight: 'bold',
  },
  customSlippage: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  customSlippageInput: {
    flex: 1,
    backgroundColor: '#0a0a0a',
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 14,
  },
  percentSymbol: {
    color: '#888',
    fontSize: 14,
    marginHorizontal: 8,
  },
  applyButton: {
    backgroundColor: '#4A90E2',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 8,
  },
  applyButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
  closeSlippageButton: {
    backgroundColor: '#333',
    padding: 14,
    borderRadius: 8,
    alignItems: 'center',
  },
  closeSlippageText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },

  // Password Modal
  passwordModal: {
    backgroundColor: '#1a1a1a',
    borderRadius: 16,
    margin: 20,
    padding: 20,
    marginTop: 'auto',
    marginBottom: 'auto',
  },
  passwordTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 16,
    textAlign: 'center',
  },
  swapSummary: {
    backgroundColor: '#0a0a0a',
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    alignItems: 'center',
  },
  swapSummaryText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  swapSummaryArrow: {
    color: '#888',
    fontSize: 12,
    marginVertical: 8,
  },
  passwordInputContainer: {
    marginBottom: 20,
  },
  passwordLabel: {
    color: '#888',
    fontSize: 12,
    marginBottom: 8,
  },
  passwordInput: {
    backgroundColor: '#0a0a0a',
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 16,
  },
  passwordActions: {
    flexDirection: 'row',
  },
  cancelPasswordButton: {
    flex: 1,
    padding: 14,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
    alignItems: 'center',
    marginRight: 8,
  },
  cancelPasswordText: {
    color: '#888',
    fontSize: 16,
  },
  confirmSwapButton: {
    flex: 2,
    backgroundColor: '#4A90E2',
    padding: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginLeft: 8,
  },
  confirmSwapText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
});

export default SwapScreen;

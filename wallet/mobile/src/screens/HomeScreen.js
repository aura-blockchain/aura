/**
 * Home Screen
 * Main dashboard showing wallet balance and quick actions
 */

import React, {useState, useEffect, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  RefreshControl,
  ScrollView,
  ActivityIndicator,
} from 'react-native';
import WalletService from '../services/WalletService';
import PawAPI from '../services/PawAPI';

function HomeScreen({navigation}) {
  const [walletInfo, setWalletInfo] = useState(null);
  const [balance, setBalance] = useState(null);
  const [recentTxs, setRecentTxs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [price, setPrice] = useState(null);

  useEffect(() => {
    loadWalletData();
  }, []);

  const loadWalletData = async () => {
    try {
      setLoading(true);

      const [info, bal, txs] = await Promise.all([
        WalletService.getWalletInfo(),
        WalletService.getBalance(),
        WalletService.getTransactionHistory(5),
      ]);

      setWalletInfo(info);
      setBalance(bal);
      setRecentTxs(txs);

      // Try to get price (may fail if oracle not available)
      try {
        const prices = await PawAPI.getOraclePrices();
        const auraPrice = prices.find(p => p.symbol === 'Aura/USD');
        if (auraPrice) {
          setPrice(parseFloat(auraPrice.price));
        }
      } catch (error) {
        console.log('Price not available');
      }
    } catch (error) {
      console.error('Error loading wallet data:', error);
    } finally {
      setLoading(false);
    }
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadWalletData();
    setRefreshing(false);
  }, []);

  const handleSend = () => {
    navigation.navigate('Send');
  };

  const handleReceive = () => {
    navigation.navigate('Receive');
  };

  const formatBalance = () => {
    if (!balance) {
      return '0.000000';
    }
    return balance.formatted;
  };

  const formatAddress = address => {
    if (!address) {
      return '';
    }
    return `${address.substring(0, 10)}...${address.substring(address.length - 8)}`;
  };

  const calculateUsdValue = () => {
    if (!balance || !price) {
      return null;
    }
    const usdValue = parseFloat(balance.formatted) * price;
    return usdValue.toFixed(2);
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
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
      <View style={styles.header}>
        <Text style={styles.walletName}>{walletInfo?.name || 'My Wallet'}</Text>
        <Text style={styles.address}>{formatAddress(walletInfo?.address)}</Text>
      </View>

      <View style={styles.balanceCard}>
        <Text style={styles.balanceLabel}>Total Balance</Text>
        <Text style={styles.balanceAmount}>{formatBalance()} AURA</Text>
        {calculateUsdValue() && (
          <Text style={styles.balanceUsd}>${calculateUsdValue()} USD</Text>
        )}
      </View>

      <View style={styles.actionsContainer}>
        <TouchableOpacity style={styles.actionButton} onPress={handleSend}>
          <View style={styles.actionIcon}>
            <Text style={styles.actionIconText}>↑</Text>
          </View>
          <Text style={styles.actionText}>Send</Text>
        </TouchableOpacity>

        <TouchableOpacity style={styles.actionButton} onPress={handleReceive}>
          <View style={styles.actionIcon}>
            <Text style={styles.actionIconText}>↓</Text>
          </View>
          <Text style={styles.actionText}>Receive</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.actionButton}
          onPress={() => navigation.navigate('History')}>
          <View style={styles.actionIcon}>
            <Text style={styles.actionIconText}>≡</Text>
          </View>
          <Text style={styles.actionText}>History</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.recentSection}>
        <Text style={styles.sectionTitle}>Recent Transactions</Text>
        {recentTxs.length === 0 ? (
          <View style={styles.emptyState}>
            <Text style={styles.emptyText}>No transactions yet</Text>
            <Text style={styles.emptySubtext}>
              Start by receiving some Aura tokens
            </Text>
          </View>
        ) : (
          recentTxs.slice(0, 3).map((tx, index) => (
            <View key={index} style={styles.txItem}>
              <View style={styles.txIcon}>
                <Text style={styles.txIconText}>
                  {tx.type === 'send' ? '↑' : '↓'}
                </Text>
              </View>
              <View style={styles.txDetails}>
                <Text style={styles.txHash}>
                  {tx.txhash ? tx.txhash.substring(0, 16) + '...' : 'Unknown'}
                </Text>
                <Text style={styles.txTime}>
                  {tx.timestamp
                    ? new Date(tx.timestamp).toLocaleDateString()
                    : 'Recent'}
                </Text>
              </View>
              <View style={styles.txAmount}>
                <Text
                  style={[
                    styles.txAmountText,
                    tx.type === 'send' ? styles.txSend : styles.txReceive,
                  ]}>
                  {tx.type === 'send' ? '-' : '+'}
                  {tx.amount || '0'} AURA
                </Text>
              </View>
            </View>
          ))
        )}
        {recentTxs.length > 0 && (
          <TouchableOpacity
            style={styles.viewAllButton}
            onPress={() => navigation.navigate('History')}>
            <Text style={styles.viewAllText}>View All Transactions</Text>
          </TouchableOpacity>
        )}
      </View>

      <View style={styles.infoSection}>
        <Text style={styles.infoTitle}>Network Status</Text>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>Status:</Text>
          <View style={styles.statusIndicator}>
            <View style={styles.statusDot} />
            <Text style={styles.statusText}>Connected</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>Network:</Text>
          <Text style={styles.infoValue}>Aura Testnet</Text>
        </View>
      </View>
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
  header: {
    padding: 20,
    paddingTop: 40,
  },
  walletName: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 5,
  },
  address: {
    fontSize: 14,
    color: '#888',
    fontFamily: 'monospace',
  },
  balanceCard: {
    backgroundColor: '#1a1a1a',
    margin: 20,
    padding: 30,
    borderRadius: 12,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  balanceLabel: {
    fontSize: 14,
    color: '#888',
    marginBottom: 10,
  },
  balanceAmount: {
    fontSize: 36,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 5,
  },
  balanceUsd: {
    fontSize: 18,
    color: '#4A90E2',
  },
  actionsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingHorizontal: 20,
    marginBottom: 30,
  },
  actionButton: {
    alignItems: 'center',
  },
  actionIcon: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: '#4A90E2',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 8,
  },
  actionIconText: {
    fontSize: 24,
    color: '#fff',
    fontWeight: 'bold',
  },
  actionText: {
    fontSize: 14,
    color: '#fff',
  },
  recentSection: {
    paddingHorizontal: 20,
    marginBottom: 20,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 15,
  },
  emptyState: {
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    color: '#666',
    marginBottom: 5,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#444',
  },
  txItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1a1a1a',
    padding: 12,
    borderRadius: 8,
    marginBottom: 8,
  },
  txIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#333',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  txIconText: {
    fontSize: 18,
    color: '#fff',
  },
  txDetails: {
    flex: 1,
  },
  txHash: {
    fontSize: 14,
    color: '#fff',
    fontFamily: 'monospace',
    marginBottom: 2,
  },
  txTime: {
    fontSize: 12,
    color: '#666',
  },
  txAmount: {
    alignItems: 'flex-end',
  },
  txAmountText: {
    fontSize: 14,
    fontWeight: 'bold',
  },
  txSend: {
    color: '#ff6b6b',
  },
  txReceive: {
    color: '#51cf66',
  },
  viewAllButton: {
    marginTop: 10,
    padding: 12,
    alignItems: 'center',
  },
  viewAllText: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: '500',
  },
  infoSection: {
    margin: 20,
    padding: 16,
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  infoTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 12,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  infoLabel: {
    fontSize: 14,
    color: '#888',
  },
  infoValue: {
    fontSize: 14,
    color: '#fff',
  },
  statusIndicator: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#51cf66',
    marginRight: 6,
  },
  statusText: {
    fontSize: 14,
    color: '#51cf66',
  },
});

export default HomeScreen;

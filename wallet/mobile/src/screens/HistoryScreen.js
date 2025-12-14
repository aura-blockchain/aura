/**
 * History Screen
 * Display transaction history with filtering and search
 */

import React, {useState, useEffect, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  RefreshControl,
  ActivityIndicator,
  TouchableOpacity,
} from 'react-native';
import WalletService from '../services/WalletService';
import TransactionList from '../components/TransactionList';

function HistoryScreen() {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [filter, setFilter] = useState('all'); // all, sent, received

  useEffect(() => {
    loadTransactions();
  }, []);

  const loadTransactions = async () => {
    try {
      setLoading(true);
      const txs = await WalletService.getTransactionHistory(50);
      setTransactions(txs);
    } catch (error) {
      console.error('Error loading transactions:', error);
    } finally {
      setLoading(false);
    }
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadTransactions();
    setRefreshing(false);
  }, []);

  const filteredTransactions = () => {
    if (filter === 'all') {
      return transactions;
    }
    return transactions.filter(tx => tx.type === filter);
  };

  const renderEmptyState = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyText}>No transactions found</Text>
      <Text style={styles.emptySubtext}>
        Your transaction history will appear here
      </Text>
    </View>
  );

  const renderHeader = () => (
    <View style={styles.header}>
      <Text style={styles.title}>Transaction History</Text>
      <View style={styles.filterContainer}>
        <TouchableOpacity
          style={[styles.filterButton, filter === 'all' && styles.filterButtonActive]}
          onPress={() => setFilter('all')}>
          <Text
            style={[
              styles.filterText,
              filter === 'all' && styles.filterTextActive,
            ]}>
            All
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.filterButton, filter === 'sent' && styles.filterButtonActive]}
          onPress={() => setFilter('sent')}>
          <Text
            style={[
              styles.filterText,
              filter === 'sent' && styles.filterTextActive,
            ]}>
            Sent
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[
            styles.filterButton,
            filter === 'received' && styles.filterButtonActive,
          ]}
          onPress={() => setFilter('received')}>
          <Text
            style={[
              styles.filterText,
              filter === 'received' && styles.filterTextActive,
            ]}>
            Received
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {renderHeader()}
      <TransactionList
        transactions={filteredTransactions()}
        onRefresh={onRefresh}
        refreshing={refreshing}
        emptyComponent={renderEmptyState()}
      />
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
  header: {
    padding: 20,
    paddingTop: 30,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 20,
  },
  filterContainer: {
    flexDirection: 'row',
    gap: 10,
  },
  filterButton: {
    flex: 1,
    padding: 10,
    borderRadius: 8,
    backgroundColor: '#1a1a1a',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  filterButtonActive: {
    backgroundColor: '#4A90E2',
    borderColor: '#4A90E2',
  },
  filterText: {
    color: '#888',
    fontSize: 14,
    fontWeight: '500',
  },
  filterTextActive: {
    color: '#fff',
  },
  emptyContainer: {
    padding: 60,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 18,
    color: '#666',
    marginBottom: 10,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#444',
    textAlign: 'center',
  },
});

export default HistoryScreen;

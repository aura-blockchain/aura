/**
 * Transaction List Component
 * Reusable component for displaying transaction history
 */

import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
} from 'react-native';

function TransactionList({
  transactions,
  onRefresh,
  refreshing = false,
  onTransactionPress,
  emptyComponent,
}) {
  const renderTransaction = ({item}) => {
    const isSend = item.type === 'send';
    const isReceive = item.type === 'receive';

    return (
      <TouchableOpacity
        style={styles.txItem}
        onPress={() => onTransactionPress && onTransactionPress(item)}>
        <View style={styles.txIcon}>
          <Text style={styles.txIconText}>
            {isSend ? '↑' : isReceive ? '↓' : '•'}
          </Text>
        </View>

        <View style={styles.txDetails}>
          <Text style={styles.txType}>
            {isSend ? 'Sent' : isReceive ? 'Received' : 'Transaction'}
          </Text>
          {item.txhash && (
            <Text style={styles.txHash}>
              {item.txhash.substring(0, 16)}...
            </Text>
          )}
          {item.timestamp && (
            <Text style={styles.txTime}>
              {formatTimestamp(item.timestamp)}
            </Text>
          )}
          {item.height && (
            <Text style={styles.txHeight}>Block #{item.height}</Text>
          )}
        </View>

        <View style={styles.txAmountContainer}>
          {item.amount && (
            <Text
              style={[
                styles.txAmount,
                isSend ? styles.txAmountSend : styles.txAmountReceive,
              ]}>
              {isSend ? '-' : '+'}
              {item.amount}
            </Text>
          )}
          {item.denom && (
            <Text style={styles.txDenom}>{item.denom.toUpperCase()}</Text>
          )}
          {item.status && (
            <View
              style={[
                styles.statusBadge,
                item.status === 'success' && styles.statusSuccess,
                item.status === 'failed' && styles.statusFailed,
                item.status === 'pending' && styles.statusPending,
              ]}>
              <Text style={styles.statusText}>{item.status}</Text>
            </View>
          )}
        </View>
      </TouchableOpacity>
    );
  };

  const formatTimestamp = timestamp => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) {
      return 'Just now';
    } else if (diffMins < 60) {
      return `${diffMins}m ago`;
    } else if (diffHours < 24) {
      return `${diffHours}h ago`;
    } else if (diffDays < 7) {
      return `${diffDays}d ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  const renderEmpty = () => {
    if (emptyComponent) {
      return emptyComponent;
    }

    return (
      <View style={styles.emptyContainer}>
        <Text style={styles.emptyText}>No transactions</Text>
      </View>
    );
  };

  return (
    <FlatList
      data={transactions}
      renderItem={renderTransaction}
      keyExtractor={(item, index) =>
        item.txhash || item.id || index.toString()
      }
      contentContainerStyle={styles.listContainer}
      refreshControl={
        onRefresh ? (
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor="#4A90E2"
          />
        ) : undefined
      }
      ListEmptyComponent={renderEmpty}
      showsVerticalScrollIndicator={false}
    />
  );
}

const styles = StyleSheet.create({
  listContainer: {
    padding: 20,
    paddingTop: 0,
  },
  txItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1a1a1a',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#333',
  },
  txIcon: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: '#333',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 14,
  },
  txIconText: {
    fontSize: 20,
    color: '#fff',
  },
  txDetails: {
    flex: 1,
  },
  txType: {
    fontSize: 16,
    fontWeight: '500',
    color: '#fff',
    marginBottom: 4,
  },
  txHash: {
    fontSize: 13,
    color: '#666',
    fontFamily: 'monospace',
    marginBottom: 2,
  },
  txTime: {
    fontSize: 12,
    color: '#666',
  },
  txHeight: {
    fontSize: 11,
    color: '#555',
    marginTop: 2,
  },
  txAmountContainer: {
    alignItems: 'flex-end',
  },
  txAmount: {
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 2,
  },
  txAmountSend: {
    color: '#ff6b6b',
  },
  txAmountReceive: {
    color: '#51cf66',
  },
  txDenom: {
    fontSize: 12,
    color: '#888',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 4,
    marginTop: 4,
  },
  statusSuccess: {
    backgroundColor: '#1a4d2e',
  },
  statusFailed: {
    backgroundColor: '#4d1a1a',
  },
  statusPending: {
    backgroundColor: '#4d4d1a',
  },
  statusText: {
    fontSize: 10,
    fontWeight: 'bold',
    textTransform: 'uppercase',
  },
  emptyContainer: {
    padding: 60,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    color: '#666',
  },
});

export default TransactionList;

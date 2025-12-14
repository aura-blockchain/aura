/**
 * Send Screen
 * Send Aura tokens to another address
 */

import React, {useState} from 'react';
import {
  View,
  Text,
  TextInput,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
} from 'react-native';
import WalletService from '../services/WalletService';

function SendScreen({navigation}) {
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [memo, setMemo] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [balance, setBalance] = useState(null);

  React.useEffect(() => {
    loadBalance();
  }, []);

  const loadBalance = async () => {
    try {
      const bal = await WalletService.getBalance();
      setBalance(bal);
    } catch (error) {
      console.error('Error loading balance:', error);
    }
  };

  const handleScan = () => {
    Alert.alert('QR Scanner', 'QR scanner would open here');
  };

  const handleMaxAmount = () => {
    if (balance) {
      setAmount(balance.formatted);
    }
  };

  const validateInputs = () => {
    if (!recipient.trim()) {
      Alert.alert('Error', 'Please enter recipient address');
      return false;
    }

    if (!recipient.startsWith('aura')) {
      Alert.alert('Error', 'Invalid Aura address');
      return false;
    }

    if (!amount || parseFloat(amount) <= 0) {
      Alert.alert('Error', 'Please enter a valid amount');
      return false;
    }

    if (balance && parseFloat(amount) > parseFloat(balance.formatted)) {
      Alert.alert('Error', 'Insufficient balance');
      return false;
    }

    if (!password) {
      Alert.alert('Error', 'Please enter your password');
      return false;
    }

    return true;
  };

  const handleSend = async () => {
    if (!validateInputs()) {
      return;
    }

    try {
      setLoading(true);

      // Verify password by attempting to unlock wallet
      await WalletService.unlockWallet(password);

      // In a real implementation, this would sign and broadcast the transaction
      // For now, we'll just show a success message
      Alert.alert(
        'Success',
        `Transaction submitted!\n\nSent ${amount} AURA to ${recipient.substring(
          0,
          16,
        )}...`,
        [
          {
            text: 'OK',
            onPress: () => navigation.goBack(),
          },
        ],
      );
    } catch (error) {
      Alert.alert('Error', error.message || 'Failed to send transaction');
    } finally {
      setLoading(false);
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.content}>
        <View style={styles.balanceCard}>
          <Text style={styles.balanceLabel}>Available Balance</Text>
          <Text style={styles.balanceAmount}>
            {balance ? balance.formatted : '0.000000'} AURA
          </Text>
        </View>

        <View style={styles.inputContainer}>
          <View style={styles.labelRow}>
            <Text style={styles.label}>Recipient Address</Text>
            <TouchableOpacity onPress={handleScan}>
              <Text style={styles.scanText}>Scan QR</Text>
            </TouchableOpacity>
          </View>
          <TextInput
            style={styles.input}
            placeholder="aura1..."
            placeholderTextColor="#666"
            value={recipient}
            onChangeText={setRecipient}
            autoCapitalize="none"
            autoCorrect={false}
          />
        </View>

        <View style={styles.inputContainer}>
          <View style={styles.labelRow}>
            <Text style={styles.label}>Amount (AURA)</Text>
            <TouchableOpacity onPress={handleMaxAmount}>
              <Text style={styles.maxText}>MAX</Text>
            </TouchableOpacity>
          </View>
          <TextInput
            style={styles.input}
            placeholder="0.000000"
            placeholderTextColor="#666"
            value={amount}
            onChangeText={setAmount}
            keyboardType="decimal-pad"
          />
        </View>

        <View style={styles.inputContainer}>
          <Text style={styles.label}>Memo (optional)</Text>
          <TextInput
            style={styles.input}
            placeholder="Enter memo"
            placeholderTextColor="#666"
            value={memo}
            onChangeText={setMemo}
            maxLength={256}
          />
        </View>

        <View style={styles.inputContainer}>
          <Text style={styles.label}>Password</Text>
          <TextInput
            style={styles.input}
            placeholder="Enter your password"
            placeholderTextColor="#666"
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            autoCapitalize="none"
          />
        </View>

        <View style={styles.feeInfo}>
          <Text style={styles.feeLabel}>Transaction Fee</Text>
          <Text style={styles.feeValue}>~0.001 AURA</Text>
        </View>

        <TouchableOpacity
          style={[styles.sendButton, loading && styles.buttonDisabled]}
          onPress={handleSend}
          disabled={loading}>
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.sendButtonText}>Send Transaction</Text>
          )}
        </TouchableOpacity>

        <View style={styles.warningBox}>
          <Text style={styles.warningText}>
            ⚠️ Double-check the recipient address. Transactions cannot be
            reversed.
          </Text>
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
  content: {
    padding: 20,
  },
  balanceCard: {
    backgroundColor: '#1a1a1a',
    padding: 20,
    borderRadius: 8,
    marginBottom: 30,
    borderWidth: 1,
    borderColor: '#333',
  },
  balanceLabel: {
    fontSize: 12,
    color: '#888',
    marginBottom: 5,
  },
  balanceAmount: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  inputContainer: {
    marginBottom: 20,
  },
  labelRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  label: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  scanText: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: '500',
  },
  maxText: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: 'bold',
  },
  input: {
    backgroundColor: '#1a1a1a',
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 16,
  },
  feeInfo: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    padding: 12,
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    marginBottom: 20,
  },
  feeLabel: {
    color: '#888',
    fontSize: 14,
  },
  feeValue: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  sendButton: {
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 20,
  },
  sendButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  warningBox: {
    backgroundColor: '#332200',
    borderLeftWidth: 4,
    borderLeftColor: '#ff9800',
    padding: 12,
    borderRadius: 4,
  },
  warningText: {
    color: '#ff9800',
    fontSize: 13,
  },
});

export default SendScreen;

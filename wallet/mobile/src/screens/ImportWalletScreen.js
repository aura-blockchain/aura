/**
 * Import Wallet Screen
 * Allows user to import an existing wallet using recovery phrase
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
import {validateMnemonic} from '../utils/crypto';
import WalletService from '../services/WalletService';
import BiometricAuth from '../services/BiometricAuth';

function ImportWalletScreen({navigation}) {
  const [walletName, setWalletName] = useState('');
  const [mnemonic, setMnemonic] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [useBiometric, setUseBiometric] = useState(false);
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    checkBiometricAvailability();
  }, []);

  const checkBiometricAvailability = async () => {
    const result = await BiometricAuth.checkAvailability();
    setBiometricAvailable(result.available);
  };

  const handleImport = async () => {
    if (!validate()) {
      return;
    }

    try {
      setLoading(true);

      await WalletService.importWallet({
        mnemonic: mnemonic.trim(),
        walletName: walletName.trim() || 'Imported Wallet',
        password,
        useBiometric,
      });

      Alert.alert(
        'Success!',
        'Your wallet has been imported successfully',
        [
          {
            text: 'OK',
            onPress: () => navigation.replace('Main'),
          },
        ],
      );
    } catch (error) {
      Alert.alert('Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  const validate = () => {
    if (!mnemonic.trim()) {
      Alert.alert('Error', 'Please enter your recovery phrase');
      return false;
    }

    const mnemonicTrimmed = mnemonic.trim();
    if (!validateMnemonic(mnemonicTrimmed)) {
      Alert.alert(
        'Error',
        'Invalid recovery phrase. Please check and try again.',
      );
      return false;
    }

    if (password.length < 8) {
      Alert.alert('Error', 'Password must be at least 8 characters');
      return false;
    }

    if (password !== confirmPassword) {
      Alert.alert('Error', 'Passwords do not match');
      return false;
    }

    return true;
  };

  const handlePasteMnemonic = async () => {
    // In a real app, use Clipboard.getString()
    // For now, just show a placeholder
    Alert.alert('Info', 'Paste functionality requires Clipboard API');
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>Import Wallet</Text>
        <Text style={styles.subtitle}>
          Enter your 12 or 24-word recovery phrase to restore your wallet
        </Text>

        <View style={styles.warningBox}>
          <Text style={styles.warningText}>
            ⚠️ Never share your recovery phrase with anyone!
          </Text>
        </View>

        <View style={styles.inputContainer}>
          <View style={styles.labelRow}>
            <Text style={styles.label}>Recovery Phrase</Text>
            <TouchableOpacity onPress={handlePasteMnemonic}>
              <Text style={styles.pasteText}>Paste</Text>
            </TouchableOpacity>
          </View>
          <TextInput
            style={[styles.input, styles.mnemonicInput]}
            placeholder="Enter your recovery phrase (12 or 24 words)"
            placeholderTextColor="#666"
            value={mnemonic}
            onChangeText={setMnemonic}
            multiline
            numberOfLines={4}
            autoCapitalize="none"
            autoCorrect={false}
            textAlignVertical="top"
          />
          {mnemonic && (
            <Text style={styles.wordCount}>
              {mnemonic.trim().split(/\s+/).length} words
            </Text>
          )}
        </View>

        <TextInput
          style={styles.input}
          placeholder="Wallet Name (optional)"
          placeholderTextColor="#666"
          value={walletName}
          onChangeText={setWalletName}
          autoCapitalize="words"
        />

        <TextInput
          style={styles.input}
          placeholder="Password (min 8 characters)"
          placeholderTextColor="#666"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          autoCapitalize="none"
        />

        <TextInput
          style={styles.input}
          placeholder="Confirm Password"
          placeholderTextColor="#666"
          value={confirmPassword}
          onChangeText={setConfirmPassword}
          secureTextEntry
          autoCapitalize="none"
        />

        {biometricAvailable && (
          <TouchableOpacity
            style={styles.checkboxContainer}
            onPress={() => setUseBiometric(!useBiometric)}>
            <View style={styles.checkbox}>
              {useBiometric && <View style={styles.checkboxChecked} />}
            </View>
            <Text style={styles.checkboxLabel}>
              Enable {BiometricAuth.getBiometryTypeName()} authentication
            </Text>
          </TouchableOpacity>
        )}

        <TouchableOpacity
          style={[styles.importButton, loading && styles.buttonDisabled]}
          onPress={handleImport}
          disabled={loading}>
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.importButtonText}>Import Wallet</Text>
          )}
        </TouchableOpacity>

        <View style={styles.securityInfo}>
          <Text style={styles.securityTitle}>Security Information</Text>
          <Text style={styles.securityText}>
            • Your recovery phrase is the only way to restore your wallet
          </Text>
          <Text style={styles.securityText}>
            • We cannot recover your wallet if you lose your recovery phrase
          </Text>
          <Text style={styles.securityText}>
            • Never share your recovery phrase with anyone
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
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 10,
  },
  subtitle: {
    fontSize: 14,
    color: '#aaa',
    marginBottom: 20,
  },
  warningBox: {
    backgroundColor: '#332200',
    borderLeftWidth: 4,
    borderLeftColor: '#ff9800',
    padding: 12,
    borderRadius: 4,
    marginBottom: 20,
  },
  warningText: {
    color: '#ff9800',
    fontSize: 14,
  },
  inputContainer: {
    marginBottom: 16,
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
  pasteText: {
    color: '#4A90E2',
    fontSize: 14,
    fontWeight: '500',
  },
  input: {
    backgroundColor: '#1a1a1a',
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 16,
    marginBottom: 16,
  },
  mnemonicInput: {
    minHeight: 100,
    paddingTop: 12,
  },
  wordCount: {
    color: '#666',
    fontSize: 12,
    marginTop: -12,
    marginBottom: 16,
  },
  checkboxContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  checkbox: {
    width: 24,
    height: 24,
    borderWidth: 2,
    borderColor: '#4A90E2',
    borderRadius: 4,
    marginRight: 10,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checkboxChecked: {
    width: 14,
    height: 14,
    backgroundColor: '#4A90E2',
    borderRadius: 2,
  },
  checkboxLabel: {
    color: '#fff',
    fontSize: 14,
  },
  importButton: {
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 30,
  },
  importButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  securityInfo: {
    backgroundColor: '#1a1a1a',
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  securityTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  securityText: {
    color: '#aaa',
    fontSize: 13,
    marginBottom: 6,
  },
});

export default ImportWalletScreen;

/**
 * Create Wallet Screen
 * Guides user through creating a new wallet
 * Steps: 1) Set password, 2) Show mnemonic, 3) Verify mnemonic
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
import BiometricAuth from '../services/BiometricAuth';

function CreateWalletScreen({navigation}) {
  const [step, setStep] = useState(1);
  const [walletName, setWalletName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [mnemonic, setMnemonic] = useState('');
  const [mnemonicWords, setMnemonicWords] = useState([]);
  const [verifyWords, setVerifyWords] = useState(['', '', '']);
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

  const handleNext = async () => {
    if (step === 1) {
      if (!validateStep1()) {
        return;
      }
      await createWallet();
    } else if (step === 2) {
      setStep(3);
    } else if (step === 3) {
      if (!validateMnemonic()) {
        Alert.alert('Error', 'Please verify your recovery phrase correctly');
        return;
      }
      completeWalletCreation();
    }
  };

  const validateStep1 = () => {
    if (!walletName.trim()) {
      Alert.alert('Error', 'Please enter a wallet name');
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

  const createWallet = async () => {
    try {
      setLoading(true);

      const result = await WalletService.createWallet({
        walletName,
        password,
        useBiometric,
      });

      setMnemonic(result.mnemonic);
      setMnemonicWords(result.mnemonic.split(' '));
      setStep(2);
    } catch (error) {
      Alert.alert('Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  const validateMnemonic = () => {
    const words = mnemonic.split(' ');
    // Verify 3 random positions
    const positions = [2, 7, 11]; // 3rd, 8th, 12th words
    return (
      verifyWords[0].toLowerCase() === words[positions[0]].toLowerCase() &&
      verifyWords[1].toLowerCase() === words[positions[1]].toLowerCase() &&
      verifyWords[2].toLowerCase() === words[positions[2]].toLowerCase()
    );
  };

  const completeWalletCreation = () => {
    Alert.alert(
      'Success!',
      'Your wallet has been created successfully',
      [
        {
          text: 'OK',
          onPress: () => navigation.replace('Main'),
        },
      ],
    );
  };

  const renderStep1 = () => (
    <ScrollView style={styles.stepContainer}>
      <Text style={styles.stepTitle}>Step 1: Set Password</Text>
      <Text style={styles.stepDescription}>
        Choose a strong password to protect your wallet
      </Text>

      <TextInput
        style={styles.input}
        placeholder="Wallet Name"
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
    </ScrollView>
  );

  const renderStep2 = () => (
    <ScrollView style={styles.stepContainer}>
      <Text style={styles.stepTitle}>Step 2: Backup Recovery Phrase</Text>
      <Text style={styles.stepDescription}>
        Write down these 24 words in order. Keep them safe and never share them.
      </Text>

      <View style={styles.warningBox}>
        <Text style={styles.warningText}>
          ⚠️ IMPORTANT: Write this down on paper. Do not screenshot or store
          digitally!
        </Text>
      </View>

      <View style={styles.mnemonicContainer}>
        {mnemonicWords.map((word, index) => (
          <View key={index} style={styles.mnemonicWord}>
            <Text style={styles.mnemonicNumber}>{index + 1}.</Text>
            <Text style={styles.mnemonicText}>{word}</Text>
          </View>
        ))}
      </View>

      <Text style={styles.confirmText}>
        I have written down my recovery phrase
      </Text>
    </ScrollView>
  );

  const renderStep3 = () => (
    <ScrollView style={styles.stepContainer}>
      <Text style={styles.stepTitle}>Step 3: Verify Recovery Phrase</Text>
      <Text style={styles.stepDescription}>
        Enter the following words from your recovery phrase to verify
      </Text>

      <View style={styles.verifyContainer}>
        <Text style={styles.verifyLabel}>Word #3:</Text>
        <TextInput
          style={styles.input}
          placeholder="Enter word 3"
          placeholderTextColor="#666"
          value={verifyWords[0]}
          onChangeText={text => {
            const newWords = [...verifyWords];
            newWords[0] = text;
            setVerifyWords(newWords);
          }}
          autoCapitalize="none"
          autoCorrect={false}
        />

        <Text style={styles.verifyLabel}>Word #8:</Text>
        <TextInput
          style={styles.input}
          placeholder="Enter word 8"
          placeholderTextColor="#666"
          value={verifyWords[1]}
          onChangeText={text => {
            const newWords = [...verifyWords];
            newWords[1] = text;
            setVerifyWords(newWords);
          }}
          autoCapitalize="none"
          autoCorrect={false}
        />

        <Text style={styles.verifyLabel}>Word #12:</Text>
        <TextInput
          style={styles.input}
          placeholder="Enter word 12"
          placeholderTextColor="#666"
          value={verifyWords[2]}
          onChangeText={text => {
            const newWords = [...verifyWords];
            newWords[2] = text;
            setVerifyWords(newWords);
          }}
          autoCapitalize="none"
          autoCorrect={false}
        />
      </View>
    </ScrollView>
  );

  return (
    <View style={styles.container}>
      <View style={styles.progressBar}>
        <View style={[styles.progressStep, step >= 1 && styles.progressStepActive]} />
        <View style={[styles.progressStep, step >= 2 && styles.progressStepActive]} />
        <View style={[styles.progressStep, step >= 3 && styles.progressStepActive]} />
      </View>

      {step === 1 && renderStep1()}
      {step === 2 && renderStep2()}
      {step === 3 && renderStep3()}

      <View style={styles.buttonContainer}>
        {step > 1 && (
          <TouchableOpacity
            style={styles.secondaryButton}
            onPress={() => setStep(step - 1)}>
            <Text style={styles.secondaryButtonText}>Back</Text>
          </TouchableOpacity>
        )}

        <TouchableOpacity
          style={[styles.primaryButton, loading && styles.buttonDisabled]}
          onPress={handleNext}
          disabled={loading}>
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.primaryButtonText}>
              {step === 3 ? 'Complete' : 'Next'}
            </Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0a0a0a',
  },
  progressBar: {
    flexDirection: 'row',
    padding: 20,
    gap: 10,
  },
  progressStep: {
    flex: 1,
    height: 4,
    backgroundColor: '#333',
    borderRadius: 2,
  },
  progressStepActive: {
    backgroundColor: '#4A90E2',
  },
  stepContainer: {
    flex: 1,
    padding: 20,
  },
  stepTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 10,
  },
  stepDescription: {
    fontSize: 14,
    color: '#aaa',
    marginBottom: 30,
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
  checkboxContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 10,
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
  mnemonicContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
    marginBottom: 20,
  },
  mnemonicWord: {
    width: '48%',
    backgroundColor: '#1a1a1a',
    padding: 10,
    borderRadius: 6,
    flexDirection: 'row',
    alignItems: 'center',
  },
  mnemonicNumber: {
    color: '#666',
    fontSize: 12,
    marginRight: 8,
  },
  mnemonicText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '500',
  },
  confirmText: {
    color: '#4A90E2',
    fontSize: 14,
    textAlign: 'center',
    marginTop: 10,
  },
  verifyContainer: {
    marginTop: 20,
  },
  verifyLabel: {
    color: '#fff',
    fontSize: 14,
    marginBottom: 8,
  },
  buttonContainer: {
    flexDirection: 'row',
    padding: 20,
    gap: 10,
  },
  primaryButton: {
    flex: 1,
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  primaryButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  secondaryButton: {
    flex: 1,
    backgroundColor: 'transparent',
    padding: 16,
    borderRadius: 8,
    borderWidth: 2,
    borderColor: '#4A90E2',
    alignItems: 'center',
  },
  secondaryButtonText: {
    color: '#4A90E2',
    fontSize: 16,
    fontWeight: 'bold',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
});

export default CreateWalletScreen;

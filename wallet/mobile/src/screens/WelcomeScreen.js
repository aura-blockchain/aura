/**
 * Welcome Screen
 * First screen shown when app launches
 * Allows user to create new wallet or import existing one
 */

import React, {useEffect, useState} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import WalletService from '../services/WalletService';

function WelcomeScreen({navigation}) {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkWalletExists();
  }, []);

  /**
   * Check if wallet already exists
   * If yes, navigate to home; if no, show welcome options
   */
  const checkWalletExists = async () => {
    try {
      const hasWallet = await WalletService.hasWallet();
      if (hasWallet) {
        // Wallet exists, go to main screen
        navigation.replace('Main');
      } else {
        setLoading(false);
      }
    } catch (error) {
      console.error('Error checking wallet:', error);
      setLoading(false);
    }
  };

  const handleCreateWallet = () => {
    navigation.navigate('CreateWallet');
  };

  const handleImportWallet = () => {
    navigation.navigate('ImportWallet');
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>Welcome to</Text>
        <Text style={styles.appName}>Aura Wallet</Text>
        <Text style={styles.subtitle}>
          Secure blockchain wallet for Aura network
        </Text>
      </View>

      <View style={styles.buttonContainer}>
        <TouchableOpacity
          style={styles.primaryButton}
          onPress={handleCreateWallet}>
          <Text style={styles.primaryButtonText}>Create New Wallet</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.secondaryButton}
          onPress={handleImportWallet}>
          <Text style={styles.secondaryButtonText}>Import Existing Wallet</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.footer}>
        <Text style={styles.footerText}>
          By continuing, you agree to our Terms of Service
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0a0a0a',
    padding: 20,
  },
  loadingContainer: {
    flex: 1,
    backgroundColor: '#0a0a0a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: '#fff',
    marginTop: 10,
    fontSize: 16,
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  title: {
    fontSize: 24,
    color: '#888',
    marginBottom: 10,
  },
  appName: {
    fontSize: 48,
    fontWeight: 'bold',
    color: '#4A90E2',
    marginBottom: 20,
  },
  subtitle: {
    fontSize: 16,
    color: '#aaa',
    textAlign: 'center',
    paddingHorizontal: 40,
  },
  buttonContainer: {
    marginBottom: 40,
  },
  primaryButton: {
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
  },
  primaryButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  secondaryButton: {
    backgroundColor: 'transparent',
    padding: 16,
    borderRadius: 8,
    borderWidth: 2,
    borderColor: '#4A90E2',
  },
  secondaryButtonText: {
    color: '#4A90E2',
    fontSize: 18,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  footer: {
    paddingBottom: 20,
  },
  footerText: {
    color: '#666',
    fontSize: 12,
    textAlign: 'center',
  },
});

export default WelcomeScreen;

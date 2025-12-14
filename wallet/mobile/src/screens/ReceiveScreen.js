/**
 * Receive Screen
 * Display wallet address and QR code for receiving tokens
 */

import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
} from 'react-native';
import QRCode from 'react-native-qrcode-svg';
import WalletService from '../services/WalletService';

function ReceiveScreen() {
  const [address, setAddress] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadAddress();
  }, []);

  const loadAddress = async () => {
    try {
      const info = await WalletService.getWalletInfo();
      setAddress(info.address);
    } catch (error) {
      console.error('Error loading address:', error);
      Alert.alert('Error', 'Failed to load wallet address');
    } finally {
      setLoading(false);
    }
  };

  const handleCopyAddress = () => {
    // In real app, use Clipboard.setString(address)
    Alert.alert('Copied', 'Address copied to clipboard');
  };

  const handleShare = () => {
    Alert.alert('Share', 'Share functionality would open here');
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#4A90E2" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>Receive Aura</Text>
        <Text style={styles.subtitle}>
          Share this address to receive Aura tokens
        </Text>

        <View style={styles.qrContainer}>
          <QRCode value={address} size={220} backgroundColor="white" />
        </View>

        <View style={styles.addressContainer}>
          <Text style={styles.addressLabel}>Your Address</Text>
          <Text style={styles.address}>{address}</Text>
        </View>

        <TouchableOpacity
          style={styles.copyButton}
          onPress={handleCopyAddress}>
          <Text style={styles.copyButtonText}>Copy Address</Text>
        </TouchableOpacity>

        <TouchableOpacity style={styles.shareButton} onPress={handleShare}>
          <Text style={styles.shareButtonText}>Share Address</Text>
        </TouchableOpacity>

        <View style={styles.infoBox}>
          <Text style={styles.infoTitle}>Important</Text>
          <Text style={styles.infoText}>
            • Only send Aura tokens to this address
          </Text>
          <Text style={styles.infoText}>
            • Sending other tokens may result in permanent loss
          </Text>
          <Text style={styles.infoText}>
            • This address is for Aura mainnet only
          </Text>
        </View>
      </View>
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
  content: {
    flex: 1,
    padding: 20,
    alignItems: 'center',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    marginTop: 20,
    marginBottom: 10,
  },
  subtitle: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
    marginBottom: 40,
  },
  qrContainer: {
    backgroundColor: '#fff',
    padding: 20,
    borderRadius: 12,
    marginBottom: 30,
  },
  addressContainer: {
    width: '100%',
    backgroundColor: '#1a1a1a',
    padding: 16,
    borderRadius: 8,
    marginBottom: 20,
    borderWidth: 1,
    borderColor: '#333',
  },
  addressLabel: {
    fontSize: 12,
    color: '#888',
    marginBottom: 8,
  },
  address: {
    fontSize: 14,
    color: '#fff',
    fontFamily: 'monospace',
    lineHeight: 20,
  },
  copyButton: {
    width: '100%',
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 12,
  },
  copyButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  shareButton: {
    width: '100%',
    backgroundColor: 'transparent',
    padding: 16,
    borderRadius: 8,
    borderWidth: 2,
    borderColor: '#4A90E2',
    alignItems: 'center',
    marginBottom: 30,
  },
  shareButtonText: {
    color: '#4A90E2',
    fontSize: 16,
    fontWeight: 'bold',
  },
  infoBox: {
    width: '100%',
    backgroundColor: '#1a1a1a',
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  infoTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 10,
  },
  infoText: {
    fontSize: 13,
    color: '#aaa',
    marginBottom: 5,
  },
});

export default ReceiveScreen;

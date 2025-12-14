/**
 * Settings Screen
 * App settings, security options, and wallet management
 */

import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Switch,
  Alert,
} from 'react-native';
import WalletService from '../services/WalletService';
import BiometricAuth from '../services/BiometricAuth';
import StorageService from '../services/StorageService';

function SettingsScreen({navigation}) {
  const [walletInfo, setWalletInfo] = useState(null);
  const [biometricEnabled, setBiometricEnabled] = useState(false);
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const [theme, setTheme] = useState('dark');
  const [notifications, setNotifications] = useState(true);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const info = await WalletService.getWalletInfo();
      setWalletInfo(info);
      setBiometricEnabled(info.biometricEnabled || false);

      const availability = await BiometricAuth.checkAvailability();
      setBiometricAvailable(availability.available);

      const currentTheme = await StorageService.getTheme();
      setTheme(currentTheme);

      const notifSettings = await StorageService.getNotificationSettings();
      setNotifications(notifSettings.enabled);
    } catch (error) {
      console.error('Error loading settings:', error);
    }
  };

  const handleToggleBiometric = async value => {
    if (!biometricAvailable) {
      Alert.alert('Error', 'Biometric authentication is not available');
      return;
    }

    try {
      if (value) {
        const authenticated = await BiometricAuth.authenticate(
          'Enable biometric authentication',
        );
        if (authenticated) {
          await BiometricAuth.createKeys();
          setBiometricEnabled(true);
          Alert.alert('Success', 'Biometric authentication enabled');
        }
      } else {
        await BiometricAuth.deleteKeys();
        setBiometricEnabled(false);
        Alert.alert('Success', 'Biometric authentication disabled');
      }
    } catch (error) {
      Alert.alert('Error', error.message);
    }
  };

  const handleToggleTheme = async value => {
    const newTheme = value ? 'dark' : 'light';
    setTheme(newTheme);
    await StorageService.setTheme(newTheme);
  };

  const handleToggleNotifications = async value => {
    setNotifications(value);
    const settings = await StorageService.getNotificationSettings();
    await StorageService.setNotificationSettings({
      ...settings,
      enabled: value,
    });
  };

  const handleViewRecoveryPhrase = () => {
    Alert.alert(
      'Security Warning',
      'Your recovery phrase gives full access to your wallet. Only view in a secure location.',
      [
        {text: 'Cancel', style: 'cancel'},
        {
          text: 'Continue',
          onPress: () => {
            Alert.alert('Info', 'Recovery phrase viewing requires password authentication');
          },
        },
      ],
    );
  };

  const handleChangePassword = () => {
    Alert.alert('Info', 'Change password functionality coming soon');
  };

  const handleDeleteWallet = () => {
    Alert.alert(
      'Delete Wallet',
      'Are you sure? This will permanently delete your wallet from this device. Make sure you have your recovery phrase backed up!',
      [
        {text: 'Cancel', style: 'cancel'},
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await WalletService.deleteWallet();
              Alert.alert('Success', 'Wallet deleted', [
                {
                  text: 'OK',
                  onPress: () => navigation.replace('Welcome'),
                },
              ]);
            } catch (error) {
              Alert.alert('Error', error.message);
            }
          },
        },
      ],
    );
  };

  const handleNetworkSettings = () => {
    Alert.alert('Info', 'Network settings coming soon');
  };

  const handleAbout = () => {
    Alert.alert(
      'About Aura Wallet',
      'Version 1.0.0\n\nSecure mobile wallet for Aura blockchain',
    );
  };

  const formatAddress = address => {
    if (!address) {
      return '';
    }
    return `${address.substring(0, 10)}...${address.substring(address.length - 8)}`;
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.content}>
        <View style={styles.walletCard}>
          <Text style={styles.walletName}>{walletInfo?.name || 'My Wallet'}</Text>
          <Text style={styles.walletAddress}>
            {formatAddress(walletInfo?.address)}
          </Text>
          {walletInfo?.createdAt && (
            <Text style={styles.walletDate}>
              Created: {new Date(walletInfo.createdAt).toLocaleDateString()}
            </Text>
          )}
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Security</Text>

          {biometricAvailable && (
            <View style={styles.settingRow}>
              <View style={styles.settingInfo}>
                <Text style={styles.settingLabel}>
                  {BiometricAuth.getBiometryTypeName()} Authentication
                </Text>
                <Text style={styles.settingDescription}>
                  Use biometrics to unlock wallet
                </Text>
              </View>
              <Switch
                value={biometricEnabled}
                onValueChange={handleToggleBiometric}
                trackColor={{false: '#333', true: '#4A90E2'}}
                thumbColor={biometricEnabled ? '#fff' : '#f4f3f4'}
              />
            </View>
          )}

          <TouchableOpacity
            style={styles.settingRow}
            onPress={handleViewRecoveryPhrase}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>View Recovery Phrase</Text>
              <Text style={styles.settingDescription}>
                Show your backup recovery phrase
              </Text>
            </View>
            <Text style={styles.settingArrow}>›</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.settingRow}
            onPress={handleChangePassword}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>Change Password</Text>
              <Text style={styles.settingDescription}>
                Update your wallet password
              </Text>
            </View>
            <Text style={styles.settingArrow}>›</Text>
          </TouchableOpacity>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Preferences</Text>

          <View style={styles.settingRow}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>Dark Theme</Text>
              <Text style={styles.settingDescription}>
                Use dark color scheme
              </Text>
            </View>
            <Switch
              value={theme === 'dark'}
              onValueChange={handleToggleTheme}
              trackColor={{false: '#333', true: '#4A90E2'}}
              thumbColor={theme === 'dark' ? '#fff' : '#f4f3f4'}
            />
          </View>

          <View style={styles.settingRow}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>Notifications</Text>
              <Text style={styles.settingDescription}>
                Receive transaction notifications
              </Text>
            </View>
            <Switch
              value={notifications}
              onValueChange={handleToggleNotifications}
              trackColor={{false: '#333', true: '#4A90E2'}}
              thumbColor={notifications ? '#fff' : '#f4f3f4'}
            />
          </View>

          <TouchableOpacity
            style={styles.settingRow}
            onPress={handleNetworkSettings}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>Network</Text>
              <Text style={styles.settingDescription}>
                Configure RPC endpoints
              </Text>
            </View>
            <Text style={styles.settingArrow}>›</Text>
          </TouchableOpacity>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>About</Text>

          <TouchableOpacity style={styles.settingRow} onPress={handleAbout}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>About Aura Wallet</Text>
              <Text style={styles.settingDescription}>Version 1.0.0</Text>
            </View>
            <Text style={styles.settingArrow}>›</Text>
          </TouchableOpacity>
        </View>

        <TouchableOpacity
          style={styles.deleteButton}
          onPress={handleDeleteWallet}>
          <Text style={styles.deleteButtonText}>Delete Wallet</Text>
        </TouchableOpacity>
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
    paddingTop: 30,
  },
  walletCard: {
    backgroundColor: '#1a1a1a',
    padding: 20,
    borderRadius: 12,
    marginBottom: 30,
    borderWidth: 1,
    borderColor: '#333',
  },
  walletName: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 5,
  },
  walletAddress: {
    fontSize: 14,
    color: '#888',
    fontFamily: 'monospace',
    marginBottom: 8,
  },
  walletDate: {
    fontSize: 12,
    color: '#666',
  },
  section: {
    marginBottom: 30,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 15,
  },
  settingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#1a1a1a',
    padding: 16,
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  settingInfo: {
    flex: 1,
    marginRight: 12,
  },
  settingLabel: {
    fontSize: 16,
    color: '#fff',
    marginBottom: 4,
  },
  settingDescription: {
    fontSize: 13,
    color: '#666',
  },
  settingArrow: {
    fontSize: 24,
    color: '#666',
  },
  deleteButton: {
    backgroundColor: '#ff4444',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 20,
    marginBottom: 40,
  },
  deleteButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
});

export default SettingsScreen;

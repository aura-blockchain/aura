/**
 * QR Scanner Component
 * Camera-based QR code scanner for addresses and transaction data
 * Note: Requires react-native-camera configuration
 */

import React, {useState} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  Modal,
  Dimensions,
} from 'react-native';
import {RNCamera} from 'react-native-camera';

const {width, height} = Dimensions.get('window');

function QRScanner({visible, onClose, onScan}) {
  const [scanned, setScanned] = useState(false);
  const [cameraReady, setCameraReady] = useState(false);

  const handleBarCodeRead = ({data, type}) => {
    if (scanned) {
      return;
    }

    setScanned(true);

    // Validate scanned data
    if (isValidAddress(data)) {
      onScan(data);
      onClose();
    } else {
      Alert.alert('Invalid QR Code', 'This does not appear to be a valid Aura address', [
        {
          text: 'OK',
          onPress: () => setScanned(false),
        },
      ]);
    }
  };

  const isValidAddress = address => {
    // Basic validation for Aura addresses
    return (
      typeof address === 'string' &&
      address.startsWith('aura') &&
      address.length >= 39
    );
  };

  const handleCameraReady = () => {
    setCameraReady(true);
  };

  const handleClose = () => {
    setScanned(false);
    onClose();
  };

  const handleToggleFlash = () => {
    Alert.alert('Info', 'Flash toggle functionality');
  };

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent={false}
      onRequestClose={handleClose}>
      <View style={styles.container}>
        {visible && (
          <RNCamera
            style={styles.camera}
            type={RNCamera.Constants.Type.back}
            flashMode={RNCamera.Constants.FlashMode.off}
            onBarCodeRead={handleBarCodeRead}
            onCameraReady={handleCameraReady}
            barCodeTypes={[RNCamera.Constants.BarCodeType.qr]}
            captureAudio={false}>
            <View style={styles.overlay}>
              <View style={styles.header}>
                <TouchableOpacity style={styles.closeButton} onPress={handleClose}>
                  <Text style={styles.closeButtonText}>✕</Text>
                </TouchableOpacity>
                <Text style={styles.title}>Scan QR Code</Text>
                <TouchableOpacity
                  style={styles.flashButton}
                  onPress={handleToggleFlash}>
                  <Text style={styles.flashButtonText}>⚡</Text>
                </TouchableOpacity>
              </View>

              <View style={styles.scanArea}>
                <View style={styles.scanFrame}>
                  <View style={[styles.corner, styles.cornerTopLeft]} />
                  <View style={[styles.corner, styles.cornerTopRight]} />
                  <View style={[styles.corner, styles.cornerBottomLeft]} />
                  <View style={[styles.corner, styles.cornerBottomRight]} />
                </View>
              </View>

              <View style={styles.instructions}>
                <Text style={styles.instructionsText}>
                  Position the QR code within the frame
                </Text>
                {!cameraReady && (
                  <Text style={styles.loadingText}>Initializing camera...</Text>
                )}
              </View>
            </View>
          </RNCamera>
        )}

        {!visible && (
          <View style={styles.fallback}>
            <Text style={styles.fallbackText}>Camera not available</Text>
            <TouchableOpacity style={styles.fallbackButton} onPress={handleClose}>
              <Text style={styles.fallbackButtonText}>Close</Text>
            </TouchableOpacity>
          </View>
        )}
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  camera: {
    flex: 1,
  },
  overlay: {
    flex: 1,
    backgroundColor: 'transparent',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    paddingTop: 50,
  },
  closeButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  closeButtonText: {
    color: '#fff',
    fontSize: 24,
    fontWeight: 'bold',
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  flashButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  flashButtonText: {
    fontSize: 20,
  },
  scanArea: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  scanFrame: {
    width: width * 0.7,
    height: width * 0.7,
    position: 'relative',
  },
  corner: {
    position: 'absolute',
    width: 40,
    height: 40,
    borderColor: '#4A90E2',
    borderWidth: 4,
  },
  cornerTopLeft: {
    top: 0,
    left: 0,
    borderRightWidth: 0,
    borderBottomWidth: 0,
  },
  cornerTopRight: {
    top: 0,
    right: 0,
    borderLeftWidth: 0,
    borderBottomWidth: 0,
  },
  cornerBottomLeft: {
    bottom: 0,
    left: 0,
    borderRightWidth: 0,
    borderTopWidth: 0,
  },
  cornerBottomRight: {
    bottom: 0,
    right: 0,
    borderLeftWidth: 0,
    borderTopWidth: 0,
  },
  instructions: {
    padding: 30,
    alignItems: 'center',
  },
  instructionsText: {
    fontSize: 16,
    color: '#fff',
    textAlign: 'center',
    marginBottom: 10,
  },
  loadingText: {
    fontSize: 14,
    color: '#888',
  },
  fallback: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  fallbackText: {
    fontSize: 16,
    color: '#fff',
    marginBottom: 20,
  },
  fallbackButton: {
    backgroundColor: '#4A90E2',
    padding: 16,
    borderRadius: 8,
  },
  fallbackButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
});

export default QRScanner;

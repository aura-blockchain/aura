import React, { useState, useEffect } from 'react';
import { hardwareWalletService } from '../services/hardwareWallet';

const HardwareWallet = ({ onAddressImported }) => {
  const [deviceType, setDeviceType] = useState('ledger');
  const [connecting, setConnecting] = useState(false);
  const [connected, setConnected] = useState(false);
  const [deviceInfo, setDeviceInfo] = useState(null);
  const [address, setAddress] = useState(null);
  const [error, setError] = useState('');
  const [derivationPath, setDerivationPath] = useState("m/44'/118'/0'/0/0");

  useEffect(() => {
    const unsubscribe = hardwareWalletService.onConnectionChange((device) => {
      setConnected(!!device);
      setDeviceInfo(device);
      if (!device) {
        setAddress(null);
      }
    });

    // Check if already connected
    if (hardwareWalletService.isConnected()) {
      setConnected(true);
      setDeviceInfo(hardwareWalletService.getDeviceInfo());
    }

    return unsubscribe;
  }, []);

  const handleConnect = async () => {
    setConnecting(true);
    setError('');

    try {
      await hardwareWalletService.connect(deviceType);
      const addrInfo = await hardwareWalletService.getAddress(derivationPath, false);
      setAddress(addrInfo);
    } catch (err) {
      setError(err.message || 'Failed to connect');
      setConnected(false);
      setDeviceInfo(null);
    } finally {
      setConnecting(false);
    }
  };

  const handleDisconnect = async () => {
    try {
      await hardwareWalletService.disconnect();
      setAddress(null);
    } catch (err) {
      setError(err.message || 'Failed to disconnect');
    }
  };

  const handleVerifyAddress = async () => {
    setError('');
    try {
      const verified = await hardwareWalletService.verifyAddress(derivationPath);
      if (verified) {
        setError('');
      }
    } catch (err) {
      setError(err.message || 'Verification failed');
    }
  };

  const handleImportAddress = () => {
    if (address && onAddressImported) {
      onAddressImported({
        address: address.address,
        publicKey: address.publicKey,
        type: 'hardware',
        device: deviceType,
        path: derivationPath
      });
    }
  };

  return (
    <div className="card">
      <h3 className="card-header">Hardware Wallet</h3>

      {error && (
        <div style={{
          padding: '12px',
          background: 'rgba(247, 118, 142, 0.1)',
          border: '1px solid var(--error)',
          borderRadius: '6px',
          marginBottom: '20px',
          color: 'var(--error)',
          fontSize: '14px'
        }}>
          {error}
        </div>
      )}

      {!connected ? (
        <>
          <p className="text-muted mb-20">
            Connect your hardware wallet to sign transactions securely.
          </p>

          <div className="form-group">
            <label className="form-label">Device Type</label>
            <select
              className="form-input"
              value={deviceType}
              onChange={(e) => setDeviceType(e.target.value)}
              disabled={connecting}
            >
              <option value="ledger">Ledger</option>
              <option value="trezor">Trezor</option>
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Derivation Path</label>
            <input
              type="text"
              className="form-input"
              value={derivationPath}
              onChange={(e) => setDerivationPath(e.target.value)}
              disabled={connecting}
              placeholder="m/44'/118'/0'/0/0"
            />
            <small className="text-muted" style={{ fontSize: '12px', marginTop: '4px', display: 'block' }}>
              Default: m/44'/118'/0'/0/0 (Cosmos/Aura)
            </small>
          </div>

          {deviceType === 'ledger' && (
            <div style={{
              padding: '12px',
              background: 'var(--bg-primary)',
              borderRadius: '6px',
              marginBottom: '20px',
              fontSize: '13px'
            }}>
              <strong>Instructions:</strong>
              <ol style={{ marginTop: '8px', paddingLeft: '20px', lineHeight: '1.6' }}>
                <li>Connect your Ledger device via USB</li>
                <li>Enter your PIN to unlock</li>
                <li>Open the Cosmos app on the device</li>
                <li>Click "Connect" below</li>
              </ol>
            </div>
          )}

          {deviceType === 'trezor' && (
            <div style={{
              padding: '12px',
              background: 'var(--bg-primary)',
              borderRadius: '6px',
              marginBottom: '20px',
              fontSize: '13px'
            }}>
              <strong>Instructions:</strong>
              <ol style={{ marginTop: '8px', paddingLeft: '20px', lineHeight: '1.6' }}>
                <li>Connect your Trezor device via USB</li>
                <li>A browser window will open for Trezor Bridge</li>
                <li>Follow the prompts to authorize the connection</li>
              </ol>
            </div>
          )}

          <button
            className="btn btn-primary"
            onClick={handleConnect}
            disabled={connecting}
          >
            {connecting ? 'Connecting...' : 'Connect'}
          </button>
        </>
      ) : (
        <>
          <div style={{
            padding: '16px',
            background: 'rgba(158, 206, 106, 0.1)',
            border: '1px solid var(--success)',
            borderRadius: '6px',
            marginBottom: '20px'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
              <span style={{
                width: '10px',
                height: '10px',
                borderRadius: '50%',
                background: 'var(--success)'
              }}></span>
              <strong style={{ color: 'var(--success)' }}>
                {deviceInfo?.name || 'Device'} Connected
              </strong>
            </div>

            {address && (
              <div style={{ fontSize: '13px' }}>
                <div style={{ marginBottom: '8px' }}>
                  <span className="text-muted">Address: </span>
                  <code style={{
                    background: 'var(--bg-primary)',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    wordBreak: 'break-all'
                  }}>
                    {address.address}
                  </code>
                </div>
                <div>
                  <span className="text-muted">Path: </span>
                  <code style={{
                    background: 'var(--bg-primary)',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontSize: '12px'
                  }}>
                    {derivationPath}
                  </code>
                </div>
              </div>
            )}
          </div>

          <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
            <button
              className="btn btn-secondary"
              onClick={handleVerifyAddress}
            >
              Verify on Device
            </button>
            {onAddressImported && address && (
              <button
                className="btn btn-primary"
                onClick={handleImportAddress}
              >
                Use This Address
              </button>
            )}
            <button
              className="btn"
              onClick={handleDisconnect}
              style={{
                background: 'var(--bg-tertiary)',
                color: 'var(--text-secondary)'
              }}
            >
              Disconnect
            </button>
          </div>
        </>
      )}
    </div>
  );
};

export default HardwareWallet;

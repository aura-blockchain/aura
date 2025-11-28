/**
 * Hardware Wallet Support for Aura Browser Extension
 * Supports Ledger and Trezor devices
 */

class HardwareWalletManager {
  constructor() {
    this.supportedDevices = {
      ledger: {
        name: 'Ledger',
        productId: 0x0001,
        vendorId: 0x2c97
      },
      trezor: {
        name: 'Trezor',
        productId: 0x0001,
        vendorId: 0x534c
      }
    };
    this.connectedDevice = null;
    this.transport = null;
  }

  async detectDevices() {
    if (!navigator.usb) {
      throw new Error('WebUSB not supported in this browser');
    }

    try {
      const devices = await navigator.usb.getDevices();
      const supported = devices.filter(device =>
        Object.values(this.supportedDevices).some(hw =>
          hw.vendorId === device.vendorId
        )
      );

      return supported.map(device => ({
        device,
        type: this.getDeviceType(device),
        connected: true
      }));
    } catch (err) {
      console.error('Failed to detect devices:', err);
      throw new Error('Failed to detect hardware wallets');
    }
  }

  getDeviceType(device) {
    for (const [type, hw] of Object.entries(this.supportedDevices)) {
      if (hw.vendorId === device.vendorId) {
        return type;
      }
    }
    return 'unknown';
  }

  async requestDevice(type = null) {
    if (!navigator.usb) {
      throw new Error('WebUSB not supported');
    }

    try {
      const filters = type
        ? [{ vendorId: this.supportedDevices[type].vendorId }]
        : Object.values(this.supportedDevices).map(hw => ({
            vendorId: hw.vendorId
          }));

      const device = await navigator.usb.requestDevice({ filters });
      return device;
    } catch (err) {
      console.error('Failed to request device:', err);
      throw new Error('User cancelled or device not found');
    }
  }

  async connect(device) {
    try {
      await device.open();

      if (device.configuration === null) {
        await device.selectConfiguration(1);
      }

      await device.claimInterface(0);

      this.connectedDevice = {
        device,
        type: this.getDeviceType(device)
      };

      return this.connectedDevice;
    } catch (err) {
      console.error('Failed to connect to device:', err);
      throw new Error('Failed to connect to hardware wallet');
    }
  }

  async disconnect() {
    if (this.connectedDevice) {
      try {
        await this.connectedDevice.device.close();
        this.connectedDevice = null;
      } catch (err) {
        console.error('Error disconnecting:', err);
      }
    }
  }

  async getAddress(path = "m/44'/118'/0'/0/0") {
    if (!this.connectedDevice) {
      throw new Error('No device connected');
    }

    const type = this.connectedDevice.type;

    switch (type) {
      case 'ledger':
        return await this.getLedgerAddress(path);
      case 'trezor':
        return await this.getTrezorAddress(path);
      default:
        throw new Error('Unsupported device type');
    }
  }

  async getLedgerAddress(path) {
    try {
      // Implement Ledger-specific address derivation
      // This is a simplified version
      const pathBuffer = this.serializePath(path);
      const command = this.buildAPDU(0xe0, 0x02, 0x00, 0x00, pathBuffer);

      const response = await this.sendCommand(command);

      if (response.statusCode !== 0x9000) {
        throw new Error('Ledger returned error');
      }

      return this.parseAddressResponse(response.data);
    } catch (err) {
      console.error('Failed to get Ledger address:', err);
      throw new Error('Failed to retrieve address from Ledger');
    }
  }

  async getTrezorAddress(path) {
    try {
      // Implement Trezor-specific address derivation
      // This would use the Trezor Connect library
      throw new Error('Trezor support not yet implemented');
    } catch (err) {
      console.error('Failed to get Trezor address:', err);
      throw err;
    }
  }

  async signTransaction(tx, path = "m/44'/118'/0'/0/0") {
    if (!this.connectedDevice) {
      throw new Error('No device connected');
    }

    const type = this.connectedDevice.type;

    switch (type) {
      case 'ledger':
        return await this.signWithLedger(tx, path);
      case 'trezor':
        return await this.signWithTrezor(tx, path);
      default:
        throw new Error('Unsupported device type');
    }
  }

  async signWithLedger(tx, path) {
    try {
      const pathBuffer = this.serializePath(path);
      const txBuffer = this.serializeTransaction(tx);

      // Send transaction in chunks if needed
      const chunks = this.chunkBuffer(txBuffer, 255);
      let response;

      for (let i = 0; i < chunks.length; i++) {
        const isLast = i === chunks.length - 1;
        const p1 = i === 0 ? 0x00 : 0x01;
        const p2 = isLast ? 0x00 : 0x01;

        const data = i === 0
          ? Buffer.concat([pathBuffer, chunks[i]])
          : chunks[i];

        const command = this.buildAPDU(0xe0, 0x04, p1, p2, data);
        response = await this.sendCommand(command);

        if (response.statusCode !== 0x9000 && !isLast) {
          throw new Error('Ledger signing failed');
        }
      }

      if (response.statusCode !== 0x9000) {
        throw new Error('Transaction rejected by user');
      }

      return this.parseSignatureResponse(response.data);
    } catch (err) {
      console.error('Failed to sign with Ledger:', err);
      throw new Error('Failed to sign transaction');
    }
  }

  async signWithTrezor(tx, path) {
    throw new Error('Trezor support not yet implemented');
  }

  serializePath(path) {
    const parts = path.split('/').slice(1);
    const buffer = Buffer.alloc(1 + parts.length * 4);

    buffer.writeUInt8(parts.length, 0);

    for (let i = 0; i < parts.length; i++) {
      let value = parseInt(parts[i].replace("'", ''));
      if (parts[i].includes("'")) {
        value += 0x80000000;
      }
      buffer.writeUInt32BE(value, 1 + i * 4);
    }

    return buffer;
  }

  serializeTransaction(tx) {
    return Buffer.from(JSON.stringify(tx), 'utf8');
  }

  chunkBuffer(buffer, size) {
    const chunks = [];
    for (let i = 0; i < buffer.length; i += size) {
      chunks.push(buffer.slice(i, i + size));
    }
    return chunks;
  }

  buildAPDU(cla, ins, p1, p2, data) {
    const header = Buffer.from([cla, ins, p1, p2, data.length]);
    return Buffer.concat([header, data]);
  }

  async sendCommand(command) {
    const device = this.connectedDevice.device;

    try {
      // Send command
      await device.transferOut(1, command);

      // Receive response
      const result = await device.transferIn(1, 256);

      if (result.status !== 'ok') {
        throw new Error('Transfer failed');
      }

      const data = Buffer.from(result.data.buffer);
      const statusCode = data.readUInt16BE(data.length - 2);

      return {
        data: data.slice(0, data.length - 2),
        statusCode
      };
    } catch (err) {
      console.error('Command failed:', err);
      throw new Error('Failed to communicate with device');
    }
  }

  parseAddressResponse(data) {
    // Parse Ledger address response
    const publicKeyLength = data[0];
    const publicKey = data.slice(1, 1 + publicKeyLength);
    const addressLength = data[1 + publicKeyLength];
    const address = data.slice(2 + publicKeyLength, 2 + publicKeyLength + addressLength).toString();

    return {
      address,
      publicKey: publicKey.toString('hex')
    };
  }

  parseSignatureResponse(data) {
    // Parse Ledger signature response
    return {
      signature: data.toString('hex')
    };
  }

  async verifyAddress(address, path) {
    // Display address on device for verification
    if (!this.connectedDevice) {
      throw new Error('No device connected');
    }

    const pathBuffer = this.serializePath(path);
    const command = this.buildAPDU(0xe0, 0x02, 0x01, 0x00, pathBuffer);

    const response = await this.sendCommand(command);

    if (response.statusCode !== 0x9000) {
      throw new Error('Address verification failed');
    }

    return true;
  }

  isConnected() {
    return this.connectedDevice !== null;
  }

  getDeviceInfo() {
    if (!this.connectedDevice) {
      return null;
    }

    return {
      type: this.connectedDevice.type,
      name: this.supportedDevices[this.connectedDevice.type].name,
      manufacturer: this.connectedDevice.device.manufacturerName,
      product: this.connectedDevice.device.productName
    };
  }
}

// Export for use in extension
if (typeof module !== 'undefined' && module.exports) {
  module.exports = HardwareWalletManager;
}

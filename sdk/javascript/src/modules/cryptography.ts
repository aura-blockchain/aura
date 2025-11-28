import { AuraClient } from '../client';

export class CryptographyModule {
  constructor(private client: AuraClient) {}

  /**
   * Generate encryption key
   */
  async generateKey(params: {
    keyType: string;
    purpose: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.cryptography.v1beta1.MsgGenerateKey',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Rotate encryption key
   */
  async rotateKey(params: {
    oldKeyId: string;
    newKeyType: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.cryptography.v1beta1.MsgRotateKey',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query public key
   */
  async queryPublicKey(keyId: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(keyId, {
      get_public_key: { key_id: keyId }
    });
  }
}

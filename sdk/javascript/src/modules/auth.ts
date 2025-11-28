import { AuraClient } from '../client';

export class AuthModule {
  constructor(private client: AuraClient) {}

  /**
   * Register a new identity on-chain
   */
  async registerIdentity(params: {
    address: string;
    identityData: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.auth.v1beta1.MsgRegisterIdentity',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Update identity information
   */
  async updateIdentity(params: {
    address: string;
    identityData: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.auth.v1beta1.MsgUpdateIdentity',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query identity by address
   */
  async queryIdentity(address: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(address, {
      get_identity: { address }
    });
  }

  /**
   * Verify identity credentials
   */
  async verifyIdentity(address: string): Promise<boolean> {
    try {
      const identity = await this.queryIdentity(address);
      return identity && identity.verified === true;
    } catch {
      return false;
    }
  }
}

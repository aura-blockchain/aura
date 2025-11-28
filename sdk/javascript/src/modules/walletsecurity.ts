import { AuraClient } from '../client';

export class WalletSecurityModule {
  constructor(private client: AuraClient) {}

  /**
   * Enable multi-signature protection
   */
  async enableMultisig(params: {
    address: string;
    signers: string[];
    threshold: number;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.walletsecurity.v1beta1.MsgEnableMultisig',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Set spending limits
   */
  async setSpendingLimits(params: {
    address: string;
    dailyLimit: string;
    transactionLimit: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.walletsecurity.v1beta1.MsgSetSpendingLimits',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query wallet security settings
   */
  async querySecuritySettings(address: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(address, {
      get_security_settings: { address }
    });
  }

  /**
   * Enable biometric authentication
   */
  async enableBiometric(params: {
    address: string;
    biometricHash: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.walletsecurity.v1beta1.MsgEnableBiometric',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }
}

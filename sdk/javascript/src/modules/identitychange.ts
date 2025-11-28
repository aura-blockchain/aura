import { AuraClient } from '../client';

export class IdentityChangeModule {
  constructor(private client: AuraClient) {}

  /**
   * Request identity change
   */
  async requestChange(params: {
    address: string;
    changeType: string;
    newData: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.identitychange.v1beta1.MsgRequestChange',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Approve identity change
   */
  async approveChange(params: {
    requestId: string;
    approver: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.identitychange.v1beta1.MsgApproveChange',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query pending changes
   */
  async queryPendingChanges(address: string): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }
}

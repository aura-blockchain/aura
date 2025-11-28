import { AuraClient } from '../client';

export class DataRegistryModule {
  constructor(private client: AuraClient) {}

  /**
   * Register data on-chain
   */
  async registerData(params: {
    dataHash: string;
    metadata: any;
    owner: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.dataregistry.v1beta1.MsgRegisterData',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Update data metadata
   */
  async updateData(params: {
    dataId: string;
    metadata: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.dataregistry.v1beta1.MsgUpdateData',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query data by ID
   */
  async queryData(dataId: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(dataId, {
      get_data: { data_id: dataId }
    });
  }

  /**
   * Verify data integrity
   */
  async verifyData(dataId: string, hash: string): Promise<boolean> {
    const data = await this.queryData(dataId);
    return data.hash === hash;
  }
}

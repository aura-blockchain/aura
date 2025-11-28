import { AuraClient } from '../client';

export class MonitoringModule {
  constructor(private client: AuraClient) {}

  /**
   * Query system metrics
   */
  async queryMetrics(): Promise<any> {
    const client = this.client.getClient();
    return {};
  }

  /**
   * Query alerts
   */
  async queryAlerts(): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }

  /**
   * Create alert rule
   */
  async createAlertRule(params: {
    name: string;
    condition: any;
    action: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.monitoring.v1beta1.MsgCreateAlertRule',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }
}

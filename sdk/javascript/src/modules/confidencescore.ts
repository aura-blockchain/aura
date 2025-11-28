import { AuraClient } from '../client';

export class ConfidenceScoreModule {
  constructor(private client: AuraClient) {}

  /**
   * Query confidence score for an address
   */
  async queryScore(address: string): Promise<number> {
    const client = this.client.getClient();
    const result = await client.queryContractSmart(address, {
      get_confidence_score: { address }
    });
    return result.score || 0;
  }

  /**
   * Update confidence score
   */
  async updateScore(params: {
    address: string;
    score: number;
    reason: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.confidencescore.v1beta1.MsgUpdateScore',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query score history
   */
  async queryScoreHistory(address: string): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }
}

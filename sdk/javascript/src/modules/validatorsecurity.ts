import { AuraClient } from '../client';

export class ValidatorSecurityModule {
  constructor(private client: AuraClient) {}

  /**
   * Query validator security status
   */
  async queryValidatorSecurity(validatorAddress: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(validatorAddress, {
      get_security_status: { validator: validatorAddress }
    });
  }

  /**
   * Report validator misbehavior
   */
  async reportMisbehavior(params: {
    reporter: string;
    validator: string;
    evidence: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.validatorsecurity.v1beta1.MsgReportMisbehavior',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query slashing history
   */
  async querySlashingHistory(validatorAddress: string): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }
}

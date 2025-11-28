import { AuraClient } from '../client';

export class InclusionRoutinesModule {
  constructor(private client: AuraClient) {}

  /**
   * Create inclusion routine
   */
  async createRoutine(params: {
    name: string;
    description: string;
    criteria: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.inclusionroutines.v1beta1.MsgCreateRoutine',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Execute inclusion routine
   */
  async executeRoutine(params: {
    routineId: string;
    params: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.inclusionroutines.v1beta1.MsgExecuteRoutine',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Query routine by ID
   */
  async queryRoutine(routineId: string): Promise<any> {
    const client = this.client.getClient();
    return await client.queryContractSmart(routineId, {
      get_routine: { routine_id: routineId }
    });
  }

  /**
   * List all routines
   */
  async listRoutines(): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }
}

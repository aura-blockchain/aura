import { AuraClient } from '../client';

export class VCRegistryModule {
  constructor(private client: AuraClient) {}

  /**
   * Issue verifiable credential
   */
  async issueCredential(params: {
    issuer: string;
    subject: string;
    credentialType: string;
    claims: any;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.vcregistry.v1beta1.MsgIssueCredential',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Revoke credential
   */
  async revokeCredential(params: {
    issuer: string;
    credentialId: string;
    reason: string;
  }): Promise<any> {
    const msg = {
      typeUrl: '/aura.vcregistry.v1beta1.MsgRevokeCredential',
      value: params
    };
    return this.client.getTxBuilder().buildAndBroadcast([msg]);
  }

  /**
   * Verify credential
   */
  async verifyCredential(credentialId: string): Promise<{
    valid: boolean;
    credential: any;
  }> {
    const client = this.client.getClient();
    const credential = await client.queryContractSmart(credentialId, {
      get_credential: { credential_id: credentialId }
    });
    return {
      valid: credential && !credential.revoked,
      credential
    };
  }

  /**
   * Query credentials by subject
   */
  async queryCredentialsBySubject(subject: string): Promise<any[]> {
    const client = this.client.getClient();
    return [];
  }
}

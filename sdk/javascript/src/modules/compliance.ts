import { AuraClient } from '../client';
import {
  ComplianceStatus,
  SubmitKYCParams,
  TransactionReport,
  ComplianceParams,
  SanctionCheckResult,
  TxResult,
  GasOptions,
} from '../types';
import axios from 'axios';

/**
 * Compliance Module
 * Handles KYC, AML, sanctions checking, and regulatory compliance
 */
export class ComplianceModule {
  constructor(private client: AuraClient) {}

  /**
   * Get compliance status for an address
   * @param address - Address to check
   * @returns Compliance status
   */
  async getComplianceStatus(address: string): Promise<ComplianceStatus> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/compliance/v1beta1/status/${address}`
      );

      const data = response.data.status;
      return {
        address: data.address,
        status: data.status,
        kycLevel: data.kyc_level,
        verifiedAt: data.verified_at ? new Date(data.verified_at) : undefined,
        expiresAt: data.expires_at ? new Date(data.expires_at) : undefined,
        country: data.country,
        riskScore: data.risk_score,
        sanctioned: data.sanctioned,
        flags: data.flags || [],
      };
    } catch (error) {
      throw new Error(`Failed to get compliance status: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Submit KYC information
   * @param params - KYC submission parameters
   * @param options - Transaction options
   * @returns Transaction result
   */
  async submitKYC(
    params: SubmitKYCParams,
    options?: GasOptions
  ): Promise<TxResult> {
    try {
      const msg = {
        typeUrl: '/aura.compliance.v1beta1.MsgSubmitKYC',
        value: {
          address: params.address,
          level: params.level,
          personalInfo: params.personalInfo,
          documents: params.documents,
          residenceInfo: params.residenceInfo,
        },
      };

      const txBuilder = this.client.getTxBuilder();
      return await txBuilder.signAndBroadcast(params.address, [msg], options);
    } catch (error) {
      throw new Error(`Failed to submit KYC: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Query transaction report
   * @param txHash - Transaction hash
   * @returns Transaction report
   */
  async queryTransactionReport(txHash: string): Promise<TransactionReport> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/compliance/v1beta1/report/${txHash}`
      );

      const data = response.data.report;
      return {
        txHash: data.tx_hash,
        sender: data.sender,
        recipient: data.recipient,
        amount: data.amount,
        denom: data.denom,
        timestamp: new Date(data.timestamp),
        complianceChecks: data.compliance_checks,
        flags: data.flags || [],
        approved: data.approved,
      };
    } catch (error) {
      throw new Error(`Failed to query transaction report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Check if address is sanctioned
   * @param address - Address to check
   * @returns Sanction check result
   */
  async checkSanctions(address: string): Promise<SanctionCheckResult> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/compliance/v1beta1/sanctions/${address}`
      );

      const data = response.data.result;
      return {
        address: data.address,
        sanctioned: data.sanctioned,
        lists: data.lists || [],
        details: data.details,
        checkedAt: new Date(data.checked_at),
      };
    } catch (error) {
      throw new Error(`Failed to check sanctions: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get compliance parameters
   * @returns Compliance parameters
   */
  async getComplianceParams(): Promise<ComplianceParams> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/compliance/v1beta1/params`
      );

      return response.data.params;
    } catch (error) {
      throw new Error(`Failed to get compliance params: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

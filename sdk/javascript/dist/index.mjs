// src/client.ts
import { SigningStargateClient, StargateClient, GasPrice } from "@cosmjs/stargate";
import { Tendermint34Client } from "@cosmjs/tendermint-rpc";

// src/tx.ts
var TxBuilder = class {
  constructor(client, gasPrice = "0.025uaura", gasAdjustment = 1.5) {
    this.client = client;
    this.defaultGasPrice = gasPrice;
    this.defaultGasAdjustment = gasAdjustment;
  }
  /**
   * Sign and broadcast a transaction
   */
  async signAndBroadcast(signerAddress, messages, options) {
    const fee = await this.calculateFee(signerAddress, messages, options);
    const memo = options?.memo || "";
    const result = await this.client.signAndBroadcast(
      signerAddress,
      messages,
      fee,
      memo
    );
    return this.formatResult(result);
  }
  /**
   * Calculate transaction fee
   */
  async calculateFee(signerAddress, messages, options) {
    const gasPrice = options?.gasPrice || this.defaultGasPrice;
    if (options?.gasLimit) {
      return {
        amount: [{ denom: this.extractDenom(gasPrice), amount: this.calculateAmount(options.gasLimit, gasPrice) }],
        gas: options.gasLimit.toString()
      };
    }
    const gasEstimate = await this.client.simulate(signerAddress, messages, "");
    const gasLimit = Math.round(gasEstimate * this.defaultGasAdjustment);
    return {
      amount: [{ denom: this.extractDenom(gasPrice), amount: this.calculateAmount(gasLimit, gasPrice) }],
      gas: gasLimit.toString()
    };
  }
  /**
   * Simulate transaction
   */
  async simulate(signerAddress, messages) {
    return await this.client.simulate(signerAddress, messages, "");
  }
  /**
   * Extract denom from gas price string
   */
  extractDenom(gasPrice) {
    const match = gasPrice.match(/[a-z]+$/);
    return match ? match[0] : "uaura";
  }
  /**
   * Calculate fee amount
   */
  calculateAmount(gasLimit, gasPrice) {
    const priceValue = parseFloat(gasPrice.replace(/[a-z]+$/, ""));
    return Math.ceil(gasLimit * priceValue).toString();
  }
  /**
   * Format transaction result
   */
  formatResult(result) {
    return {
      transactionHash: result.transactionHash,
      height: result.height,
      code: result.code,
      rawLog: result.rawLog,
      gasUsed: Number(result.gasUsed),
      gasWanted: Number(result.gasWanted)
    };
  }
};

// src/modules/bank.ts
var BankModule = class {
  constructor(client) {
    this.client = client;
  }
  /**
   * Get account balance for a specific denom
   */
  async getBalance(address, denom) {
    const client = this.client.getClient();
    return await client.getBalance(address, denom);
  }
  /**
   * Get all account balances
   */
  async getAllBalances(address) {
    const client = this.client.getClient();
    return await client.getAllBalances(address);
  }
  /**
   * Send tokens to another address
   */
  async send(senderAddress, params, options) {
    const denom = params.denom || "uaura";
    const message = {
      typeUrl: "/cosmos.bank.v1beta1.MsgSend",
      value: {
        fromAddress: senderAddress,
        toAddress: params.recipient,
        amount: [{ denom, amount: params.amount }]
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(senderAddress, [message], {
      ...options,
      memo: params.memo
    });
  }
  /**
   * Multi-send tokens to multiple recipients
   */
  async multiSend(senderAddress, recipients, options) {
    const outputs = recipients.map((recipient) => ({
      address: recipient.address,
      coins: [{ denom: recipient.denom || "uaura", amount: recipient.amount }]
    }));
    const message = {
      typeUrl: "/cosmos.bank.v1beta1.MsgMultiSend",
      value: {
        inputs: [{
          address: senderAddress,
          coins: outputs.flatMap((o) => o.coins)
        }],
        outputs
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(senderAddress, [message], options);
  }
  /**
   * Get total supply of a denom
   */
  async getTotalSupply(denom) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/bank/v1beta1/supply/${denom}`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.amount || null;
    } catch (error) {
      console.error("Error fetching total supply:", error);
      return null;
    }
  }
  /**
   * Get all denoms
   */
  async getAllDenoms() {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/bank/v1beta1/supply`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return (data.supply || []).map((coin) => coin.denom);
    } catch (error) {
      console.error("Error fetching denoms:", error);
      return [];
    }
  }
  /**
   * Get spendable balance (available for spending)
   */
  async getSpendableBalance(address, denom) {
    return await this.getBalance(address, denom);
  }
  /**
   * Format balance for display
   */
  formatBalance(balance, decimals = 6) {
    const amount = parseInt(balance.amount) / Math.pow(10, decimals);
    return `${amount.toFixed(decimals)} ${balance.denom.replace("u", "").toUpperCase()}`;
  }
};

// src/modules/dex.ts
var DexModule = class {
  constructor(client) {
    this.client = client;
  }
  /**
   * Create a new liquidity pool
   */
  async createPool(creator, params, options) {
    const message = {
      typeUrl: "/paw.dex.v1.MsgCreatePool",
      value: {
        creator,
        tokenA: params.tokenA,
        tokenB: params.tokenB,
        amountA: params.amountA,
        amountB: params.amountB
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(creator, [message], options);
  }
  /**
   * Add liquidity to an existing pool
   */
  async addLiquidity(sender, params, options) {
    const message = {
      typeUrl: "/paw.dex.v1.MsgAddLiquidity",
      value: {
        sender,
        poolId: params.poolId,
        amountA: params.amountA,
        amountB: params.amountB,
        minShares: params.minShares
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(sender, [message], options);
  }
  /**
   * Remove liquidity from a pool
   */
  async removeLiquidity(sender, params, options) {
    const message = {
      typeUrl: "/paw.dex.v1.MsgRemoveLiquidity",
      value: {
        sender,
        poolId: params.poolId,
        shares: params.shares,
        minAmountA: params.minAmountA,
        minAmountB: params.minAmountB
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(sender, [message], options);
  }
  /**
   * Swap tokens
   */
  async swap(sender, params, options) {
    const message = {
      typeUrl: "/paw.dex.v1.MsgSwap",
      value: {
        sender,
        poolId: params.poolId,
        tokenIn: params.tokenIn,
        amountIn: params.amountIn,
        minAmountOut: params.minAmountOut,
        recipient: params.recipient || sender
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(sender, [message], options);
  }
  /**
   * Get pool by ID
   */
  async getPool(poolId) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/paw/dex/v1/pools/${poolId}`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.pool || null;
    } catch (error) {
      console.error("Error fetching pool:", error);
      return null;
    }
  }
  /**
   * Get all pools
   */
  async getAllPools() {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/paw/dex/v1/pools`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.pools || [];
    } catch (error) {
      console.error("Error fetching pools:", error);
      return [];
    }
  }
  /**
   * Get pool for token pair
   */
  async getPoolByTokens(tokenA, tokenB) {
    const pools = await this.getAllPools();
    return pools.find(
      (pool) => pool.tokenA === tokenA && pool.tokenB === tokenB || pool.tokenA === tokenB && pool.tokenB === tokenA
    ) || null;
  }
  /**
   * Calculate swap output amount
   */
  calculateSwapOutput(amountIn, reserveIn, reserveOut, swapFee = "0.003") {
    const amountInBig = BigInt(amountIn);
    const reserveInBig = BigInt(reserveIn);
    const reserveOutBig = BigInt(reserveOut);
    const feeBig = BigInt(Math.floor(parseFloat(swapFee) * 1e4));
    const amountInWithFee = amountInBig * (10000n - feeBig) / 10000n;
    const numerator = amountInWithFee * reserveOutBig;
    const denominator = reserveInBig + amountInWithFee;
    return (numerator / denominator).toString();
  }
  /**
   * Calculate price impact
   */
  calculatePriceImpact(amountIn, reserveIn, reserveOut) {
    const amountOut = this.calculateSwapOutput(amountIn, reserveIn, reserveOut, "0");
    const priceBeforeSwap = parseFloat(reserveOut) / parseFloat(reserveIn);
    const priceAfterSwap = parseFloat(amountOut) / parseFloat(amountIn);
    return Math.abs((priceAfterSwap - priceBeforeSwap) / priceBeforeSwap) * 100;
  }
  /**
   * Calculate shares for liquidity addition
   */
  calculateShares(amountA, amountB, reserveA, _reserveB, totalShares) {
    if (totalShares === "0") {
      const amountABig2 = BigInt(amountA);
      const amountBBig = BigInt(amountB);
      return (amountABig2 * amountBBig).toString();
    }
    const amountABig = BigInt(amountA);
    const reserveABig = BigInt(reserveA);
    const totalSharesBig = BigInt(totalShares);
    return (amountABig * totalSharesBig / reserveABig).toString();
  }
};

// src/modules/staking.ts
var StakingModule = class {
  constructor(client) {
    this.client = client;
  }
  /**
   * Delegate tokens to a validator
   */
  async delegate(delegator, params, options) {
    const denom = params.denom || "uaura";
    const message = {
      typeUrl: "/cosmos.staking.v1beta1.MsgDelegate",
      value: {
        delegatorAddress: delegator,
        validatorAddress: params.validatorAddress,
        amount: { denom, amount: params.amount }
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(delegator, [message], options);
  }
  /**
   * Undelegate tokens from a validator
   */
  async undelegate(delegator, params, options) {
    const denom = params.denom || "uaura";
    const message = {
      typeUrl: "/cosmos.staking.v1beta1.MsgUndelegate",
      value: {
        delegatorAddress: delegator,
        validatorAddress: params.validatorAddress,
        amount: { denom, amount: params.amount }
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(delegator, [message], options);
  }
  /**
   * Redelegate tokens from one validator to another
   */
  async redelegate(delegator, params, options) {
    const denom = params.denom || "uaura";
    const message = {
      typeUrl: "/cosmos.staking.v1beta1.MsgBeginRedelegate",
      value: {
        delegatorAddress: delegator,
        validatorSrcAddress: params.srcValidatorAddress,
        validatorDstAddress: params.dstValidatorAddress,
        amount: { denom, amount: params.amount }
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(delegator, [message], options);
  }
  /**
   * Withdraw delegation rewards from a validator
   */
  async withdrawRewards(delegator, validatorAddress, options) {
    const message = {
      typeUrl: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
      value: {
        delegatorAddress: delegator,
        validatorAddress
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(delegator, [message], options);
  }
  /**
   * Withdraw all delegation rewards
   */
  async withdrawAllRewards(delegator, options) {
    const delegations = await this.getDelegations(delegator);
    const messages = delegations.map((delegation) => ({
      typeUrl: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
      value: {
        delegatorAddress: delegator,
        validatorAddress: delegation.delegation.validatorAddress
      }
    }));
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(delegator, messages, options);
  }
  /**
   * Get all validators
   */
  async getValidators() {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.validators || [];
    } catch (error) {
      console.error("Error fetching validators:", error);
      return [];
    }
  }
  /**
   * Get validator by address
   */
  async getValidator(validatorAddress) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/staking/v1beta1/validators/${validatorAddress}`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.validator || null;
    } catch (error) {
      console.error("Error fetching validator:", error);
      return null;
    }
  }
  /**
   * Get delegations for a delegator
   */
  async getDelegations(delegator) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/staking/v1beta1/delegations/${delegator}`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.delegation_responses || [];
    } catch (error) {
      console.error("Error fetching delegations:", error);
      return [];
    }
  }
  /**
   * Get delegation to a specific validator
   */
  async getDelegation(delegator, validatorAddress) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(
        `${restEndpoint}/cosmos/staking/v1beta1/validators/${validatorAddress}/delegations/${delegator}`
      );
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.delegation_response || null;
    } catch (error) {
      console.error("Error fetching delegation:", error);
      return null;
    }
  }
  /**
   * Get unbonding delegations
   */
  async getUnbondingDelegations(delegator) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(
        `${restEndpoint}/cosmos/staking/v1beta1/delegators/${delegator}/unbonding_delegations`
      );
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.unbonding_responses || [];
    } catch (error) {
      console.error("Error fetching unbonding delegations:", error);
      return [];
    }
  }
  /**
   * Get rewards for a delegator
   */
  async getRewards(delegator) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(
        `${restEndpoint}/cosmos/distribution/v1beta1/delegators/${delegator}/rewards`
      );
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.total || [];
    } catch (error) {
      console.error("Error fetching rewards:", error);
      return [];
    }
  }
  /**
   * Get staking pool
   */
  async getPool() {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/staking/v1beta1/pool`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.pool || null;
    } catch (error) {
      console.error("Error fetching pool:", error);
      return null;
    }
  }
  /**
   * Calculate APY for a validator
   */
  calculateAPY(validator, annualProvisions, totalBondedTokens) {
    const commission = parseFloat(validator.commission.rate);
    const inflation = parseFloat(annualProvisions) / parseFloat(totalBondedTokens);
    return inflation * (1 - commission) * 100;
  }
};

// src/types/bridge.ts
var BridgeTransferStatus = /* @__PURE__ */ ((BridgeTransferStatus2) => {
  BridgeTransferStatus2[BridgeTransferStatus2["PENDING"] = 0] = "PENDING";
  BridgeTransferStatus2[BridgeTransferStatus2["COMPLETED"] = 1] = "COMPLETED";
  BridgeTransferStatus2[BridgeTransferStatus2["FAILED"] = 2] = "FAILED";
  BridgeTransferStatus2[BridgeTransferStatus2["REFUNDED"] = 3] = "REFUNDED";
  return BridgeTransferStatus2;
})(BridgeTransferStatus || {});

// src/types/compliance.ts
var ComplianceStatusType = /* @__PURE__ */ ((ComplianceStatusType2) => {
  ComplianceStatusType2[ComplianceStatusType2["UNKNOWN"] = 0] = "UNKNOWN";
  ComplianceStatusType2[ComplianceStatusType2["PENDING"] = 1] = "PENDING";
  ComplianceStatusType2[ComplianceStatusType2["APPROVED"] = 2] = "APPROVED";
  ComplianceStatusType2[ComplianceStatusType2["REJECTED"] = 3] = "REJECTED";
  ComplianceStatusType2[ComplianceStatusType2["REVOKED"] = 4] = "REVOKED";
  return ComplianceStatusType2;
})(ComplianceStatusType || {});
var KYCLevel = /* @__PURE__ */ ((KYCLevel2) => {
  KYCLevel2[KYCLevel2["NONE"] = 0] = "NONE";
  KYCLevel2[KYCLevel2["BASIC"] = 1] = "BASIC";
  KYCLevel2[KYCLevel2["INTERMEDIATE"] = 2] = "INTERMEDIATE";
  KYCLevel2[KYCLevel2["ADVANCED"] = 3] = "ADVANCED";
  return KYCLevel2;
})(KYCLevel || {});

// src/types/cryptography.ts
var KeyStatus = /* @__PURE__ */ ((KeyStatus2) => {
  KeyStatus2[KeyStatus2["ACTIVE"] = 0] = "ACTIVE";
  KeyStatus2[KeyStatus2["ROTATED"] = 1] = "ROTATED";
  KeyStatus2[KeyStatus2["REVOKED"] = 2] = "REVOKED";
  KeyStatus2[KeyStatus2["EXPIRED"] = 3] = "EXPIRED";
  return KeyStatus2;
})(KeyStatus || {});

// src/types/data-registry.ts
var DataItemStatus = /* @__PURE__ */ ((DataItemStatus2) => {
  DataItemStatus2[DataItemStatus2["ACTIVE"] = 0] = "ACTIVE";
  DataItemStatus2[DataItemStatus2["ARCHIVED"] = 1] = "ARCHIVED";
  DataItemStatus2[DataItemStatus2["DELETED"] = 2] = "DELETED";
  return DataItemStatus2;
})(DataItemStatus || {});

// src/types/economic-security.ts
var MEVProtectionLevel = /* @__PURE__ */ ((MEVProtectionLevel2) => {
  MEVProtectionLevel2[MEVProtectionLevel2["NONE"] = 0] = "NONE";
  MEVProtectionLevel2[MEVProtectionLevel2["LOW"] = 1] = "LOW";
  MEVProtectionLevel2[MEVProtectionLevel2["MEDIUM"] = 2] = "MEDIUM";
  MEVProtectionLevel2[MEVProtectionLevel2["HIGH"] = 3] = "HIGH";
  return MEVProtectionLevel2;
})(MEVProtectionLevel || {});

// src/types/identity-change.ts
var ChangeRequestStatus = /* @__PURE__ */ ((ChangeRequestStatus2) => {
  ChangeRequestStatus2[ChangeRequestStatus2["PENDING"] = 0] = "PENDING";
  ChangeRequestStatus2[ChangeRequestStatus2["APPROVED"] = 1] = "APPROVED";
  ChangeRequestStatus2[ChangeRequestStatus2["REJECTED"] = 2] = "REJECTED";
  ChangeRequestStatus2[ChangeRequestStatus2["EXPIRED"] = 3] = "EXPIRED";
  ChangeRequestStatus2[ChangeRequestStatus2["CANCELLED"] = 4] = "CANCELLED";
  return ChangeRequestStatus2;
})(ChangeRequestStatus || {});
var ChangeType = /* @__PURE__ */ ((ChangeType2) => {
  ChangeType2[ChangeType2["NAME"] = 0] = "NAME";
  ChangeType2[ChangeType2["EMAIL"] = 1] = "EMAIL";
  ChangeType2[ChangeType2["PHONE"] = 2] = "PHONE";
  ChangeType2[ChangeType2["ADDRESS"] = 3] = "ADDRESS";
  ChangeType2[ChangeType2["DOCUMENTS"] = 4] = "DOCUMENTS";
  ChangeType2[ChangeType2["OTHER"] = 5] = "OTHER";
  return ChangeType2;
})(ChangeType || {});

// src/types/inclusion-routines.ts
var RoutineStatus = /* @__PURE__ */ ((RoutineStatus2) => {
  RoutineStatus2[RoutineStatus2["PENDING"] = 0] = "PENDING";
  RoutineStatus2[RoutineStatus2["IN_PROGRESS"] = 1] = "IN_PROGRESS";
  RoutineStatus2[RoutineStatus2["COMPLETED"] = 2] = "COMPLETED";
  RoutineStatus2[RoutineStatus2["FAILED"] = 3] = "FAILED";
  RoutineStatus2[RoutineStatus2["EXPIRED"] = 4] = "EXPIRED";
  return RoutineStatus2;
})(RoutineStatus || {});
var RoutineType = /* @__PURE__ */ ((RoutineType2) => {
  RoutineType2[RoutineType2["VERIFICATION"] = 0] = "VERIFICATION";
  RoutineType2[RoutineType2["STAKING"] = 1] = "STAKING";
  RoutineType2[RoutineType2["GOVERNANCE"] = 2] = "GOVERNANCE";
  RoutineType2[RoutineType2["SOCIAL"] = 3] = "SOCIAL";
  RoutineType2[RoutineType2["EDUCATIONAL"] = 4] = "EDUCATIONAL";
  RoutineType2[RoutineType2["CUSTOM"] = 5] = "CUSTOM";
  return RoutineType2;
})(RoutineType || {});

// src/types/monitoring.ts
var AlertSeverity = /* @__PURE__ */ ((AlertSeverity2) => {
  AlertSeverity2[AlertSeverity2["INFO"] = 0] = "INFO";
  AlertSeverity2[AlertSeverity2["WARNING"] = 1] = "WARNING";
  AlertSeverity2[AlertSeverity2["ERROR"] = 2] = "ERROR";
  AlertSeverity2[AlertSeverity2["CRITICAL"] = 3] = "CRITICAL";
  return AlertSeverity2;
})(AlertSeverity || {});
var AlertStatus = /* @__PURE__ */ ((AlertStatus2) => {
  AlertStatus2[AlertStatus2["ACTIVE"] = 0] = "ACTIVE";
  AlertStatus2[AlertStatus2["ACKNOWLEDGED"] = 1] = "ACKNOWLEDGED";
  AlertStatus2[AlertStatus2["RESOLVED"] = 2] = "RESOLVED";
  return AlertStatus2;
})(AlertStatus || {});

// src/types/network-security.ts
var ThreatLevel = /* @__PURE__ */ ((ThreatLevel2) => {
  ThreatLevel2[ThreatLevel2["NONE"] = 0] = "NONE";
  ThreatLevel2[ThreatLevel2["LOW"] = 1] = "LOW";
  ThreatLevel2[ThreatLevel2["MEDIUM"] = 2] = "MEDIUM";
  ThreatLevel2[ThreatLevel2["HIGH"] = 3] = "HIGH";
  ThreatLevel2[ThreatLevel2["CRITICAL"] = 4] = "CRITICAL";
  return ThreatLevel2;
})(ThreatLevel || {});

// src/types/privacy.ts
var PrivacyLevel = /* @__PURE__ */ ((PrivacyLevel2) => {
  PrivacyLevel2[PrivacyLevel2["PUBLIC"] = 0] = "PUBLIC";
  PrivacyLevel2[PrivacyLevel2["PSEUDONYMOUS"] = 1] = "PSEUDONYMOUS";
  PrivacyLevel2[PrivacyLevel2["PRIVATE"] = 2] = "PRIVATE";
  PrivacyLevel2[PrivacyLevel2["ANONYMOUS"] = 3] = "ANONYMOUS";
  return PrivacyLevel2;
})(PrivacyLevel || {});

// src/types/validator-security.ts
var SlashReason = /* @__PURE__ */ ((SlashReason2) => {
  SlashReason2[SlashReason2["DOUBLE_SIGN"] = 0] = "DOUBLE_SIGN";
  SlashReason2[SlashReason2["DOWNTIME"] = 1] = "DOWNTIME";
  SlashReason2[SlashReason2["BYZANTINE"] = 2] = "BYZANTINE";
  SlashReason2[SlashReason2["CENSORSHIP"] = 3] = "CENSORSHIP";
  SlashReason2[SlashReason2["OTHER"] = 4] = "OTHER";
  return SlashReason2;
})(SlashReason || {});

// src/types/vc-registry.ts
var VCStatus = /* @__PURE__ */ ((VCStatus2) => {
  VCStatus2[VCStatus2["ACTIVE"] = 0] = "ACTIVE";
  VCStatus2[VCStatus2["REVOKED"] = 1] = "REVOKED";
  VCStatus2[VCStatus2["EXPIRED"] = 2] = "EXPIRED";
  VCStatus2[VCStatus2["SUSPENDED"] = 3] = "SUSPENDED";
  return VCStatus2;
})(VCStatus || {});

// src/types/wallet-security.ts
var MultisigTransactionStatus = /* @__PURE__ */ ((MultisigTransactionStatus2) => {
  MultisigTransactionStatus2[MultisigTransactionStatus2["PENDING"] = 0] = "PENDING";
  MultisigTransactionStatus2[MultisigTransactionStatus2["APPROVED"] = 1] = "APPROVED";
  MultisigTransactionStatus2[MultisigTransactionStatus2["EXECUTED"] = 2] = "EXECUTED";
  MultisigTransactionStatus2[MultisigTransactionStatus2["REJECTED"] = 3] = "REJECTED";
  MultisigTransactionStatus2[MultisigTransactionStatus2["EXPIRED"] = 4] = "EXPIRED";
  return MultisigTransactionStatus2;
})(MultisigTransactionStatus || {});
var ApprovalStatus = /* @__PURE__ */ ((ApprovalStatus2) => {
  ApprovalStatus2[ApprovalStatus2["PENDING"] = 0] = "PENDING";
  ApprovalStatus2[ApprovalStatus2["APPROVED"] = 1] = "APPROVED";
  ApprovalStatus2[ApprovalStatus2["REJECTED"] = 2] = "REJECTED";
  ApprovalStatus2[ApprovalStatus2["EXPIRED"] = 3] = "EXPIRED";
  return ApprovalStatus2;
})(ApprovalStatus || {});

// src/types/index.ts
var VoteOption = /* @__PURE__ */ ((VoteOption2) => {
  VoteOption2[VoteOption2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  VoteOption2[VoteOption2["YES"] = 1] = "YES";
  VoteOption2[VoteOption2["ABSTAIN"] = 2] = "ABSTAIN";
  VoteOption2[VoteOption2["NO"] = 3] = "NO";
  VoteOption2[VoteOption2["NO_WITH_VETO"] = 4] = "NO_WITH_VETO";
  return VoteOption2;
})(VoteOption || {});

// src/modules/governance.ts
var GovernanceModule = class {
  constructor(client) {
    this.client = client;
  }
  /**
   * Submit a text proposal
   */
  async submitTextProposal(proposer, title, description, initialDeposit, denom = "uaura", options) {
    const message = {
      typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
      value: {
        content: {
          typeUrl: "/cosmos.gov.v1beta1.TextProposal",
          value: {
            title,
            description
          }
        },
        initialDeposit: [{ denom, amount: initialDeposit }],
        proposer
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(proposer, [message], options);
  }
  /**
   * Vote on a proposal
   */
  async vote(voter, params, options) {
    const message = {
      typeUrl: "/cosmos.gov.v1beta1.MsgVote",
      value: {
        proposalId: params.proposalId,
        voter,
        option: params.option,
        metadata: params.metadata || ""
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(voter, [message], options);
  }
  /**
   * Deposit to a proposal
   */
  async deposit(depositor, params, options) {
    const denom = params.denom || "uaura";
    const message = {
      typeUrl: "/cosmos.gov.v1beta1.MsgDeposit",
      value: {
        proposalId: params.proposalId,
        depositor,
        amount: [{ denom, amount: params.amount }]
      }
    };
    const txBuilder = this.client.getTxBuilder();
    return await txBuilder.signAndBroadcast(depositor, [message], options);
  }
  /**
   * Get all proposals
   */
  async getProposals(status) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const statusParam = status !== void 0 ? `?proposal_status=${status}` : "";
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/proposals${statusParam}`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.proposals || [];
    } catch (error) {
      console.error("Error fetching proposals:", error);
      return [];
    }
  }
  /**
   * Get proposal by ID
   */
  async getProposal(proposalId) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.proposal || null;
    } catch (error) {
      console.error("Error fetching proposal:", error);
      return null;
    }
  }
  /**
   * Get votes for a proposal
   */
  async getVotes(proposalId) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/votes`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.votes || [];
    } catch (error) {
      console.error("Error fetching votes:", error);
      return [];
    }
  }
  /**
   * Get vote for a specific voter
   */
  async getVote(proposalId, voter) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(
        `${restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/votes/${voter}`
      );
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.vote || null;
    } catch (error) {
      console.error("Error fetching vote:", error);
      return null;
    }
  }
  /**
   * Get deposits for a proposal
   */
  async getDeposits(proposalId) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/deposits`);
      if (!response.ok) {
        return [];
      }
      const data = await response.json();
      return data.deposits || [];
    } catch (error) {
      console.error("Error fetching deposits:", error);
      return [];
    }
  }
  /**
   * Get tally for a proposal
   */
  async getTally(proposalId) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/proposals/${proposalId}/tally`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data.tally || null;
    } catch (error) {
      console.error("Error fetching tally:", error);
      return null;
    }
  }
  /**
   * Get governance parameters
   */
  async getParams(paramsType) {
    try {
      const config = this.client.getConfig();
      const restEndpoint = config.restEndpoint || config.rpcEndpoint.replace(":26657", ":1317");
      const response = await fetch(`${restEndpoint}/cosmos/gov/v1beta1/params/${paramsType}`);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      return data[`${paramsType}_params`] || null;
    } catch (error) {
      console.error("Error fetching params:", error);
      return null;
    }
  }
  /**
   * Check if proposal has passed quorum
   */
  async hasQuorum(proposalId) {
    const tally = await this.getTally(proposalId);
    const tallyParams = await this.getParams("tallying");
    if (!tally || !tallyParams) {
      return false;
    }
    const totalVotes = BigInt(tally.yes) + BigInt(tally.no) + BigInt(tally.abstain) + BigInt(tally.no_with_veto);
    const quorum = BigInt(tallyParams.quorum);
    return totalVotes > quorum;
  }
  /**
   * Get vote option name
   */
  getVoteOptionName(option) {
    switch (option) {
      case 1 /* YES */:
        return "Yes";
      case 3 /* NO */:
        return "No";
      case 2 /* ABSTAIN */:
        return "Abstain";
      case 4 /* NO_WITH_VETO */:
        return "No with Veto";
      default:
        return "Unspecified";
    }
  }
};

// src/client.ts
var AuraClient = class {
  constructor(config) {
    this.client = null;
    this.signingClient = null;
    this.txBuilder = null;
    this.config = {
      prefix: "aura",
      gasPrice: "0.025uaura",
      gasAdjustment: 1.5,
      ...config
    };
    this.bank = new BankModule(this);
    this.dex = new DexModule(this);
    this.staking = new StakingModule(this);
    this.governance = new GovernanceModule(this);
  }
  /**
   * Connect to the blockchain without signing capabilities
   */
  async connect() {
    const tmClient = await Tendermint34Client.connect(this.config.rpcEndpoint);
    this.client = await StargateClient.create(tmClient);
  }
  /**
   * Connect with a wallet for signing transactions
   */
  async connectWithWallet(wallet) {
    const signer = wallet.getSigner();
    this.signingClient = await SigningStargateClient.connectWithSigner(
      this.config.rpcEndpoint,
      signer,
      {
        gasPrice: this.config.gasPrice ? GasPrice.fromString(this.config.gasPrice) : void 0
      }
    );
    this.txBuilder = new TxBuilder(
      this.signingClient,
      this.config.gasPrice,
      this.config.gasAdjustment
    );
  }
  /**
   * Get the read-only client
   */
  getClient() {
    if (!this.client) {
      throw new Error("Client not connected. Call connect() first");
    }
    return this.client;
  }
  /**
   * Get the signing client
   */
  getSigningClient() {
    if (!this.signingClient) {
      throw new Error("Signing client not connected. Call connectWithWallet() first");
    }
    return this.signingClient;
  }
  /**
   * Get the transaction builder
   */
  getTxBuilder() {
    if (!this.txBuilder) {
      throw new Error("Transaction builder not available. Call connectWithWallet() first");
    }
    return this.txBuilder;
  }
  /**
   * Get chain configuration
   */
  getConfig() {
    return this.config;
  }
  /**
   * Get current block height
   */
  async getHeight() {
    return await this.getClient().getHeight();
  }
  /**
   * Get chain ID
   */
  async getChainId() {
    return await this.getClient().getChainId();
  }
  /**
   * Disconnect from the blockchain
   */
  async disconnect() {
    if (this.client) {
      this.client.disconnect();
      this.client = null;
    }
    if (this.signingClient) {
      this.signingClient.disconnect();
      this.signingClient = null;
    }
    this.txBuilder = null;
  }
  /**
   * Check if client is connected
   */
  isConnected() {
    return this.client !== null || this.signingClient !== null;
  }
  /**
   * Check if client has signing capabilities
   */
  canSign() {
    return this.signingClient !== null;
  }
};

// src/wallet.ts
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { stringToPath } from "@cosmjs/crypto";
import * as bip39 from "bip39";
var AuraWallet = class _AuraWallet {
  constructor(prefix = "paw") {
    this.wallet = null;
    this.prefix = prefix;
  }
  /**
   * Generate a new 24-word mnemonic
   */
  static generateMnemonic() {
    return bip39.generateMnemonic(256);
  }
  /**
   * Validate a mnemonic phrase
   */
  static validateMnemonic(mnemonic) {
    return bip39.validateMnemonic(mnemonic);
  }
  /**
   * Create wallet from mnemonic
   */
  async fromMnemonic(mnemonic, hdPath) {
    if (!_AuraWallet.validateMnemonic(mnemonic)) {
      throw new Error("Invalid mnemonic phrase");
    }
    const options = hdPath ? { hdPaths: [stringToPath(hdPath)], prefix: this.prefix } : { prefix: this.prefix };
    this.wallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, options);
  }
  /**
   * Get wallet accounts
   */
  async getAccounts() {
    if (!this.wallet) {
      throw new Error("Wallet not initialized");
    }
    const accounts = await this.wallet.getAccounts();
    return accounts.map((account) => ({
      address: account.address,
      pubkey: account.pubkey,
      algo: account.algo
    }));
  }
  /**
   * Get first account address
   */
  async getAddress() {
    const accounts = await this.getAccounts();
    if (accounts.length === 0) {
      throw new Error("No accounts in wallet");
    }
    return accounts[0].address;
  }
  /**
   * Get the offline signer for transaction signing
   */
  getSigner() {
    if (!this.wallet) {
      throw new Error("Wallet not initialized");
    }
    return this.wallet;
  }
  /**
   * Export mnemonic (use with caution!)
   */
  async exportMnemonic() {
    if (!this.wallet) {
      throw new Error("Wallet not initialized");
    }
    return this.wallet.mnemonic;
  }
};
export {
  AlertSeverity,
  AlertStatus,
  ApprovalStatus,
  AuraClient,
  AuraWallet,
  BankModule,
  BridgeTransferStatus,
  ChangeRequestStatus,
  ChangeType,
  ComplianceStatusType,
  DataItemStatus,
  DexModule,
  GovernanceModule,
  KYCLevel,
  KeyStatus,
  MEVProtectionLevel,
  MultisigTransactionStatus,
  PrivacyLevel,
  RoutineStatus,
  RoutineType,
  SlashReason,
  StakingModule,
  ThreatLevel,
  TxBuilder,
  VCStatus,
  VoteOption
};

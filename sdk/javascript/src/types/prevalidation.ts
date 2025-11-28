/**
 * Validation result
 */
export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  gasEstimate?: number;
  feeEstimate?: string;
  confidence: number;
}

/**
 * Validation error
 */
export interface ValidationError {
  code: string;
  message: string;
  field?: string;
  severity: 'error' | 'critical';
}

/**
 * Validation warning
 */
export interface ValidationWarning {
  code: string;
  message: string;
  field?: string;
  suggestion?: string;
}

/**
 * Transaction validation rules
 */
export interface ValidationRules {
  minGasPrice: string;
  maxGasLimit: number;
  maxMemoLength: number;
  allowedMsgTypes: string[];
  requiredFields: string[];
  customRules: CustomRule[];
}

/**
 * Custom validation rule
 */
export interface CustomRule {
  name: string;
  description: string;
  validator: string; // Function name or expression
  enabled: boolean;
  severity: 'error' | 'warning';
}

/**
 * Prevalidation parameters
 */
export interface PrevalidationParams {
  enabled: boolean;
  strictMode: boolean;
  cacheResults: boolean;
  cacheDuration: number;
  maxValidationTime: number;
  asyncValidation: boolean;
}

/**
 * Balance check result
 */
export interface BalanceCheckResult {
  sufficient: boolean;
  required: string;
  available: string;
  deficit?: string;
}

/**
 * Signature validation result
 */
export interface SignatureValidationResult {
  valid: boolean;
  publicKey: string;
  algorithm: string;
  message?: string;
}

/**
 * Transaction structure validation
 */
export interface StructureValidationResult {
  valid: boolean;
  issues: {
    field: string;
    issue: string;
    fix?: string;
  }[];
}

/**
 * Compliance validation result
 */
export interface ComplianceValidationResult {
  compliant: boolean;
  checks: {
    name: string;
    passed: boolean;
    details?: string;
  }[];
  requiredActions?: string[];
}

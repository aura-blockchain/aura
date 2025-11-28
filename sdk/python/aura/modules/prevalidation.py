"""Module for transaction prevalidation operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    PrevalidationParams,
    PrevalidationResult,
    ValidationRule,
    ValidationStatus,
    ValidationLevel,
    TransactionCheck,
    ValidationPolicy,
    GasEstimation,
    NonceValidation,
    BalanceValidation,
    TxResult,
    GasOptions
)


class PrevalidationModule:
    """Prevalidation module for validating transactions before submission."""

    def __init__(self, client):
        """Initialize prevalidation module."""
        self.client = client

    async def prevalidate_transaction(
        self,
        params: PrevalidationParams
    ) -> PrevalidationResult:
        """Prevalidate a transaction.

        Args:
            params: Validation parameters

        Returns:
            Validation result
        """
        if not params.transaction_data:
            raise ValueError("Transaction data is required")

        try:
            request_data = {
                "transaction_data": params.transaction_data,
                "validation_level": params.validation_level.value if isinstance(params.validation_level, ValidationLevel) else params.validation_level,
                "check_balance": params.check_balance,
                "check_nonce": params.check_nonce,
                "check_signature": params.check_signature,
                "check_gas": params.check_gas
            }

            data = await self.client.post("/aura/prevalidation/v1beta1/validate", request_data)
            result_data = data.get("result", {})

            return PrevalidationResult(
                status=ValidationStatus(result_data.get("status", "pending")),
                valid=result_data.get("valid", False),
                errors=result_data.get("errors", []),
                warnings=result_data.get("warnings", []),
                estimated_gas=result_data.get("estimated_gas", 0),
                estimated_fee=result_data.get("estimated_fee", "0"),
                validated_at=datetime.fromisoformat(result_data.get("validated_at")) if result_data.get("validated_at") else datetime.now(),
                checks_performed=result_data.get("checks_performed", []),
                metadata=result_data.get("metadata")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to prevalidate transaction: {e}")

    async def get_validation_rules(self) -> List[ValidationRule]:
        """Get validation rules.

        Returns:
            List of validation rules
        """
        try:
            data = await self.client.get("/aura/prevalidation/v1beta1/rules")

            rules = []
            for rule_data in data.get("rules", []):
                rules.append(ValidationRule(
                    rule_id=rule_data.get("rule_id", ""),
                    name=rule_data.get("name", ""),
                    description=rule_data.get("description", ""),
                    rule_type=rule_data.get("rule_type", ""),
                    condition=rule_data.get("condition", ""),
                    severity=rule_data.get("severity", "warning"),
                    enabled=rule_data.get("enabled", True),
                    error_message=rule_data.get("error_message", "")
                ))

            return rules
        except Exception as e:
            raise RuntimeError(f"Failed to get validation rules: {e}")

    async def update_validation_params(
        self,
        policy: ValidationPolicy,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update validation parameters (admin only).

        Args:
            policy: Validation policy
            options: Transaction options

        Returns:
            Transaction result
        """
        if not policy.policy_id:
            raise ValueError("Policy ID is required")
        if not policy.name:
            raise ValueError("Policy name is required")

        message = {
            "@type": "/aura.prevalidation.v1beta1.MsgUpdateValidationParams",
            "policy_id": policy.policy_id,
            "name": policy.name,
            "rules": policy.rules,
            "strict_mode": policy.strict_mode,
            "auto_reject": policy.auto_reject,
            "notify_on_fail": policy.notify_on_fail
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def estimate_gas(
        self,
        transaction_data: str
    ) -> GasEstimation:
        """Estimate gas for a transaction.

        Args:
            transaction_data: Transaction data

        Returns:
            Gas estimation
        """
        if not transaction_data:
            raise ValueError("Transaction data is required")

        try:
            request_data = {"transaction_data": transaction_data}
            data = await self.client.post("/aura/prevalidation/v1beta1/estimate-gas", request_data)

            estimation_data = data.get("estimation", {})

            return GasEstimation(
                base_gas=estimation_data.get("base_gas", 0),
                computation_gas=estimation_data.get("computation_gas", 0),
                storage_gas=estimation_data.get("storage_gas", 0),
                total_gas=estimation_data.get("total_gas", 0),
                gas_price=estimation_data.get("gas_price", "0"),
                total_fee=estimation_data.get("total_fee", "0"),
                confidence=estimation_data.get("confidence", 0.0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to estimate gas: {e}")

    async def validate_nonce(
        self,
        address: str,
        nonce: int
    ) -> NonceValidation:
        """Validate nonce for an address.

        Args:
            address: Address to check
            nonce: Nonce to validate

        Returns:
            Nonce validation result
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/prevalidation/v1beta1/validate-nonce/{address}/{nonce}")
            validation_data = data.get("validation", {})

            return NonceValidation(
                address=address,
                current_nonce=validation_data.get("current_nonce", 0),
                expected_nonce=validation_data.get("expected_nonce", 0),
                valid=validation_data.get("valid", False),
                gap=validation_data.get("gap", 0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to validate nonce: {e}")

    async def validate_balance(
        self,
        address: str,
        required_amount: str
    ) -> BalanceValidation:
        """Validate balance for a transaction.

        Args:
            address: Address to check
            required_amount: Required amount

        Returns:
            Balance validation result
        """
        if not address:
            raise ValueError("Address is required")
        if not required_amount:
            raise ValueError("Required amount is required")

        try:
            data = await self.client.get(f"/aura/prevalidation/v1beta1/validate-balance/{address}?amount={required_amount}")
            validation_data = data.get("validation", {})

            return BalanceValidation(
                address=address,
                required_balance=required_amount,
                actual_balance=validation_data.get("actual_balance", "0"),
                sufficient=validation_data.get("sufficient", False),
                shortfall=validation_data.get("shortfall")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to validate balance: {e}")

    async def batch_prevalidate(
        self,
        transactions: List[str],
        validation_level: ValidationLevel = ValidationLevel.STANDARD
    ) -> List[PrevalidationResult]:
        """Prevalidate multiple transactions.

        Args:
            transactions: List of transaction data
            validation_level: Validation level

        Returns:
            List of validation results
        """
        if not transactions or len(transactions) == 0:
            raise ValueError("At least one transaction is required")

        try:
            request_data = {
                "transactions": transactions,
                "validation_level": validation_level.value if isinstance(validation_level, ValidationLevel) else validation_level
            }

            data = await self.client.post("/aura/prevalidation/v1beta1/batch-validate", request_data)

            results = []
            for result_data in data.get("results", []):
                results.append(PrevalidationResult(
                    status=ValidationStatus(result_data.get("status", "pending")),
                    valid=result_data.get("valid", False),
                    errors=result_data.get("errors", []),
                    warnings=result_data.get("warnings", []),
                    estimated_gas=result_data.get("estimated_gas", 0),
                    estimated_fee=result_data.get("estimated_fee", "0"),
                    validated_at=datetime.fromisoformat(result_data.get("validated_at")) if result_data.get("validated_at") else datetime.now(),
                    checks_performed=result_data.get("checks_performed", []),
                    metadata=result_data.get("metadata")
                ))

            return results
        except Exception as e:
            raise RuntimeError(f"Failed to batch prevalidate: {e}")

    async def add_validation_rule(
        self,
        rule: ValidationRule,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Add a custom validation rule.

        Args:
            rule: Validation rule
            options: Transaction options

        Returns:
            Transaction result
        """
        if not rule.name:
            raise ValueError("Rule name is required")
        if not rule.condition:
            raise ValueError("Rule condition is required")

        message = {
            "@type": "/aura.prevalidation.v1beta1.MsgAddValidationRule",
            "name": rule.name,
            "description": rule.description,
            "rule_type": rule.rule_type,
            "condition": rule.condition,
            "severity": rule.severity,
            "error_message": rule.error_message
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_validation_policy(self, policy_id: str) -> Optional[ValidationPolicy]:
        """Get a validation policy.

        Args:
            policy_id: Policy ID

        Returns:
            Validation policy or None
        """
        if not policy_id:
            raise ValueError("Policy ID is required")

        try:
            data = await self.client.get(f"/aura/prevalidation/v1beta1/policies/{policy_id}")
            policy_data = data.get("policy")

            if not policy_data:
                return None

            return ValidationPolicy(
                policy_id=policy_data.get("policy_id", policy_id),
                name=policy_data.get("name", ""),
                rules=policy_data.get("rules", []),
                strict_mode=policy_data.get("strict_mode", False),
                auto_reject=policy_data.get("auto_reject", False),
                notify_on_fail=policy_data.get("notify_on_fail", False)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get validation policy: {e}")

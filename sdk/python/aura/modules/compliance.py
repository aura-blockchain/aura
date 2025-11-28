"""Module for compliance and KYC/AML operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    ComplianceCheckParams,
    ComplianceReport,
    ComplianceStatus,
    ComplianceLevel,
    KYCAMLStatus,
    SanctionCheck,
    TransactionMonitoring,
    TxResult,
    GasOptions
)


class ComplianceModule:
    """Compliance module for KYC/AML operations."""

    def __init__(self, client):
        """Initialize compliance module."""
        self.client = client

    async def get_kyc_status(self, address: str) -> Optional[KYCAMLStatus]:
        """Get KYC/AML status for an address.

        Args:
            address: Address to check

        Returns:
            KYC/AML status or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/compliance/v1beta1/kyc/{address}")
            kyc_data = data.get("kyc_status")

            if not kyc_data:
                return None

            return KYCAMLStatus(
                address=kyc_data.get("address", address),
                kyc_verified=kyc_data.get("kyc_verified", False),
                aml_cleared=kyc_data.get("aml_cleared", False),
                verification_date=datetime.fromisoformat(kyc_data.get("verification_date")) if kyc_data.get("verification_date") else None,
                expiry_date=datetime.fromisoformat(kyc_data.get("expiry_date")) if kyc_data.get("expiry_date") else None,
                provider=kyc_data.get("provider")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get KYC status: {e}")

    async def perform_aml_check(
        self,
        address: str,
        transaction_amount: Optional[str] = None
    ) -> Dict[str, Any]:
        """Perform AML check on an address.

        Args:
            address: Address to check
            transaction_amount: Optional transaction amount

        Returns:
            AML check result
        """
        if not address:
            raise ValueError("Address is required")

        try:
            params = {"address": address}
            if transaction_amount:
                params["amount"] = transaction_amount

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/compliance/v1beta1/aml/check?{query_str}")

            return {
                "address": address,
                "cleared": data.get("cleared", False),
                "risk_score": data.get("risk_score", 0.0),
                "risk_level": data.get("risk_level", "unknown"),
                "flags": data.get("flags", []),
                "checked_at": datetime.fromisoformat(data.get("checked_at")) if data.get("checked_at") else datetime.now(),
                "details": data.get("details", {})
            }
        except Exception as e:
            raise RuntimeError(f"Failed to perform AML check: {e}")

    async def check_sanctions(self, address: str) -> SanctionCheck:
        """Check if an address is on sanctions lists.

        Args:
            address: Address to check

        Returns:
            Sanction check result
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/compliance/v1beta1/sanctions/{address}")

            return SanctionCheck(
                address=address,
                is_sanctioned=data.get("is_sanctioned", False),
                lists_matched=data.get("lists_matched", []),
                checked_at=datetime.fromisoformat(data.get("checked_at")) if data.get("checked_at") else datetime.now(),
                details=data.get("details")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to check sanctions: {e}")

    async def get_compliance_report(
        self,
        address: str,
        level: ComplianceLevel = ComplianceLevel.STANDARD
    ) -> ComplianceReport:
        """Get comprehensive compliance report for an address.

        Args:
            address: Address to check
            level: Compliance check level

        Returns:
            Compliance report
        """
        if not address:
            raise ValueError("Address is required")

        try:
            params = {
                "address": address,
                "level": level.value if isinstance(level, ComplianceLevel) else level
            }
            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/compliance/v1beta1/reports?{query_str}")

            report_data = data.get("report", {})

            return ComplianceReport(
                report_id=report_data.get("report_id", f"{address}_{datetime.now().timestamp()}"),
                address=address,
                status=ComplianceStatus(report_data.get("status", "pending")),
                level=ComplianceLevel(report_data.get("level", "standard")),
                checks_passed=report_data.get("checks_passed", []),
                checks_failed=report_data.get("checks_failed", []),
                risk_score=report_data.get("risk_score", 0.0),
                created_at=datetime.fromisoformat(report_data.get("created_at")) if report_data.get("created_at") else datetime.now(),
                expires_at=datetime.fromisoformat(report_data.get("expires_at")) if report_data.get("expires_at") else None,
                notes=report_data.get("notes")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get compliance report: {e}")

    async def register_compliance_officer(
        self,
        officer_address: str,
        jurisdiction: str,
        credentials: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Register a compliance officer.

        Args:
            officer_address: Officer's address
            jurisdiction: Jurisdiction code
            credentials: Officer credentials
            options: Transaction options

        Returns:
            Transaction result
        """
        if not officer_address:
            raise ValueError("Officer address is required")
        if not jurisdiction:
            raise ValueError("Jurisdiction is required")
        if not credentials:
            raise ValueError("Credentials are required")

        message = {
            "@type": "/aura.compliance.v1beta1.MsgRegisterOfficer",
            "officer_address": officer_address,
            "jurisdiction": jurisdiction,
            "credentials": credentials
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_compliance_rules(self, jurisdiction: Optional[str] = None) -> List[Dict[str, Any]]:
        """Get compliance rules.

        Args:
            jurisdiction: Optional jurisdiction filter

        Returns:
            List of compliance rules
        """
        try:
            path = "/aura/compliance/v1beta1/rules"
            if jurisdiction:
                path += f"?jurisdiction={jurisdiction}"

            data = await self.client.get(path)
            return data.get("rules", [])
        except Exception as e:
            raise RuntimeError(f"Failed to get compliance rules: {e}")

    async def get_monitoring_alerts(
        self,
        address: Optional[str] = None,
        severity: Optional[str] = None,
        limit: int = 100
    ) -> List[TransactionMonitoring]:
        """Get transaction monitoring alerts.

        Args:
            address: Optional address filter
            severity: Optional severity filter
            limit: Maximum number of results

        Returns:
            List of monitoring alerts
        """
        try:
            params = {"limit": limit}
            if address:
                params["address"] = address
            if severity:
                params["severity"] = severity

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/compliance/v1beta1/monitoring/alerts?{query_str}")

            alerts = []
            for alert_data in data.get("alerts", []):
                alerts.append(TransactionMonitoring(
                    tx_hash=alert_data.get("tx_hash", ""),
                    address=alert_data.get("address", ""),
                    amount=alert_data.get("amount", "0"),
                    risk_level=alert_data.get("risk_level", "low"),
                    flags=alert_data.get("flags", []),
                    timestamp=datetime.fromisoformat(alert_data.get("timestamp")) if alert_data.get("timestamp") else datetime.now(),
                    reviewed=alert_data.get("reviewed", False)
                ))

            return alerts
        except Exception as e:
            raise RuntimeError(f"Failed to get monitoring alerts: {e}")

    async def submit_compliance_check(
        self,
        params: ComplianceCheckParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Submit a compliance check request.

        Args:
            params: Compliance check parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.check_type:
            raise ValueError("Check type is required")

        message = {
            "@type": "/aura.compliance.v1beta1.MsgSubmitCheck",
            "address": params.address,
            "check_type": params.check_type,
            "level": params.level.value if isinstance(params.level, ComplianceLevel) else params.level,
            "metadata": params.metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def approve_compliance(
        self,
        report_id: str,
        officer_address: str,
        notes: Optional[str] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Approve a compliance report.

        Args:
            report_id: Report ID
            officer_address: Compliance officer address
            notes: Optional notes
            options: Transaction options

        Returns:
            Transaction result
        """
        if not report_id:
            raise ValueError("Report ID is required")
        if not officer_address:
            raise ValueError("Officer address is required")

        message = {
            "@type": "/aura.compliance.v1beta1.MsgApproveCompliance",
            "report_id": report_id,
            "officer_address": officer_address,
            "notes": notes or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def flag_transaction(
        self,
        tx_hash: str,
        reason: str,
        severity: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Flag a transaction for compliance review.

        Args:
            tx_hash: Transaction hash
            reason: Flag reason
            severity: Severity level
            options: Transaction options

        Returns:
            Transaction result
        """
        if not tx_hash:
            raise ValueError("Transaction hash is required")
        if not reason:
            raise ValueError("Reason is required")
        if not severity:
            raise ValueError("Severity is required")

        message = {
            "@type": "/aura.compliance.v1beta1.MsgFlagTransaction",
            "tx_hash": tx_hash,
            "reason": reason,
            "severity": severity
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

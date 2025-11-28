"""Module for network monitoring operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    MonitoringAlert,
    MonitoringMetric,
    SystemHealth,
    AnomalyDetectionParams,
    AnomalyEvent,
    PerformanceMetrics,
    AlertRule,
    MonitoringDashboard,
    AlertSeverity,
    AlertStatus,
    HealthStatus,
    TxResult,
    GasOptions
)


class MonitoringModule:
    """Monitoring module for network health and alerting."""

    def __init__(self, client):
        """Initialize monitoring module."""
        self.client = client

    async def get_system_metrics(self) -> SystemHealth:
        """Get current system health metrics.

        Returns:
            System health information
        """
        try:
            data = await self.client.get("/aura/monitoring/v1beta1/health")
            health_data = data.get("health", {})

            return SystemHealth(
                overall_status=HealthStatus(health_data.get("overall_status", "healthy")),
                components=health_data.get("components", {}),
                uptime=health_data.get("uptime", 0),
                last_check=datetime.fromisoformat(health_data.get("last_check")) if health_data.get("last_check") else datetime.now(),
                active_alerts=health_data.get("active_alerts", 0),
                cpu_usage=health_data.get("cpu_usage", 0.0),
                memory_usage=health_data.get("memory_usage", 0.0),
                disk_usage=health_data.get("disk_usage", 0.0),
                network_latency=health_data.get("network_latency", 0.0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get system metrics: {e}")

    async def get_alerts(
        self,
        severity: Optional[AlertSeverity] = None,
        status: Optional[AlertStatus] = None,
        limit: int = 100
    ) -> List[MonitoringAlert]:
        """Get monitoring alerts.

        Args:
            severity: Optional severity filter
            status: Optional status filter
            limit: Maximum number of results

        Returns:
            List of alerts
        """
        try:
            params = {"limit": limit}
            if severity:
                params["severity"] = severity.value if isinstance(severity, AlertSeverity) else severity
            if status:
                params["status"] = status.value if isinstance(status, AlertStatus) else status

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/monitoring/v1beta1/alerts?{query_str}")

            alerts = []
            for alert_data in data.get("alerts", []):
                alerts.append(MonitoringAlert(
                    alert_id=alert_data.get("alert_id", ""),
                    title=alert_data.get("title", ""),
                    description=alert_data.get("description", ""),
                    severity=AlertSeverity(alert_data.get("severity", "info")),
                    status=AlertStatus(alert_data.get("status", "active")),
                    source=alert_data.get("source", ""),
                    created_at=datetime.fromisoformat(alert_data.get("created_at")) if alert_data.get("created_at") else datetime.now(),
                    updated_at=datetime.fromisoformat(alert_data.get("updated_at")) if alert_data.get("updated_at") else datetime.now(),
                    acknowledged_by=alert_data.get("acknowledged_by"),
                    resolved_at=datetime.fromisoformat(alert_data.get("resolved_at")) if alert_data.get("resolved_at") else None,
                    metadata=alert_data.get("metadata")
                ))

            return alerts
        except Exception as e:
            raise RuntimeError(f"Failed to get alerts: {e}")

    async def create_alert(
        self,
        title: str,
        description: str,
        severity: AlertSeverity,
        source: str,
        metadata: Optional[Dict[str, Any]] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a monitoring alert.

        Args:
            title: Alert title
            description: Alert description
            severity: Alert severity
            source: Alert source
            metadata: Optional metadata
            options: Transaction options

        Returns:
            Transaction result
        """
        if not title:
            raise ValueError("Title is required")
        if not description:
            raise ValueError("Description is required")
        if not source:
            raise ValueError("Source is required")

        message = {
            "@type": "/aura.monitoring.v1beta1.MsgCreateAlert",
            "title": title,
            "description": description,
            "severity": severity.value if isinstance(severity, AlertSeverity) else severity,
            "source": source,
            "metadata": metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def acknowledge_alert(
        self,
        alert_id: str,
        acknowledger: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Acknowledge an alert.

        Args:
            alert_id: Alert ID
            acknowledger: Acknowledger address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not alert_id:
            raise ValueError("Alert ID is required")
        if not acknowledger:
            raise ValueError("Acknowledger address is required")

        message = {
            "@type": "/aura.monitoring.v1beta1.MsgAcknowledgeAlert",
            "alert_id": alert_id,
            "acknowledger": acknowledger
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def resolve_alert(
        self,
        alert_id: str,
        resolver: str,
        resolution_notes: Optional[str] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Resolve an alert.

        Args:
            alert_id: Alert ID
            resolver: Resolver address
            resolution_notes: Optional notes
            options: Transaction options

        Returns:
            Transaction result
        """
        if not alert_id:
            raise ValueError("Alert ID is required")
        if not resolver:
            raise ValueError("Resolver address is required")

        message = {
            "@type": "/aura.monitoring.v1beta1.MsgResolveAlert",
            "alert_id": alert_id,
            "resolver": resolver,
            "resolution_notes": resolution_notes or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_performance_stats(self) -> PerformanceMetrics:
        """Get network performance statistics.

        Returns:
            Performance metrics
        """
        try:
            data = await self.client.get("/aura/monitoring/v1beta1/performance")
            perf_data = data.get("metrics", {})

            return PerformanceMetrics(
                average_block_time=perf_data.get("average_block_time", 0.0),
                transactions_per_second=perf_data.get("transactions_per_second", 0.0),
                average_gas_price=perf_data.get("average_gas_price", "0"),
                network_hashrate=perf_data.get("network_hashrate", 0.0),
                active_validators=perf_data.get("active_validators", 0),
                timestamp=datetime.fromisoformat(perf_data.get("timestamp")) if perf_data.get("timestamp") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get performance stats: {e}")

    async def get_network_health(self) -> Dict[str, Any]:
        """Get comprehensive network health information.

        Returns:
            Network health data
        """
        try:
            data = await self.client.get("/aura/monitoring/v1beta1/network/health")
            return data.get("health", {})
        except Exception as e:
            raise RuntimeError(f"Failed to get network health: {e}")

    async def submit_metric(
        self,
        metric: MonitoringMetric,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Submit a custom monitoring metric.

        Args:
            metric: Metric data
            options: Transaction options

        Returns:
            Transaction result
        """
        if not metric.metric_name:
            raise ValueError("Metric name is required")

        message = {
            "@type": "/aura.monitoring.v1beta1.MsgSubmitMetric",
            "metric_name": metric.metric_name,
            "value": str(metric.value),
            "unit": metric.unit,
            "labels": metric.labels or {},
            "tags": metric.tags or []
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def configure_anomaly_detection(
        self,
        params: AnomalyDetectionParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure anomaly detection.

        Args:
            params: Detection parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.metric_name:
            raise ValueError("Metric name is required")

        message = {
            "@type": "/aura.monitoring.v1beta1.MsgConfigureAnomalyDetection",
            "metric_name": params.metric_name,
            "threshold": str(params.threshold),
            "window_size": params.window_size,
            "sensitivity": str(params.sensitivity)
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_anomaly_events(self, limit: int = 100) -> List[AnomalyEvent]:
        """Get detected anomaly events.

        Args:
            limit: Maximum number of results

        Returns:
            List of anomaly events
        """
        try:
            data = await self.client.get(f"/aura/monitoring/v1beta1/anomalies?limit={limit}")

            events = []
            for event_data in data.get("events", []):
                events.append(AnomalyEvent(
                    event_id=event_data.get("event_id", ""),
                    metric_name=event_data.get("metric_name", ""),
                    detected_value=event_data.get("detected_value", 0.0),
                    expected_range=event_data.get("expected_range", ""),
                    severity=AlertSeverity(event_data.get("severity", "warning")),
                    detected_at=datetime.fromisoformat(event_data.get("detected_at")) if event_data.get("detected_at") else datetime.now(),
                    confidence=event_data.get("confidence", 0.0)
                ))

            return events
        except Exception as e:
            raise RuntimeError(f"Failed to get anomaly events: {e}")

    async def get_dashboard_data(self) -> MonitoringDashboard:
        """Get dashboard metrics summary.

        Returns:
            Dashboard data
        """
        try:
            data = await self.client.get("/aura/monitoring/v1beta1/dashboard")
            dashboard_data = data.get("dashboard", {})

            return MonitoringDashboard(
                total_nodes=dashboard_data.get("total_nodes", 0),
                active_nodes=dashboard_data.get("active_nodes", 0),
                total_transactions=dashboard_data.get("total_transactions", 0),
                failed_transactions=dashboard_data.get("failed_transactions", 0),
                average_response_time=dashboard_data.get("average_response_time", 0.0),
                error_rate=dashboard_data.get("error_rate", 0.0),
                timestamp=datetime.fromisoformat(dashboard_data.get("timestamp")) if dashboard_data.get("timestamp") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get dashboard data: {e}")

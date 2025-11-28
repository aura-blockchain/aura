"""Type definitions for Monitoring module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class AlertSeverity(Enum):
    """Alert severity levels."""

    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


class AlertStatus(Enum):
    """Alert status."""

    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"


class HealthStatus(Enum):
    """System health status."""

    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"
    CRITICAL = "critical"


@dataclass
class MonitoringAlert:
    """Monitoring alert."""

    alert_id: str
    title: str
    description: str
    severity: AlertSeverity
    status: AlertStatus
    source: str
    created_at: datetime
    updated_at: datetime
    acknowledged_by: Optional[str] = None
    resolved_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class MonitoringMetric:
    """Monitoring metric data."""

    metric_name: str
    value: float
    unit: str
    timestamp: datetime
    labels: Optional[Dict[str, str]] = None
    tags: Optional[List[str]] = None


@dataclass
class SystemHealth:
    """System health information."""

    overall_status: HealthStatus
    components: Dict[str, str]
    uptime: int
    last_check: datetime
    active_alerts: int
    cpu_usage: float
    memory_usage: float
    disk_usage: float
    network_latency: float


@dataclass
class AnomalyDetectionParams:
    """Anomaly detection parameters."""

    metric_name: str
    threshold: float
    window_size: int
    sensitivity: float


@dataclass
class AnomalyEvent:
    """Detected anomaly event."""

    event_id: str
    metric_name: str
    detected_value: float
    expected_range: str
    severity: AlertSeverity
    detected_at: datetime
    confidence: float


@dataclass
class PerformanceMetrics:
    """Performance metrics."""

    average_block_time: float
    transactions_per_second: float
    average_gas_price: str
    network_hashrate: float
    active_validators: int
    timestamp: datetime


@dataclass
class AlertRule:
    """Alert rule configuration."""

    rule_id: str
    name: str
    condition: str
    severity: AlertSeverity
    enabled: bool
    notification_channels: List[str]


@dataclass
class MonitoringDashboard:
    """Dashboard metrics summary."""

    total_nodes: int
    active_nodes: int
    total_transactions: int
    failed_transactions: int
    average_response_time: float
    error_rate: float
    timestamp: datetime

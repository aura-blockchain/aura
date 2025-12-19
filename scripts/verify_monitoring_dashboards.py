#!/usr/bin/env python3
"""
Verifies that Grafana dashboards, alert rules, and Prometheus configuration
reference the Aura monitoring metrics that are defined in code.

Checks performed:
1. Parse chain/x/monitoring/metrics/prometheus.go to collect all metric names.
2. Scan Grafana dashboards and alert rules for aura_monitoring_* references
   and ensure every reference maps to a defined metric (including histogram
   bucket/count/sum helpers).
3. Ensure key metrics appear in at least one dashboard or alert.
4. Confirm Prometheus is configured with the aura-monitoring job.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Dict, Iterable, List, Set, Tuple


REPO_ROOT = Path(__file__).resolve().parents[1]
METRICS_SRC = REPO_ROOT / "chain" / "x" / "monitoring" / "metrics" / "prometheus.go"
PROMETHEUS_CONFIG = REPO_ROOT / "prometheus" / "prometheus.yml"
GRAFANA_DIRS = [
    REPO_ROOT / "grafana" / "dashboards",
    REPO_ROOT / "docker" / "monitoring" / "grafana" / "dashboards",
]
ALERT_FILES = [
    REPO_ROOT / "prometheus" / "rules" / "monitoring-alerts.yml",
    REPO_ROOT / "docs" / "ops" / "MONITORING_ALERTING.md",
]
REQUIRED_METRICS = {
    "aura_monitoring_validator_uptime_percentage",
    "aura_monitoring_active_validators",
    "aura_monitoring_block_time_seconds",
    "aura_monitoring_transactions_per_second",
    "aura_monitoring_total_tvl",
    "aura_monitoring_security_events_total",
    "aura_monitoring_mempool_size",
    "aura_monitoring_alerts_total",
}
METRIC_PATTERN = re.compile(r"aura_monitoring[_0-9a-zA-Z]*")
defined_metrics: Set[str] = set()


class VerificationError(RuntimeError):
    """Raised when verification fails."""


def extract_metric_names() -> Tuple[Set[str], Dict[str, str]]:
    """Return defined metric names (with namespace) and their Go types."""
    if not METRICS_SRC.exists():
        raise VerificationError(f"Missing metrics source: {METRICS_SRC}")

    defined: Dict[str, str] = {}
    histogram_bases: Set[str] = set()
    current_type: str | None = None

    for raw_line in METRICS_SRC.read_text().splitlines():
        line = raw_line.strip()
        new_match = re.search(r"promauto\.New([A-Za-z]+)", line)
        if new_match:
            current_type = new_match.group(1)
            continue

        if current_type:
            name_match = re.search(r'Name:\s*"([^"]+)"', line)
            if name_match:
                metric = f"aura_monitoring_{name_match.group(1)}"
                defined[metric] = current_type
                if "Histogram" in current_type:
                    histogram_bases.add(metric)
                current_type = None

    if not defined:
        raise VerificationError("No metrics were parsed from prometheus.go")

    # Prometheus automatically emits bucket/sum/count for histograms; add them
    for hist in histogram_bases:
        defined[f"{hist}_bucket"] = "HistogramBucket"
        defined[f"{hist}_sum"] = "HistogramSum"
        defined[f"{hist}_count"] = "HistogramCount"

    return set(defined.keys()), defined


def collect_files() -> List[Path]:
    """Gather Grafana dashboard JSON files and alerting files to scan."""
    files: List[Path] = []
    for folder in GRAFANA_DIRS:
        if folder.exists():
            files.extend(sorted(folder.rglob("*.json")))

    for path in ALERT_FILES:
        if path.exists():
            files.append(path)

    if not files:
        raise VerificationError("No Grafana or alert files found to verify")
    return files


def scan_metric_usage(
    files: Iterable[Path],
) -> Tuple[Set[str], Dict[Path, List[str]]]:
    """Scan files for aura_monitoring_* references."""
    referenced: Set[str] = set()
    unknown: Dict[Path, List[str]] = {}

    for path in files:
        content = path.read_text()
        matches = {match.lower() for match in METRIC_PATTERN.findall(content)}
        referenced |= matches

        invalid = sorted(match for match in matches if match not in defined_metrics)
        if invalid:
            unknown[path] = invalid

    return referenced, unknown


def verify_prometheus_job() -> None:
    """Ensure the Prometheus scrape config includes the aura-monitoring job."""
    text = PROMETHEUS_CONFIG.read_text()
    if "job_name: 'aura-monitoring'" not in text and "job_name: \"aura-monitoring\"" not in text:
        raise VerificationError("Prometheus config missing aura-monitoring job")
    if "localhost:9090" not in text:
        raise VerificationError("Prometheus config missing localhost:9090 target for monitoring job")


def main() -> int:
    global defined_metrics
    defined_metrics, typed_metrics = extract_metric_names()
    files = collect_files()
    referenced, unknown_refs = scan_metric_usage(files)

    if unknown_refs:
        details = "\n".join(
            f"- {path}: {', '.join(metrics)}" for path, metrics in unknown_refs.items()
        )
        raise VerificationError(
            "Found references to undefined metrics:\n" + details
        )

    missing_required = sorted(metric for metric in REQUIRED_METRICS if metric not in referenced)
    if missing_required:
        raise VerificationError(
            "Required monitoring metrics are missing from dashboards/alerts: "
            + ", ".join(missing_required)
        )

    verify_prometheus_job()

    unused = sorted(defined_metrics - referenced)
    print(f"Defined metrics: {len(defined_metrics)}")
    print(f"Referenced metrics: {len(referenced)}")
    print(f"Files scanned: {len(files)}")
    if unused:
        print(f"Note: {len(unused)} metrics are not referenced by dashboards/alerts yet.")

    print("Monitoring dashboards and Prometheus configuration verified successfully.")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except VerificationError as exc:
        print(f"[monitoring verification] {exc}", file=sys.stderr)
        sys.exit(1)

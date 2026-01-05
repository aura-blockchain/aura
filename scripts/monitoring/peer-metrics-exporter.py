#!/usr/bin/env python3
"""
Cosmos/CometBFT Peer Metrics Exporter

Lightweight Prometheus exporter that queries CometBFT RPC /net_info
and exposes peer count and peer details as Prometheus metrics.

Usage:
    python3 peer-metrics-exporter.py --rpc http://localhost:26657 --port 9600 --chain aura

This follows the standard approach used by production Cosmos chains.
"""

import argparse
import json
import time
import logging
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.request import urlopen
from urllib.error import URLError
from typing import Dict, Any, Optional

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


class PeerMetrics:
    """Fetches and caches peer metrics from CometBFT RPC."""

    def __init__(self, rpc_url: str, chain_id: str, network: str = "testnet"):
        self.rpc_url = rpc_url.rstrip('/')
        self.chain_id = chain_id
        self.network = network
        self._cache: Dict[str, Any] = {}
        self._cache_time: float = 0
        self._cache_ttl: float = 5.0  # 5 second cache

    def fetch_net_info(self) -> Optional[Dict[str, Any]]:
        """Fetch /net_info from CometBFT RPC."""
        now = time.time()
        if now - self._cache_time < self._cache_ttl and self._cache:
            return self._cache

        try:
            url = f"{self.rpc_url}/net_info"
            with urlopen(url, timeout=10) as response:
                data = json.loads(response.read().decode())
                self._cache = data.get('result', {})
                self._cache_time = now
                return self._cache
        except URLError as e:
            logger.error(f"Failed to fetch net_info: {e}")
            return None
        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse net_info: {e}")
            return None

    def fetch_status(self) -> Optional[Dict[str, Any]]:
        """Fetch /status from CometBFT RPC for additional info."""
        try:
            url = f"{self.rpc_url}/status"
            with urlopen(url, timeout=10) as response:
                data = json.loads(response.read().decode())
                return data.get('result', {})
        except (URLError, json.JSONDecodeError) as e:
            logger.error(f"Failed to fetch status: {e}")
            return None

    def get_metrics(self) -> str:
        """Generate Prometheus metrics output."""
        lines = []

        # Metadata
        lines.append("# HELP cosmos_peer_count Number of connected P2P peers")
        lines.append("# TYPE cosmos_peer_count gauge")

        lines.append("# HELP cosmos_peer_info Information about connected peers")
        lines.append("# TYPE cosmos_peer_info gauge")

        lines.append("# HELP cosmos_peer_bytes_received Total bytes received from peer")
        lines.append("# TYPE cosmos_peer_bytes_received counter")

        lines.append("# HELP cosmos_peer_bytes_sent Total bytes sent to peer")
        lines.append("# TYPE cosmos_peer_bytes_sent counter")

        lines.append("# HELP cosmos_peer_exporter_up Whether the exporter can reach the RPC")
        lines.append("# TYPE cosmos_peer_exporter_up gauge")

        net_info = self.fetch_net_info()

        if net_info is None:
            lines.append(f'cosmos_peer_exporter_up{{chain="{self.chain_id}",network="{self.network}"}} 0')
            lines.append(f'cosmos_peer_count{{chain="{self.chain_id}",network="{self.network}"}} 0')
            return '\n'.join(lines) + '\n'

        lines.append(f'cosmos_peer_exporter_up{{chain="{self.chain_id}",network="{self.network}"}} 1')

        # Total peer count
        n_peers = int(net_info.get('n_peers', 0))
        lines.append(f'cosmos_peer_count{{chain="{self.chain_id}",network="{self.network}"}} {n_peers}')

        # Individual peer metrics
        peers = net_info.get('peers', [])
        for peer in peers:
            node_info = peer.get('node_info', {})
            remote_ip = peer.get('remote_ip', 'unknown')
            peer_id = node_info.get('id', 'unknown')[:16]  # Truncate for readability
            moniker = node_info.get('moniker', 'unknown')
            listen_addr = node_info.get('listen_addr', 'unknown')

            # Sanitize values for Prometheus labels
            moniker = moniker.replace('"', '\\"').replace('\\', '\\\\')[:32]

            # Peer info (always 1, carries labels)
            lines.append(
                f'cosmos_peer_info{{chain="{self.chain_id}",network="{self.network}",'
                f'peer_id="{peer_id}",moniker="{moniker}",remote_ip="{remote_ip}"}} 1'
            )

            # Connection metrics
            conn_status = peer.get('connection_status', {})

            # Bytes received/sent
            recv_monitor = conn_status.get('RecvMonitor', {})
            send_monitor = conn_status.get('SendMonitor', {})

            bytes_recv = recv_monitor.get('Bytes', 0)
            bytes_sent = send_monitor.get('Bytes', 0)

            lines.append(
                f'cosmos_peer_bytes_received{{chain="{self.chain_id}",network="{self.network}",'
                f'peer_id="{peer_id}"}} {bytes_recv}'
            )
            lines.append(
                f'cosmos_peer_bytes_sent{{chain="{self.chain_id}",network="{self.network}",'
                f'peer_id="{peer_id}"}} {bytes_sent}'
            )

        return '\n'.join(lines) + '\n'


class MetricsHandler(BaseHTTPRequestHandler):
    """HTTP handler for Prometheus metrics endpoint."""

    metrics: PeerMetrics = None

    def do_GET(self):
        if self.path == '/metrics':
            content = self.metrics.get_metrics()
            self.send_response(200)
            self.send_header('Content-Type', 'text/plain; charset=utf-8')
            self.send_header('Content-Length', len(content))
            self.end_headers()
            self.wfile.write(content.encode())
        elif self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'OK')
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # Suppress default logging for /metrics requests
        if '/metrics' not in args[0]:
            logger.info("%s - %s", self.address_string(), format % args)


def main():
    parser = argparse.ArgumentParser(description='Cosmos/CometBFT Peer Metrics Exporter')
    parser.add_argument('--rpc', default='http://localhost:26657', help='CometBFT RPC URL')
    parser.add_argument('--port', type=int, default=9600, help='Metrics port')
    parser.add_argument('--chain', default='cosmos', help='Chain identifier (e.g., aura, paw)')
    parser.add_argument('--network', default='testnet', help='Network type (testnet/mainnet)')
    args = parser.parse_args()

    MetricsHandler.metrics = PeerMetrics(args.rpc, args.chain, args.network)

    server = HTTPServer(('0.0.0.0', args.port), MetricsHandler)
    logger.info(f"Starting peer metrics exporter on port {args.port}")
    logger.info(f"RPC: {args.rpc}, Chain: {args.chain}, Network: {args.network}")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Shutting down")
        server.shutdown()


if __name__ == '__main__':
    main()

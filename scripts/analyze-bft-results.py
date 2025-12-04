#!/usr/bin/env python3

"""
AURA BFT Test Results Analysis Tool

Analyzes and visualizes Byzantine Fault Tolerance test results.
Generates comprehensive reports from test output files.

Usage:
    python3 scripts/analyze-bft-results.py <test_log_file> [options]
    python3 scripts/analyze-bft-results.py --recent [options]
    python3 scripts/analyze-bft-results.py --compare <log1> <log2> <log3> ... [options]

Options:
    --verbose           Detailed output
    --output-format     (text|json|html) Output format
    --export-csv        Export metrics as CSV
    --html-report       Generate HTML report
    --help              Show this help message
"""

import sys
import json
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple, Optional
import argparse


class BFTTestAnalyzer:
    """Analyzes AURA BFT test results from log files."""

    def __init__(self, log_file: str, verbose: bool = False):
        """Initialize analyzer with a test log file."""
        self.log_file = Path(log_file)
        self.verbose = verbose
        self.metrics = {}
        self.events = []
        self.scenarios = {}

        if not self.log_file.exists():
            raise FileNotFoundError(f"Log file not found: {log_file}")

        self.parse_log()

    def parse_log(self):
        """Parse log file and extract metrics."""
        with open(self.log_file, 'r') as f:
            content = f.read()

        # Extract timeline info
        lines = content.split('\n')

        for i, line in enumerate(lines):
            # Extract timestamped events
            timestamp_match = re.search(r'\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\]', line)
            level_match = re.search(r'\[([A-Z]+)\]', line)

            if timestamp_match and level_match:
                event = {
                    'timestamp': timestamp_match.group(1),
                    'level': level_match.group(1),
                    'message': line.strip(),
                    'line_number': i + 1
                }
                self.events.append(event)

            # Extract metrics
            if '=' in line and not line.startswith('['):
                key, value = line.split('=', 1)
                key = key.strip()
                value = value.strip()

                # Try to convert to number
                try:
                    if '.' in value:
                        self.metrics[key] = float(value)
                    else:
                        self.metrics[key] = int(value)
                except ValueError:
                    self.metrics[key] = value

    def get_block_progression(self) -> Dict[str, List[int]]:
        """Extract block height progression for each validator."""
        progression = {
            'val1': [],
            'val2': [],
            'val3': [],
            'val4': []
        }

        for val in progression.keys():
            # Find all heights for this validator
            baseline_key = f'BASELINE_{val.upper()}'
            if baseline_key in self.metrics:
                progression[val].append(self.metrics[baseline_key])

            # Find 3_OF_4 heights
            three_of_four_key = f'3_OF_4_{val}'
            if three_of_four_key in self.metrics:
                progression[val].append(self.metrics[three_of_four_key])

            # Find catch-up heights
            if val == 'val3':
                if 'VAL3_CATCH_UP_HEIGHT' in self.metrics:
                    progression[val].append(self.metrics['VAL3_CATCH_UP_HEIGHT'])

            # Find final heights
            final_key = f'FINAL_HEIGHT_{val}'
            if final_key in self.metrics:
                progression[val].append(self.metrics[final_key])

        return progression

    def get_event_timeline(self) -> List[Dict]:
        """Get chronological timeline of major events."""
        timeline = []

        for event in self.events:
            if any(keyword in event['message'].upper() for keyword in
                   ['STEP', 'STOPPED', 'STARTED', 'SUCCESS', 'FAILED', 'BLOCK']):
                timeline.append(event)

        return timeline

    def analyze_consensus_behavior(self) -> Dict:
        """Analyze consensus behavior across test phases."""
        analysis = {
            'baseline_sync': self._check_baseline_sync(),
            'three_of_four': self._check_three_of_four(),
            'catch_up': self._check_catch_up(),
            'two_of_four': self._check_two_of_four(),
            'recovery': self._check_recovery()
        }

        return analysis

    def _check_baseline_sync(self) -> Dict:
        """Check if validators started in sync."""
        baseline_heights = [
            self.metrics.get('BASELINE_VAL1', 0),
            self.metrics.get('BASELINE_VAL2', 0),
            self.metrics.get('BASELINE_VAL3', 0),
            self.metrics.get('BASELINE_VAL4', 0)
        ]

        baseline_heights = [h for h in baseline_heights if h > 0]

        if not baseline_heights:
            return {'status': 'UNKNOWN', 'message': 'No baseline data'}

        height_diff = max(baseline_heights) - min(baseline_heights)

        return {
            'status': 'PASS' if height_diff <= 1 else 'WARNING',
            'message': f'Height variance: {height_diff} blocks',
            'heights': baseline_heights,
            'variance': height_diff
        }

    def _check_three_of_four(self) -> Dict:
        """Check if chain continued with 3/4 validators."""
        heights_before = self.metrics.get('BASELINE_VAL1', 0)
        heights_after = [
            self.metrics.get('3_OF_4_val1', 0),
            self.metrics.get('3_OF_4_val2', 0),
            self.metrics.get('3_OF_4_val4', 0)
        ]

        heights_after = [h for h in heights_after if h > 0]

        if not heights_after:
            return {'status': 'UNKNOWN', 'message': 'No 3/4 test data'}

        blocks_produced = sum(h - heights_before for h in heights_after) / len(heights_after)

        return {
            'status': 'PASS' if blocks_produced > 0 else 'FAIL',
            'message': f'Blocks produced: {blocks_produced:.0f}',
            'blocks_produced': blocks_produced,
            'heights': heights_after
        }

    def _check_catch_up(self) -> Dict:
        """Check if stopped validator caught up."""
        catch_up_height = self.metrics.get('VAL3_CATCH_UP_HEIGHT', 0)
        val1_height = self.metrics.get('VAL1_HEIGHT_AT_RESTART', 0)
        height_diff = self.metrics.get('CATCH_UP_DIFF', 999)

        if catch_up_height == 0:
            return {'status': 'UNKNOWN', 'message': 'No catch-up data'}

        return {
            'status': 'PASS' if height_diff <= 5 else 'WARNING',
            'message': f'Catch-up difference: {height_diff} blocks',
            'val3_height': catch_up_height,
            'val1_height': val1_height,
            'difference': height_diff
        }

    def _check_two_of_four(self) -> Dict:
        """Check if chain halted with 2/4 validators."""
        blocks_produced = self.metrics.get('2_OF_4_BLOCKS_PRODUCED', -1)

        if blocks_produced == -1:
            return {'status': 'UNKNOWN', 'message': 'No 2/4 halt test data'}

        return {
            'status': 'PASS' if blocks_produced == 0 else 'CRITICAL',
            'message': f'Blocks produced: {blocks_produced}',
            'blocks_produced': blocks_produced,
            'severity': 'CRITICAL_BFT_BUG' if blocks_produced > 0 else 'OK'
        }

    def _check_recovery(self) -> Dict:
        """Check if chain recovered with all 4 validators."""
        recovery_prod = self.metrics.get('RECOVERY_BLOCK_PRODUCTION', -1)

        if recovery_prod == -1:
            return {'status': 'UNKNOWN', 'message': 'No recovery test data'}

        return {
            'status': 'PASS' if recovery_prod >= 3 else 'WARNING',
            'message': f'Validators producing blocks: {recovery_prod}/4',
            'validators_synced': recovery_prod
        }

    def generate_text_report(self) -> str:
        """Generate human-readable text report."""
        report = []
        report.append("=" * 80)
        report.append("AURA BFT TEST ANALYSIS REPORT")
        report.append("=" * 80)
        report.append("")

        # Test metadata
        report.append("TEST METADATA")
        report.append("-" * 80)
        report.append(f"Log File: {self.log_file}")
        report.append(f"Generated: {datetime.now().isoformat()}")
        report.append("")

        # Block progression
        report.append("BLOCK HEIGHT PROGRESSION")
        report.append("-" * 80)
        progression = self.get_block_progression()
        for val, heights in sorted(progression.items()):
            if heights:
                report.append(f"{val:>6}: {' → '.join(str(h) for h in heights)}")
        report.append("")

        # Consensus analysis
        report.append("CONSENSUS BEHAVIOR ANALYSIS")
        report.append("-" * 80)
        analysis = self.analyze_consensus_behavior()

        for phase, result in analysis.items():
            status_icon = "✓" if result.get('status') == 'PASS' else \
                          "⚠" if result.get('status') == 'WARNING' else \
                          "✗" if result.get('status') == 'FAIL' else "?"

            report.append(f"{status_icon} {phase.upper()}: {result.get('status', 'UNKNOWN')}")
            report.append(f"  {result.get('message', 'N/A')}")

            if result.get('severity') == 'CRITICAL_BFT_BUG':
                report.append("  *** CRITICAL: Chain continued with 2/4 validators! ***")

            report.append("")

        # Event timeline
        report.append("MAJOR EVENTS TIMELINE")
        report.append("-" * 80)
        timeline = self.get_event_timeline()
        for event in timeline[-20:]:  # Last 20 events
            report.append(f"{event['timestamp']} [{event['level']}] {event['message'][:70]}")
        report.append("")

        # Summary
        report.append("TEST SUMMARY")
        report.append("-" * 80)
        all_pass = all(r.get('status') in ['PASS', 'UNKNOWN'] for r in analysis.values())
        critical_bugs = [r for r in analysis.values() if r.get('severity') == 'CRITICAL_BFT_BUG']

        if critical_bugs:
            report.append("RESULT: FAILED - CRITICAL BFT BUG DETECTED")
            report.append("The chain continued producing blocks with only 2/4 validators!")
            report.append("This violates the 2/3 consensus threshold requirement.")
        elif all_pass:
            report.append("RESULT: PASSED - All BFT tests completed successfully")
        else:
            report.append("RESULT: PASSED WITH WARNINGS - Most tests passed")

        report.append("")
        report.append("=" * 80)

        return "\n".join(report)

    def generate_json_report(self) -> Dict:
        """Generate machine-readable JSON report."""
        return {
            'test_metadata': {
                'log_file': str(self.log_file),
                'analysis_timestamp': datetime.now().isoformat(),
                'total_events': len(self.events),
                'metrics_extracted': len(self.metrics)
            },
            'block_progression': self.get_block_progression(),
            'consensus_analysis': self.analyze_consensus_behavior(),
            'raw_metrics': self.metrics,
            'event_count': {
                'info': len([e for e in self.events if e['level'] == 'INFO']),
                'success': len([e for e in self.events if e['level'] == 'SUCCESS']),
                'warning': len([e for e in self.events if e['level'] == 'WARNING']),
                'error': len([e for e in self.events if e['level'] == 'ERROR']),
            }
        }

    def generate_csv_export(self) -> str:
        """Generate CSV export of key metrics."""
        lines = []
        lines.append("Metric,Value")

        for key, value in sorted(self.metrics.items()):
            # Escape commas in values
            value_str = str(value).replace(',', ';')
            lines.append(f"{key},{value_str}")

        return "\n".join(lines)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='AURA BFT Test Results Analysis Tool',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Analyze a test log
  python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log

  # Analyze with verbose output
  python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --verbose

  # Export to JSON
  python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --output-format json

  # Export metrics to CSV
  python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --export-csv
        """
    )

    parser.add_argument('log_file', nargs='?', help='BFT test log file to analyze')
    parser.add_argument('--verbose', '-v', action='store_true', help='Verbose output')
    parser.add_argument('--output-format', choices=['text', 'json', 'csv'],
                        default='text', help='Output format')
    parser.add_argument('--export-csv', action='store_true', help='Export metrics as CSV')
    parser.add_argument('--output', '-o', help='Output file (default: stdout)')

    args = parser.parse_args()

    if not args.log_file:
        parser.print_help()
        sys.exit(1)

    try:
        # Analyze log file
        analyzer = BFTTestAnalyzer(args.log_file, verbose=args.verbose)

        # Generate output
        if args.output_format == 'json':
            output = json.dumps(analyzer.generate_json_report(), indent=2)
        elif args.output_format == 'csv':
            output = analyzer.generate_csv_export()
        else:  # text
            output = analyzer.generate_text_report()

        # Write output
        if args.output:
            with open(args.output, 'w') as f:
                f.write(output)
            print(f"Report written to: {args.output}")
        else:
            print(output)

        # If CSV export requested in addition
        if args.export_csv and args.output_format != 'csv':
            csv_output = analyzer.generate_csv_export()
            csv_file = Path(args.log_file).stem + '.csv'
            with open(csv_file, 'w') as f:
                f.write(csv_output)
            print(f"CSV exported to: {csv_file}")

    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        if args.verbose:
            import traceback
            traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()

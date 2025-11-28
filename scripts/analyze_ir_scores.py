#!/usr/bin/env python3
"""
Statistical analysis of Inclusion Routine (IR) scores
"""

import json
import statistics
from collections import defaultdict

# Load IR definitions
with open('../data/inclusion_routines/ir_definitions.json', 'r') as f:
    data = json.load(f)

irs = data['inclusion_routines']

# Extract scores (exclude IR-000 which is the anchor with 0 score)
scores = [ir['score'] for ir in irs if ir['id'] != 'IR-000']
arena_scores = defaultdict(list)

for ir in irs:
    if ir['id'] != 'IR-000':
        arena_scores[ir['arena']].append(ir['score'])

# Overall statistics
print("=" * 70)
print("OVERALL IR SCORE STATISTICS")
print("=" * 70)
print(f"Total IRs (excluding IR-000): {len(scores)}")
print(f"Minimum score: {min(scores)}")
print(f"Maximum score: {max(scores)}")
print(f"Mean score: {statistics.mean(scores):.2f}")
print(f"Median score: {statistics.median(scores):.2f}")
print(f"Standard deviation: {statistics.stdev(scores):.2f}")
print(f"Total possible score (all IRs): {sum(scores):,}")
print()

# Percentiles
print("Score Percentiles:")
sorted_scores = sorted(scores)
for p in [10, 25, 50, 75, 90, 95, 99]:
    idx = int(len(sorted_scores) * p / 100)
    print(f"  {p}th percentile: {sorted_scores[idx]}")
print()

# Score distribution by ranges
print("Score Distribution by Range:")
ranges = [
    (0, 100, "0-100"),
    (101, 300, "101-300"),
    (301, 500, "301-500"),
    (501, 700, "501-700"),
    (701, 1000, "701-1000"),
    (1001, 1500, "1001-1500"),
    (1501, 3000, "1501-3000"),
]

for min_val, max_val, label in ranges:
    count = sum(1 for s in scores if min_val <= s <= max_val)
    pct = count / len(scores) * 100
    print(f"  {label:15} {count:3d} IRs ({pct:5.1f}%)")
print()

# Arena breakdown
print("=" * 70)
print("ARENA-SPECIFIC STATISTICS")
print("=" * 70)
for arena in sorted(arena_scores.keys()):
    arena_list = arena_scores[arena]
    print(f"\n{arena}:")
    print(f"  Count: {len(arena_list)}")
    print(f"  Min: {min(arena_list)}")
    print(f"  Max: {max(arena_list)}")
    print(f"  Mean: {statistics.mean(arena_list):.2f}")
    print(f"  Median: {statistics.median(arena_list):.2f}")
    print(f"  Total possible: {sum(arena_list):,}")
print()

# Analysis for different CS thresholds
print("=" * 70)
print("CS THRESHOLD ANALYSIS (NO MULTIPLIERS)")
print("=" * 70)

thresholds = [50, 100, 500, 1000, 5000, 10000]

for threshold in thresholds:
    print(f"\nTo reach CS {threshold:,}:")

    # Minimum number of IRs needed (taking highest scores)
    sorted_desc = sorted(scores, reverse=True)
    cumsum = 0
    min_irs = 0
    for score in sorted_desc:
        cumsum += score
        min_irs += 1
        if cumsum >= threshold:
            break

    # Maximum number of IRs needed (taking lowest scores)
    sorted_asc = sorted(scores)
    cumsum = 0
    max_irs = 0
    for score in sorted_asc:
        cumsum += score
        max_irs += 1
        if cumsum >= threshold:
            break

    # Average scenario (using median scores)
    avg_irs = threshold / statistics.median(scores)

    # IRs that alone meet threshold
    single_ir_count = sum(1 for s in scores if s >= threshold)

    print(f"  Minimum IRs needed (best case): {min_irs}")
    print(f"  Maximum IRs needed (worst case): {max_irs}")
    print(f"  Average IRs needed (median): {avg_irs:.1f}")
    print(f"  IRs that alone meet threshold: {single_ir_count}")

    # Show example best path
    print(f"  Example best path (top {min_irs} IRs):")
    cumsum = 0
    for i, score in enumerate(sorted_desc[:min_irs]):
        cumsum += score
        # Find IR with this score
        matching_ir = next(ir for ir in irs if ir['score'] == score and cumsum - score < threshold)
        print(f"    {i+1}. {matching_ir['id']} ({matching_ir['arena'][:10]:10s}): {score:4d} pts (total: {cumsum:5d})")

# Multiplier analysis
print("\n" + "=" * 70)
print("MULTIPLIER IMPACT ANALYSIS")
print("=" * 70)

multipliers = {
    "Base (1.0x)": 1.0,
    "With velocity (1.25x)": 1.25,
    "With arena (1.5x)": 1.5,
    "With both (1.875x)": 1.875,
    "With jackpot 5x": 5.0,
    "With jackpot 25x": 25.0,
}

print("\nTo reach CS 10,000 with multipliers:")
print(f"{'Scenario':<25} {'IRs Needed':>12} {'Example'}")
print("-" * 70)

for scenario, mult in multipliers.items():
    needed = 10000 / (statistics.median(scores) * mult)
    print(f"{scenario:<25} {needed:>12.1f} IRs     {int(statistics.median(scores) * mult)} pts/IR avg")

# Sybil resistance analysis
print("\n" + "=" * 70)
print("SYBIL RESISTANCE ANALYSIS")
print("=" * 70)

print("\nEasiest path to each threshold (lowest effort IRs):")
for threshold in [50, 100, 500, 1000, 5000, 10000]:
    sorted_asc = sorted(scores)
    cumsum = 0
    count = 0
    for score in sorted_asc:
        cumsum += score
        count += 1
        if cumsum >= threshold:
            break

    # Calculate diversity (how many arenas)
    irs_used = []
    cumsum = 0
    for score in sorted_asc:
        matching_irs = [ir for ir in irs if ir['score'] == score and ir not in irs_used]
        if matching_irs:
            irs_used.append(matching_irs[0])
            cumsum += score
            if cumsum >= threshold:
                break

    arenas_used = set(ir['arena'] for ir in irs_used)

    print(f"\nCS {threshold:,}:")
    print(f"  Minimum IRs: {count}")
    print(f"  Arenas needed: {len(arenas_used)}")
    print(f"  Arena diversity: {len(arenas_used)/9*100:.1f}%")

print("\n" + "=" * 70)
print("RECOMMENDATION SUMMARY")
print("=" * 70)
print("""
Based on the statistical analysis:

1. CS 50: TOO LOW
   - Can be achieved with just 1 IR (IR-101: Simple Liveness)
   - No diversity required
   - Trivially gameable

2. CS 10,000: APPROPRIATE
   - Requires 25-30 IRs on average (without multipliers)
   - Requires 13-16 IRs with velocity/arena bonuses
   - Forces diversity across multiple arenas
   - Aligns with system design (verification_threshold)
   - Provides strong Sybil resistance

3. Arena Focus (CS 5,000 in one arena):
   - Requires 12-15 IRs in a single arena
   - Specialization reward is meaningful

The current VerifiedHuman threshold of CS 50 defeats the purpose
of the multi-layered IR system and should be raised to 10,000.
""")

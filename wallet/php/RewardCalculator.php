<?php

/**
 * Reward-focused helpers for wallet tooling.
 *
 * @package Aequitas\Wallet
 */

declare(strict_types=1);

namespace Aequitas\Wallet;

/**
 * Aggregates reward transactions into per-category/inflow/outflow summaries.
 */
final class RewardCalculator
{
    /**
     * Summaries reward transactions grouped by categories and direction.
     *
     * @param array<array{amount?: float, category?: string, direction?: 'in'|'out'}> $transactions
     *
     * @return array{
     *     total_in: float,
     *     total_out: float,
     *     net: float,
     *     categories: array<string, float>
     * }
     */
    public function summarize(array $transactions): array
    {
        $summary = [
            'total_in' => 0.0,
            'total_out' => 0.0,
            'categories' => [],
            'net' => 0.0,
        ];

        foreach ($transactions as $entry) {
            $amount = $entry['amount'] ?? 0.0;
            $direction = $entry['direction'] ?? 'in';
            $category = $entry['category'] ?? 'uncategorized';
            $summary['net'] += $direction === 'out' ? -$amount : $amount;

            if (!isset($summary['categories'][$category])) {
                $summary['categories'][$category] = 0.0;
            }
            $summary['categories'][$category] += $direction === 'out' ? -$amount : $amount;

            if ($direction === 'out') {
                $summary['total_out'] += $amount;
            } else {
                $summary['total_in'] += $amount;
            }
        }

        return $summary;
    }
}

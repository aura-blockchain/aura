<?php

/**
 * Utility helpers for wallet balance calculations.
 *
 * @package Aequitas\Wallet
 * @subpackage BalanceHelpers
 * @author Squiz Pty Ltd <products@squiz.net>
 * @copyright Copyright (c) 2025 Squiz Pty Ltd (ABN 77 084 670 600)
 */

declare(strict_types=1);

namespace Aequitas\Wallet;

/**
 * Simple helper that derives a wallet's net balance from transaction data.
 */
final class BalanceCalculator
{
    /**
     * @param array<array{amount?: float, direction?: 'in'|'out'}> $transactions
     */
    public function calculateNetBalance(array $transactions): float
    {
        $balance = 0.0;
        foreach ($transactions as $transaction) {
            $amount = $transaction['amount'] ?? 0.0;
            $direction = $transaction['direction'] ?? 'in';
            $balance += $direction === 'out' ? -$amount : $amount;
        }

        return $balance;
    }
}

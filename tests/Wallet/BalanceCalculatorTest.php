<?php

declare(strict_types=1);

namespace Tests\Wallet;

use Aequitas\Wallet\BalanceCalculator;
use PHPUnit\Framework\TestCase;

final class BalanceCalculatorTest extends TestCase
{
    public function testCalculateNetBalanceWithMixedDirections(): void
    {
        $calculator = new BalanceCalculator();
        $transactions = [
            ['amount' => 50.0, 'direction' => 'in'],
            ['amount' => 25.5, 'direction' => 'out'],
            ['amount' => 10.0, 'direction' => 'in'],
        ];

        self::assertSame(34.5, $calculator->calculateNetBalance($transactions));
    }

    public function testCalculateNetBalanceIgnoresUnknownDirection(): void
    {
        $calculator = new BalanceCalculator();
        $transactions = [
            ['amount' => 100.0, 'direction' => 'in'],
            ['amount' => 20.0, 'direction' => 'fee'],
        ];

        self::assertSame(120.0, $calculator->calculateNetBalance($transactions));
    }
}

<?php

declare(strict_types=1);

namespace Tests\Wallet;

use Aequitas\Wallet\RewardCalculator;
use PHPUnit\Framework\TestCase;

final class RewardCalculatorTest extends TestCase
{
    public function testSummarizeInflowsAndOutflows(): void
    {
        $calculator = new RewardCalculator();
        $transactions = [
            ['amount' => 100.0, 'category' => 'staking', 'direction' => 'in'],
            ['amount' => 30.5, 'category' => 'staking', 'direction' => 'out'],
            ['amount' => 10.0, 'category' => 'fees', 'direction' => 'out'],
        ];

        $summary = $calculator->summarize($transactions);

        self::assertSame(100.0, $summary['total_in']);
        self::assertSame(40.5, $summary['total_out']);
        self::assertSame(59.5, $summary['net']);
        self::assertSame(69.5, $summary['categories']['staking']);
        self::assertSame(-10.0, $summary['categories']['fees']);
    }

    public function testSummarizeHandlesMissingMetadata(): void
    {
        $calculator = new RewardCalculator();
        $summary = $calculator->summarize([
            ['amount' => 5.0],
            ['amount' => 2.5, 'direction' => 'out', 'category' => 'adjustment'],
        ]);

        self::assertSame(5.0, $summary['total_in']);
        self::assertSame(2.5, $summary['total_out']);
        self::assertSame(2.5, $summary['net']);
        self::assertSame(5.0, $summary['categories']['uncategorized']);
        self::assertSame(-2.5, $summary['categories']['adjustment']);
    }

    public function testSummarizeReturnsZeroedStructureForEmptyInput(): void
    {
        $calculator = new RewardCalculator();
        $summary = $calculator->summarize([]);

        self::assertSame(0.0, $summary['total_in']);
        self::assertSame(0.0, $summary['total_out']);
        self::assertSame(0.0, $summary['net']);
        self::assertSame([], $summary['categories']);
    }
}

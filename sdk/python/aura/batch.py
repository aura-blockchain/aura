"""
Batch Operation Helpers for Aura SDK

Provides utilities for batching multiple transactions and queries.
"""

import asyncio
from dataclasses import dataclass, field
from typing import Any, Callable, Generic, Optional, TypeVar, Union

T = TypeVar("T")


# ============================================================================
# Types
# ============================================================================


@dataclass
class Coin:
    """Coin type for token amounts."""

    denom: str
    amount: str


@dataclass
class BatchTransferItem:
    """Item for batch transfer."""

    recipient: str
    amount: list[Coin]


@dataclass
class BatchResult(Generic[T]):
    """Result of a batch operation."""

    success: bool
    results: list[T]
    errors: list[tuple[int, Exception]]
    tx_hash: Optional[str] = None


@dataclass
class BatchQueryResult(Generic[T]):
    """Result of batch queries."""

    results: list[T]
    errors: list[tuple[int, Exception]]


@dataclass
class BatchConfig:
    """Configuration for batch operations."""

    max_batch_size: int = 50
    parallel_queries: int = 10
    retry_on_failure: bool = True
    retry_attempts: int = 3
    delay_between_retries: float = 1.0


@dataclass
class Message:
    """Generic message for transactions."""

    type_url: str
    value: dict[str, Any]


# ============================================================================
# Transaction Batching
# ============================================================================


class BatchTransactionBuilder:
    """Builder for multi-message transactions."""

    def __init__(self) -> None:
        self._messages: list[Message] = []
        self._memo: str = ""

    def add_message(self, msg: Message) -> "BatchTransactionBuilder":
        """Add a message to the batch."""
        self._messages.append(msg)
        return self

    def add_messages(self, msgs: list[Message]) -> "BatchTransactionBuilder":
        """Add multiple messages to the batch."""
        self._messages.extend(msgs)
        return self

    def add_send(
        self, sender: str, recipient: str, amount: list[Coin]
    ) -> "BatchTransactionBuilder":
        """Add a bank send message."""
        self._messages.append(
            Message(
                type_url="/cosmos.bank.v1beta1.MsgSend",
                value={
                    "from_address": sender,
                    "to_address": recipient,
                    "amount": [{"denom": c.denom, "amount": c.amount} for c in amount],
                },
            )
        )
        return self

    def add_batch_sends(
        self, sender: str, transfers: list[BatchTransferItem]
    ) -> "BatchTransactionBuilder":
        """Add multiple send messages for batch transfers."""
        for transfer in transfers:
            self.add_send(sender, transfer.recipient, transfer.amount)
        return self

    def add_delegate(
        self, delegator: str, validator: str, amount: Coin
    ) -> "BatchTransactionBuilder":
        """Add a delegate message."""
        self._messages.append(
            Message(
                type_url="/cosmos.staking.v1beta1.MsgDelegate",
                value={
                    "delegator_address": delegator,
                    "validator_address": validator,
                    "amount": {"denom": amount.denom, "amount": amount.amount},
                },
            )
        )
        return self

    def add_undelegate(
        self, delegator: str, validator: str, amount: Coin
    ) -> "BatchTransactionBuilder":
        """Add an undelegate message."""
        self._messages.append(
            Message(
                type_url="/cosmos.staking.v1beta1.MsgUndelegate",
                value={
                    "delegator_address": delegator,
                    "validator_address": validator,
                    "amount": {"denom": amount.denom, "amount": amount.amount},
                },
            )
        )
        return self

    def add_vote(
        self, voter: str, proposal_id: str, option: int
    ) -> "BatchTransactionBuilder":
        """Add a governance vote message."""
        self._messages.append(
            Message(
                type_url="/cosmos.gov.v1beta1.MsgVote",
                value={
                    "proposal_id": proposal_id,
                    "voter": voter,
                    "option": option,
                },
            )
        )
        return self

    def with_memo(self, memo: str) -> "BatchTransactionBuilder":
        """Set transaction memo."""
        self._memo = memo
        return self

    def get_messages(self) -> list[Message]:
        """Get all messages in the batch."""
        return list(self._messages)

    def get_memo(self) -> str:
        """Get the memo."""
        return self._memo

    def size(self) -> int:
        """Get the number of messages."""
        return len(self._messages)

    def clear(self) -> "BatchTransactionBuilder":
        """Clear all messages."""
        self._messages = []
        self._memo = ""
        return self


# ============================================================================
# Query Batching
# ============================================================================


QueryFunction = Callable[[], T]
AsyncQueryFunction = Callable[[], Any]  # Coroutine returning T


async def batch_queries(
    queries: list[Union[QueryFunction[T], AsyncQueryFunction]],
    config: Optional[BatchConfig] = None,
) -> BatchQueryResult[T]:
    """Execute multiple queries in parallel with batching."""
    cfg = config or BatchConfig()
    results: list[T] = [None] * len(queries)  # type: ignore
    errors: list[tuple[int, Exception]] = []

    # Split into batches
    batches = list(chunk(list(enumerate(queries)), cfg.parallel_queries))

    for batch in batches:
        tasks = []
        for index, query in batch:
            tasks.append(_execute_query(index, query, cfg))

        batch_results = await asyncio.gather(*tasks, return_exceptions=True)

        for result in batch_results:
            if isinstance(result, Exception):
                # This shouldn't happen as we catch exceptions in _execute_query
                continue
            idx, res, err = result
            if err:
                errors.append((idx, err))
            else:
                results[idx] = res

    return BatchQueryResult(results=results, errors=errors)


async def _execute_query(
    index: int,
    query: Union[QueryFunction[T], AsyncQueryFunction],
    config: BatchConfig,
) -> tuple[int, Optional[T], Optional[Exception]]:
    """Execute a single query with retry logic."""
    last_error: Optional[Exception] = None

    for attempt in range(config.retry_attempts):
        try:
            result = query()
            if asyncio.iscoroutine(result):
                result = await result
            return (index, result, None)
        except Exception as e:
            last_error = e
            if not config.retry_on_failure or attempt == config.retry_attempts - 1:
                return (index, None, e)
            await asyncio.sleep(config.delay_between_retries * (2**attempt))

    return (index, None, last_error)


def batch_queries_sync(
    queries: list[QueryFunction[T]],
    config: Optional[BatchConfig] = None,
) -> BatchQueryResult[T]:
    """Execute multiple queries synchronously (one at a time)."""
    cfg = config or BatchConfig()
    results: list[T] = []
    errors: list[tuple[int, Exception]] = []

    for index, query in enumerate(queries):
        try:
            result = _execute_with_retry_sync(query, cfg)
            results.append(result)
        except Exception as e:
            errors.append((index, e))
            results.append(None)  # type: ignore

    return BatchQueryResult(results=results, errors=errors)


def _execute_with_retry_sync(
    fn: QueryFunction[T],
    config: BatchConfig,
) -> T:
    """Execute a function with retry logic (synchronous)."""
    last_error: Optional[Exception] = None

    for attempt in range(config.retry_attempts):
        try:
            return fn()
        except Exception as e:
            last_error = e
            if not config.retry_on_failure or attempt == config.retry_attempts - 1:
                raise
            import time

            time.sleep(config.delay_between_retries * (2**attempt))

    raise last_error  # type: ignore


# ============================================================================
# Utility Functions
# ============================================================================


def chunk(items: list[T], size: int) -> list[list[T]]:
    """Split a list into chunks of specified size."""
    return [items[i : i + size] for i in range(0, len(items), size)]


def create_batch_transfers() -> BatchTransactionBuilder:
    """Create a batch transfer builder."""
    return BatchTransactionBuilder()


def validate_batch_size(items: list[Any], max_size: int = 50) -> None:
    """Validate batch size."""
    if len(items) > max_size:
        raise ValueError(
            f"Batch size {len(items)} exceeds maximum allowed size of {max_size}"
        )


def estimate_batch_gas(
    message_count: int,
    base_gas_per_message: int = 100000,
    overhead_gas: int = 50000,
) -> int:
    """Calculate estimated gas for batch operations."""
    return overhead_gas + message_count * base_gas_per_message


# ============================================================================
# Multi-Sig Batch Helpers
# ============================================================================


@dataclass
class MultiSigBatchItem:
    """Item for multi-sig batch."""

    messages: list[Message]
    memo: str = ""


@dataclass
class MultiSigBatch:
    """Result of creating a multi-sig batch."""

    transactions: list[dict[str, Any]]
    total_messages: int


def create_multi_sig_batch(items: list[MultiSigBatchItem]) -> MultiSigBatch:
    """Create multiple transactions for multi-sig signing."""
    transactions = [
        {"messages": item.messages, "memo": item.memo} for item in items
    ]
    total_messages = sum(len(item.messages) for item in items)
    return MultiSigBatch(transactions=transactions, total_messages=total_messages)

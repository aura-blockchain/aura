"""
Event Subscription System for Aura SDK

Provides typed event subscriptions for blockchain events.
"""

import asyncio
import json
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Optional, Union
import websockets
from websockets.exceptions import ConnectionClosed


# ============================================================================
# Event Types
# ============================================================================


class EventType(str, Enum):
    """Enum of event types."""

    BLOCK = "block"
    TX = "tx"
    TRANSFER = "transfer"
    BRIDGE_TRANSFER = "bridge_transfer"
    IDENTITY = "identity"
    GOVERNANCE = "governance"
    DEX = "dex"


@dataclass
class BlockEvent:
    """New block event."""

    type: str = field(default="block")
    height: int = 0
    hash: str = ""
    timestamp: str = ""
    tx_count: int = 0


@dataclass
class TxEvent:
    """Transaction event."""

    type: str = field(default="tx")
    hash: str = ""
    height: int = 0
    code: int = 0
    gas_used: int = 0
    gas_wanted: int = 0
    events: list[dict[str, Any]] = field(default_factory=list)


@dataclass
class TransferEvent:
    """Token transfer event."""

    type: str = field(default="transfer")
    sender: str = ""
    recipient: str = ""
    amount: str = ""
    denom: str = ""


@dataclass
class BridgeTransferEvent:
    """Bridge transfer event."""

    type: str = field(default="bridge_transfer")
    transfer_id: str = ""
    sender: str = ""
    recipient: str = ""
    amount: str = ""
    target_chain: str = ""
    status: str = ""  # initiated, completed, failed


@dataclass
class IdentityEvent:
    """Identity event."""

    type: str = field(default="identity")
    action: str = ""  # created, updated, deleted
    did: str = ""
    owner: str = ""


@dataclass
class GovernanceEvent:
    """Governance event."""

    type: str = field(default="governance")
    action: str = ""  # proposal_submitted, vote_cast, proposal_passed, proposal_rejected
    proposal_id: str = ""
    voter: Optional[str] = None


@dataclass
class DEXEvent:
    """DEX event."""

    type: str = field(default="dex")
    action: str = ""  # swap, add_liquidity, remove_liquidity, order_created, order_filled
    pool_id: Optional[str] = None
    sender: str = ""
    amounts: list[str] = field(default_factory=list)


AuraEvent = Union[
    BlockEvent,
    TxEvent,
    TransferEvent,
    BridgeTransferEvent,
    IdentityEvent,
    GovernanceEvent,
    DEXEvent,
]


# ============================================================================
# Event Filter
# ============================================================================


@dataclass
class EventFilter:
    """Filter for event subscriptions."""

    type: Optional[Union[str, list[str]]] = None
    sender: Optional[str] = None
    recipient: Optional[str] = None
    module: Optional[str] = None
    action: Optional[str] = None
    min_height: Optional[int] = None
    max_height: Optional[int] = None


EventHandler = Callable[[AuraEvent], None]
AsyncEventHandler = Callable[[AuraEvent], Any]  # Coroutine
ErrorHandler = Callable[[Exception], None]


# ============================================================================
# Subscription
# ============================================================================


@dataclass
class Subscription:
    """Event subscription."""

    id: str
    filter: EventFilter
    handler: Union[EventHandler, AsyncEventHandler]


# ============================================================================
# Event Subscriber
# ============================================================================


class EventSubscriber:
    """Manages WebSocket connections and event subscriptions."""

    def __init__(self, ws_endpoint: str):
        self.ws_endpoint = ws_endpoint
        self._ws: Optional[websockets.WebSocketClientProtocol] = None
        self._subscriptions: dict[str, Subscription] = {}
        self._error_handler: Optional[ErrorHandler] = None
        self._reconnect_attempts = 0
        self._max_reconnect_attempts = 5
        self._reconnect_delay = 1.0
        self._is_connected = False
        self._subscription_counter = 0
        self._running = False

    async def connect(self) -> None:
        """Connect to the WebSocket endpoint."""
        try:
            self._ws = await websockets.connect(self.ws_endpoint)
            self._is_connected = True
            self._reconnect_attempts = 0
            await self._resubscribe_all()
        except Exception as e:
            if self._error_handler:
                self._error_handler(e)
            raise

    async def disconnect(self) -> None:
        """Disconnect from the WebSocket endpoint."""
        self._running = False
        if self._ws:
            await self._ws.close()
            self._ws = None
        self._is_connected = False
        self._subscriptions.clear()

    def subscribe(
        self,
        filter_: EventFilter,
        handler: Union[EventHandler, AsyncEventHandler],
    ) -> str:
        """Subscribe to events matching the filter."""
        self._subscription_counter += 1
        sub_id = f"sub_{self._subscription_counter}"

        self._subscriptions[sub_id] = Subscription(
            id=sub_id,
            filter=filter_,
            handler=handler,
        )

        return sub_id

    def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from events."""
        if subscription_id in self._subscriptions:
            del self._subscriptions[subscription_id]
            return True
        return False

    def on_block(self, handler: Callable[[BlockEvent], None]) -> str:
        """Subscribe to new blocks."""
        return self.subscribe(EventFilter(type="block"), handler)

    def on_transaction(
        self,
        handler: Callable[[TxEvent], None],
        filter_: Optional[EventFilter] = None,
    ) -> str:
        """Subscribe to transactions."""
        f = filter_ or EventFilter()
        f.type = "tx"
        return self.subscribe(f, handler)

    def on_transfer(
        self,
        address: str,
        handler: Callable[[TransferEvent], None],
    ) -> str:
        """Subscribe to transfers involving an address."""
        return self.subscribe(
            EventFilter(type="transfer", sender=address),
            handler,
        )

    def on_bridge_transfer(
        self, handler: Callable[[BridgeTransferEvent], None]
    ) -> str:
        """Subscribe to bridge transfer events."""
        return self.subscribe(
            EventFilter(type="bridge_transfer", module="bridge"),
            handler,
        )

    def on_identity(self, handler: Callable[[IdentityEvent], None]) -> str:
        """Subscribe to identity events."""
        return self.subscribe(
            EventFilter(type="identity", module="identity"),
            handler,
        )

    def on_governance(self, handler: Callable[[GovernanceEvent], None]) -> str:
        """Subscribe to governance events."""
        return self.subscribe(
            EventFilter(type="governance", module="governance"),
            handler,
        )

    def on_dex(self, handler: Callable[[DEXEvent], None]) -> str:
        """Subscribe to DEX events."""
        return self.subscribe(
            EventFilter(type="dex", module="dex"),
            handler,
        )

    def on_error(self, handler: ErrorHandler) -> None:
        """Set error handler."""
        self._error_handler = handler

    @property
    def connected(self) -> bool:
        """Check if connected."""
        return self._is_connected

    async def run(self) -> None:
        """Run the event loop to process incoming messages."""
        self._running = True

        while self._running:
            try:
                if not self._ws or not self._is_connected:
                    await self.connect()

                async for message in self._ws:  # type: ignore
                    if not self._running:
                        break
                    await self._handle_message(str(message))

            except ConnectionClosed:
                self._is_connected = False
                if self._running:
                    await self._attempt_reconnect()
            except Exception as e:
                if self._error_handler:
                    self._error_handler(e)
                if self._running:
                    await self._attempt_reconnect()

    async def _handle_message(self, data: str) -> None:
        """Handle incoming WebSocket message."""
        try:
            message = json.loads(data)
            event = self._parse_event(message)

            if event:
                for subscription in self._subscriptions.values():
                    if self._matches_filter(event, subscription.filter):
                        try:
                            result = subscription.handler(event)
                            if asyncio.iscoroutine(result):
                                await result
                        except Exception as e:
                            if self._error_handler:
                                self._error_handler(e)
        except Exception as e:
            if self._error_handler:
                self._error_handler(e)

    def _parse_event(self, message: dict[str, Any]) -> Optional[AuraEvent]:
        """Parse WebSocket message into event."""
        result = message.get("result", {})
        data = result.get("data", {})

        if not data:
            return None

        event_type = data.get("type", "")

        if event_type == "tendermint/event/NewBlock":
            return self._parse_block_event(data)
        elif event_type == "tendermint/event/Tx":
            return self._parse_tx_event(data)

        return None

    def _parse_block_event(self, data: dict[str, Any]) -> Optional[BlockEvent]:
        """Parse block event."""
        value = data.get("value", {})
        block = value.get("block", {})
        header = block.get("header", {})

        if not header:
            return None

        return BlockEvent(
            height=int(header.get("height", 0)),
            hash=block.get("last_commit", {}).get("block_id", ""),
            timestamp=header.get("time", ""),
            tx_count=len(block.get("data", {}).get("txs", [])),
        )

    def _parse_tx_event(self, data: dict[str, Any]) -> Optional[TxEvent]:
        """Parse transaction event."""
        value = data.get("value", {})
        tx_result = value.get("TxResult", {})

        if not tx_result:
            return None

        result = tx_result.get("result", {})

        return TxEvent(
            hash=tx_result.get("hash", ""),
            height=int(tx_result.get("height", 0)),
            code=result.get("code", 0),
            gas_used=int(result.get("gas_used", 0)),
            gas_wanted=int(result.get("gas_wanted", 0)),
            events=result.get("events", []),
        )

    def _matches_filter(self, event: AuraEvent, filter_: EventFilter) -> bool:
        """Check if event matches filter."""
        if filter_.type:
            types = [filter_.type] if isinstance(filter_.type, str) else filter_.type
            if event.type not in types:
                return False

        if filter_.min_height and hasattr(event, "height"):
            if event.height < filter_.min_height:  # type: ignore
                return False

        if filter_.max_height and hasattr(event, "height"):
            if event.height > filter_.max_height:  # type: ignore
                return False

        if filter_.sender and hasattr(event, "sender"):
            if event.sender != filter_.sender:  # type: ignore
                return False

        if filter_.recipient and hasattr(event, "recipient"):
            if event.recipient != filter_.recipient:  # type: ignore
                return False

        return True

    async def _send_subscription(self, filter_: EventFilter) -> None:
        """Send subscription request."""
        if not self._ws or not self._is_connected:
            return

        query = self._build_query(filter_)
        message = {
            "jsonrpc": "2.0",
            "method": "subscribe",
            "id": int(asyncio.get_event_loop().time() * 1000),
            "params": {"query": query},
        }

        await self._ws.send(json.dumps(message))

    def _build_query(self, filter_: EventFilter) -> str:
        """Build Tendermint query string."""
        conditions: list[str] = []

        if filter_.type == "block":
            conditions.append("tm.event='NewBlock'")
        elif filter_.type == "tx":
            conditions.append("tm.event='Tx'")
        else:
            conditions.append("tm.event='Tx'")

        if filter_.module:
            conditions.append(f"message.module='{filter_.module}'")

        if filter_.action:
            conditions.append(f"message.action='{filter_.action}'")

        if filter_.sender:
            conditions.append(f"message.sender='{filter_.sender}'")

        return " AND ".join(conditions)

    async def _resubscribe_all(self) -> None:
        """Resubscribe to all subscriptions after reconnection."""
        for subscription in self._subscriptions.values():
            await self._send_subscription(subscription.filter)

    async def _attempt_reconnect(self) -> None:
        """Attempt to reconnect with exponential backoff."""
        if self._reconnect_attempts >= self._max_reconnect_attempts:
            if self._error_handler:
                self._error_handler(Exception("Max reconnection attempts reached"))
            self._running = False
            return

        self._reconnect_attempts += 1
        delay = self._reconnect_delay * (2 ** (self._reconnect_attempts - 1))

        await asyncio.sleep(delay)

        try:
            await self.connect()
        except Exception as e:
            if self._error_handler:
                self._error_handler(e)


def create_event_subscriber(ws_endpoint: str) -> EventSubscriber:
    """Create an event subscriber for the given WebSocket endpoint."""
    return EventSubscriber(ws_endpoint)

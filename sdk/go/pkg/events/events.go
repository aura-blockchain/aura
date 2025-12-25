// Package events provides event subscription functionality for the Aura SDK.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EventType represents the type of blockchain event
type EventType string

const (
	EventTypeBlock          EventType = "block"
	EventTypeTx             EventType = "tx"
	EventTypeTransfer       EventType = "transfer"
	EventTypeBridgeTransfer EventType = "bridge_transfer"
	EventTypeIdentity       EventType = "identity"
	EventTypeGovernance     EventType = "governance"
	EventTypeDEX            EventType = "dex"
)

// Event is the interface for all event types
type Event interface {
	Type() EventType
}

// BlockEvent represents a new block event
type BlockEvent struct {
	Height    int64  `json:"height"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
	TxCount   int    `json:"tx_count"`
}

func (e BlockEvent) Type() EventType { return EventTypeBlock }

// TxEvent represents a transaction event
type TxEvent struct {
	Hash      string           `json:"hash"`
	Height    int64            `json:"height"`
	Code      uint32           `json:"code"`
	GasUsed   int64            `json:"gas_used"`
	GasWanted int64            `json:"gas_wanted"`
	Events    []EventAttribute `json:"events"`
}

func (e TxEvent) Type() EventType { return EventTypeTx }

// EventAttribute represents an event attribute
type EventAttribute struct {
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes"`
}

// TransferEvent represents a token transfer event
type TransferEvent struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
	Denom     string `json:"denom"`
}

func (e TransferEvent) Type() EventType { return EventTypeTransfer }

// BridgeTransferEvent represents a bridge transfer event
type BridgeTransferEvent struct {
	TransferID  string `json:"transfer_id"`
	Sender      string `json:"sender"`
	Recipient   string `json:"recipient"`
	Amount      string `json:"amount"`
	TargetChain string `json:"target_chain"`
	Status      string `json:"status"` // initiated, completed, failed
}

func (e BridgeTransferEvent) Type() EventType { return EventTypeBridgeTransfer }

// IdentityEvent represents an identity event
type IdentityEvent struct {
	Action string `json:"action"` // created, updated, deleted
	DID    string `json:"did"`
	Owner  string `json:"owner"`
}

func (e IdentityEvent) Type() EventType { return EventTypeIdentity }

// GovernanceEvent represents a governance event
type GovernanceEvent struct {
	Action     string `json:"action"` // proposal_submitted, vote_cast, proposal_passed, proposal_rejected
	ProposalID string `json:"proposal_id"`
	Voter      string `json:"voter,omitempty"`
}

func (e GovernanceEvent) Type() EventType { return EventTypeGovernance }

// DEXEvent represents a DEX event
type DEXEvent struct {
	Action  string   `json:"action"` // swap, add_liquidity, remove_liquidity, order_created, order_filled
	PoolID  string   `json:"pool_id,omitempty"`
	Sender  string   `json:"sender"`
	Amounts []string `json:"amounts,omitempty"`
}

func (e DEXEvent) Type() EventType { return EventTypeDEX }

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Type      EventType `json:"type,omitempty"`
	Sender    string    `json:"sender,omitempty"`
	Recipient string    `json:"recipient,omitempty"`
	Module    string    `json:"module,omitempty"`
	Action    string    `json:"action,omitempty"`
	MinHeight int64     `json:"min_height,omitempty"`
	MaxHeight int64     `json:"max_height,omitempty"`
}

// EventHandler is a callback function for handling events
type EventHandler func(Event)

// ErrorHandler is a callback function for handling errors
type ErrorHandler func(error)

// Subscription represents an active event subscription
type Subscription struct {
	ID      string
	Filter  EventFilter
	Handler EventHandler
}

// Subscriber manages WebSocket connections and event subscriptions
type Subscriber struct {
	wsEndpoint           string
	conn                 *websocket.Conn
	subscriptions        map[string]*Subscription
	errorHandler         ErrorHandler
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	isConnected          bool
	subscriptionCounter  int
	mu                   sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
}

// NewSubscriber creates a new event subscriber
func NewSubscriber(wsEndpoint string) *Subscriber {
	return &Subscriber{
		wsEndpoint:           wsEndpoint,
		subscriptions:        make(map[string]*Subscription),
		maxReconnectAttempts: 5,
		reconnectDelay:       time.Second,
	}
}

// Connect establishes a WebSocket connection
func (s *Subscriber) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(ctx)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(s.ctx, s.wsEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	s.conn = conn
	s.isConnected = true
	s.reconnectAttempts = 0

	// Resubscribe to all existing subscriptions
	for _, sub := range s.subscriptions {
		if err := s.sendSubscription(sub.Filter); err != nil {
			return fmt.Errorf("failed to resubscribe: %w", err)
		}
	}

	return nil
}

// Disconnect closes the WebSocket connection
func (s *Subscriber) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.isConnected = false
		return err
	}

	return nil
}

// Subscribe adds a new event subscription
func (s *Subscriber) Subscribe(filter EventFilter, handler EventHandler) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptionCounter++
	id := fmt.Sprintf("sub_%d", s.subscriptionCounter)

	s.subscriptions[id] = &Subscription{
		ID:      id,
		Filter:  filter,
		Handler: handler,
	}

	if s.isConnected {
		_ = s.sendSubscription(filter)
	}

	return id
}

// Unsubscribe removes an event subscription
func (s *Subscriber) Unsubscribe(subscriptionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subscriptions[subscriptionID]; exists {
		delete(s.subscriptions, subscriptionID)
		return true
	}
	return false
}

// OnBlock subscribes to new block events
func (s *Subscriber) OnBlock(handler func(BlockEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeBlock}, func(e Event) {
		if be, ok := e.(BlockEvent); ok {
			handler(be)
		}
	})
}

// OnTransaction subscribes to transaction events
func (s *Subscriber) OnTransaction(handler func(TxEvent), filter *EventFilter) string {
	f := EventFilter{Type: EventTypeTx}
	if filter != nil {
		f.Sender = filter.Sender
		f.MinHeight = filter.MinHeight
		f.MaxHeight = filter.MaxHeight
	}
	return s.Subscribe(f, func(e Event) {
		if te, ok := e.(TxEvent); ok {
			handler(te)
		}
	})
}

// OnTransfer subscribes to transfer events for an address
func (s *Subscriber) OnTransfer(address string, handler func(TransferEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeTransfer, Sender: address}, func(e Event) {
		if te, ok := e.(TransferEvent); ok {
			handler(te)
		}
	})
}

// OnBridgeTransfer subscribes to bridge transfer events
func (s *Subscriber) OnBridgeTransfer(handler func(BridgeTransferEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeBridgeTransfer, Module: "bridge"}, func(e Event) {
		if bte, ok := e.(BridgeTransferEvent); ok {
			handler(bte)
		}
	})
}

// OnIdentity subscribes to identity events
func (s *Subscriber) OnIdentity(handler func(IdentityEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeIdentity, Module: "identity"}, func(e Event) {
		if ie, ok := e.(IdentityEvent); ok {
			handler(ie)
		}
	})
}

// OnGovernance subscribes to governance events
func (s *Subscriber) OnGovernance(handler func(GovernanceEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeGovernance, Module: "governance"}, func(e Event) {
		if ge, ok := e.(GovernanceEvent); ok {
			handler(ge)
		}
	})
}

// OnDEX subscribes to DEX events
func (s *Subscriber) OnDEX(handler func(DEXEvent)) string {
	return s.Subscribe(EventFilter{Type: EventTypeDEX, Module: "dex"}, func(e Event) {
		if de, ok := e.(DEXEvent); ok {
			handler(de)
		}
	})
}

// OnError sets the error handler
func (s *Subscriber) OnError(handler ErrorHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorHandler = handler
}

// IsConnected returns connection status
func (s *Subscriber) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected
}

// Run starts the event loop
func (s *Subscriber) Run() error {
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
			if !s.isConnected {
				if err := s.Connect(s.ctx); err != nil {
					s.handleError(err)
					if err := s.attemptReconnect(); err != nil {
						return err
					}
					continue
				}
			}

			_, message, err := s.conn.ReadMessage()
			if err != nil {
				s.mu.Lock()
				s.isConnected = false
				s.mu.Unlock()
				s.handleError(err)
				if err := s.attemptReconnect(); err != nil {
					return err
				}
				continue
			}

			s.handleMessage(message)
		}
	}
}

func (s *Subscriber) handleMessage(data []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		s.handleError(err)
		return
	}

	event := s.parseEvent(msg)
	if event == nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sub := range s.subscriptions {
		if s.matchesFilter(event, sub.Filter) {
			go sub.Handler(event)
		}
	}
}

func (s *Subscriber) parseEvent(msg map[string]interface{}) Event {
	result, ok := msg["result"].(map[string]interface{})
	if !ok {
		return nil
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	eventType, _ := data["type"].(string)

	switch eventType {
	case "tendermint/event/NewBlock":
		return s.parseBlockEvent(data)
	case "tendermint/event/Tx":
		return s.parseTxEvent(data)
	}

	return nil
}

func (s *Subscriber) parseBlockEvent(data map[string]interface{}) Event {
	value, ok := data["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	block, ok := value["block"].(map[string]interface{})
	if !ok {
		return nil
	}

	header, ok := block["header"].(map[string]interface{})
	if !ok {
		return nil
	}

	height, _ := header["height"].(float64)
	timestamp, _ := header["time"].(string)

	var txCount int
	if blockData, ok := block["data"].(map[string]interface{}); ok {
		if txs, ok := blockData["txs"].([]interface{}); ok {
			txCount = len(txs)
		}
	}

	return BlockEvent{
		Height:    int64(height),
		Timestamp: timestamp,
		TxCount:   txCount,
	}
}

func (s *Subscriber) parseTxEvent(data map[string]interface{}) Event {
	value, ok := data["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	txResult, ok := value["TxResult"].(map[string]interface{})
	if !ok {
		return nil
	}

	hash, _ := txResult["hash"].(string)
	height, _ := txResult["height"].(float64)

	result, _ := txResult["result"].(map[string]interface{})
	code, _ := result["code"].(float64)
	gasUsed, _ := result["gas_used"].(float64)
	gasWanted, _ := result["gas_wanted"].(float64)

	return TxEvent{
		Hash:      hash,
		Height:    int64(height),
		Code:      uint32(code),
		GasUsed:   int64(gasUsed),
		GasWanted: int64(gasWanted),
	}
}

func (s *Subscriber) matchesFilter(event Event, filter EventFilter) bool {
	if filter.Type != "" && filter.Type != event.Type() {
		return false
	}

	switch e := event.(type) {
	case BlockEvent:
		if filter.MinHeight > 0 && e.Height < filter.MinHeight {
			return false
		}
		if filter.MaxHeight > 0 && e.Height > filter.MaxHeight {
			return false
		}
	case TxEvent:
		if filter.MinHeight > 0 && e.Height < filter.MinHeight {
			return false
		}
		if filter.MaxHeight > 0 && e.Height > filter.MaxHeight {
			return false
		}
	case TransferEvent:
		if filter.Sender != "" && e.Sender != filter.Sender {
			return false
		}
		if filter.Recipient != "" && e.Recipient != filter.Recipient {
			return false
		}
	}

	return true
}

func (s *Subscriber) sendSubscription(filter EventFilter) error {
	if s.conn == nil {
		return fmt.Errorf("not connected")
	}

	query := s.buildQuery(filter)
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "subscribe",
		"id":      time.Now().UnixNano(),
		"params":  map[string]string{"query": query},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return s.conn.WriteMessage(websocket.TextMessage, data)
}

func (s *Subscriber) buildQuery(filter EventFilter) string {
	var conditions []string

	switch filter.Type {
	case EventTypeBlock:
		conditions = append(conditions, "tm.event='NewBlock'")
	default:
		conditions = append(conditions, "tm.event='Tx'")
	}

	if filter.Module != "" {
		conditions = append(conditions, fmt.Sprintf("message.module='%s'", filter.Module))
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("message.action='%s'", filter.Action))
	}

	if filter.Sender != "" {
		conditions = append(conditions, fmt.Sprintf("message.sender='%s'", filter.Sender))
	}

	query := ""
	for i, c := range conditions {
		if i > 0 {
			query += " AND "
		}
		query += c
	}

	return query
}

func (s *Subscriber) handleError(err error) {
	s.mu.RLock()
	handler := s.errorHandler
	s.mu.RUnlock()

	if handler != nil {
		handler(err)
	}
}

func (s *Subscriber) attemptReconnect() error {
	if s.reconnectAttempts >= s.maxReconnectAttempts {
		return fmt.Errorf("max reconnection attempts reached")
	}

	s.reconnectAttempts++
	delay := s.reconnectDelay * time.Duration(1<<(s.reconnectAttempts-1))

	select {
	case <-time.After(delay):
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(wsEndpoint string) *Subscriber {
	return NewSubscriber(wsEndpoint)
}

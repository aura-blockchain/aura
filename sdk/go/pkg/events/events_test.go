package events

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{}

func startWSServer(t *testing.T, message []byte) (string, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		// Wait for subscription message then send test payload
		_, _, _ = conn.ReadMessage()
		_ = conn.WriteMessage(websocket.TextMessage, message)
	}))

	return "ws" + server.URL[4:], server.Close
}

func TestSubscriberReceivesBlockEvent(t *testing.T) {
	blockMsg := map[string]interface{}{
		"result": map[string]interface{}{
			"data": map[string]interface{}{
				"type": "tendermint/event/NewBlock",
				"value": map[string]interface{}{
					"block": map[string]interface{}{
						"header": map[string]interface{}{
							"height": float64(5),
							"time":   "2026-01-15T00:00:00Z",
						},
						"data": map[string]interface{}{
							"txs": []interface{}{},
						},
					},
				},
			},
		},
	}
	payload, _ := json.Marshal(blockMsg)

	wsURL, closeFn := startWSServer(t, payload)
	defer closeFn()

	sub := NewSubscriber(wsURL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var got BlockEvent
	done := make(chan struct{})

	sub.OnBlock(func(be BlockEvent) {
		got = be
		close(done)
	})

	require.NoError(t, sub.Connect(ctx))
	go func() {
		_ = sub.Run()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("timeout waiting for block event")
	}

	assert.Equal(t, int64(5), got.Height)
	assert.Equal(t, "2026-01-15T00:00:00Z", got.Timestamp)
}

func TestSubscriberFiltersTxHeight(t *testing.T) {
	txMsg := map[string]interface{}{
		"result": map[string]interface{}{
			"data": map[string]interface{}{
				"type": "tendermint/event/Tx",
				"value": map[string]interface{}{
					"TxResult": map[string]interface{}{
						"hash":   "ABC123",
						"height": float64(10),
						"result": map[string]interface{}{
							"code":       float64(0),
							"gas_used":   float64(1000),
							"gas_wanted": float64(2000),
						},
					},
				},
			},
		},
	}
	payload, _ := json.Marshal(txMsg)

	wsURL, closeFn := startWSServer(t, payload)
	defer closeFn()

	sub := NewSubscriber(wsURL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var mux sync.Mutex
	var received []TxEvent

	sub.OnTransaction(func(te TxEvent) {
		mux.Lock()
		defer mux.Unlock()
		received = append(received, te)
	}, &EventFilter{MinHeight: 5, MaxHeight: 15})

	require.NoError(t, sub.Connect(ctx))
	go func() { _ = sub.Run() }()

	<-time.After(200 * time.Millisecond)

	mux.Lock()
	defer mux.Unlock()
	require.Len(t, received, 1)
	assert.Equal(t, int64(10), received[0].Height)
	assert.Equal(t, "ABC123", received[0].Hash)
}

func TestBuildQuery(t *testing.T) {
	sub := NewSubscriber("ws://example")
	query := sub.buildQuery(EventFilter{
		Type:   EventTypeTx,
		Module: "identity",
		Action: "create",
		Sender: "aura1abc",
	})

	assert.Contains(t, query, "tm.event='Tx'")
	assert.Contains(t, query, "message.module='identity'")
	assert.Contains(t, query, "message.action='create'")
	assert.Contains(t, query, "message.sender='aura1abc'")
}

func TestUnsubscribe(t *testing.T) {
	sub := NewSubscriber("ws://example")
	id := sub.Subscribe(EventFilter{Type: EventTypeBlock}, func(Event) {})
	assert.True(t, sub.Unsubscribe(id))
	assert.False(t, sub.Unsubscribe(id)) // already removed
}

/**
 * WebSocket Client for Verifier Portal Real-time Updates
 */

class VerifierWebSocketClient {
  constructor(url = 'ws://localhost:8080/ws') {
    this.url = url;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.listeners = new Map();
    this.subscriptions = new Set();
    this.connected = false;
  }

  connect() {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('[WS] Connected to verifier portal');
        this.connected = true;
        this.reconnectAttempts = 0;
        this.emit('connected');
        this.resubscribe();
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (err) {
          console.error('[WS] Failed to parse message:', err);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[WS] Error:', error);
        this.emit('error', error);
      };

      this.ws.onclose = () => {
        console.log('[WS] Connection closed');
        this.connected = false;
        this.emit('disconnected');
        this.attemptReconnect();
      };
    } catch (err) {
      console.error('[WS] Connection failed:', err);
      this.attemptReconnect();
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnection attempts reached');
      this.emit('reconnect_failed');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`[WS] Reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
    setTimeout(() => this.connect(), delay);
  }

  resubscribe() {
    for (const subscription of this.subscriptions) {
      this.send(subscription);
    }
  }

  handleMessage(message) {
    const { type, data } = message;

    switch (type) {
      case 'assistant_update':
        this.emit('assistant_update', data);
        break;
      case 'score_update':
        this.emit('score_update', data);
        break;
      case 'ir_completion':
        this.emit('ir_completion', data);
        break;
      case 'heartbeat_alert':
        this.emit('heartbeat_alert', data);
        break;
      case 'misbehavior_report':
        this.emit('misbehavior_report', data);
        break;
      case 'subscribed':
        console.log('[WS] Subscribed to:', data.channel);
        break;
      case 'unsubscribed':
        console.log('[WS] Unsubscribed from:', data.channel);
        break;
      case 'error':
        console.error('[WS] Server error:', data.error);
        this.emit('server_error', data);
        break;
      case 'pong':
        this.emit('pong');
        break;
      default:
        console.warn('[WS] Unknown message type:', type);
    }
  }

  subscribe(channel, params = {}) {
    const subscription = {
      type: 'subscribe',
      data: { channel, ...params }
    };

    this.subscriptions.add(JSON.stringify(subscription));
    this.send(subscription);
  }

  unsubscribe(channel, params = {}) {
    const subscription = {
      type: 'unsubscribe',
      data: { channel, ...params }
    };

    this.subscriptions.delete(JSON.stringify(subscription));
    this.send(subscription);
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn('[WS] Cannot send message - not connected');
    }
  }

  ping() {
    this.send({ type: 'ping', data: {} });
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  off(event, callback) {
    if (!this.listeners.has(event)) return;

    const callbacks = this.listeners.get(event);
    const index = callbacks.indexOf(callback);
    if (index > -1) {
      callbacks.splice(index, 1);
    }
  }

  emit(event, data = null) {
    if (!this.listeners.has(event)) return;

    for (const callback of this.listeners.get(event)) {
      try {
        callback(data);
      } catch (err) {
        console.error(`[WS] Error in ${event} listener:`, err);
      }
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected = false;
    this.subscriptions.clear();
  }

  isConnected() {
    return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN;
  }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = VerifierWebSocketClient;
}

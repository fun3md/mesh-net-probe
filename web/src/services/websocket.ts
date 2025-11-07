import { io, Socket } from 'socket.io-client';
import type { Probe, Alert } from '@/types';

// Extend the ImportMeta interface to include env
declare global {
  interface ImportMeta {
    env: Record<string, string>;
  }
}

export type WebSocketEventType =
  | 'probe_update'
  | 'measurement'
  | 'health_status'
  | 'config_change'
  | 'alert';

export type WebSocketEventHandler<T = any> = (data: T) => void;

export type WebSocketEvent =
  | 'probe_update'
  | 'measurement'
  | 'health_status'
  | 'config_change'
  | 'alert';

export interface WebSocketMessage {
  type: WebSocketEvent;
  data: any;
  timestamp?: string;
}

export interface WebSocketOptions {
  onMessage?: (msg: WebSocketMessage) => void;
  onError?: (err: Event | ErrorEvent) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

/**
 * Phase 5.2.7 - WebSocket Strategy (Experimental)
 *
 * - This client is intentionally optional.
 * - Controlled by VITE_WS_ENABLED / VITE_WEBSOCKET_ENABLED (default: false).
 * - When disabled, nothing connects and core flows rely on HTTP polling.
 * - When enabled, it attempts to connect to `${VITE_WS_BASE_URL || API_BASE_URL}/ws`.
 * - Frontend code MUST treat this as best-effort; no hard dependency.
 */

const WS_ENABLED: boolean =
  (import.meta as any).env?.VITE_WS_ENABLED === 'true' ||
  (import.meta as any).env?.VITE_WEBSOCKET_ENABLED === 'true';

const WS_PATH: string =
  (import.meta as any).env?.VITE_WS_PATH || '/ws';

const resolveWsUrl = (): string | null => {
  const httpBase: string =
    (import.meta as any).env?.VITE_WS_BASE_URL ||
    (import.meta as any).env?.VITE_API_BASE_URL ||
    'http://localhost:8080/api';

  try {
    const url = new URL(httpBase);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';

    if (WS_PATH.startsWith('ws://') || WS_PATH.startsWith('wss://')) {
      return WS_PATH;
    }

    if (url.pathname.endsWith('/api') && WS_PATH.startsWith('/')) {
      url.pathname = url.pathname.replace(/\/api$/, '') + WS_PATH;
    } else {
      url.pathname = WS_PATH;
    }

    return url.toString();
  } catch {
    return null;
  }
};

class WebSocketService {
  private socket: Socket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private eventHandlers: Map<WebSocketEventType, Set<WebSocketEventHandler>> = new Map();

  constructor() {
    // Initialize event handlers map
    this.eventHandlers.set('probe_update', new Set());
    this.eventHandlers.set('measurement', new Set());
    this.eventHandlers.set('health_status', new Set());
    this.eventHandlers.set('config_change', new Set());
    this.eventHandlers.set('alert', new Set());
  }

  connect(url?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const socketUrl = url || `${(import.meta as any).env?.VITE_WS_BASE_URL || 'ws://localhost:8080'}`;
        
        this.socket = io(socketUrl, {
          autoConnect: true,
          transports: ['websocket'],
          timeout: 5000,
          auth: {
            token: localStorage.getItem('auth_token')
          }
        });

        this.socket.on('connect', () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          resolve();
        });

        this.socket.on('disconnect', (reason) => {
          console.log('WebSocket disconnected:', reason);
          this.handleDisconnect();
        });

        this.socket.on('connect_error', (error) => {
          console.error('WebSocket connection error:', error);
          reject(error);
        });

        this.socket.on('probe_update', (data: Probe[]) => {
          this.emit('probe_update', data);
        });

        this.socket.on('measurement', (data: any) => {
          this.emit('measurement', data);
        });

        this.socket.on('health_status', (data: any) => {
          this.emit('health_status', data);
        });

        this.socket.on('config_change', (data: any) => {
          this.emit('config_change', data);
        });

        this.socket.on('alert', (data: Alert) => {
          this.emit('alert', data);
        });

      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
  }

  private handleDisconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
      setTimeout(() => {
        console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, delay);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  // Subscribe to event types
  on<T = any>(eventType: WebSocketEventType, handler: WebSocketEventHandler<T>): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      handlers.add(handler as WebSocketEventHandler);
    }
  }

  // Unsubscribe from event types
  off<T = any>(eventType: WebSocketEventType, handler: WebSocketEventHandler<T>): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      handlers.delete(handler as WebSocketEventHandler);
    }
  }

  // Emit events to handlers
  private emit<T = any>(eventType: WebSocketEventType, data: T): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${eventType}:`, error);
        }
      });
    }
  }

  // Send message to server (if needed)
  send(message: any): void {
    if (this.socket && this.socket.connected) {
      this.socket.emit('message', message);
    }
  }

  // Connection status
  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  // Get connection ID
  getId(): string | undefined {
    return this.socket?.id;
  }

  // Event listeners for specific endpoints
  subscribeToProbeUpdates(): void {
    if (this.socket) {
      this.socket.emit('subscribe', 'probe_updates');
    }
  }

  unsubscribeFromProbeUpdates(): void {
    if (this.socket) {
      this.socket.emit('unsubscribe', 'probe_updates');
    }
  }

  subscribeToMeasurementStream(probeId?: string): void {
    if (this.socket) {
      this.socket.emit('subscribe', 'measurements', { probeId });
    }
  }

  unsubscribeFromMeasurementStream(probeId?: string): void {
    if (this.socket) {
      this.socket.emit('unsubscribe', 'measurements', { probeId });
    }
  }

  subscribeToHealthUpdates(): void {
    if (this.socket) {
      this.socket.emit('subscribe', 'health_updates');
    }
  }

  unsubscribeFromHealthUpdates(): void {
    if (this.socket) {
      this.socket.emit('unsubscribe', 'health_updates');
    }
  }

  // Utility methods
  cleanup(): void {
    this.eventHandlers.forEach(handlers => {
      handlers.clear();
    });
  }
}

/**
 * Experimental WebSocket client, controlled by VITE_WS_ENABLED.
 * - When disabled, does not attempt to connect.
 * - Provides best-effort connectivity with HTTP fallback via API.
 * - Frontend code must treat this as best-effort; no hard dependency.
 */
class WebSocketExperimental {
  private socket: WebSocket | null = null;
  private options: WebSocketOptions | null = null;
  private reconnectTimer: number | null = null;
  private readonly reconnectDelay = 5000;
  private readonly maxReconnectAttempts = 5;
  private reconnectAttempts = 0;

  connect(options: WebSocketOptions = {}): void {
    this.options = options;

    if (!WS_ENABLED) {
      // WebSocket is experimental and disabled by default; rely on HTTP polling.
      console.info(
        '[WebSocketService] Disabled by configuration (VITE_WS_ENABLED=false). Using HTTP polling only.'
      );
      return;
    }

    const url = resolveWsUrl();
    if (!url) {
      console.warn(
        '[WebSocketService] Failed to resolve WS URL. Check VITE_WS_BASE_URL / VITE_WS_PATH.'
      );
      return;
    }

    if (this.socket) {
      if (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING) {
        console.info('[WebSocketService] WebSocket already connected or connecting.');
        return;
      }
    }

    try {
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        this.reconnectAttempts = 0;
        console.info('[WebSocketService] Connected to', url);
        this.options?.onOpen?.();
      };

      this.socket.onmessage = (event: MessageEvent) => {
        try {
          const raw = JSON.parse(event.data);
          const msg: WebSocketMessage = {
            type: raw.type,
            data: raw.data,
            timestamp: raw.timestamp || new Date().toISOString(),
          };
          this.options?.onMessage?.(msg);
        } catch (err) {
          console.warn('[WebSocketService] Failed to parse WS message', err);
        }
      };

      this.socket.onerror = (err: Event | ErrorEvent) => {
        console.warn('[WebSocketService] Error', err);
        this.options?.onError?.(err);
      };

      this.socket.onclose = () => {
        this.options?.onClose?.();
        this.socket = null;
        if (this.reconnectAttempts < this.maxReconnectAttempts && WS_ENABLED) {
          this.scheduleReconnect();
        } else {
          console.warn(
            '[WebSocketService] Max reconnect attempts reached or WS disabled. Staying disconnected.'
          );
        }
      };
    } catch (err) {
      console.warn('[WebSocketService] Failed to open WebSocket', err);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer != null) {
      return;
    }
    this.reconnectAttempts += 1;
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      console.info(
        `[WebSocketService] Reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`
      );
      this.connect(this.options || {});
    }, this.reconnectDelay);
  }

  disconnect(): void {
    if (this.reconnectTimer != null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  /**
   * Best-effort send; no-op if WS is disabled or not connected.
   */
  send(event: WebSocketEvent, data: any): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    const payload: WebSocketMessage = {
      type: event,
      data,
      timestamp: new Date().toISOString(),
    };
    this.socket.send(JSON.stringify(payload));
  }
}

export const websocketService = new WebSocketExperimental();
export const webSocketService = new WebSocketService();
export default webSocketService;
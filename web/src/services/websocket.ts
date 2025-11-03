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

export const webSocketService = new WebSocketService();
export default webSocketService;
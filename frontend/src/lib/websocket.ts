const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080';

export interface NewEmailPayload {
  emailID: string;
  from: string;
  subject: string;
  receivedAt: string;
}

export interface ErrorPayload {
  message: string;
  code: string;
}

export interface WebSocketMessage {
  type: 'new_email' | 'error';
  payload: NewEmailPayload | ErrorPayload;
}

export type MessageHandler = (message: WebSocketMessage) => void;
export type ErrorHandler = (error: Event) => void;
export type CloseHandler = (event: CloseEvent) => void;

export class MailboxWebSocket {
  private ws: WebSocket | null = null;
  private address: string;
  private messageHandlers: Set<MessageHandler> = new Set();
  private errorHandlers: Set<ErrorHandler> = new Set();
  private closeHandlers: Set<CloseHandler> = new Set();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isManualClose = false;

  constructor(address: string) {
    this.address = address;
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    this.isManualClose = false;
    const url = `${WS_URL}/api/ws/mailbox/${this.address}`;
    
    console.log(`Connecting to WebSocket: ${url}`);
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log('✅ WebSocket connected');
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        console.log('📩 WebSocket message received:', message);
        
        this.messageHandlers.forEach((handler) => {
          try {
            handler(message);
          } catch (error) {
            console.error('Error in message handler:', error);
          }
        });
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.ws.onerror = (event) => {
      console.error('❌ WebSocket error:', event);
      this.errorHandlers.forEach((handler) => {
        try {
          handler(event);
        } catch (error) {
          console.error('Error in error handler:', error);
        }
      });
    };

    this.ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      this.closeHandlers.forEach((handler) => {
        try {
          handler(event);
        } catch (error) {
          console.error('Error in close handler:', error);
        }
      });

      // Attempt to reconnect if not manually closed
      if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`Reconnecting... Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
        
        this.reconnectTimer = setTimeout(() => {
          this.connect();
        }, this.reconnectDelay);

        // Exponential backoff
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000); // Max 30 seconds
      } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        console.error('❌ Max reconnection attempts reached');
      }
    };
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.isManualClose = true;
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    console.log('WebSocket disconnected');
  }

  /**
   * Add message handler
   */
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  /**
   * Add error handler
   */
  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  /**
   * Add close handler
   */
  onClose(handler: CloseHandler): () => void {
    this.closeHandlers.add(handler);
    return () => this.closeHandlers.delete(handler);
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Get current connection state
   */
  getState(): number | null {
    return this.ws?.readyState ?? null;
  }
}

/**
 * Create and connect to mailbox WebSocket
 */
export function connectToMailbox(address: string): MailboxWebSocket {
  const ws = new MailboxWebSocket(address);
  ws.connect();
  return ws;
}


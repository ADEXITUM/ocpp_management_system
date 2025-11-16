/**
 * OCPP WebSocket Server
 * Accepts connections from OCPP 1.6J charge points
 */

import { WebSocketServer, WebSocket } from 'ws';
import { IncomingMessage } from 'http';
import { connectionManager } from './connection-manager';

export class OCPPServer {
  private wss: WebSocketServer | null = null;
  private port: number;

  constructor(port: number = 9000) {
    this.port = port;
  }

  /**
   * Start the OCPP WebSocket server
   */
  start(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.wss = new WebSocketServer({ port: this.port });

        this.wss.on('connection', (ws: WebSocket, request: IncomingMessage) => {
          this.handleConnection(ws, request);
        });

        this.wss.on('listening', () => {
          console.log(`🔌 OCPP Server listening on ws://0.0.0.0:${this.port}`);
          console.log(`   Waiting for charge points to connect...`);
          resolve();
        });

        this.wss.on('error', (error) => {
          console.error('WebSocket Server Error:', error);
          reject(error);
        });
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Handle new charge point connection
   */
  private handleConnection(ws: WebSocket, request: IncomingMessage): void {
    // Extract charge point ID from URL path
    // Expected format: /CP001 or /ocpp16/CP001
    const chargePointId = this.extractChargePointId(request.url || '');

    if (!chargePointId) {
      console.error('[OCPPServer] Connection rejected: No charge point ID in URL');
      ws.close(1008, 'Charge point ID required in URL path (e.g., /CP001)');
      return;
    }

    console.log(`[OCPPServer] New connection from ${chargePointId}`);
    console.log(`            Remote address: ${request.socket.remoteAddress}`);

    // Register the connection with the connection manager
    connectionManager.registerConnection(chargePointId, ws);
  }

  /**
   * Extract charge point ID from URL path
   */
  private extractChargePointId(url: string): string | null {
    // Handle formats like:
    // /CP001
    // /ocpp16/CP001
    // /CP001/
    const parts = url.split('/').filter(Boolean);

    if (parts.length === 0) {
      return null;
    }

    // Return the last non-empty part
    return parts[parts.length - 1] || null;
  }

  /**
   * Stop the server
   */
  stop(): Promise<void> {
    return new Promise((resolve) => {
      if (this.wss) {
        this.wss.close(() => {
          console.log('[OCPPServer] Server stopped');
          resolve();
        });
      } else {
        resolve();
      }
    });
  }

  /**
   * Get server status
   */
  getStatus() {
    return {
      port: this.port,
      running: this.wss !== null,
      connections: connectionManager.getStats()
    };
  }
}

/**
 * OCPP Connection Manager
 * Manages WebSocket connections to multiple charge points
 */

import { WebSocket } from 'ws';
import { v4 as uuidv4 } from 'uuid';
import { EventEmitter } from 'events';
import {
  OCPPMessage,
  OCPPCall,
  OCPPCallResult,
  OCPPCallError,
  MessageType,
  OCPPAction,
  RemoteStartTransactionRequest,
  RemoteStartTransactionResponse,
  RemoteStopTransactionRequest,
  RemoteStopTransactionResponse
} from '../types/ocpp-types';
import { messageHandlers } from './message-handlers';
import { database } from '../database/mock-database';

interface PendingRequest {
  resolve: (response: any) => void;
  reject: (error: Error) => void;
  timeout: NodeJS.Timeout;
}

export class ConnectionManager extends EventEmitter {
  private connections: Map<string, WebSocket> = new Map();
  private pendingRequests: Map<string, PendingRequest> = new Map();
  private readonly REQUEST_TIMEOUT = 30000; // 30 seconds

  /**
   * Register a new charge point connection
   */
  registerConnection(chargePointId: string, ws: WebSocket): void {
    // Close existing connection if any
    const existingWs = this.connections.get(chargePointId);
    if (existingWs) {
      console.log(`[ConnectionManager] Closing existing connection for ${chargePointId}`);
      existingWs.close();
    }

    this.connections.set(chargePointId, ws);
    database.updateChargePointStatus(chargePointId, 'online');

    console.log(`[ConnectionManager] Registered connection for ${chargePointId}`);
    console.log(`[ConnectionManager] Total connections: ${this.connections.size}`);

    // Set up WebSocket event handlers
    ws.on('message', (data: Buffer) => {
      this.handleMessage(chargePointId, data);
    });

    ws.on('close', () => {
      this.handleDisconnect(chargePointId);
    });

    ws.on('error', (error) => {
      console.error(`[ConnectionManager] WebSocket error for ${chargePointId}:`, error.message);
    });
  }

  /**
   * Handle incoming OCPP message
   */
  private handleMessage(chargePointId: string, data: Buffer): void {
    try {
      const message: OCPPMessage = JSON.parse(data.toString());
      const messageType = message[0];

      if (messageType === MessageType.CALL) {
        this.handleCall(chargePointId, message as OCPPCall);
      } else if (messageType === MessageType.CALLRESULT) {
        this.handleCallResult(message as OCPPCallResult);
      } else if (messageType === MessageType.CALLERROR) {
        this.handleCallError(message as OCPPCallError);
      }
    } catch (error) {
      console.error(
        `[ConnectionManager] Failed to parse message from ${chargePointId}:`,
        error instanceof Error ? error.message : error
      );
    }
  }

  /**
   * Handle CALL (request from charge point)
   */
  private async handleCall(chargePointId: string, message: OCPPCall): Promise<void> {
    const [, messageId, action, payload] = message;

    console.log(`[ConnectionManager] <- ${chargePointId}: ${action}`);

    let response: any;
    let error: { code: string; description: string } | null = null;

    try {
      // Route to appropriate handler
      switch (action) {
        case OCPPAction.BootNotification:
          response = messageHandlers.handleBootNotification(chargePointId, payload);
          break;
        case OCPPAction.Heartbeat:
          response = messageHandlers.handleHeartbeat(chargePointId);
          break;
        case OCPPAction.StatusNotification:
          response = messageHandlers.handleStatusNotification(chargePointId, payload);
          break;
        case OCPPAction.MeterValues:
          response = messageHandlers.handleMeterValues(chargePointId, payload);
          break;
        case OCPPAction.StartTransaction:
          response = messageHandlers.handleStartTransaction(chargePointId, payload);
          break;
        case OCPPAction.StopTransaction:
          response = messageHandlers.handleStopTransaction(chargePointId, payload);
          break;
        case OCPPAction.Authorize:
          response = messageHandlers.handleAuthorize(chargePointId, payload);
          break;
        default:
          error = {
            code: 'NotImplemented',
            description: `Action ${action} is not implemented`
          };
      }
    } catch (err) {
      console.error(`[ConnectionManager] Error handling ${action}:`, err);
      error = {
        code: 'InternalError',
        description: err instanceof Error ? err.message : 'Unknown error'
      };
    }

    // Send response
    if (error) {
      this.sendCallError(chargePointId, messageId, error.code, error.description);
    } else {
      this.sendCallResult(chargePointId, messageId, response);
    }
  }

  /**
   * Handle CALLRESULT (response to our request)
   */
  private handleCallResult(message: OCPPCallResult): void {
    const [, messageId, payload] = message;

    const pending = this.pendingRequests.get(messageId);
    if (pending) {
      clearTimeout(pending.timeout);
      pending.resolve(payload);
      this.pendingRequests.delete(messageId);
    }
  }

  /**
   * Handle CALLERROR (error response to our request)
   */
  private handleCallError(message: OCPPCallError): void {
    const [, messageId, errorCode, errorDescription] = message;

    const pending = this.pendingRequests.get(messageId);
    if (pending) {
      clearTimeout(pending.timeout);
      pending.reject(new Error(`${errorCode}: ${errorDescription}`));
      this.pendingRequests.delete(messageId);
    }
  }

  /**
   * Handle charge point disconnect
   */
  private handleDisconnect(chargePointId: string): void {
    this.connections.delete(chargePointId);
    database.updateChargePointStatus(chargePointId, 'offline');

    console.log(`[ConnectionManager] ${chargePointId} disconnected`);
    console.log(`[ConnectionManager] Total connections: ${this.connections.size}`);
  }

  /**
   * Send CALLRESULT to charge point
   */
  private sendCallResult(chargePointId: string, messageId: string, payload: any): void {
    const message: OCPPCallResult = [MessageType.CALLRESULT, messageId, payload];
    this.sendMessage(chargePointId, message);
  }

  /**
   * Send CALLERROR to charge point
   */
  private sendCallError(
    chargePointId: string,
    messageId: string,
    errorCode: string,
    errorDescription: string
  ): void {
    const message: OCPPCallError = [
      MessageType.CALLERROR,
      messageId,
      errorCode,
      errorDescription,
      {}
    ];
    this.sendMessage(chargePointId, message);
  }

  /**
   * Send CALL (request) to charge point
   */
  sendCall<T = any>(
    chargePointId: string,
    action: string,
    payload: any,
    timeout: number = this.REQUEST_TIMEOUT
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const messageId = uuidv4();
      const message: OCPPCall = [MessageType.CALL, messageId, action, payload];

      // Set up timeout
      const timeoutHandle = setTimeout(() => {
        this.pendingRequests.delete(messageId);
        reject(new Error(`Request timeout: ${action} to ${chargePointId}`));
      }, timeout);

      // Store pending request
      this.pendingRequests.set(messageId, {
        resolve,
        reject,
        timeout: timeoutHandle
      });

      // Send message
      console.log(`[ConnectionManager] -> ${chargePointId}: ${action}`);
      this.sendMessage(chargePointId, message);
    });
  }

  /**
   * Send raw message to charge point
   */
  private sendMessage(chargePointId: string, message: OCPPMessage): void {
    const ws = this.connections.get(chargePointId);
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.error(`[ConnectionManager] Cannot send to ${chargePointId}: not connected`);
      return;
    }

    ws.send(JSON.stringify(message));
  }

  /**
   * Check if charge point is connected
   */
  isConnected(chargePointId: string): boolean {
    const ws = this.connections.get(chargePointId);
    return ws !== undefined && ws.readyState === WebSocket.OPEN;
  }

  /**
   * Get all connected charge point IDs
   */
  getConnectedChargePoints(): string[] {
    return Array.from(this.connections.keys());
  }

  /**
   * Send RemoteStartTransaction
   */
  async remoteStartTransaction(
    chargePointId: string,
    idTag: string,
    connectorId?: number
  ): Promise<RemoteStartTransactionResponse> {
    const request: RemoteStartTransactionRequest = {
      idTag,
      connectorId
    };

    return this.sendCall<RemoteStartTransactionResponse>(
      chargePointId,
      OCPPAction.RemoteStartTransaction,
      request
    );
  }

  /**
   * Send RemoteStopTransaction
   */
  async remoteStopTransaction(
    chargePointId: string,
    transactionId: number
  ): Promise<RemoteStopTransactionResponse> {
    const request: RemoteStopTransactionRequest = {
      transactionId
    };

    return this.sendCall<RemoteStopTransactionResponse>(
      chargePointId,
      OCPPAction.RemoteStopTransaction,
      request
    );
  }

  /**
   * Get connection statistics
   */
  getStats() {
    return {
      totalConnections: this.connections.size,
      pendingRequests: this.pendingRequests.size,
      connectedChargePoints: this.getConnectedChargePoints()
    };
  }
}

export const connectionManager = new ConnectionManager();

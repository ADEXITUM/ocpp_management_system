/**
 * Charging Service - Simple API for External Services
 *
 * This is the main interface for payment services and other parts
 * of the application to control charging sessions without knowing
 * OCPP protocol details.
 */

import { database } from '../database/mock-database';
import { connectionManager } from '../ocpp/connection-manager';
import {
  ChargePointNotFoundError,
  ChargePointOfflineError,
  ConnectorNotFoundError,
  ConnectorUnavailableError,
  SessionNotFoundError,
  SessionAlreadyActiveError,
  RemoteOperationFailedError,
  OperationTimeoutError,
  ChargePointNotConfiguredError
} from '../errors/custom-errors';
import {
  TurnOnResult,
  TurnOffResult,
  EnergyConsumptionResult
} from '../types/domain-types';
import { RemoteStartStopStatus } from '../types/ocpp-types';

export class ChargingService {
  /**
   * Turn on a charging socket (start a charging session)
   *
   * @param chargePointId - ID of the charge point
   * @param connectorId - Connector number (1, 2, etc.)
   * @param userId - User/customer ID who is starting the session
   * @returns Promise with transaction details
   * @throws ChargePointNotFoundError - If charge point doesn't exist
   * @throws ChargePointOfflineError - If charge point is not connected
   * @throws ConnectorNotFoundError - If connector doesn't exist
   * @throws ConnectorUnavailableError - If connector is busy/faulted
   * @throws SessionAlreadyActiveError - If connector already has active session
   * @throws RemoteOperationFailedError - If charge point rejects the command
   * @throws OperationTimeoutError - If operation times out
   */
  async turnOn(
    chargePointId: string,
    connectorId: number,
    userId: string
  ): Promise<TurnOnResult> {
    // 1. Check if charge point exists in database
    const chargePoint = database.getChargePoint(chargePointId);
    if (!chargePoint) {
      throw new ChargePointNotFoundError(chargePointId);
    }

    // 2. Check if charge point is configured (not pending)
    if (chargePoint.registrationStatus !== 'accepted') {
      throw new ChargePointNotConfiguredError(chargePointId);
    }

    // 3. Check if charge point is connected
    if (!connectionManager.isConnected(chargePointId)) {
      throw new ChargePointOfflineError(chargePointId);
    }

    // 4. Check if connector exists
    const connector = database.getConnector(chargePointId, connectorId);
    if (!connector) {
      throw new ConnectorNotFoundError(chargePointId, connectorId);
    }

    // 5. Check connector status
    if (connector.status === 'Faulted') {
      throw new ConnectorUnavailableError(chargePointId, connectorId, 'Faulted');
    }

    if (connector.status === 'Unavailable') {
      throw new ConnectorUnavailableError(chargePointId, connectorId, 'Unavailable');
    }

    // 6. Check if connector already has an active session
    const existingSession = database.getActiveSessionByConnector(chargePointId, connectorId);
    if (existingSession) {
      throw new SessionAlreadyActiveError(chargePointId, connectorId);
    }

    // 7. Send RemoteStartTransaction to charge point
    try {
      const response = await connectionManager.remoteStartTransaction(
        chargePointId,
        userId,
        connectorId
      );

      if (response.status === RemoteStartStopStatus.Rejected) {
        throw new RemoteOperationFailedError('start', chargePointId, 'Rejected by charge point');
      }

      // 8. Wait a bit for the StartTransaction message from the charge point
      await this.waitForSessionStart(chargePointId, connectorId, 10000);

      // 9. Get the created transaction
      const session = database.getActiveSessionByConnector(chargePointId, connectorId);
      if (!session) {
        throw new RemoteOperationFailedError(
          'start',
          chargePointId,
          'Session not created after remote start'
        );
      }

      return {
        success: true,
        transactionId: session.transactionId,
        chargePointId,
        connectorId,
        userId,
        message: 'Charging session started successfully'
      };
    } catch (error) {
      if (error instanceof Error && error.message.includes('timeout')) {
        throw new OperationTimeoutError('RemoteStartTransaction', chargePointId);
      }
      throw error;
    }
  }

  /**
   * Turn off a charging socket (stop a charging session)
   *
   * @param chargePointId - ID of the charge point
   * @param transactionId - Transaction ID to stop
   * @returns Promise with energy consumption details
   * @throws ChargePointNotFoundError - If charge point doesn't exist
   * @throws ChargePointOfflineError - If charge point is not connected
   * @throws SessionNotFoundError - If transaction doesn't exist
   * @throws RemoteOperationFailedError - If charge point rejects the command
   * @throws OperationTimeoutError - If operation times out
   */
  async turnOff(chargePointId: string, transactionId: number): Promise<TurnOffResult> {
    // 1. Check if session exists
    const session = database.getSession(transactionId);
    if (!session) {
      throw new SessionNotFoundError(transactionId);
    }

    // Verify it's the right charge point
    if (session.chargePointId !== chargePointId) {
      throw new SessionNotFoundError(transactionId);
    }

    // 2. Check if session is still active
    if (session.status !== 'active') {
      // Session already completed, return the final data
      const energyConsumed = (session.endMeterValue || session.currentMeterValue) - session.startMeterValue;
      const duration = session.endTime
        ? (session.endTime.getTime() - session.startTime.getTime()) / 1000
        : 0;

      return {
        success: true,
        transactionId,
        chargePointId,
        energyConsumed,
        duration,
        message: 'Session was already stopped'
      };
    }

    // 3. Check if charge point is connected
    if (!connectionManager.isConnected(chargePointId)) {
      throw new ChargePointOfflineError(chargePointId);
    }

    // 4. Send RemoteStopTransaction to charge point
    try {
      const response = await connectionManager.remoteStopTransaction(chargePointId, transactionId);

      if (response.status === RemoteStartStopStatus.Rejected) {
        throw new RemoteOperationFailedError('stop', chargePointId, 'Rejected by charge point');
      }

      // 5. Wait for the StopTransaction message from the charge point
      await this.waitForSessionStop(transactionId, 15000);

      // 6. Get final session data
      const updatedSession = database.getSession(transactionId);
      if (!updatedSession) {
        throw new SessionNotFoundError(transactionId);
      }

      const energyConsumed =
        (updatedSession.endMeterValue || updatedSession.currentMeterValue) -
        updatedSession.startMeterValue;

      const duration = updatedSession.endTime
        ? (updatedSession.endTime.getTime() - updatedSession.startTime.getTime()) / 1000
        : (new Date().getTime() - updatedSession.startTime.getTime()) / 1000;

      return {
        success: true,
        transactionId,
        chargePointId,
        energyConsumed,
        duration,
        message: 'Charging session stopped successfully'
      };
    } catch (error) {
      if (error instanceof Error && error.message.includes('timeout')) {
        throw new OperationTimeoutError('RemoteStopTransaction', chargePointId);
      }
      throw error;
    }
  }

  /**
   * Get current energy consumption for an active charging session
   *
   * @param chargePointId - ID of the charge point
   * @param transactionId - Transaction ID
   * @returns Promise with energy consumption data
   * @throws SessionNotFoundError - If transaction doesn't exist
   */
  async getEnergyConsumption(
    chargePointId: string,
    transactionId: number
  ): Promise<EnergyConsumptionResult> {
    // 1. Get session
    const session = database.getSession(transactionId);
    if (!session) {
      throw new SessionNotFoundError(transactionId);
    }

    // Verify it's the right charge point
    if (session.chargePointId !== chargePointId) {
      throw new SessionNotFoundError(transactionId);
    }

    // 2. Calculate current energy consumption
    const currentMeterValue =
      session.status === 'completed' && session.endMeterValue
        ? session.endMeterValue
        : session.currentMeterValue;

    const energyConsumed = currentMeterValue - session.startMeterValue;

    const endTime = session.status === 'completed' && session.endTime ? session.endTime : new Date();
    const durationSeconds = Math.floor((endTime.getTime() - session.startTime.getTime()) / 1000);

    // 3. Get latest energy reading for additional details
    const latestReading = database.getLatestEnergyReading(transactionId);

    return {
      transactionId,
      chargePointId: session.chargePointId,
      connectorId: session.connectorId,
      userId: session.userId,
      startTime: session.startTime,
      currentEnergyWh: energyConsumed,
      durationSeconds,
      status: session.status === 'completed' ? 'completed' : 'active',
      lastUpdate: latestReading?.timestamp || session.startTime
    };
  }

  /**
   * Helper: Wait for session to start
   */
  private async waitForSessionStart(
    chargePointId: string,
    connectorId: number,
    timeoutMs: number
  ): Promise<void> {
    const startTime = Date.now();
    const pollInterval = 500; // Check every 500ms

    while (Date.now() - startTime < timeoutMs) {
      const session = database.getActiveSessionByConnector(chargePointId, connectorId);
      if (session) {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }

    // Timeout
    throw new Error('Timeout waiting for session to start');
  }

  /**
   * Helper: Wait for session to stop
   */
  private async waitForSessionStop(transactionId: number, timeoutMs: number): Promise<void> {
    const startTime = Date.now();
    const pollInterval = 500; // Check every 500ms

    while (Date.now() - startTime < timeoutMs) {
      const session = database.getSession(transactionId);
      if (session && session.status === 'completed') {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }

    // Timeout - session might still be stopping
    // We'll just return what we have
  }

  /**
   * Get all active sessions
   */
  async getActiveSessions(): Promise<EnergyConsumptionResult[]> {
    const sessions = database.getAllActiveSessions();
    return sessions.map((session) => {
      const energyConsumed = session.currentMeterValue - session.startMeterValue;
      const durationSeconds = Math.floor((Date.now() - session.startTime.getTime()) / 1000);
      const latestReading = database.getLatestEnergyReading(session.transactionId);

      return {
        transactionId: session.transactionId,
        chargePointId: session.chargePointId,
        connectorId: session.connectorId,
        userId: session.userId,
        startTime: session.startTime,
        currentEnergyWh: energyConsumed,
        durationSeconds,
        status: 'active',
        lastUpdate: latestReading?.timestamp || session.startTime
      };
    });
  }
}

export const chargingService = new ChargingService();

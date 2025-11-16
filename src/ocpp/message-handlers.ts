/**
 * OCPP Message Handlers
 * Process incoming OCPP messages from charge points
 */

import { database } from '../database/mock-database';
import {
  OCPPAction,
  BootNotificationRequest,
  BootNotificationResponse,
  HeartbeatResponse,
  StatusNotificationRequest,
  StatusNotificationResponse,
  MeterValuesRequest,
  MeterValuesResponse,
  StartTransactionRequest,
  StartTransactionResponse,
  StopTransactionRequest,
  StopTransactionResponse,
  AuthorizeRequest,
  AuthorizeResponse,
  RegistrationStatus,
  AuthorizationStatus,
  ChargePointStatus
} from '../types/ocpp-types';
import { ChargePoint, ConnectorStatus, EnergyReading } from '../types/domain-types';

export class MessageHandlers {
  /**
   * Handle BootNotification - Charge point is connecting
   */
  handleBootNotification(
    chargePointId: string,
    request: BootNotificationRequest
  ): BootNotificationResponse {
    console.log(`[BootNotification] Charge point ${chargePointId} is booting...`);

    const existingCP = database.getChargePoint(chargePointId);

    if (!existingCP) {
      // Charge point not in database - print warning
      console.warn(`⚠️  WARNING: Charge point '${chargePointId}' is NOT CONFIGURED!`);
      console.warn(`   Please configure this charge point in the database before use.`);
      console.warn(`   Vendor: ${request.chargePointVendor}, Model: ${request.chargePointModel}`);

      // We'll still accept it temporarily but mark it
      const newCP: ChargePoint = {
        id: chargePointId,
        name: `Unconfigured - ${chargePointId}`,
        vendor: request.chargePointVendor,
        model: request.chargePointModel,
        serialNumber: request.chargePointSerialNumber,
        firmwareVersion: request.firmwareVersion,
        numberOfConnectors: 1, // Default, will be updated via StatusNotification
        status: 'online',
        registrationStatus: 'pending', // Mark as pending until configured
        lastSeen: new Date(),
        createdAt: new Date()
      };
      database.upsertChargePoint(newCP);
    } else {
      // Update existing charge point
      existingCP.vendor = request.chargePointVendor;
      existingCP.model = request.chargePointModel;
      existingCP.serialNumber = request.chargePointSerialNumber;
      existingCP.firmwareVersion = request.firmwareVersion;
      existingCP.status = 'online';
      existingCP.lastSeen = new Date();
      database.upsertChargePoint(existingCP);

      console.log(`[BootNotification] ${chargePointId} registered successfully`);
    }

    return {
      status: existingCP?.registrationStatus === 'accepted'
        ? RegistrationStatus.Accepted
        : RegistrationStatus.Pending,
      currentTime: new Date().toISOString(),
      interval: 300 // Heartbeat every 5 minutes
    };
  }

  /**
   * Handle Heartbeat - Keep connection alive
   */
  handleHeartbeat(chargePointId: string): HeartbeatResponse {
    database.updateChargePointStatus(chargePointId, 'online');
    return {
      currentTime: new Date().toISOString()
    };
  }

  /**
   * Handle StatusNotification - Connector status update
   */
  handleStatusNotification(
    chargePointId: string,
    request: StatusNotificationRequest
  ): StatusNotificationResponse {
    const { connectorId, status, errorCode } = request;

    console.log(
      `[StatusNotification] ${chargePointId} Connector ${connectorId}: ${status} (error: ${errorCode})`
    );

    // Update connector status in database
    database.updateConnectorStatus(chargePointId, connectorId, status as ConnectorStatus);

    // If this is the first time we're seeing this connector, update charge point connector count
    const cp = database.getChargePoint(chargePointId);
    if (cp) {
      const connectors = database.getConnectorsByChargePoint(chargePointId);
      const maxConnectorId = Math.max(...connectors.map((c) => c.connectorId), 0);
      if (maxConnectorId > cp.numberOfConnectors) {
        cp.numberOfConnectors = maxConnectorId;
        database.upsertChargePoint(cp);
      }
    }

    return {};
  }

  /**
   * Handle MeterValues - Energy consumption updates
   */
  handleMeterValues(
    chargePointId: string,
    request: MeterValuesRequest
  ): MeterValuesResponse {
    const { connectorId, transactionId, meterValue } = request;

    if (!transactionId) {
      // No transaction, just periodic meter values
      return {};
    }

    // Process each meter value
    for (const mv of meterValue) {
      const timestamp = new Date(mv.timestamp);

      // Extract energy and power readings
      let energyWh: number | undefined;
      let powerW: number | undefined;
      let currentA: number | undefined;
      let voltageV: number | undefined;

      for (const sample of mv.sampledValue) {
        const value = parseFloat(sample.value);

        if (sample.measurand === 'Energy.Active.Import.Register') {
          energyWh = sample.unit === 'kWh' ? value * 1000 : value;
        } else if (sample.measurand === 'Power.Active.Import') {
          powerW = sample.unit === 'kW' ? value * 1000 : value;
        } else if (sample.measurand === 'Current.Import') {
          currentA = value;
        } else if (sample.measurand === 'Voltage') {
          voltageV = value;
        }
      }

      // Store energy reading if we have energy data
      if (energyWh !== undefined) {
        const reading: EnergyReading = {
          timestamp,
          transactionId,
          energyWh,
          powerW,
          currentA,
          voltageV
        };
        database.addEnergyReading(reading);

        console.log(
          `[MeterValues] ${chargePointId} Transaction ${transactionId}: ${energyWh} Wh` +
          (powerW ? `, ${powerW} W` : '')
        );
      }
    }

    return {};
  }

  /**
   * Handle StartTransaction - Charging session started
   */
  handleStartTransaction(
    chargePointId: string,
    request: StartTransactionRequest
  ): StartTransactionResponse {
    const { connectorId, idTag, meterStart, timestamp } = request;

    console.log(
      `[StartTransaction] ${chargePointId} Connector ${connectorId}: User ${idTag}, Meter: ${meterStart} Wh`
    );

    // Check if there's already an active session on this connector
    const existingSession = database.getActiveSessionByConnector(chargePointId, connectorId);
    if (existingSession) {
      console.warn(
        `[StartTransaction] WARNING: Connector already has active session ${existingSession.transactionId}`
      );
      // Complete the old session
      database.completeSession(existingSession.transactionId, meterStart, 'Replaced');
    }

    // Create new session
    const transactionId = database.createSession({
      chargePointId,
      connectorId,
      userId: idTag,
      startTime: new Date(timestamp),
      startMeterValue: meterStart,
      currentMeterValue: meterStart,
      status: 'active'
    });

    // Update connector
    database.setConnectorTransaction(chargePointId, connectorId, transactionId);
    database.updateConnectorStatus(chargePointId, connectorId, 'Charging');

    console.log(`[StartTransaction] Created transaction ${transactionId}`);

    return {
      transactionId,
      idTagInfo: {
        status: AuthorizationStatus.Accepted
      }
    };
  }

  /**
   * Handle StopTransaction - Charging session ended
   */
  handleStopTransaction(
    chargePointId: string,
    request: StopTransactionRequest
  ): StopTransactionResponse {
    const { transactionId, meterStop, timestamp, reason, transactionData } = request;

    console.log(
      `[StopTransaction] Transaction ${transactionId}: Meter: ${meterStop} Wh, Reason: ${reason || 'User'}`
    );

    const session = database.getSession(transactionId);
    if (session) {
      // Complete the session
      database.completeSession(transactionId, meterStop, reason);

      // Clear connector transaction
      database.setConnectorTransaction(session.chargePointId, session.connectorId, undefined);
      database.updateConnectorStatus(session.chargePointId, session.connectorId, 'Available');

      const energyConsumed = meterStop - session.startMeterValue;
      const duration = (new Date(timestamp).getTime() - session.startTime.getTime()) / 1000;

      console.log(
        `[StopTransaction] Session completed: ${energyConsumed} Wh consumed in ${Math.round(duration)}s`
      );

      // Process transaction data if present
      if (transactionData && transactionData.length > 0) {
        // Store final meter readings
        this.handleMeterValues(chargePointId, {
          connectorId: session.connectorId,
          transactionId,
          meterValue: transactionData
        });
      }
    } else {
      console.warn(`[StopTransaction] WARNING: Transaction ${transactionId} not found`);
    }

    return {
      idTagInfo: {
        status: AuthorizationStatus.Accepted
      }
    };
  }

  /**
   * Handle Authorize - Validate user credentials
   */
  handleAuthorize(chargePointId: string, request: AuthorizeRequest): AuthorizeResponse {
    const { idTag } = request;

    console.log(`[Authorize] ${chargePointId}: Checking authorization for ${idTag}`);

    // In a real system, you'd check against a user database
    // For now, we'll accept all users except those starting with "BLOCKED"
    const isBlocked = idTag.startsWith('BLOCKED');

    return {
      idTagInfo: {
        status: isBlocked ? AuthorizationStatus.Blocked : AuthorizationStatus.Accepted
      }
    };
  }
}

export const messageHandlers = new MessageHandlers();

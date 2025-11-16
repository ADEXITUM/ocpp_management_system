/**
 * Mock In-Memory Database for OCPP Management System
 * In production, this would be replaced with PostgreSQL, MongoDB, etc.
 */

import { ChargePoint, Connector, ChargingSession, EnergyReading } from '../types/domain-types';

export class MockDatabase {
  private chargePoints: Map<string, ChargePoint> = new Map();
  private connectors: Map<string, Connector> = new Map(); // key: "cpId:connectorId"
  private sessions: Map<number, ChargingSession> = new Map();
  private energyReadings: Map<number, EnergyReading[]> = new Map(); // key: transactionId
  private transactionIdCounter = 1;

  constructor() {
    this.initializeMockData();
  }

  /**
   * Initialize with some sample charge points
   */
  private initializeMockData(): void {
    // Sample Charge Point 1
    const cp1: ChargePoint = {
      id: 'CP001',
      name: 'Main Street Station 1',
      vendor: 'EVBox',
      model: 'Elvi',
      serialNumber: 'EVB-001-2024',
      firmwareVersion: '1.0.0',
      numberOfConnectors: 2,
      status: 'offline',
      registrationStatus: 'accepted',
      lastSeen: new Date(),
      createdAt: new Date()
    };
    this.chargePoints.set(cp1.id, cp1);

    // Create connectors for CP1
    for (let i = 1; i <= cp1.numberOfConnectors; i++) {
      const connector: Connector = {
        chargePointId: cp1.id,
        connectorId: i,
        status: 'Available',
        lastStatusUpdate: new Date()
      };
      this.connectors.set(`${cp1.id}:${i}`, connector);
    }

    // Sample Charge Point 2
    const cp2: ChargePoint = {
      id: 'CP002',
      name: 'Shopping Mall Station',
      vendor: 'ABB',
      model: 'Terra AC',
      serialNumber: 'ABB-002-2024',
      firmwareVersion: '2.1.3',
      numberOfConnectors: 1,
      status: 'offline',
      registrationStatus: 'accepted',
      lastSeen: new Date(),
      createdAt: new Date()
    };
    this.chargePoints.set(cp2.id, cp2);

    const connector2: Connector = {
      chargePointId: cp2.id,
      connectorId: 1,
      status: 'Available',
      lastStatusUpdate: new Date()
    };
    this.connectors.set(`${cp2.id}:1`, connector2);

    // Sample Charge Point 3
    const cp3: ChargePoint = {
      id: 'CP003',
      name: 'Office Parking Charger',
      vendor: 'ChargePoint',
      model: 'CPE250',
      numberOfConnectors: 2,
      status: 'offline',
      registrationStatus: 'accepted',
      lastSeen: new Date(),
      createdAt: new Date()
    };
    this.chargePoints.set(cp3.id, cp3);

    for (let i = 1; i <= cp3.numberOfConnectors; i++) {
      const connector: Connector = {
        chargePointId: cp3.id,
        connectorId: i,
        status: 'Available',
        lastStatusUpdate: new Date()
      };
      this.connectors.set(`${cp3.id}:${i}`, connector);
    }
  }

  // Charge Point Methods

  getChargePoint(id: string): ChargePoint | undefined {
    return this.chargePoints.get(id);
  }

  getAllChargePoints(): ChargePoint[] {
    return Array.from(this.chargePoints.values());
  }

  upsertChargePoint(chargePoint: ChargePoint): void {
    this.chargePoints.set(chargePoint.id, chargePoint);
  }

  updateChargePointStatus(id: string, status: 'online' | 'offline'): void {
    const cp = this.chargePoints.get(id);
    if (cp) {
      cp.status = status;
      cp.lastSeen = new Date();
    }
  }

  // Connector Methods

  getConnector(chargePointId: string, connectorId: number): Connector | undefined {
    return this.connectors.get(`${chargePointId}:${connectorId}`);
  }

  getConnectorsByChargePoint(chargePointId: string): Connector[] {
    const connectors: Connector[] = [];
    this.connectors.forEach((connector) => {
      if (connector.chargePointId === chargePointId) {
        connectors.push(connector);
      }
    });
    return connectors;
  }

  updateConnectorStatus(
    chargePointId: string,
    connectorId: number,
    status: Connector['status']
  ): void {
    const key = `${chargePointId}:${connectorId}`;
    const connector = this.connectors.get(key);
    if (connector) {
      connector.status = status;
      connector.lastStatusUpdate = new Date();
    } else {
      // Create connector if it doesn't exist
      const newConnector: Connector = {
        chargePointId,
        connectorId,
        status,
        lastStatusUpdate: new Date()
      };
      this.connectors.set(key, newConnector);
    }
  }

  setConnectorTransaction(
    chargePointId: string,
    connectorId: number,
    transactionId: number | undefined
  ): void {
    const connector = this.connectors.get(`${chargePointId}:${connectorId}`);
    if (connector) {
      connector.currentTransaction = transactionId;
    }
  }

  // Session Methods

  createSession(session: Omit<ChargingSession, 'transactionId'>): number {
    const transactionId = this.transactionIdCounter++;
    const fullSession: ChargingSession = {
      ...session,
      transactionId
    };
    this.sessions.set(transactionId, fullSession);
    this.energyReadings.set(transactionId, []);
    return transactionId;
  }

  getSession(transactionId: number): ChargingSession | undefined {
    return this.sessions.get(transactionId);
  }

  getActiveSessionByConnector(
    chargePointId: string,
    connectorId: number
  ): ChargingSession | undefined {
    for (const session of this.sessions.values()) {
      if (
        session.chargePointId === chargePointId &&
        session.connectorId === connectorId &&
        session.status === 'active'
      ) {
        return session;
      }
    }
    return undefined;
  }

  getAllActiveSessions(): ChargingSession[] {
    return Array.from(this.sessions.values()).filter((s) => s.status === 'active');
  }

  updateSessionMeterValue(transactionId: number, meterValue: number): void {
    const session = this.sessions.get(transactionId);
    if (session) {
      session.currentMeterValue = meterValue;
    }
  }

  completeSession(
    transactionId: number,
    endMeterValue: number,
    stopReason?: string
  ): void {
    const session = this.sessions.get(transactionId);
    if (session) {
      session.status = 'completed';
      session.endTime = new Date();
      session.endMeterValue = endMeterValue;
      session.stopReason = stopReason;
    }
  }

  // Energy Reading Methods

  addEnergyReading(reading: EnergyReading): void {
    const readings = this.energyReadings.get(reading.transactionId) || [];
    readings.push(reading);
    this.energyReadings.set(reading.transactionId, readings);

    // Also update session's current meter value
    this.updateSessionMeterValue(reading.transactionId, reading.energyWh);
  }

  getEnergyReadings(transactionId: number): EnergyReading[] {
    return this.energyReadings.get(transactionId) || [];
  }

  getLatestEnergyReading(transactionId: number): EnergyReading | undefined {
    const readings = this.energyReadings.get(transactionId);
    if (!readings || readings.length === 0) {
      return undefined;
    }
    return readings[readings.length - 1];
  }

  // Utility Methods

  reset(): void {
    this.sessions.clear();
    this.energyReadings.clear();
    this.transactionIdCounter = 1;
    // Reset connector statuses
    this.connectors.forEach((connector) => {
      connector.status = 'Available';
      connector.currentTransaction = undefined;
    });
  }

  getStats() {
    return {
      totalChargePoints: this.chargePoints.size,
      onlineChargePoints: Array.from(this.chargePoints.values()).filter(
        (cp) => cp.status === 'online'
      ).length,
      totalConnectors: this.connectors.size,
      activeSessions: Array.from(this.sessions.values()).filter((s) => s.status === 'active')
        .length,
      totalSessions: this.sessions.size
    };
  }
}

// Singleton instance
export const database = new MockDatabase();

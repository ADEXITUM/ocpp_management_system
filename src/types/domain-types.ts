/**
 * Domain Types for the OCPP Management System
 */

export interface ChargePoint {
  id: string;
  name: string;
  vendor: string;
  model: string;
  serialNumber?: string;
  firmwareVersion?: string;
  numberOfConnectors: number;
  status: 'online' | 'offline';
  registrationStatus: 'accepted' | 'pending' | 'rejected';
  lastSeen: Date;
  createdAt: Date;
}

export interface Connector {
  chargePointId: string;
  connectorId: number;
  status: ConnectorStatus;
  currentTransaction?: number;
  lastStatusUpdate: Date;
}

export type ConnectorStatus =
  | 'Available'
  | 'Preparing'
  | 'Charging'
  | 'SuspendedEVSE'
  | 'SuspendedEV'
  | 'Finishing'
  | 'Reserved'
  | 'Unavailable'
  | 'Faulted';

export interface ChargingSession {
  transactionId: number;
  chargePointId: string;
  connectorId: number;
  userId: string;
  startTime: Date;
  startMeterValue: number;  // in Wh
  currentMeterValue: number; // in Wh
  endTime?: Date;
  endMeterValue?: number;    // in Wh
  status: 'active' | 'completed' | 'failed';
  stopReason?: string;
}

export interface EnergyReading {
  timestamp: Date;
  transactionId: number;
  energyWh: number;       // Total energy in Wh
  powerW?: number;        // Instantaneous power in W
  currentA?: number;      // Current in Amperes
  voltageV?: number;      // Voltage in Volts
}

export interface TurnOnResult {
  success: boolean;
  transactionId?: number;
  chargePointId: string;
  connectorId: number;
  userId: string;
  message: string;
}

export interface TurnOffResult {
  success: boolean;
  transactionId: number;
  chargePointId: string;
  energyConsumed: number;  // in Wh
  duration: number;        // in seconds
  message: string;
}

export interface EnergyConsumptionResult {
  transactionId: number;
  chargePointId: string;
  connectorId: number;
  userId: string;
  startTime: Date;
  currentEnergyWh: number;
  durationSeconds: number;
  status: 'active' | 'completed';
  lastUpdate: Date;
}

/**
 * OCPP 1.6J Message Types and Interfaces
 */

// OCPP Message Type IDs
export enum MessageType {
  CALL = 2,        // Request
  CALLRESULT = 3,  // Response
  CALLERROR = 4    // Error
}

// OCPP Message Format: [MessageTypeId, UniqueId, Action, Payload]
export type OCPPCall = [MessageType.CALL, string, string, any];
export type OCPPCallResult = [MessageType.CALLRESULT, string, any];
export type OCPPCallError = [MessageType.CALLERROR, string, string, string, any];
export type OCPPMessage = OCPPCall | OCPPCallResult | OCPPCallError;

// Charge Point Status
export enum ChargePointStatus {
  Available = 'Available',
  Preparing = 'Preparing',
  Charging = 'Charging',
  SuspendedEVSE = 'SuspendedEVSE',
  SuspendedEV = 'SuspendedEV',
  Finishing = 'Finishing',
  Reserved = 'Reserved',
  Unavailable = 'Unavailable',
  Faulted = 'Faulted'
}

// Registration Status
export enum RegistrationStatus {
  Accepted = 'Accepted',
  Pending = 'Pending',
  Rejected = 'Rejected'
}

// Authorization Status
export enum AuthorizationStatus {
  Accepted = 'Accepted',
  Blocked = 'Blocked',
  Expired = 'Expired',
  Invalid = 'Invalid',
  ConcurrentTx = 'ConcurrentTx'
}

// Remote Start/Stop Status
export enum RemoteStartStopStatus {
  Accepted = 'Accepted',
  Rejected = 'Rejected'
}

// OCPP Actions
export enum OCPPAction {
  // CP to CS
  BootNotification = 'BootNotification',
  Heartbeat = 'Heartbeat',
  StatusNotification = 'StatusNotification',
  MeterValues = 'MeterValues',
  StartTransaction = 'StartTransaction',
  StopTransaction = 'StopTransaction',
  Authorize = 'Authorize',

  // CS to CP
  RemoteStartTransaction = 'RemoteStartTransaction',
  RemoteStopTransaction = 'RemoteStopTransaction',
  ChangeAvailability = 'ChangeAvailability',
  Reset = 'Reset',
  GetConfiguration = 'GetConfiguration',
  ChangeConfiguration = 'ChangeConfiguration'
}

// Message Payloads

export interface BootNotificationRequest {
  chargePointVendor: string;
  chargePointModel: string;
  chargePointSerialNumber?: string;
  chargeBoxSerialNumber?: string;
  firmwareVersion?: string;
  iccid?: string;
  imsi?: string;
  meterType?: string;
  meterSerialNumber?: string;
}

export interface BootNotificationResponse {
  status: RegistrationStatus;
  currentTime: string;
  interval: number;
}

export interface HeartbeatRequest {}

export interface HeartbeatResponse {
  currentTime: string;
}

export interface StatusNotificationRequest {
  connectorId: number;
  errorCode: string;
  status: ChargePointStatus;
  timestamp?: string;
  info?: string;
  vendorId?: string;
  vendorErrorCode?: string;
}

export interface StatusNotificationResponse {}

export interface MeterValue {
  timestamp: string;
  sampledValue: SampledValue[];
}

export interface SampledValue {
  value: string;
  context?: string;
  format?: string;
  measurand?: string;
  phase?: string;
  location?: string;
  unit?: string;
}

export interface MeterValuesRequest {
  connectorId: number;
  transactionId?: number;
  meterValue: MeterValue[];
}

export interface MeterValuesResponse {}

export interface StartTransactionRequest {
  connectorId: number;
  idTag: string;
  meterStart: number;
  timestamp: string;
  reservationId?: number;
}

export interface StartTransactionResponse {
  transactionId: number;
  idTagInfo: IdTagInfo;
}

export interface StopTransactionRequest {
  transactionId: number;
  idTag?: string;
  meterStop: number;
  timestamp: string;
  reason?: string;
  transactionData?: MeterValue[];
}

export interface StopTransactionResponse {
  idTagInfo?: IdTagInfo;
}

export interface IdTagInfo {
  status: AuthorizationStatus;
  expiryDate?: string;
  parentIdTag?: string;
}

export interface AuthorizeRequest {
  idTag: string;
}

export interface AuthorizeResponse {
  idTagInfo: IdTagInfo;
}

export interface RemoteStartTransactionRequest {
  idTag: string;
  connectorId?: number;
  chargingProfile?: any;
}

export interface RemoteStartTransactionResponse {
  status: RemoteStartStopStatus;
}

export interface RemoteStopTransactionRequest {
  transactionId: number;
}

export interface RemoteStopTransactionResponse {
  status: RemoteStartStopStatus;
}

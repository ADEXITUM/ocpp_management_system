# OCPP 1.6J Protocol Guide

## Overview

OCPP (Open Charge Point Protocol) 1.6J is a communication protocol between Electric Vehicle (EV) charging stations and Central Management Systems. The "J" stands for JSON - messages are formatted in JSON and transmitted over WebSocket connections.

## Key Concepts

### Architecture
```
[Charge Point] <---WebSocket---> [Central System/Backend]
```

- **Charge Point (CP)**: The physical charging station with one or more connectors
- **Central System (CS)**: The backend management system that controls charge points
- **Connector**: Individual charging outlet on a charge point (e.g., a station might have 2 connectors)

### Connection Flow
1. Charge Point initiates WebSocket connection to Central System
2. CP sends `BootNotification` to register itself
3. CS responds with Accept/Reject and heartbeat interval
4. CP periodically sends `Heartbeat` messages to maintain connection
5. Bidirectional communication for operations and status updates

## Message Structure

All OCPP 1.6J messages follow this JSON array format:

```json
[MessageTypeId, UniqueId, Action, Payload]
```

**Message Types:**
- `2` - CALL (request from CP or CS)
- `3` - CALLRESULT (successful response)
- `4` - CALLERROR (error response)

**Example - BootNotification Request:**
```json
[2, "unique-id-123", "BootNotification", {
  "chargePointVendor": "VendorName",
  "chargePointModel": "Model1"
}]
```

**Example - Response:**
```json
[3, "unique-id-123", {
  "status": "Accepted",
  "currentTime": "2025-11-16T10:30:00Z",
  "interval": 300
}]
```

## Core Operations

### 1. Boot & Connection

**BootNotification** - CP registers with CS
- Sent by: Charge Point
- Purpose: Inform CS about charge point details
- Response: Accepted/Pending/Rejected + heartbeat interval

**Heartbeat** - Keep connection alive
- Sent by: Charge Point (periodically)
- Purpose: Maintain connection and sync time
- Response: Current server time

### 2. Status & Monitoring

**StatusNotification** - Report connector status
- Sent by: Charge Point
- Purpose: Update connector status (Available, Preparing, Charging, Faulted, etc.)
- Statuses: Available, Preparing, Charging, SuspendedEVSE, SuspendedEV, Finishing, Reserved, Unavailable, Faulted

**MeterValues** - Energy consumption data
- Sent by: Charge Point
- Purpose: Send meter readings (energy, power, voltage, etc.)
- Contains: Timestamp, sampled values (Wh, W, V, A, etc.)

### 3. Transaction Management

**StartTransaction** - Begin charging session
- Sent by: Charge Point
- Contains: ConnectorId, IdTag (user ID), MeterStart, Timestamp
- Response: TransactionId (unique session identifier)

**StopTransaction** - End charging session
- Sent by: Charge Point
- Contains: TransactionId, MeterStop, Timestamp, Reason
- Response: IdTagInfo (authorization status)

### 4. Remote Control (From CS to CP)

**RemoteStartTransaction** - Start charging remotely
- Sent by: Central System
- Contains: IdTag, ConnectorId (optional)
- Response: Accepted/Rejected

**RemoteStopTransaction** - Stop charging remotely
- Sent by: Central System
- Contains: TransactionId
- Response: Accepted/Rejected

**ChangeAvailability** - Make connector available/unavailable
- Sent by: Central System
- Contains: ConnectorId, Type (Operative/Inoperative)
- Response: Accepted/Scheduled/Rejected

### 5. Authorization

**Authorize** - Validate user credentials
- Sent by: Charge Point
- Contains: IdTag (RFID card, user ID, etc.)
- Response: IdTagInfo with status (Accepted, Blocked, Expired, Invalid)

## Energy Metering

**MeterValues Message Example:**
```json
{
  "connectorId": 1,
  "transactionId": 12345,
  "meterValue": [{
    "timestamp": "2025-11-16T10:35:00Z",
    "sampledValue": [
      {
        "value": "1500.5",
        "context": "Sample.Periodic",
        "measurand": "Energy.Active.Import.Register",
        "unit": "Wh"
      },
      {
        "value": "7200",
        "measurand": "Power.Active.Import",
        "unit": "W"
      }
    ]
  }]
}
```

**Key Measurands:**
- `Energy.Active.Import.Register` - Total energy consumed (Wh or kWh)
- `Power.Active.Import` - Current power (W or kW)
- `Current.Import` - Current flow (A)
- `Voltage` - Voltage (V)
- `SoC` - State of Charge (%)

## Common Workflows

### Starting a Charging Session
1. User presents authorization (RFID card)
2. CP sends `Authorize` request
3. CS responds with Accepted/Rejected
4. If accepted, CP sends `StatusNotification` (Preparing)
5. CP sends `StartTransaction` with IdTag and initial meter reading
6. CS responds with TransactionId
7. CP sends `StatusNotification` (Charging)
8. CP periodically sends `MeterValues` during charging

### Stopping a Charging Session
1. User ends session (unplugs or swipes card)
2. CP sends `StatusNotification` (Finishing)
3. CP sends `StopTransaction` with final meter reading
4. CS responds with IdTagInfo
5. CP sends `StatusNotification` (Available)

### Remote Control from Backend
1. Backend calls `RemoteStartTransaction` with user IdTag
2. CP validates and responds Accepted
3. CP automatically starts charging (follows StartTransaction flow)
4. Backend can monitor via `MeterValues` messages
5. Backend calls `RemoteStopTransaction` when needed
6. CP stops charging and follows StopTransaction flow

## Configuration

Charge Points have configurable parameters accessible via:
- `GetConfiguration` - Retrieve settings
- `ChangeConfiguration` - Modify settings

**Common Configuration Keys:**
- `HeartbeatInterval` - Seconds between heartbeats (default: 300)
- `MeterValueSampleInterval` - Seconds between meter readings
- `ClockAlignedDataInterval` - Interval for clock-aligned meter values
- `AuthorizeRemoteTxRequests` - Require authorization for remote starts
- `NumberOfConnectors` - Total connectors on charge point

## Error Handling

**Common Error Codes:**
- `NotImplemented` - Action not supported
- `NotSupported` - Request not supported
- `InternalError` - Internal charge point error
- `ProtocolError` - Invalid message format
- `SecurityError` - Authentication/authorization failure
- `FormationViolation` - Message doesn't match schema
- `PropertyConstraintViolation` - Invalid property value
- `OccurenceConstraintViolation` - Required field missing
- `TypeConstraintViolation` - Wrong data type

## Summary

OCPP 1.6J enables:
- ✅ Real-time bidirectional communication over WebSocket
- ✅ Remote control of charging sessions
- ✅ Energy consumption monitoring and metering
- ✅ User authorization and authentication
- ✅ Status monitoring and diagnostics
- ✅ Configuration management

**Key Points for Implementation:**
1. Charge Point initiates WebSocket connection
2. Use unique message IDs to match requests/responses
3. TransactionId is critical for tracking charging sessions
4. MeterValues provide energy consumption data
5. Remote operations (RemoteStart/Stop) enable backend control
6. Always handle errors gracefully with appropriate error codes

# OCPP Management System Architecture

## Overview

This system manages multiple OCPP 1.6J charging stations, providing a simple API for controlling charging sessions and monitoring energy consumption.

## Architecture Layers

```
┌─────────────────────────────────────────────┐
│     Application Layer (Payment Service)     │
│  Uses: turnOn(), turnOff(), getEnergy()     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│         Service Wrapper (Simple API)        │
│   - Session Management                      │
│   - Error Translation                       │
│   - Business Logic                          │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│        OCPP Connection Manager              │
│   - Multiple CP Connections                 │
│   - Message Routing                         │
│   - State Management                        │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│          WebSocket Server (OCPP)            │
│   - Accept CP Connections                   │
│   - Protocol Handling                       │
│   - Message Parsing                         │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│            Mock Database                    │
│   - Charge Point Configs                    │
│   - Session Data                            │
│   - Energy Readings                         │
└─────────────────────────────────────────────┘
```

## Core Components

### 1. Mock Database (`database.ts`)
- In-memory storage for charge point configurations
- Session tracking (active transactions)
- Energy consumption history
- Validates charge point registration

### 2. OCPP WebSocket Server (`ocpp-server.ts`)
- Listens for charge point connections
- Handles OCPP 1.6J protocol messages
- Validates message format
- Routes messages to appropriate handlers

### 3. Connection Manager (`connection-manager.ts`)
- Maintains connections to multiple charge points
- Tracks connection status
- Handles reconnection logic
- Manages charge point state

### 4. Message Handlers (`message-handlers.ts`)
- Processes OCPP messages (BootNotification, StatusNotification, etc.)
- Sends responses according to OCPP spec
- Updates internal state

### 5. Service Wrapper (`charging-service.ts`)
- **Simple Public API** for external services
- Methods:
  - `turnOn(chargePointId, connectorId, userId)` - Start charging
  - `turnOff(chargePointId, transactionId)` - Stop charging
  - `getEnergyConsumption(chargePointId, transactionId)` - Get energy used
- Handles all OCPP complexity internally
- Returns clear errors for calling services

### 6. Error Handling (`errors.ts`)
- Custom error types for different scenarios
- User-friendly error messages
- Error codes for programmatic handling

## Data Flow

### Starting a Charging Session
```
Payment Service
  └─> chargingService.turnOn(cpId, connectorId, userId)
        └─> Validate charge point exists in DB
        └─> Check if CP is connected
        └─> Send RemoteStartTransaction via OCPP
        └─> Wait for CP response
        └─> Return success/error to payment service
```

### Getting Energy Consumption
```
Payment Service
  └─> chargingService.getEnergyConsumption(cpId, txId)
        └─> Look up transaction in database
        └─> Return latest energy reading
        └─> Return error if transaction not found
```

### Stopping a Charging Session
```
Payment Service
  └─> chargingService.turnOff(cpId, txId)
        └─> Send RemoteStopTransaction via OCPP
        └─> Wait for CP to stop and send final meter reading
        └─> Return final energy consumption
```

## Database Schema (Mock)

### ChargePoint
```typescript
{
  id: string;              // Unique charge point ID
  name: string;            // Friendly name
  connectors: number;      // Number of connectors
  status: 'online' | 'offline';
  lastSeen: Date;
}
```

### Session
```typescript
{
  transactionId: number;
  chargePointId: string;
  connectorId: number;
  userId: string;          // Who started the session
  startTime: Date;
  startMeterValue: number; // Wh at start
  currentMeterValue: number; // Current Wh
  endTime?: Date;
  endMeterValue?: number;  // Wh at end
  status: 'active' | 'completed';
}
```

## Error Types

- `ChargePointNotFoundError` - CP not in database
- `ChargePointOfflineError` - CP not connected
- `ConnectorUnavailableError` - Connector busy/faulted
- `SessionNotFoundError` - Transaction ID not found
- `RemoteOperationFailedError` - CP rejected remote command
- `TimeoutError` - CP didn't respond in time

## Usage Example

```typescript
// In payment service after payment approved
try {
  const result = await chargingService.turnOn(
    'CP001',      // charge point ID
    1,            // connector ID
    'user-123'    // user ID
  );

  console.log(`Session started: ${result.transactionId}`);

} catch (error) {
  if (error instanceof ChargePointNotFoundError) {
    displayToUser("Charge point not configured. Please contact support.");
  } else if (error instanceof ChargePointOfflineError) {
    displayToUser("Charge point is offline. Please try another station.");
  } else if (error instanceof ConnectorUnavailableError) {
    displayToUser("This connector is currently in use. Try another connector.");
  }
}
```

## Scalability Considerations

- Connection pooling for multiple charge points
- Message queue for high-volume scenarios
- Database could be replaced with PostgreSQL/MongoDB
- WebSocket server can be clustered
- Service wrapper remains unchanged (stable API)

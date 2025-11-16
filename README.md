# OCPP Management System

A comprehensive OCPP 1.6J management system for EV charging stations with a simple, easy-to-use API for integration with payment services and other applications.

## Features

- **OCPP 1.6J Protocol Support** - Full WebSocket-based OCPP implementation
- **Multiple Charge Point Management** - Handle connections from many charging stations simultaneously
- **Simple API** - Easy-to-use methods for controlling charging sessions:
  - `turnOn()` - Start charging
  - `turnOff()` - Stop charging
  - `getEnergyConsumption()` - Get real-time energy usage
- **Comprehensive Error Handling** - Clear, user-friendly error messages
- **Mock Database** - In-memory database (easily replaceable with PostgreSQL/MongoDB)
- **Real-time Monitoring** - Track energy consumption and session status
- **Unconfigured CP Detection** - Warns when unknown charge points connect

## Quick Start

### 1. Installation

```bash
npm install
```

### 2. Start the OCPP Server

```bash
npm run dev
```

The server will start on `ws://0.0.0.0:9000` and display configured charge points:

```
═══════════════════════════════════════════════════════════
         OCPP Management System v1.0.0
         OCPP 1.6J WebSocket Server
═══════════════════════════════════════════════════════════

📊 Configured Charge Points: 3
   - CP001: Main Street Station 1 (EVBox Elvi)
     Connectors: 2, Status: offline
   - CP002: Shopping Mall Station (ABB Terra AC)
     Connectors: 1, Status: offline
   - CP003: Office Parking Charger (ChargePoint CPE250)
     Connectors: 2, Status: offline

🔌 OCPP Server listening on ws://0.0.0.0:9000
   Waiting for charge points to connect...
```

### 3. Connect Charge Points

Charge points should connect to:
```
ws://localhost:9000/<charge-point-id>
```

Examples:
- `ws://localhost:9000/CP001`
- `ws://localhost:9000/CP002`
- `ws://localhost:9000/CP003`

### 4. Use the Emulator for Testing

Clone and use the OCPP virtual charge point emulator:

```bash
git clone https://github.com/solidstudiosh/ocpp-virtual-charge-point.git
cd ocpp-virtual-charge-point
npm install

# Connect as CP001
WS_URL=ws://localhost:9000/CP001 npx tsx index_16.ts
```

## Using the API

### In Your Payment Service

```typescript
import { chargingService } from './src/services/charging-service';
import { getUserFriendlyErrorMessage } from './src/errors/custom-errors';

// After payment is approved, start charging
async function startCharging(userId: string) {
  try {
    const result = await chargingService.turnOn(
      'CP001',      // charge point ID
      1,            // connector ID
      userId        // user ID
    );

    console.log(`Session started: ${result.transactionId}`);
    return result.transactionId;

  } catch (error) {
    // Get user-friendly error message to display
    const message = getUserFriendlyErrorMessage(error);
    console.error(message);
    throw error;
  }
}

// Monitor energy consumption
async function checkProgress(transactionId: number) {
  const data = await chargingService.getEnergyConsumption('CP001', transactionId);

  console.log(`Energy consumed: ${data.currentEnergyWh / 1000} kWh`);
  console.log(`Duration: ${data.durationSeconds} seconds`);
  console.log(`Status: ${data.status}`);
}

// Stop charging
async function stopCharging(transactionId: number) {
  const result = await chargingService.turnOff('CP001', transactionId);

  console.log(`Energy consumed: ${result.energyConsumed / 1000} kWh`);
  console.log(`Duration: ${result.duration} seconds`);
}
```

### Error Handling

The system provides clear, user-friendly error messages:

```typescript
import {
  ChargePointNotFoundError,
  ChargePointOfflineError,
  ConnectorUnavailableError,
  SessionNotFoundError
} from './src/errors/custom-errors';

try {
  await chargingService.turnOn('CP001', 1, 'user-123');
} catch (error) {
  if (error instanceof ChargePointNotFoundError) {
    // Display: "This charging station is not configured in the system..."
  } else if (error instanceof ChargePointOfflineError) {
    // Display: "This charging station is currently offline..."
  } else if (error instanceof ConnectorUnavailableError) {
    // Display: "This connector is currently in use..."
  }
}
```

## API Reference

### `chargingService.turnOn(chargePointId, connectorId, userId)`

Start a charging session.

**Parameters:**
- `chargePointId` (string) - ID of the charge point
- `connectorId` (number) - Connector number (1, 2, etc.)
- `userId` (string) - User/customer ID

**Returns:** `Promise<TurnOnResult>`
```typescript
{
  success: boolean;
  transactionId: number;
  chargePointId: string;
  connectorId: number;
  userId: string;
  message: string;
}
```

**Throws:**
- `ChargePointNotFoundError` - Charge point not in database
- `ChargePointOfflineError` - Charge point not connected
- `ConnectorNotFoundError` - Connector doesn't exist
- `ConnectorUnavailableError` - Connector busy/faulted
- `SessionAlreadyActiveError` - Connector already has active session

### `chargingService.turnOff(chargePointId, transactionId)`

Stop a charging session.

**Parameters:**
- `chargePointId` (string) - ID of the charge point
- `transactionId` (number) - Transaction ID to stop

**Returns:** `Promise<TurnOffResult>`
```typescript
{
  success: boolean;
  transactionId: number;
  chargePointId: string;
  energyConsumed: number;  // in Wh
  duration: number;        // in seconds
  message: string;
}
```

### `chargingService.getEnergyConsumption(chargePointId, transactionId)`

Get current energy consumption for a session.

**Parameters:**
- `chargePointId` (string) - ID of the charge point
- `transactionId` (number) - Transaction ID

**Returns:** `Promise<EnergyConsumptionResult>`
```typescript
{
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
```

## Project Structure

```
ocpp_management_system/
├── src/
│   ├── types/
│   │   ├── ocpp-types.ts          # OCPP protocol types
│   │   └── domain-types.ts        # Domain model types
│   ├── errors/
│   │   └── custom-errors.ts       # Custom error classes
│   ├── database/
│   │   └── mock-database.ts       # In-memory database
│   ├── ocpp/
│   │   ├── ocpp-server.ts         # WebSocket server
│   │   ├── connection-manager.ts  # Connection management
│   │   └── message-handlers.ts    # OCPP message handlers
│   ├── services/
│   │   └── charging-service.ts    # Simple API (turnOn/turnOff/getEnergy)
│   └── index.ts                   # Main entry point
├── examples/
│   └── payment-service-example.ts # Example integration
├── OCPP_1.6J_GUIDE.md            # OCPP protocol documentation
├── ARCHITECTURE.md                # System architecture
└── README.md                      # This file
```

## Architecture

The system is built in layers:

```
┌─────────────────────────────────────────────┐
│     Application Layer (Payment Service)     │
│  Uses: turnOn(), turnOff(), getEnergy()     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│         Service Wrapper (Simple API)        │
│   Hides OCPP complexity                     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│        OCPP Connection Manager              │
│   Manages multiple charge point connections │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│          WebSocket Server                   │
│   OCPP 1.6J protocol handling               │
└─────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

Create a `.env` file:

```env
OCPP_PORT=9000
```

### Adding Charge Points

Edit `src/database/mock-database.ts` to add your charge points:

```typescript
const cp: ChargePoint = {
  id: 'CP004',
  name: 'My New Station',
  vendor: 'MyVendor',
  model: 'Model-X',
  numberOfConnectors: 2,
  status: 'offline',
  registrationStatus: 'accepted',
  lastSeen: new Date(),
  createdAt: new Date()
};
```

Or replace the mock database with a real database (PostgreSQL, MongoDB, etc.).

## OCPP Protocol

This system implements **OCPP 1.6J** (JSON over WebSocket). See [OCPP_1.6J_GUIDE.md](./OCPP_1.6J_GUIDE.md) for detailed protocol documentation.

### Key OCPP Messages Handled

**From Charge Points:**
- `BootNotification` - Registration
- `Heartbeat` - Keep-alive
- `StatusNotification` - Connector status updates
- `MeterValues` - Energy consumption data
- `StartTransaction` - Session started
- `StopTransaction` - Session stopped
- `Authorize` - User authorization

**To Charge Points:**
- `RemoteStartTransaction` - Start charging remotely
- `RemoteStopTransaction` - Stop charging remotely

## Unconfigured Charge Points

If a charge point connects but is not in the database, the system will:

1. Accept the connection temporarily
2. Print a warning:
   ```
   ⚠️  WARNING: Charge point 'CP999' is NOT CONFIGURED!
      Please configure this charge point in the database before use.
   ```
3. Reject API calls with `ChargePointNotConfiguredError`

## Testing

### Run Example

```bash
npm run dev

# In another terminal
npx ts-node examples/payment-service-example.ts
```

### Manual Testing

```bash
# Start server
npm run dev

# In another terminal, connect emulator
cd ocpp-virtual-charge-point
WS_URL=ws://localhost:9000/CP001 npx tsx index_16.ts

# The charge point will boot and send StatusNotification
# You can then use the API to control it
```

## Production Deployment

For production:

1. **Replace Mock Database** - Use PostgreSQL, MongoDB, or your preferred database
2. **Add Authentication** - Implement charge point authentication
3. **Add User Database** - Store and validate user credentials
4. **Scale WebSocket Server** - Use clustering for high availability
5. **Add Monitoring** - Prometheus, Grafana, etc.
6. **SSL/TLS** - Use WSS (secure WebSocket)

## Extensibility

The system is designed to be easily extended:

- **Add new OCPP messages** - Edit `message-handlers.ts`
- **Add business logic** - Extend `charging-service.ts`
- **Replace database** - Implement the same interface as `mock-database.ts`
- **Add middleware** - Modify `ocpp-server.ts` for authentication, logging, etc.

## Troubleshooting

### Charge point won't connect

1. Check the WebSocket URL format: `ws://ip:9000/CP001`
2. Ensure firewall allows port 9000
3. Check server logs for connection attempts

### "Charge point not configured" error

Add the charge point to the database in `src/database/mock-database.ts`

### Energy consumption not updating

1. Ensure charge point sends `MeterValues` messages
2. Check that `transactionId` is included in `MeterValues`
3. Verify measurand is `Energy.Active.Import.Register`

## License

MIT

## Support

For issues, questions, or contributions, please open an issue on GitHub.

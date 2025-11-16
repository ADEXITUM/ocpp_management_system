# OCPP Management System (Go)

A comprehensive OCPP 1.6J management system for EV charging stations written in **Go**, with a simple, easy-to-use API for integration with payment services and other applications.

## Features

- **OCPP 1.6J Protocol Support** - Full WebSocket-based OCPP implementation
- **Multiple Charge Point Management** - Handle connections from many charging stations simultaneously
- **REST API** - HTTP endpoints for easy integration (see [API_EXAMPLES.md](./API_EXAMPLES.md)):
  - `POST /sessions/start` - Start charging with payment
  - `GET /sessions/{id}/energy` - Get current energy consumption
  - `POST /sessions/{id}/stop` - Stop charging session
- **Go Service API** - Programmatic API for Go applications:
  - `TurnOn()` - Start charging
  - `TurnOff()` - Stop charging
  - `GetEnergyConsumption()` - Get real-time energy usage
- **Comprehensive Error Handling** - Clear, user-friendly error messages
- **Mock Database** - In-memory database (easily replaceable with PostgreSQL/MongoDB)
- **Real-time Monitoring** - Track energy consumption and session status
- **Unconfigured CP Detection** - Warns when unknown charge points connect
- **High Performance** - Written in Go for excellent concurrency and performance

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone <repo-url>
cd ocpp_management_system

# Download dependencies
go mod download
```

### 2. Build the Server

```bash
go build -o ocpp-server cmd/server/main.go
```

### 3. Start the OCPP Server

```bash
./ocpp-server
# Or run directly:
go run cmd/server/main.go
```

The server will start two services:
- **OCPP WebSocket** on port `9000` (for charge points)
- **REST API** on port `8080` (for your application)

```
═══════════════════════════════════════════════════════
         OCPP Management System v1.0.0
         OCPP 1.6J WebSocket Server (Go)
═══════════════════════════════════════════════════════

📊 Configured Charge Points: 3
   - CP001: Main Street Station 1 (EVBox Elvi)
     Connectors: 2, Status: offline

✅ System ready!

🌐 REST API listening on http://0.0.0.0:8080
   POST   http://0.0.0.0:8080/sessions/start
   GET    http://0.0.0.0:8080/sessions/{transactionId}/energy
   POST   http://0.0.0.0:8080/sessions/{transactionId}/stop

🔌 OCPP Server listening on ws://0.0.0.0:9000
   Waiting for charge points to connect...
```

### 4. Connect Charge Points

Charge points should connect to:
```
ws://localhost:9000/<charge-point-id>
```

Examples:
- `ws://localhost:9000/CP001`
- `ws://localhost:9000/CP002`
- `ws://localhost:9000/CP003`

### 5. Use the Emulator for Testing

Clone and use the OCPP virtual charge point emulator:

```bash
git clone https://github.com/solidstudiosh/ocpp-virtual-charge-point.git
cd ocpp-virtual-charge-point
npm install

# Connect as CP001
WS_URL=ws://localhost:9000/CP001 npx tsx index_16.ts
```

## Using the REST API (Recommended)

The easiest way to control charging sessions is via REST API. See [API_EXAMPLES.md](./API_EXAMPLES.md) for complete examples.

### Quick Example

```bash
# 1. Start charging session (with payment)
curl -X POST http://localhost:8080/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "chargePointId": "CP001",
    "connectorId": 1,
    "userId": "user-123",
    "amountPaid": 25.00
  }'

# Response:
# {
#   "success": true,
#   "transactionId": 1,
#   "paymentStatus": "ОПЛАТА ПРОШЛА УСПЕШНО ✓"
# }

# 2. Get energy consumption
curl "http://localhost:8080/sessions/1/energy?chargePointId=CP001"

# 3. Stop charging
curl -X POST "http://localhost:8080/sessions/1/stop?chargePointId=CP001"
```

See [API_EXAMPLES.md](./API_EXAMPLES.md) for:
- Complete curl examples
- JavaScript/TypeScript integration
- Python integration
- Error handling
- Postman collection

## Using the Go Service API

### In Your Payment Service (Go)

```go
package main

import (
    "fmt"
    "log"

    "github.com/ADEXITUM/ocpp_management_system/pkg/database"
    "github.com/ADEXITUM/ocpp_management_system/pkg/errors"
    "github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
    "github.com/ADEXITUM/ocpp_management_system/pkg/service"
)

func main() {
    // Initialize
    db := database.NewMockDatabase()
    server := ocpp.NewServer(9000, db)
    chargingService := service.NewChargingService(db, server.GetConnectionManager())

    // Start charging after payment approved
    result, err := chargingService.TurnOn("CP001", 1, "user-123")
    if err != nil {
        handleError(err)
        return
    }

    fmt.Printf("Session started: %d\n", result.TransactionID)

    // Monitor energy consumption
    consumption, err := chargingService.GetEnergyConsumption("CP001", result.TransactionID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Energy consumed: %.2f kWh\n", consumption.CurrentEnergyWh/1000)
    fmt.Printf("Duration: %d seconds\n", consumption.DurationSeconds)

    // Stop charging
    stopResult, err := chargingService.TurnOff("CP001", result.TransactionID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Final energy: %.2f kWh\n", stopResult.EnergyConsumed/1000)
}

func handleError(err error) {
    if ocppErr, ok := err.(*errors.OCPPError); ok {
        fmt.Println("Error:", ocppErr.UserMessage)

        switch ocppErr.Code {
        case "CHARGE_POINT_NOT_FOUND":
            // Handle: refund payment, notify user
        case "CHARGE_POINT_OFFLINE":
            // Handle: refund payment, suggest another station
        case "CONNECTOR_UNAVAILABLE":
            // Handle: refund payment, suggest another connector
        }
    }
}
```

### Error Handling

The system provides clear, user-friendly error messages:

```go
import "github.com/ADEXITUM/ocpp_management_system/pkg/errors"

result, err := chargingService.TurnOn("CP001", 1, "user-123")
if err != nil {
    // Get user-friendly message
    userMessage := errors.GetUserFriendlyMessage(err)
    fmt.Println(userMessage)

    // Or type-check for specific errors
    if ocppErr, ok := err.(*errors.OCPPError); ok {
        switch ocppErr.Code {
        case "CHARGE_POINT_NOT_FOUND":
            // "This charging station is not configured in the system..."
        case "CHARGE_POINT_OFFLINE":
            // "This charging station is currently offline..."
        case "CONNECTOR_UNAVAILABLE":
            // "This connector is currently in use..."
        }
    }
}
```

## API Reference

### `TurnOn(chargePointID, connectorID, userID)`

Start a charging session.

**Parameters:**
- `chargePointID` (string) - ID of the charge point
- `connectorID` (int) - Connector number (1, 2, etc.)
- `userID` (string) - User/customer ID

**Returns:** `*TurnOnResult, error`
```go
type TurnOnResult struct {
    Success       bool
    TransactionID int
    ChargePointID string
    ConnectorID   int
    UserID        string
    Message       string
}
```

**Errors:**
- `ChargePointNotFoundError` - Charge point not in database
- `ChargePointOfflineError` - Charge point not connected
- `ConnectorNotFoundError` - Connector doesn't exist
- `ConnectorUnavailableError` - Connector busy/faulted
- `SessionAlreadyActiveError` - Connector already has active session

### `TurnOff(chargePointID, transactionID)`

Stop a charging session.

**Parameters:**
- `chargePointID` (string) - ID of the charge point
- `transactionID` (int) - Transaction ID to stop

**Returns:** `*TurnOffResult, error`
```go
type TurnOffResult struct {
    Success        bool
    TransactionID  int
    ChargePointID  string
    EnergyConsumed float64  // in Wh
    Duration       int      // in seconds
    Message        string
}
```

### `GetEnergyConsumption(chargePointID, transactionID)`

Get current energy consumption for a session.

**Parameters:**
- `chargePointID` (string) - ID of the charge point
- `transactionID` (int) - Transaction ID

**Returns:** `*EnergyConsumptionResult, error`
```go
type EnergyConsumptionResult struct {
    TransactionID   int
    ChargePointID   string
    ConnectorID     int
    UserID          string
    StartTime       time.Time
    CurrentEnergyWh float64
    DurationSeconds int
    Status          string  // "active" or "completed"
    LastUpdate      time.Time
}
```

## Project Structure

```
ocpp_management_system/
├── cmd/
│   └── server/
│       └── main.go                # Main entry point
├── pkg/
│   ├── types/
│   │   └── ocpp_types.go          # OCPP protocol types & domain types
│   ├── errors/
│   │   └── errors.go              # Custom error types
│   ├── database/
│   │   └── mock_database.go       # In-memory database
│   ├── ocpp/
│   │   ├── server.go              # WebSocket server
│   │   ├── connection_manager.go  # Connection management
│   │   └── message_handlers.go    # OCPP message handlers
│   └── service/
│       └── charging_service.go    # Simple API (TurnOn/TurnOff/GetEnergy)
├── examples/
│   └── payment_service_example.go # Example integration
├── OCPP_1.6J_GUIDE.md            # OCPP protocol documentation
├── ARCHITECTURE.md                # System architecture
├── go.mod                         # Go module file
└── README.md                      # This file
```

## Architecture

The system is built in layers:

```
┌─────────────────────────────────────────────┐
│     Application Layer (Payment Service)     │
│  Uses: TurnOn(), TurnOff(), GetEnergy()     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│         Service Package (Simple API)        │
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

Set environment variables or create a `.env` file:

```bash
export OCPP_PORT=9000
```

### Adding Charge Points

Edit `pkg/database/mock_database.go` to add your charge points:

```go
cp := &types.ChargePoint{
    ID:                 "CP004",
    Name:               "My New Station",
    Vendor:             "MyVendor",
    Model:              "Model-X",
    NumberOfConnectors: 2,
    Status:             "offline",
    RegistrationStatus: "accepted",
    LastSeen:           time.Now(),
    CreatedAt:          time.Now(),
}
db.chargePoints[cp.ID] = cp
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
# Terminal 1: Start server
go run cmd/server/main.go

# Terminal 2: Run example
go run examples/payment_service_example.go
```

### Manual Testing

```bash
# Start server
go run cmd/server/main.go

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
4. **Scale WebSocket Server** - Use load balancing for high availability
5. **Add Monitoring** - Prometheus, Grafana, etc.
6. **SSL/TLS** - Use WSS (secure WebSocket)
7. **Build optimized binary** - `go build -ldflags="-s -w" cmd/server/main.go`

## Extensibility

The system is designed to be easily extended:

- **Add new OCPP messages** - Edit `pkg/ocpp/message_handlers.go`
- **Add business logic** - Extend `pkg/service/charging_service.go`
- **Replace database** - Implement the same interface as `pkg/database/mock_database.go`
- **Add middleware** - Modify `pkg/ocpp/server.go` for authentication, logging, etc.

## Troubleshooting

### Charge point won't connect

1. Check the WebSocket URL format: `ws://ip:9000/CP001`
2. Ensure firewall allows port 9000
3. Check server logs for connection attempts

### "Charge point not configured" error

Add the charge point to the database in `pkg/database/mock_database.go`

### Energy consumption not updating

1. Ensure charge point sends `MeterValues` messages
2. Check that `transactionId` is included in `MeterValues`
3. Verify measurand is `Energy.Active.Import.Register`

## Performance

Go's excellent concurrency support makes this system highly performant:
- Goroutines handle each charge point connection independently
- Efficient memory usage with typed structs
- Fast JSON marshaling/unmarshaling
- Low latency WebSocket communication

## License

MIT

## Support

For issues, questions, or contributions, please open an issue on GitHub.

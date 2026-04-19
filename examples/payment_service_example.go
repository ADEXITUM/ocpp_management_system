package main

import (
	"fmt"

	"github.com/ADEXITUM/ocpp_management_system/pkg/errors"
)

/*
This example demonstrates how to integrate the OCPP Charging Service
into your payment/billing application.

IMPORTANT: This is example code showing the API usage.
To actually run charging operations, you need:
1. The OCPP server running: go run cmd/server/main.go
2. A charge point connected via the emulator
3. Access to the server's ChargingService instance

For working integration tests, see: pkg/service/charging_service_test.go
Run with: go test ./pkg/service -v
*/

// PaymentService demonstrates integration with charging service
type PaymentService struct {
	// In your real application, you would inject the charging service
	// from the running server instance, not create a new one
	// chargingService *service.ChargingService
}

// HandlePaymentApproved shows how to start charging after payment
func (ps *PaymentService) HandlePaymentApproved(
	userID string,
	chargePointID string,
	connectorID int,
	amountPaid float64,
) {
	fmt.Printf("\n💳 Payment approved for user %s: $%.2f\n", userID, amountPaid)
	fmt.Printf("   Starting charging session on %s connector %d...\n", chargePointID, connectorID)

	/*
		// In your real application:
		result, err := ps.chargingService.TurnOn(chargePointID, connectorID, userID)

		if err != nil {
			// Handle errors with user-friendly messages
			userMessage := errors.GetUserFriendlyMessage(err)

			switch err.(type) {
			case *errors.OCPPError:
				ocppErr := err.(*errors.OCPPError)
				switch ocppErr.Code {
				case "CHARGE_POINT_NOT_FOUND":
					ps.refundPayment(userID, amountPaid)
					ps.notifyUser(userID, userMessage)
				case "CHARGE_POINT_OFFLINE":
					ps.refundPayment(userID, amountPaid)
					ps.notifyUser(userID, userMessage)
				case "CONNECTOR_UNAVAILABLE":
					ps.refundPayment(userID, amountPaid)
					ps.notifyUser(userID, userMessage)
				default:
					ps.refundPayment(userID, amountPaid)
				}
			default:
				ps.refundPayment(userID, amountPaid)
			}
			return
		}

		fmt.Printf("✅ %s\n", result.Message)
		fmt.Printf("   Transaction ID: %d\n", result.TransactionID)
		fmt.Printf("   User can now charge their vehicle!\n")

		// Store transaction ID for later
		ps.storeTransaction(userID, result.TransactionID, amountPaid)
	*/

	// Example output
	fmt.Println("\n   [Example Code - See charging_service_test.go for working tests]")
	fmt.Println("   ✓ Would call: chargingService.TurnOn(chargePointID, connectorID, userID)")
	fmt.Println("   ✓ Would receive: TransactionID and confirmation")
	fmt.Println("   ✓ Would store: Transaction for billing")
}

// CheckChargingProgress shows how to monitor energy consumption
func (ps *PaymentService) CheckChargingProgress(transactionID int, chargePointID string) {
	fmt.Printf("\n⚡ Checking Charging Progress:\n")

	/*
		// In your real application:
		consumption, err := ps.chargingService.GetEnergyConsumption(chargePointID, transactionID)
		if err != nil {
			// Handle error
			return
		}

		energyKwh := consumption.CurrentEnergyWh / 1000
		durationMinutes := consumption.DurationSeconds / 60
		cost := ps.calculateCost(consumption.CurrentEnergyWh)

		fmt.Printf("   Transaction: %d\n", transactionID)
		fmt.Printf("   User: %s\n", consumption.UserID)
		fmt.Printf("   Energy: %.2f kWh\n", energyKwh)
		fmt.Printf("   Duration: %d minutes\n", durationMinutes)
		fmt.Printf("   Cost so far: $%.2f\n", cost)
		fmt.Printf("   Status: %s\n", consumption.Status)
	*/

	// Example output
	fmt.Println("   [Example Code]")
	fmt.Println("   ✓ Would call: chargingService.GetEnergyConsumption(chargePointID, txID)")
	fmt.Println("   ✓ Would receive: Current energy, duration, status")
	fmt.Println("   ✓ Would calculate: Cost based on energy consumed")
}

// StopChargingSession shows how to stop a session
func (ps *PaymentService) StopChargingSession(
	transactionID int,
	chargePointID string,
	reason string,
) {
	fmt.Printf("\n🛑 Stopping charging session %d...\n", transactionID)
	fmt.Printf("   Reason: %s\n", reason)

	/*
		// In your real application:
		result, err := ps.chargingService.TurnOff(chargePointID, transactionID)
		if err != nil {
			userMessage := errors.GetUserFriendlyMessage(err)
			fmt.Printf("❌ Error stopping session: %s\n", userMessage)
			return
		}

		energyKwh := result.EnergyConsumed / 1000
		durationMinutes := result.Duration / 60
		finalCost := ps.calculateCost(result.EnergyConsumed)

		fmt.Printf("✅ %s\n", result.Message)
		fmt.Printf("\n📊 Session Summary:\n")
		fmt.Printf("   Total Energy: %.2f kWh\n", energyKwh)
		fmt.Printf("   Duration: %d minutes\n", durationMinutes)
		fmt.Printf("   Final Cost: $%.2f\n", finalCost)

		ps.processFinalPayment(transactionID, finalCost)
	*/

	// Example output
	fmt.Println("\n   [Example Code]")
	fmt.Println("   ✓ Would call: chargingService.TurnOff(chargePointID, transactionID)")
	fmt.Println("   ✓ Would receive: Total energy, duration, final cost")
	fmt.Println("   ✓ Would process: Final payment/billing")
}

// Error handling example
func ExampleErrorHandling() {
	fmt.Println("\n📋 Error Handling Examples:")
	fmt.Println()

	// Example 1: Charge point not found
	err := errors.NewChargePointNotFoundError("CP999")
	fmt.Printf("Error Type: ChargePointNotFoundError\n")
	fmt.Printf("Error Code: %s\n", err.Code)
	fmt.Printf("User Message: %s\n\n", err.UserMessage)

	// Example 2: Charge point offline
	err = errors.NewChargePointOfflineError("CP001")
	fmt.Printf("Error Type: ChargePointOfflineError\n")
	fmt.Printf("Error Code: %s\n", err.Code)
	fmt.Printf("User Message: %s\n\n", err.UserMessage)

	// Example 3: Connector unavailable
	err = errors.NewConnectorUnavailableError("CP001", 1, "Charging")
	fmt.Printf("Error Type: ConnectorUnavailableError\n")
	fmt.Printf("Error Code: %s\n", err.Code)
	fmt.Printf("User Message: %s\n\n", err.UserMessage)
}

func main() {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("    OCPP Charging Service - Integration Examples")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("This file shows example code for integrating the")
	fmt.Println("OCPP Charging Service into your payment application.")
	fmt.Println()
	fmt.Println("📌 For WORKING TESTS, run:")
	fmt.Println("   go test ./pkg/service -v")
	fmt.Println()
	fmt.Println("📌 Prerequisites for integration tests:")
	fmt.Println("   1. Start OCPP server: go run cmd/server/main.go")
	fmt.Println("   2. Connect emulator: WS_URL=ws://localhost:9005/CP001 npx tsx index_16.ts")
	fmt.Println("   3. Run tests: go test ./pkg/service -v -run TestChargingServiceFlow")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")

	paymentService := &PaymentService{}

	// Example 1: Start charging after payment
	paymentService.HandlePaymentApproved("user-123", "CP001", 1, 25.0)

	// Example 2: Check charging progress
	paymentService.CheckChargingProgress(1, "CP001")

	// Example 3: Stop charging
	paymentService.StopChargingSession(1, "CP001", "User requested stop")

	// Example 4: Error handling
	ExampleErrorHandling()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("    API Summary")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("// Start charging")
	fmt.Println("result, err := chargingService.TurnOn(chargePointID, connectorID, userID)")
	fmt.Println()
	fmt.Println("// Get energy consumption")
	fmt.Println("consumption, err := chargingService.GetEnergyConsumption(chargePointID, txID)")
	fmt.Println()
	fmt.Println("// Stop charging")
	fmt.Println("result, err := chargingService.TurnOff(chargePointID, transactionID)")
	fmt.Println()
	fmt.Println("// Handle errors")
	fmt.Println("if ocppErr, ok := err.(*errors.OCPPError); ok {")
	fmt.Println("    userMessage := ocppErr.UserMessage")
	fmt.Println("    // Display to user or handle based on error code")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
}

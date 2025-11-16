package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/errors"
	"github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
	"github.com/ADEXITUM/ocpp_management_system/pkg/service"
)

// PaymentService is an example payment service
type PaymentService struct {
	chargingService *service.ChargingService
}

// NewPaymentService creates a new payment service
func NewPaymentService(chargingService *service.ChargingService) *PaymentService {
	return &PaymentService{
		chargingService: chargingService,
	}
}

// HandlePaymentApproved is called after user payment is approved
func (ps *PaymentService) HandlePaymentApproved(
	userID string,
	chargePointID string,
	connectorID int,
	amountPaid float64,
) {
	fmt.Printf("\n💳 Payment approved for user %s: $%.2f\n", userID, amountPaid)
	fmt.Printf("   Starting charging session on %s connector %d...\n", chargePointID, connectorID)

	// Simple API - just turn on!
	result, err := ps.chargingService.TurnOn(chargePointID, connectorID, userID)

	if err != nil {
		// Handle errors with user-friendly messages
		userMessage := errors.GetUserFriendlyMessage(err)

		switch err.(type) {
		case *errors.OCPPError:
			ocppErr := err.(*errors.OCPPError)
			switch ocppErr.Code {
			case "CHARGE_POINT_NOT_FOUND", "CHARGE_POINT_OFFLINE", "CONNECTOR_UNAVAILABLE":
				fmt.Printf("❌ Error: %s\n", userMessage)
				ps.refundPayment(userID, amountPaid)
				ps.notifyUser(userID, userMessage)
			default:
				fmt.Printf("❌ Error: %s\n", userMessage)
				ps.refundPayment(userID, amountPaid)
			}
		default:
			fmt.Printf("❌ Unexpected error: %v\n", err)
			ps.refundPayment(userID, amountPaid)
		}
		return
	}

	fmt.Printf("✅ %s\n", result.Message)
	fmt.Printf("   Transaction ID: %d\n", result.TransactionID)
	fmt.Printf("   User can now charge their vehicle!\n")

	// Store transaction ID for later (in your database)
	ps.storeTransaction(userID, result.TransactionID, amountPaid)
}

// CheckChargingProgress monitors energy consumption during charging
func (ps *PaymentService) CheckChargingProgress(transactionID int, chargePointID string) {
	consumption, err := ps.chargingService.GetEnergyConsumption(chargePointID, transactionID)
	if err != nil {
		if _, ok := err.(*errors.OCPPError); ok {
			fmt.Printf("❌ Session not found\n")
		} else {
			fmt.Printf("❌ Error checking progress: %v\n", err)
		}
		return
	}

	energyKwh := consumption.CurrentEnergyWh / 1000
	durationMinutes := consumption.DurationSeconds / 60
	cost := ps.calculateCost(consumption.CurrentEnergyWh)

	fmt.Printf("\n⚡ Charging Progress:\n")
	fmt.Printf("   Transaction: %d\n", transactionID)
	fmt.Printf("   User: %s\n", consumption.UserID)
	fmt.Printf("   Energy: %.2f kWh\n", energyKwh)
	fmt.Printf("   Duration: %d minutes\n", durationMinutes)
	fmt.Printf("   Cost so far: $%.2f\n", cost)
	fmt.Printf("   Status: %s\n", consumption.Status)
}

// StopChargingSession stops a charging session
func (ps *PaymentService) StopChargingSession(
	transactionID int,
	chargePointID string,
	reason string,
) {
	fmt.Printf("\n🛑 Stopping charging session %d...\n", transactionID)
	fmt.Printf("   Reason: %s\n", reason)

	// Simple API - just turn off!
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

	// Process final payment
	ps.processFinalPayment(transactionID, finalCost)
}

// Mock helper methods

func (ps *PaymentService) storeTransaction(userID string, transactionID int, amountPaid float64) {
	// In real app: store in your database
	fmt.Printf("   💾 Stored transaction %d in database\n", transactionID)
}

func (ps *PaymentService) refundPayment(userID string, amount float64) {
	// In real app: process refund
	fmt.Printf("   💰 Refunded $%.2f to user %s\n", amount, userID)
}

func (ps *PaymentService) notifyUser(userID, message string) {
	// In real app: send push notification, SMS, etc.
	fmt.Printf("   📧 Notified user %s: %s\n", userID, message)
}

func (ps *PaymentService) calculateCost(energyWh float64) float64 {
	// $0.30 per kWh
	pricePerKwh := 0.3
	energyKwh := energyWh / 1000
	return energyKwh * pricePerKwh
}

func (ps *PaymentService) processFinalPayment(transactionID int, amount float64) {
	// In real app: charge credit card, update balance, etc.
	fmt.Printf("   💳 Processed final payment: $%.2f\n", amount)
}

// Demo scenario
func demo(chargingService *service.ChargingService) {
	paymentService := NewPaymentService(chargingService)

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("         Payment Service Integration Demo")
	fmt.Println("═══════════════════════════════════════════════════════")

	// Scenario: User pays and starts charging
	paymentService.HandlePaymentApproved("user-123", "CP001", 1, 25.0)

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Note: In real usage with actual charge points, there would be real transactions
	// paymentService.CheckChargingProgress(1, "CP001")
	// paymentService.StopChargingSession(1, "CP001", "User requested stop")

	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("         Demo Complete")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
}

func main() {
	fmt.Println("⚠️  Make sure the OCPP server is running (go run cmd/server/main.go)")
	fmt.Println("    And charge points are connected before running this demo\n")

	// Set up services (same as server)
	db := database.NewMockDatabase()
	server := ocpp.NewServer(9000, db)
	chargingService := service.NewChargingService(db, server.GetConnectionManager())

	time.Sleep(1 * time.Second)

	demo(chargingService)
}

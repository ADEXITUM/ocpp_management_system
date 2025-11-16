package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
	"github.com/ADEXITUM/ocpp_management_system/pkg/service"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("         OCPP Management System v1.0.0")
	fmt.Println("         OCPP 1.6J WebSocket Server (Go)")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Get port from environment or use default
	port := 9000
	if portEnv := os.Getenv("OCPP_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	}

	// Initialize database
	db := database.NewMockDatabase()

	// Display configured charge points
	chargePoints := db.GetAllChargePoints()
	fmt.Printf("📊 Configured Charge Points: %d\n", len(chargePoints))
	for _, cp := range chargePoints {
		fmt.Printf("   - %s: %s (%s %s)\n", cp.ID, cp.Name, cp.Vendor, cp.Model)
		fmt.Printf("     Connectors: %d, Status: %s\n", cp.NumberOfConnectors, cp.Status)
	}
	fmt.Println()

	// Create OCPP server
	server := ocpp.NewServer(port, db)

	// Create charging service (for API usage)
	chargingService := service.NewChargingService(db, server.GetConnectionManager())
	_ = chargingService // Available for use

	// Start periodic status display
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			displayStatus(db, server.GetConnectionManager())
		}
	}()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Shutting down OCPP Management System...")
		os.Exit(0)
	}()

	// Display API info
	fmt.Println("📡 Service API Available:")
	fmt.Println("   - chargingService.TurnOn(chargePointId, connectorId, userId)")
	fmt.Println("   - chargingService.TurnOff(chargePointId, transactionId)")
	fmt.Println("   - chargingService.GetEnergyConsumption(chargePointId, transactionId)")
	fmt.Println()

	fmt.Println("✅ System ready to accept charge point connections")
	fmt.Printf("   Charge points should connect to: ws://<server-ip>:%d/<charge-point-id>\n", port)
	fmt.Printf("   Example: ws://localhost:%d/CP001\n\n", port)

	// Start server (blocking)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start OCPP server: %v", err)
	}
}

func displayStatus(db *database.MockDatabase, cm *ocpp.ConnectionManager) {
	stats := db.GetStats()

	fmt.Println("\n─────────────────────────────────────────────────────")
	fmt.Printf("📊 System Status - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Charge Points: %d/%d online\n", stats.OnlineChargePoints, stats.TotalChargePoints)
	fmt.Printf("   Active Sessions: %d\n", stats.ActiveSessions)
	fmt.Printf("   WebSocket Connections: %d\n", len(cm.GetConnectedChargePoints()))
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()
}

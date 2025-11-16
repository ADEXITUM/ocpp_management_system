package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/api"
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

	// Get ports from environment or use defaults
	ocppPort := 9000
	if portEnv := os.Getenv("OCPP_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			ocppPort = p
		}
	}

	apiPort := 8080
	if portEnv := os.Getenv("API_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			apiPort = p
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

	// Create OCPP WebSocket server
	ocppServer := ocpp.NewServer(ocppPort, db)

	// Create charging service
	chargingService := service.NewChargingService(db, ocppServer.GetConnectionManager())

	// Create REST API server
	apiServer := api.NewServer(apiPort, chargingService)

	// Start periodic status display
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			displayStatus(db, ocppServer.GetConnectionManager())
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

	fmt.Println("✅ System ready!")
	fmt.Println()

	// Start REST API server in goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("❌ Failed to start API server: %v", err)
		}
	}()

	// Give API server time to start
	time.Sleep(500 * time.Millisecond)

	// Start OCPP WebSocket server (blocking)
	if err := ocppServer.Start(); err != nil {
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

package service

import (
	"testing"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
)

// TestChargingServiceFlow tests the complete charging flow
// NOTE: This requires a charge point to be connected to test RemoteStart/Stop
// For manual testing: run server, connect emulator, then run this test
func TestChargingServiceFlow(t *testing.T) {
	// Setup
	db := database.NewMockDatabase()
	server := ocpp.NewServer(9005, db)
	cm := server.GetConnectionManager()
	service := NewChargingService(db, cm)

	chargePointID := "my-test"
	connectorID := 1
	userID := "test-user-123"

	t.Run("1. Check charge point exists", func(t *testing.T) {
		cp := db.GetChargePoint(chargePointID)
		if cp == nil {
			t.Fatalf("Charge point %s not found in database", chargePointID)
		}
		t.Logf("✓ Charge point exists: %s - %s", cp.ID, cp.Name)
	})

	t.Run("2. Check charge point is configured", func(t *testing.T) {
		cp := db.GetChargePoint(chargePointID)
		if cp.RegistrationStatus != "accepted" {
			t.Fatalf("Charge point %s not configured (status: %s)", chargePointID, cp.RegistrationStatus)
		}
		t.Logf("✓ Charge point is configured and accepted")
	})

	t.Run("3. Start charging session (requires connected charge point)", func(t *testing.T) {
		// Check if charge point is connected
		if !cm.IsConnected(chargePointID) {
			t.Skip("Skipping: Charge point not connected. Connect emulator to run this test.")
		}

		// Start charging
		result, err := service.TurnOn(chargePointID, connectorID, userID)
		if err != nil {
			t.Fatalf("Failed to start charging: %v", err)
		}

		if !result.Success {
			t.Fatalf("TurnOn failed: %s", result.Message)
		}

		t.Logf("✓ Charging session started")
		t.Logf("  Transaction ID: %d", result.TransactionID)
		t.Logf("  Charge Point: %s", result.ChargePointID)
		t.Logf("  Connector: %d", result.ConnectorID)
		t.Logf("  User: %s", result.UserID)

		// Store transaction ID for next tests
		transactionID := result.TransactionID

		t.Run("4. Get current energy consumption", func(t *testing.T) {
			// Wait a bit for meter values
			time.Sleep(2 * time.Second)

			consumption, err := service.GetEnergyConsumption(chargePointID, transactionID)
			if err != nil {
				t.Fatalf("Failed to get energy consumption: %v", err)
			}

			t.Logf("✓ Energy consumption retrieved")
			t.Logf("  Transaction ID: %d", consumption.TransactionID)
			t.Logf("  Energy: %.2f Wh", consumption.CurrentEnergyWh)
			t.Logf("  Duration: %d seconds", consumption.DurationSeconds)
			t.Logf("  Status: %s", consumption.Status)
			t.Logf("  Last Update: %s", consumption.LastUpdate.Format(time.RFC3339))
		})

		t.Run("5. Stop charging session", func(t *testing.T) {
			// Wait a bit more
			time.Sleep(2 * time.Second)

			result, err := service.TurnOff(chargePointID, transactionID)
			if err != nil {
				t.Fatalf("Failed to stop charging: %v", err)
			}

			if !result.Success {
				t.Fatalf("TurnOff failed: %s", result.Message)
			}

			t.Logf("✓ Charging session stopped")
			t.Logf("  Transaction ID: %d", result.TransactionID)
			t.Logf("  Energy Consumed: %.2f Wh (%.3f kWh)", result.EnergyConsumed, result.EnergyConsumed/1000)
			t.Logf("  Duration: %d seconds", result.Duration)
			t.Logf("  Message: %s", result.Message)
		})
	})
}

// TestChargingServiceErrors tests error handling
func TestChargingServiceErrors(t *testing.T) {
	db := database.NewMockDatabase()
	server := ocpp.NewServer(9001, db) // Different port
	cm := server.GetConnectionManager()
	service := NewChargingService(db, cm)

	t.Run("ChargePointNotFound", func(t *testing.T) {
		_, err := service.TurnOn("NONEXISTENT", 1, "user-123")
		if err == nil {
			t.Fatal("Expected error for non-existent charge point")
		}
		t.Logf("✓ Correct error: %v", err)
	})

	t.Run("ChargePointOffline", func(t *testing.T) {
		// my-test exists but is not connected
		_, err := service.TurnOn("my-test", 1, "user-123")
		if err == nil {
			t.Fatal("Expected error for offline charge point")
		}
		t.Logf("✓ Correct error: %v", err)
	})

	t.Run("ConnectorNotFound", func(t *testing.T) {
		_, err := service.TurnOn("my-test", 99, "user-123")
		if err == nil {
			t.Fatal("Expected error for non-existent connector")
		}
		t.Logf("✓ Correct error: %v", err)
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		_, err := service.GetEnergyConsumption("my-test", 99999)
		if err == nil {
			t.Fatal("Expected error for non-existent session")
		}
		t.Logf("✓ Correct error: %v", err)
	})
}

// TestDatabaseOperations tests database functionality
func TestDatabaseOperations(t *testing.T) {
	db := database.NewMockDatabase()

	t.Run("Charge points initialized", func(t *testing.T) {
		chargePoints := db.GetAllChargePoints()
		if len(chargePoints) != 1 {
			t.Fatalf("Expected 1 charge point, got %d", len(chargePoints))
		}
		t.Logf("✓ Found %d charge points", len(chargePoints))
	})

	t.Run("Create and retrieve session", func(t *testing.T) {
		session := &types.ChargingSession{
			ChargePointID:     "my-test",
			ConnectorID:       1,
			UserID:            "test-user",
			StartTime:         time.Now(),
			StartMeterValue:   0,
			CurrentMeterValue: 0,
			Status:            "active",
		}

		txID := db.CreateSession(session)
		if txID == 0 {
			t.Fatal("Failed to create session")
		}

		retrieved := db.GetSession(txID)
		if retrieved == nil {
			t.Fatal("Failed to retrieve session")
		}

		if retrieved.UserID != "test-user" {
			t.Fatalf("Expected user 'test-user', got '%s'", retrieved.UserID)
		}

		t.Logf("✓ Session created and retrieved: ID=%d", txID)
	})

	t.Run("Update connector status", func(t *testing.T) {
		db.UpdateConnectorStatus("my-test", 1, types.StatusCharging)

		connector := db.GetConnector("my-test", 1)
		if connector == nil {
			t.Fatal("Connector not found")
		}

		if connector.Status != types.StatusCharging {
			t.Fatalf("Expected status 'Charging', got '%s'", connector.Status)
		}

		t.Logf("✓ Connector status updated to: %s", connector.Status)
	})
}

// TestEnergyCalculations tests energy consumption calculations
func TestEnergyCalculations(t *testing.T) {
	db := database.NewMockDatabase()

	// Create a session
	session := &types.ChargingSession{
		ChargePointID:     "my-test",
		ConnectorID:       1,
		UserID:            "test-user",
		StartTime:         time.Now().Add(-30 * time.Minute),
		StartMeterValue:   1000, // 1000 Wh
		CurrentMeterValue: 5000, // 5000 Wh
		Status:            "active",
	}

	txID := db.CreateSession(session)

	// Add energy readings
	readings := []*types.EnergyReading{
		{
			Timestamp:     time.Now().Add(-20 * time.Minute),
			TransactionID: txID,
			EnergyWh:      2000,
			PowerW:        floatPtr(7200),
		},
		{
			Timestamp:     time.Now().Add(-10 * time.Minute),
			TransactionID: txID,
			EnergyWh:      3500,
			PowerW:        floatPtr(7200),
		},
		{
			Timestamp:     time.Now(),
			TransactionID: txID,
			EnergyWh:      5000,
			PowerW:        floatPtr(7200),
		},
	}

	for _, reading := range readings {
		db.AddEnergyReading(reading)
	}

	// Get latest reading
	latest := db.GetLatestEnergyReading(txID)
	if latest == nil {
		t.Fatal("No energy readings found")
	}

	expectedEnergy := 5000.0
	if latest.EnergyWh != expectedEnergy {
		t.Fatalf("Expected energy %.0f Wh, got %.0f Wh", expectedEnergy, latest.EnergyWh)
	}

	// Check total consumption
	updatedSession := db.GetSession(txID)
	energyConsumed := updatedSession.CurrentMeterValue - updatedSession.StartMeterValue

	expectedConsumed := 4000 // 5000 - 1000
	if energyConsumed != expectedConsumed {
		t.Fatalf("Expected %d Wh consumed, got %d Wh", expectedConsumed, energyConsumed)
	}

	t.Logf("✓ Energy calculations correct")
	t.Logf("  Start: %d Wh", updatedSession.StartMeterValue)
	t.Logf("  Current: %d Wh", updatedSession.CurrentMeterValue)
	t.Logf("  Consumed: %d Wh (%.3f kWh)", energyConsumed, float64(energyConsumed)/1000)
}

func floatPtr(f float64) *float64 {
	return &f
}

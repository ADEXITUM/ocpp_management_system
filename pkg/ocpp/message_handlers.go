package ocpp

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
)

// MessageHandlers handles incoming OCPP messages from charge points
type MessageHandlers struct {
	db *database.MockDatabase
}

// NewMessageHandlers creates a new message handlers instance
func NewMessageHandlers(db *database.MockDatabase) *MessageHandlers {
	return &MessageHandlers{db: db}
}

// HandleBootNotification processes BootNotification from charge point
func (h *MessageHandlers) HandleBootNotification(
	chargePointID string,
	req types.BootNotificationRequest,
) types.BootNotificationResponse {
	log.Printf("[BootNotification] Charge point %s is booting...", chargePointID)

	existingCP := h.db.GetChargePoint(chargePointID)

	if existingCP == nil {
		// Charge point not in database - print warning
		log.Printf("⚠️  WARNING: Charge point '%s' is NOT CONFIGURED!", chargePointID)
		log.Printf("   Please configure this charge point in the database before use.")
		log.Printf("   Vendor: %s, Model: %s", req.ChargePointVendor, req.ChargePointModel)

		// Create temporarily but mark as pending
		newCP := &types.ChargePoint{
			ID:                 chargePointID,
			Name:               fmt.Sprintf("Unconfigured - %s", chargePointID),
			Vendor:             req.ChargePointVendor,
			Model:              req.ChargePointModel,
			SerialNumber:       getStringOrEmpty(req.ChargePointSerialNumber),
			FirmwareVersion:    getStringOrEmpty(req.FirmwareVersion),
			NumberOfConnectors: 1,
			Status:             "online",
			RegistrationStatus: "pending",
			LastSeen:           time.Now(),
			CreatedAt:          time.Now(),
		}
		h.db.UpsertChargePoint(newCP)
		existingCP = newCP
	} else {
		// Update existing charge point
		existingCP.Vendor = req.ChargePointVendor
		existingCP.Model = req.ChargePointModel
		existingCP.SerialNumber = getStringOrEmpty(req.ChargePointSerialNumber)
		existingCP.FirmwareVersion = getStringOrEmpty(req.FirmwareVersion)
		existingCP.Status = "online"
		existingCP.LastSeen = time.Now()
		h.db.UpsertChargePoint(existingCP)

		log.Printf("[BootNotification] %s registered successfully", chargePointID)
	}

	status := types.RegistrationPending
	if existingCP.RegistrationStatus == "accepted" {
		status = types.RegistrationAccepted
	}

	return types.BootNotificationResponse{
		Status:      status,
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
		Interval:    300, // Heartbeat every 5 minutes
	}
}

// HandleHeartbeat processes Heartbeat from charge point
func (h *MessageHandlers) HandleHeartbeat(chargePointID string) types.HeartbeatResponse {
	h.db.UpdateChargePointStatus(chargePointID, "online")
	return types.HeartbeatResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
	}
}

// HandleStatusNotification processes StatusNotification from charge point
func (h *MessageHandlers) HandleStatusNotification(
	chargePointID string,
	req types.StatusNotificationRequest,
) types.StatusNotificationResponse {
	log.Printf("[StatusNotification] %s Connector %d: %s (error: %s)",
		chargePointID, req.ConnectorID, req.Status, req.ErrorCode)

	h.db.UpdateConnectorStatus(chargePointID, req.ConnectorID, req.Status)

	// Update charge point connector count if needed
	cp := h.db.GetChargePoint(chargePointID)
	if cp != nil {
		connectors := h.db.GetConnectorsByChargePoint(chargePointID)
		maxConnectorID := 0
		for _, conn := range connectors {
			if conn.ConnectorID > maxConnectorID {
				maxConnectorID = conn.ConnectorID
			}
		}
		if maxConnectorID > cp.NumberOfConnectors {
			cp.NumberOfConnectors = maxConnectorID
			h.db.UpsertChargePoint(cp)
		}
	}

	return types.StatusNotificationResponse{}
}

// HandleMeterValues processes MeterValues from charge point
func (h *MessageHandlers) HandleMeterValues(
	chargePointID string,
	req types.MeterValuesRequest,
) types.MeterValuesResponse {
	if req.TransactionID == nil {
		// No transaction, just periodic meter values
		return types.MeterValuesResponse{}
	}

	transactionID := *req.TransactionID

	// Process each meter value
	for _, mv := range req.MeterValue {
		timestamp, err := time.Parse(time.RFC3339, mv.Timestamp)
		if err != nil {
			timestamp = time.Now()
		}

		var energyWh, powerW, currentA, voltageV *float64

		// Extract values from sampled data
		for _, sample := range mv.SampledValue {
			value, err := strconv.ParseFloat(sample.Value, 64)
			if err != nil {
				continue
			}

			measurand := getStringOrEmpty(sample.Measurand)
			unit := getStringOrEmpty(sample.Unit)

			if measurand == "Energy.Active.Import.Register" {
				if unit == "kWh" {
					val := value * 1000
					energyWh = &val
				} else {
					energyWh = &value
				}
			} else if measurand == "Power.Active.Import" {
				if unit == "kW" {
					val := value * 1000
					powerW = &val
				} else {
					powerW = &value
				}
			} else if measurand == "Current.Import" {
				currentA = &value
			} else if measurand == "Voltage" {
				voltageV = &value
			}
		}

		// Store energy reading if we have energy data
		if energyWh != nil {
			reading := &types.EnergyReading{
				Timestamp:     timestamp,
				TransactionID: transactionID,
				EnergyWh:      *energyWh,
				PowerW:        powerW,
				CurrentA:      currentA,
				VoltageV:      voltageV,
			}
			h.db.AddEnergyReading(reading)

			powerStr := ""
			if powerW != nil {
				powerStr = fmt.Sprintf(", %.0f W", *powerW)
			}
			log.Printf("[MeterValues] %s Transaction %d: %.0f Wh%s",
				chargePointID, transactionID, *energyWh, powerStr)
		}
	}

	return types.MeterValuesResponse{}
}

// HandleStartTransaction processes StartTransaction from charge point
func (h *MessageHandlers) HandleStartTransaction(
	chargePointID string,
	req types.StartTransactionRequest,
) types.StartTransactionResponse {
	log.Printf("[StartTransaction] %s Connector %d: User %s, Meter: %d Wh",
		chargePointID, req.ConnectorID, req.IDTag, req.MeterStart)

	// Check for existing session
	existingSession := h.db.GetActiveSessionByConnector(chargePointID, req.ConnectorID)
	if existingSession != nil {
		log.Printf("[StartTransaction] WARNING: Connector already has active session %d",
			existingSession.TransactionID)
		h.db.CompleteSession(existingSession.TransactionID, req.MeterStart, "Replaced")
	}

	// Create new session
	startTime, _ := time.Parse(time.RFC3339, req.Timestamp)
	session := &types.ChargingSession{
		ChargePointID:     chargePointID,
		ConnectorID:       req.ConnectorID,
		UserID:            req.IDTag,
		StartTime:         startTime,
		StartMeterValue:   req.MeterStart,
		CurrentMeterValue: req.MeterStart,
		Status:            "active",
	}

	transactionID := h.db.CreateSession(session)

	// Update connector
	h.db.SetConnectorTransaction(chargePointID, req.ConnectorID, &transactionID)
	h.db.UpdateConnectorStatus(chargePointID, req.ConnectorID, types.StatusCharging)

	log.Printf("[StartTransaction] Created transaction %d", transactionID)

	return types.StartTransactionResponse{
		TransactionID: transactionID,
		IDTagInfo: types.IDTagInfo{
			Status: types.AuthStatusAccepted,
		},
	}
}

// HandleStopTransaction processes StopTransaction from charge point
func (h *MessageHandlers) HandleStopTransaction(
	chargePointID string,
	req types.StopTransactionRequest,
) types.StopTransactionResponse {
	reason := getStringOrEmpty(req.Reason)
	if reason == "" {
		reason = "User"
	}

	log.Printf("[StopTransaction] Transaction %d: Meter: %d Wh, Reason: %s",
		req.TransactionID, req.MeterStop, reason)

	session := h.db.GetSession(req.TransactionID)
	if session != nil {
		// Complete the session
		h.db.CompleteSession(req.TransactionID, req.MeterStop, reason)

		// Clear connector transaction
		h.db.SetConnectorTransaction(session.ChargePointID, session.ConnectorID, nil)
		h.db.UpdateConnectorStatus(session.ChargePointID, session.ConnectorID, types.StatusAvailable)

		energyConsumed := req.MeterStop - session.StartMeterValue
		duration := time.Since(session.StartTime).Seconds()

		log.Printf("[StopTransaction] Session completed: %d Wh consumed in %.0fs",
			energyConsumed, duration)

		// Process transaction data if present
		if len(req.TransactionData) > 0 {
			h.HandleMeterValues(chargePointID, types.MeterValuesRequest{
				ConnectorID:   session.ConnectorID,
				TransactionID: &req.TransactionID,
				MeterValue:    req.TransactionData,
			})
		}
	} else {
		log.Printf("[StopTransaction] WARNING: Transaction %d not found", req.TransactionID)
	}

	return types.StopTransactionResponse{
		IDTagInfo: &types.IDTagInfo{
			Status: types.AuthStatusAccepted,
		},
	}
}

// HandleAuthorize processes Authorize from charge point
func (h *MessageHandlers) HandleAuthorize(
	chargePointID string,
	req types.AuthorizeRequest,
) types.AuthorizeResponse {
	log.Printf("[Authorize] %s: Checking authorization for %s", chargePointID, req.IDTag)

	// In a real system, check against user database
	// For now, block users starting with "BLOCKED"
	status := types.AuthStatusAccepted
	if len(req.IDTag) >= 7 && req.IDTag[:7] == "BLOCKED" {
		status = types.AuthStatusBlocked
	}

	return types.AuthorizeResponse{
		IDTagInfo: types.IDTagInfo{
			Status: status,
		},
	}
}

// Helper function
func getStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

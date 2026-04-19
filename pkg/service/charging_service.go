package service

import (
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/errors"
	"github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
)

// ChargingService provides simple API for controlling charging sessions
type ChargingService struct {
	db                *database.MockDatabase
	connectionManager *ocpp.ConnectionManager
}

// NewChargingService creates a new charging service
func NewChargingService(db *database.MockDatabase, cm *ocpp.ConnectionManager) *ChargingService {
	return &ChargingService{
		db:                db,
		connectionManager: cm,
	}
}

// TurnOn starts a charging session
func (s *ChargingService) TurnOn(
	chargePointID string,
	connectorID int,
	userID string,
) (*types.TurnOnResult, error) {
	// 1. Check if charge point exists in database
	chargePoint := s.db.GetChargePoint(chargePointID)
	if chargePoint == nil {
		return nil, errors.NewChargePointNotFoundError(chargePointID)
	}

	// 2. Check if charge point is configured
	if chargePoint.RegistrationStatus != "accepted" {
		return nil, errors.NewChargePointNotConfiguredError(chargePointID)
	}

	// 3. Check if charge point is connected
	if !s.connectionManager.IsConnected(chargePointID) {
		return nil, errors.NewChargePointOfflineError(chargePointID)
	}

	// 4. Check if connector exists
	connector := s.db.GetConnector(chargePointID, connectorID)
	if connector == nil {
		return nil, errors.NewConnectorNotFoundError(chargePointID, connectorID)
	}

	// 5. Check connector status
	if connector.Status == types.StatusFaulted {
		return nil, errors.NewConnectorUnavailableError(chargePointID, connectorID, "Faulted")
	}

	if connector.Status == types.StatusUnavailable {
		return nil, errors.NewConnectorUnavailableError(chargePointID, connectorID, "Unavailable")
	}

	// 6. Check if connector already has an active session
	existingSession := s.db.GetActiveSessionByConnector(chargePointID, connectorID)
	if existingSession != nil {
		return nil, errors.NewSessionAlreadyActiveError(chargePointID, connectorID)
	}

	// 7. Send RemoteStartTransaction to charge point
	response, err := s.connectionManager.RemoteStartTransaction(
		chargePointID,
		userID,
		&connectorID,
	)

	if err != nil {
		if isTimeoutError(err) {
			return nil, errors.NewOperationTimeoutError("RemoteStartTransaction", chargePointID)
		}
		return nil, err
	}

	if response.Status == types.RemoteStartStopRejected {
		return nil, errors.NewRemoteOperationFailedError("start", chargePointID, "Rejected by charge point")
	}

	// 8. Wait for the StartTransaction message from the charge point.
	// Real devices/testers can take longer between RemoteStart acceptance and StartTransaction.
	session, err := s.waitForSessionStart(chargePointID, connectorID, 90*time.Second)
	if err != nil {
		return nil, err
	}

	return &types.TurnOnResult{
		Success:       true,
		TransactionID: session.TransactionID,
		ChargePointID: chargePointID,
		ConnectorID:   connectorID,
		UserID:        userID,
		Message:       "Charging session started successfully",
	}, nil
}

// TurnOff stops a charging session
func (s *ChargingService) TurnOff(
	chargePointID string,
	transactionID int,
) (*types.TurnOffResult, error) {
	// 1. Check if session exists
	session := s.db.GetSession(transactionID)
	if session == nil {
		return nil, errors.NewSessionNotFoundError(transactionID)
	}

	// Verify it's the right charge point
	if session.ChargePointID != chargePointID {
		return nil, errors.NewSessionNotFoundError(transactionID)
	}

	// 2. Check if session is still active
	if session.Status != "active" {
		// Session already completed, return the final data
		energyConsumed := float64(getEndMeterValue(session) - session.StartMeterValue)
		duration := getDuration(session)

		return &types.TurnOffResult{
			Success:        true,
			TransactionID:  transactionID,
			ChargePointID:  chargePointID,
			EnergyConsumed: energyConsumed,
			Duration:       duration,
			Message:        "Session was already stopped",
		}, nil
	}

	// 3. Check if charge point is connected
	if !s.connectionManager.IsConnected(chargePointID) {
		return nil, errors.NewChargePointOfflineError(chargePointID)
	}

	// 4. Send RemoteStopTransaction to charge point
	response, err := s.connectionManager.RemoteStopTransaction(chargePointID, transactionID)
	if err != nil {
		if isTimeoutError(err) {
			return nil, errors.NewOperationTimeoutError("RemoteStopTransaction", chargePointID)
		}
		return nil, err
	}

	if response.Status == types.RemoteStartStopRejected {
		return nil, errors.NewRemoteOperationFailedError("stop", chargePointID, "Rejected by charge point")
	}

	// 5. Wait for the StopTransaction message from the charge point
	s.waitForSessionStop(transactionID, 15*time.Second)

	// 6. Get final session data
	updatedSession := s.db.GetSession(transactionID)
	if updatedSession == nil {
		return nil, errors.NewSessionNotFoundError(transactionID)
	}

	energyConsumed := float64(getEndMeterValue(updatedSession) - updatedSession.StartMeterValue)
	duration := getDuration(updatedSession)

	return &types.TurnOffResult{
		Success:        true,
		TransactionID:  transactionID,
		ChargePointID:  chargePointID,
		EnergyConsumed: energyConsumed,
		Duration:       duration,
		Message:        "Charging session stopped successfully",
	}, nil
}

// GetEnergyConsumption gets current energy consumption for a session
func (s *ChargingService) GetEnergyConsumption(
	chargePointID string,
	transactionID int,
) (*types.EnergyConsumptionResult, error) {
	// 1. Get session
	session := s.db.GetSession(transactionID)
	if session == nil {
		return nil, errors.NewSessionNotFoundError(transactionID)
	}

	// Verify it's the right charge point
	if session.ChargePointID != chargePointID {
		return nil, errors.NewSessionNotFoundError(transactionID)
	}

	// 2. Calculate current energy consumption
	currentMeterValue := session.CurrentMeterValue
	if session.Status == "completed" && session.EndMeterValue != nil {
		currentMeterValue = *session.EndMeterValue
	}

	energyConsumed := float64(currentMeterValue - session.StartMeterValue)
	durationSeconds := getDuration(session)

	// 3. Get latest energy reading for additional details
	latestReading := s.db.GetLatestEnergyReading(transactionID)
	lastUpdate := session.StartTime
	if latestReading != nil {
		lastUpdate = latestReading.Timestamp
	}

	status := "active"
	if session.Status == "completed" {
		status = "completed"
	}

	return &types.EnergyConsumptionResult{
		TransactionID:   transactionID,
		ChargePointID:   session.ChargePointID,
		ConnectorID:     session.ConnectorID,
		UserID:          session.UserID,
		StartTime:       session.StartTime,
		CurrentEnergyWh: energyConsumed,
		DurationSeconds: durationSeconds,
		Status:          status,
		LastUpdate:      lastUpdate,
	}, nil
}

// GetActiveSessions returns all active charging sessions
func (s *ChargingService) GetActiveSessions() []*types.EnergyConsumptionResult {
	sessions := s.db.GetAllActiveSessions()
	results := make([]*types.EnergyConsumptionResult, 0, len(sessions))

	for _, session := range sessions {
		energyConsumed := float64(session.CurrentMeterValue - session.StartMeterValue)
		durationSeconds := int(time.Since(session.StartTime).Seconds())

		latestReading := s.db.GetLatestEnergyReading(session.TransactionID)
		lastUpdate := session.StartTime
		if latestReading != nil {
			lastUpdate = latestReading.Timestamp
		}

		results = append(results, &types.EnergyConsumptionResult{
			TransactionID:   session.TransactionID,
			ChargePointID:   session.ChargePointID,
			ConnectorID:     session.ConnectorID,
			UserID:          session.UserID,
			StartTime:       session.StartTime,
			CurrentEnergyWh: energyConsumed,
			DurationSeconds: durationSeconds,
			Status:          "active",
			LastUpdate:      lastUpdate,
		})
	}

	return results
}

// Helper functions

func (s *ChargingService) waitForSessionStart(
	chargePointID string,
	connectorID int,
	timeout time.Duration,
) (*types.ChargingSession, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		session := s.db.GetActiveSessionByConnector(chargePointID, connectorID)
		if session != nil {
			return session, nil
		}
		time.Sleep(pollInterval)
	}

	return nil, errors.NewRemoteOperationFailedError(
		"start",
		chargePointID,
		"Session not created after remote start",
	)
}

func (s *ChargingService) waitForSessionStop(transactionID int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		session := s.db.GetSession(transactionID)
		if session != nil && session.Status == "completed" {
			return
		}
		time.Sleep(pollInterval)
	}
}

func getEndMeterValue(session *types.ChargingSession) int {
	if session.EndMeterValue != nil {
		return *session.EndMeterValue
	}
	return session.CurrentMeterValue
}

func getDuration(session *types.ChargingSession) int {
	if session.EndTime != nil {
		return int(session.EndTime.Sub(session.StartTime).Seconds())
	}
	return int(time.Since(session.StartTime).Seconds())
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "request timeout" ||
		(len(err.Error()) > 15 && err.Error()[:15] == "request timeout")
}

package database

import (
	"sync"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
)

// MockDatabase is an in-memory database for charge points and sessions
type MockDatabase struct {
	mu                   sync.RWMutex
	chargePoints         map[string]*types.ChargePoint
	connectors           map[string]*types.Connector // key: "cpId:connectorId"
	sessions             map[int]*types.ChargingSession
	energyReadings       map[int][]*types.EnergyReading // key: transactionId
	transactionIDCounter int
}

// NewMockDatabase creates a new mock database with sample data
func NewMockDatabase() *MockDatabase {
	db := &MockDatabase{
		chargePoints:         make(map[string]*types.ChargePoint),
		connectors:           make(map[string]*types.Connector),
		sessions:             make(map[int]*types.ChargingSession),
		energyReadings:       make(map[int][]*types.EnergyReading),
		transactionIDCounter: 1,
	}
	db.initializeMockData()
	return db
}

func (db *MockDatabase) initializeMockData() {
	now := time.Now()

	// Sample Charge Point 1
	cp1 := &types.ChargePoint{
		ID:                 "CP001",
		Name:               "Main Street Station 1",
		Vendor:             "EVBox",
		Model:              "Elvi",
		SerialNumber:       "EVB-001-2024",
		FirmwareVersion:    "1.0.0",
		NumberOfConnectors: 2,
		Status:             "offline",
		RegistrationStatus: "accepted",
		LastSeen:           now,
		CreatedAt:          now,
	}
	db.chargePoints[cp1.ID] = cp1

	for i := 1; i <= cp1.NumberOfConnectors; i++ {
		connector := &types.Connector{
			ChargePointID:    cp1.ID,
			ConnectorID:      i,
			Status:           types.StatusAvailable,
			LastStatusUpdate: now,
		}
		db.connectors[connectorKey(cp1.ID, i)] = connector
	}

	// Sample Charge Point 2
	cp2 := &types.ChargePoint{
		ID:                 "CP002",
		Name:               "Shopping Mall Station",
		Vendor:             "ABB",
		Model:              "Terra AC",
		SerialNumber:       "ABB-002-2024",
		FirmwareVersion:    "2.1.3",
		NumberOfConnectors: 1,
		Status:             "offline",
		RegistrationStatus: "accepted",
		LastSeen:           now,
		CreatedAt:          now,
	}
	db.chargePoints[cp2.ID] = cp2

	connector2 := &types.Connector{
		ChargePointID:    cp2.ID,
		ConnectorID:      1,
		Status:           types.StatusAvailable,
		LastStatusUpdate: now,
	}
	db.connectors[connectorKey(cp2.ID, 1)] = connector2

	// Sample Charge Point 3
	cp3 := &types.ChargePoint{
		ID:                 "CP003",
		Name:               "Office Parking Charger",
		Vendor:             "ChargePoint",
		Model:              "CPE250",
		NumberOfConnectors: 2,
		Status:             "offline",
		RegistrationStatus: "accepted",
		LastSeen:           now,
		CreatedAt:          now,
	}
	db.chargePoints[cp3.ID] = cp3

	for i := 1; i <= cp3.NumberOfConnectors; i++ {
		connector := &types.Connector{
			ChargePointID:    cp3.ID,
			ConnectorID:      i,
			Status:           types.StatusAvailable,
			LastStatusUpdate: now,
		}
		db.connectors[connectorKey(cp3.ID, i)] = connector
	}
}

func connectorKey(chargePointID string, connectorID int) string {
	return chargePointID + ":" + string(rune(connectorID+'0'))
}

// ChargePoint methods

func (db *MockDatabase) GetChargePoint(id string) *types.ChargePoint {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.chargePoints[id]
}

func (db *MockDatabase) GetAllChargePoints() []*types.ChargePoint {
	db.mu.RLock()
	defer db.mu.RUnlock()

	chargePoints := make([]*types.ChargePoint, 0, len(db.chargePoints))
	for _, cp := range db.chargePoints {
		chargePoints = append(chargePoints, cp)
	}
	return chargePoints
}

func (db *MockDatabase) UpsertChargePoint(cp *types.ChargePoint) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.chargePoints[cp.ID] = cp
}

func (db *MockDatabase) UpdateChargePointStatus(id, status string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if cp, ok := db.chargePoints[id]; ok {
		cp.Status = status
		cp.LastSeen = time.Now()
	}
}

// Connector methods

func (db *MockDatabase) GetConnector(chargePointID string, connectorID int) *types.Connector {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.connectors[connectorKey(chargePointID, connectorID)]
}

func (db *MockDatabase) GetConnectorsByChargePoint(chargePointID string) []*types.Connector {
	db.mu.RLock()
	defer db.mu.RUnlock()

	connectors := make([]*types.Connector, 0)
	for _, connector := range db.connectors {
		if connector.ChargePointID == chargePointID {
			connectors = append(connectors, connector)
		}
	}
	return connectors
}

func (db *MockDatabase) UpdateConnectorStatus(chargePointID string, connectorID int, status string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	key := connectorKey(chargePointID, connectorID)
	if connector, ok := db.connectors[key]; ok {
		connector.Status = status
		connector.LastStatusUpdate = time.Now()
	} else {
		// Create connector if it doesn't exist
		db.connectors[key] = &types.Connector{
			ChargePointID:    chargePointID,
			ConnectorID:      connectorID,
			Status:           status,
			LastStatusUpdate: time.Now(),
		}
	}
}

func (db *MockDatabase) SetConnectorTransaction(chargePointID string, connectorID int, transactionID *int) {
	db.mu.Lock()
	defer db.mu.Unlock()

	key := connectorKey(chargePointID, connectorID)
	if connector, ok := db.connectors[key]; ok {
		connector.CurrentTransaction = transactionID
	}
}

// Session methods

func (db *MockDatabase) CreateSession(session *types.ChargingSession) int {
	db.mu.Lock()
	defer db.mu.Unlock()

	transactionID := db.transactionIDCounter
	db.transactionIDCounter++

	session.TransactionID = transactionID
	db.sessions[transactionID] = session
	db.energyReadings[transactionID] = make([]*types.EnergyReading, 0)

	return transactionID
}

func (db *MockDatabase) GetSession(transactionID int) *types.ChargingSession {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.sessions[transactionID]
}

func (db *MockDatabase) GetActiveSessionByConnector(chargePointID string, connectorID int) *types.ChargingSession {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, session := range db.sessions {
		if session.ChargePointID == chargePointID &&
			session.ConnectorID == connectorID &&
			session.Status == "active" {
			return session
		}
	}
	return nil
}

func (db *MockDatabase) GetAllActiveSessions() []*types.ChargingSession {
	db.mu.RLock()
	defer db.mu.RUnlock()

	sessions := make([]*types.ChargingSession, 0)
	for _, session := range db.sessions {
		if session.Status == "active" {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

func (db *MockDatabase) UpdateSessionMeterValue(transactionID, meterValue int) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if session, ok := db.sessions[transactionID]; ok {
		session.CurrentMeterValue = meterValue
	}
}

func (db *MockDatabase) CompleteSession(transactionID, endMeterValue int, stopReason string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if session, ok := db.sessions[transactionID]; ok {
		session.Status = "completed"
		now := time.Now()
		session.EndTime = &now
		session.EndMeterValue = &endMeterValue
		session.StopReason = stopReason
	}
}

// Energy Reading methods

func (db *MockDatabase) AddEnergyReading(reading *types.EnergyReading) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.energyReadings[reading.TransactionID] = append(
		db.energyReadings[reading.TransactionID],
		reading,
	)

	// Update session's current meter value
	if session, ok := db.sessions[reading.TransactionID]; ok {
		session.CurrentMeterValue = int(reading.EnergyWh)
	}
}

func (db *MockDatabase) GetEnergyReadings(transactionID int) []*types.EnergyReading {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.energyReadings[transactionID]
}

func (db *MockDatabase) GetLatestEnergyReading(transactionID int) *types.EnergyReading {
	db.mu.RLock()
	defer db.mu.RUnlock()

	readings := db.energyReadings[transactionID]
	if len(readings) == 0 {
		return nil
	}
	return readings[len(readings)-1]
}

// Stats

type Stats struct {
	TotalChargePoints  int
	OnlineChargePoints int
	TotalConnectors    int
	ActiveSessions     int
	TotalSessions      int
}

func (db *MockDatabase) GetStats() Stats {
	db.mu.RLock()
	defer db.mu.RUnlock()

	onlineCount := 0
	for _, cp := range db.chargePoints {
		if cp.Status == "online" {
			onlineCount++
		}
	}

	activeCount := 0
	for _, session := range db.sessions {
		if session.Status == "active" {
			activeCount++
		}
	}

	return Stats{
		TotalChargePoints:  len(db.chargePoints),
		OnlineChargePoints: onlineCount,
		TotalConnectors:    len(db.connectors),
		ActiveSessions:     activeCount,
		TotalSessions:      len(db.sessions),
	}
}

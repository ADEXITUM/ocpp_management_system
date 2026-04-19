package database

import (
	"fmt"
	"strconv"
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
	events               []*TimelineEvent
	transactionIDCounter int
	eventIDCounter       int
}

// TimelineEvent stores operational history for the learning dashboard.
type TimelineEvent struct {
	ID            int                    `json:"id"`
	Timestamp     string                 `json:"timestamp"`
	Source        string                 `json:"source"`
	Level         string                 `json:"level"`
	Action        string                 `json:"action"`
	ChargePointID string                 `json:"chargePointId,omitempty"`
	ConnectorID   *int                   `json:"connectorId,omitempty"`
	TransactionID *int                   `json:"transactionId,omitempty"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// NewMockDatabase creates a new mock database with sample data
func NewMockDatabase() *MockDatabase {
	db := &MockDatabase{
		chargePoints:         make(map[string]*types.ChargePoint),
		connectors:           make(map[string]*types.Connector),
		sessions:             make(map[int]*types.ChargingSession),
		energyReadings:       make(map[int][]*types.EnergyReading),
		events:               make([]*TimelineEvent, 0, 256),
		transactionIDCounter: 1,
		eventIDCounter:       1,
	}
	db.initializeMockData()
	return db
}

func (db *MockDatabase) initializeMockData() {
	now := time.Now()

	// Single preconfigured charge point for focused station testing.
	cp := &types.ChargePoint{
		ID:                 "my-test",
		Name:               "Local Test Charge Point",
		Vendor:             "ETEK",
		Model:              "TestModel",
		NumberOfConnectors: 0,
		Status:             "offline",
		RegistrationStatus: "accepted",
		LastSeen:           now,
		CreatedAt:          now,
	}
	db.chargePoints[cp.ID] = cp

	db.AddEvent(
		"system",
		"info",
		"DatabaseInitialized",
		"Mock database initialized with only my-test charge point (connectors are discovered dynamically)",
		"my-test",
		nil,
		nil,
		map[string]interface{}{
			"numberOfConnectors": 0,
			"registrationStatus": "accepted",
		},
	)
}

func connectorKey(chargePointID string, connectorID int) string {
	return chargePointID + ":" + strconv.Itoa(connectorID)
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
	// connectorId=0 is a charge-point level status in OCPP 1.6, not a physical connector.
	if connectorID <= 0 {
		return
	}

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
	if connectorID <= 0 {
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	key := connectorKey(chargePointID, connectorID)
	if connector, ok := db.connectors[key]; ok {
		connector.CurrentTransaction = transactionID
		return
	}

	db.connectors[key] = &types.Connector{
		ChargePointID:      chargePointID,
		ConnectorID:        connectorID,
		Status:             "Unknown",
		CurrentTransaction: transactionID,
		LastStatusUpdate:   time.Now(),
	}
}

func (db *MockDatabase) SyncConnectorCount(chargePointID string, count int) (int, int, error) {
	if count < 0 {
		return 0, 0, fmt.Errorf("connector count cannot be negative")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	cp, ok := db.chargePoints[chargePointID]
	if !ok {
		return 0, 0, fmt.Errorf("charge point %s not found", chargePointID)
	}

	created := 0
	removed := 0
	now := time.Now()

	for id := 1; id <= count; id++ {
		key := connectorKey(chargePointID, id)
		if _, exists := db.connectors[key]; !exists {
			db.connectors[key] = &types.Connector{
				ChargePointID:    chargePointID,
				ConnectorID:      id,
				Status:           "Unknown",
				LastStatusUpdate: now,
			}
			created++
		}
	}

	for key, connector := range db.connectors {
		if connector.ChargePointID == chargePointID && connector.ConnectorID > count {
			delete(db.connectors, key)
			removed++
		}
	}

	cp.NumberOfConnectors = count
	cp.LastSeen = now
	db.chargePoints[chargePointID] = cp

	return created, removed, nil
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

func (db *MockDatabase) AddEvent(
	source, level, action, message, chargePointID string,
	connectorID, transactionID *int,
	details map[string]interface{},
) {
	db.mu.Lock()
	defer db.mu.Unlock()

	event := &TimelineEvent{
		ID:            db.eventIDCounter,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Source:        source,
		Level:         level,
		Action:        action,
		ChargePointID: chargePointID,
		ConnectorID:   cloneIntPtr(connectorID),
		TransactionID: cloneIntPtr(transactionID),
		Message:       message,
		Details:       cloneDetails(details),
	}
	db.eventIDCounter++
	db.events = append(db.events, event)

	// Keep memory bounded.
	const maxEvents = 2000
	if len(db.events) > maxEvents {
		db.events = append([]*TimelineEvent(nil), db.events[len(db.events)-maxEvents:]...)
	}
}

func (db *MockDatabase) GetRecentEvents(limit int) []TimelineEvent {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}
	if limit > len(db.events) {
		limit = len(db.events)
	}

	out := make([]TimelineEvent, 0, limit)
	start := len(db.events) - limit
	for i := len(db.events) - 1; i >= start; i-- {
		ev := db.events[i]
		out = append(out, TimelineEvent{
			ID:            ev.ID,
			Timestamp:     ev.Timestamp,
			Source:        ev.Source,
			Level:         ev.Level,
			Action:        ev.Action,
			ChargePointID: ev.ChargePointID,
			ConnectorID:   cloneIntPtr(ev.ConnectorID),
			TransactionID: cloneIntPtr(ev.TransactionID),
			Message:       ev.Message,
			Details:       cloneDetails(ev.Details),
		})
	}
	return out
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
	TotalEvents        int
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
		TotalEvents:        len(db.events),
	}
}

func intPtr(v int) *int {
	return &v
}

func cloneIntPtr(v *int) *int {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneDetails(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

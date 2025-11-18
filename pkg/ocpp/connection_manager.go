package ocpp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const requestTimeout = 30 * time.Second

type pendingRequest struct {
	responseChan chan interface{}
	errorChan    chan error
}

// ConnectionManager manages WebSocket connections to charge points
type ConnectionManager struct {
	mu              sync.RWMutex
	connections     map[string]*websocket.Conn
	pendingRequests map[string]*pendingRequest
	handlers        *MessageHandlers
	db              *database.MockDatabase
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(db *database.MockDatabase) *ConnectionManager {
	return &ConnectionManager{
		connections:     make(map[string]*websocket.Conn),
		pendingRequests: make(map[string]*pendingRequest),
		handlers:        NewMessageHandlers(db),
		db:              db,
	}
}

// RegisterConnection registers a new charge point connection
func (cm *ConnectionManager) RegisterConnection(chargePointID string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close existing connection if any
	if existingConn, ok := cm.connections[chargePointID]; ok {
		log.Printf("[ConnectionManager] Closing existing connection for %s", chargePointID)
		existingConn.Close()
	}

	cm.connections[chargePointID] = conn
	cm.db.UpdateChargePointStatus(chargePointID, "online")

	log.Printf("[ConnectionManager] Registered connection for %s", chargePointID)
	log.Printf("[ConnectionManager] Total connections: %d", len(cm.connections))

	// Start message handler goroutine
	go cm.handleMessages(chargePointID, conn)
}

// handleMessages processes incoming messages from a charge point
func (cm *ConnectionManager) handleMessages(chargePointID string, conn *websocket.Conn) {
	defer func() {
		cm.handleDisconnect(chargePointID)
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ConnectionManager] WebSocket error for %s: %v", chargePointID, err)
			}
			return
		}

		var message types.OCPPMessage
		if err := json.Unmarshal(data, &message); err != nil {
			log.Printf("[ConnectionManager] Failed to parse message from %s: %v", chargePointID, err)
			continue
		}

		if len(message) < 3 {
			log.Printf("[ConnectionManager] Invalid message format from %s", chargePointID)
			continue
		}

		messageType := int(message[0].(float64))

		switch messageType {
		case types.MessageTypeCall:
			cm.handleCall(chargePointID, message)
		case types.MessageTypeCallResult:
			cm.handleCallResult(message)
		case types.MessageTypeCallError:
			cm.handleCallError(message)
		}
	}
}

// handleCall processes a CALL (request) from charge point
func (cm *ConnectionManager) handleCall(chargePointID string, message types.OCPPMessage) {
	if len(message) < 4 {
		return
	}

	messageID := message[1].(string)
	action := message[2].(string)
	payload := message[3]

	log.Printf("[ConnectionManager] <- %s: %s", chargePointID, action)

	var response interface{}
	var err error

	// Marshal payload to JSON and unmarshal to specific type
	payloadJSON, _ := json.Marshal(payload)

	switch action {
	case types.ActionBootNotification:
		var req types.BootNotificationRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleBootNotification(chargePointID, req)

	case types.ActionHeartbeat:
		response = cm.handlers.HandleHeartbeat(chargePointID)

	case types.ActionStatusNotification:
		var req types.StatusNotificationRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleStatusNotification(chargePointID, req)

	case types.ActionMeterValues:
		var req types.MeterValuesRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleMeterValues(chargePointID, req)

	case types.ActionStartTransaction:
		var req types.StartTransactionRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleStartTransaction(chargePointID, req)

	case types.ActionStopTransaction:
		var req types.StopTransactionRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleStopTransaction(chargePointID, req)

	case types.ActionAuthorize:
		var req types.AuthorizeRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleAuthorize(chargePointID, req)

	default:
		err = fmt.Errorf("action %s not implemented", action)
	}

	if err != nil {
		cm.sendCallError(chargePointID, messageID, "NotImplemented", err.Error())
	} else {
		cm.sendCallResult(chargePointID, messageID, response)
	}
}

// handleCallResult processes a CALLRESULT (response to our request)
func (cm *ConnectionManager) handleCallResult(message types.OCPPMessage) {
	if len(message) < 3 {
		return
	}

	messageID := message[1].(string)
	payload := message[2]

	cm.mu.Lock()
	pending, ok := cm.pendingRequests[messageID]
	if ok {
		delete(cm.pendingRequests, messageID)
	}
	cm.mu.Unlock()

	if ok {
		pending.responseChan <- payload
	}
}

// handleCallError processes a CALLERROR (error response to our request)
func (cm *ConnectionManager) handleCallError(message types.OCPPMessage) {
	if len(message) < 4 {
		return
	}

	messageID := message[1].(string)
	errorCode := message[2].(string)
	errorDescription := message[3].(string)

	cm.mu.Lock()
	pending, ok := cm.pendingRequests[messageID]
	if ok {
		delete(cm.pendingRequests, messageID)
	}
	cm.mu.Unlock()

	if ok {
		pending.errorChan <- fmt.Errorf("%s: %s", errorCode, errorDescription)
	}
}

// handleDisconnect handles charge point disconnection
func (cm *ConnectionManager) handleDisconnect(chargePointID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.connections, chargePointID)
	cm.db.UpdateChargePointStatus(chargePointID, "offline")

	log.Printf("[ConnectionManager] %s disconnected", chargePointID)
	log.Printf("[ConnectionManager] Total connections: %d", len(cm.connections))
}

// sendCallResult sends a CALLRESULT to charge point
func (cm *ConnectionManager) sendCallResult(chargePointID, messageID string, payload interface{}) {
	message := types.OCPPMessage{
		types.MessageTypeCallResult,
		messageID,
		payload,
	}
	cm.sendMessage(chargePointID, message)
}

// sendCallError sends a CALLERROR to charge point
func (cm *ConnectionManager) sendCallError(chargePointID, messageID, errorCode, errorDescription string) {
	message := types.OCPPMessage{
		types.MessageTypeCallError,
		messageID,
		errorCode,
		errorDescription,
		map[string]interface{}{},
	}
	cm.sendMessage(chargePointID, message)
}

// SendCall sends a CALL (request) to charge point
func (cm *ConnectionManager) SendCall(chargePointID, action string, payload interface{}) (interface{}, error) {
	messageID := uuid.New().String()

	message := types.OCPPMessage{
		types.MessageTypeCall,
		messageID,
		action,
		payload,
	}

	// Create pending request
	pending := &pendingRequest{
		responseChan: make(chan interface{}, 1),
		errorChan:    make(chan error, 1),
	}

	cm.mu.Lock()
	cm.pendingRequests[messageID] = pending
	cm.mu.Unlock()

	// Send message
	log.Printf("[ConnectionManager] -> %s: %s", chargePointID, action)
	if err := cm.sendMessage(chargePointID, message); err != nil {
		cm.mu.Lock()
		delete(cm.pendingRequests, messageID)
		cm.mu.Unlock()
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-pending.responseChan:
		return response, nil
	case err := <-pending.errorChan:
		return nil, err
	case <-time.After(requestTimeout):
		cm.mu.Lock()
		delete(cm.pendingRequests, messageID)
		cm.mu.Unlock()
		return nil, fmt.Errorf("request timeout: %s to %s", action, chargePointID)
	}
}

// sendMessage sends a message to charge point
func (cm *ConnectionManager) sendMessage(chargePointID string, message types.OCPPMessage) error {
	cm.mu.RLock()
	conn, ok := cm.connections[chargePointID]
	cm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("charge point %s not connected", chargePointID)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// IsConnected checks if charge point is connected
func (cm *ConnectionManager) IsConnected(chargePointID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, ok := cm.connections[chargePointID]
	return ok
}

// GetConnectedChargePoints returns list of connected charge point IDs
func (cm *ConnectionManager) GetConnectedChargePoints() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	chargePoints := make([]string, 0, len(cm.connections))
	for cpID := range cm.connections {
		chargePoints = append(chargePoints, cpID)
	}
	return chargePoints
}

// RemoteStartTransaction sends RemoteStartTransaction to charge point
func (cm *ConnectionManager) RemoteStartTransaction(
	chargePointID, idTag string,
	connectorID *int,
) (*types.RemoteStartTransactionResponse, error) {
	req := types.RemoteStartTransactionRequest{
		IDTag:       idTag,
		ConnectorID: connectorID,
	}

	response, err := cm.SendCall(chargePointID, types.ActionRemoteStartTransaction, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	responseJSON, _ := json.Marshal(response)
	var result types.RemoteStartTransactionResponse
	json.Unmarshal(responseJSON, &result)

	return &result, nil
}

// RemoteStopTransaction sends RemoteStopTransaction to charge point
func (cm *ConnectionManager) RemoteStopTransaction(
	chargePointID string,
	transactionID int,
) (*types.RemoteStopTransactionResponse, error) {
	req := types.RemoteStopTransactionRequest{
		TransactionID: transactionID,
	}

	response, err := cm.SendCall(chargePointID, types.ActionRemoteStopTransaction, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	responseJSON, _ := json.Marshal(response)
	var result types.RemoteStopTransactionResponse
	json.Unmarshal(responseJSON, &result)

	return &result, nil
}

// GetStats returns connection statistics
func (cm *ConnectionManager) GetStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return map[string]interface{}{
		"totalConnections":       len(cm.connections),
		"pendingRequests":        len(cm.pendingRequests),
		"connectedChargePoints": cm.GetConnectedChargePoints(),
	}
}

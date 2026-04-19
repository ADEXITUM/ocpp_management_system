package ocpp

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const requestTimeout = 30 * time.Second

type pendingRequest struct {
	responseChan  chan interface{}
	errorChan     chan error
	action        string
	chargePointID string
	connectorID   *int
	transactionID *int
	sentAt        time.Time
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
	cm.db.AddEvent(
		"ocpp",
		"info",
		"ConnectionOpened",
		"WebSocket connection opened",
		chargePointID,
		nil,
		nil,
		nil,
	)

	log.Printf("[ConnectionManager] Registered connection for %s", chargePointID)
	log.Printf("[ConnectionManager] Total connections: %d", len(cm.connections))

	// Try early connector sync right after WebSocket is established.
	// Some stations respond to GetConfiguration before sending BootNotification.
	go cm.syncConnectorCountFromChargePoint(chargePointID)

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
				cm.db.AddEvent(
					"ocpp",
					"error",
					"WebSocketReadError",
					err.Error(),
					chargePointID,
					nil,
					nil,
					nil,
				)
			}
			return
		}

		var message types.OCPPMessage
		if err := json.Unmarshal(data, &message); err != nil {
			log.Printf("[ConnectionManager] Failed to parse message from %s: %v", chargePointID, err)
			cm.db.AddEvent(
				"ocpp",
				"warn",
				"InvalidJSON",
				"Failed to parse incoming OCPP message",
				chargePointID,
				nil,
				nil,
				map[string]interface{}{"raw": string(data)},
			)
			continue
		}

		if len(message) < 3 {
			log.Printf("[ConnectionManager] Invalid message format from %s", chargePointID)
			cm.db.AddEvent(
				"ocpp",
				"warn",
				"InvalidMessageFormat",
				"Incoming OCPP message has invalid array length",
				chargePointID,
				nil,
				nil,
				map[string]interface{}{"length": len(message)},
			)
			continue
		}

		rawMessageType, ok := message[0].(float64)
		if !ok {
			cm.db.AddEvent(
				"ocpp",
				"warn",
				"InvalidMessageType",
				"Incoming OCPP message type is not numeric",
				chargePointID,
				nil,
				nil,
				nil,
			)
			continue
		}
		messageType := int(rawMessageType)

		switch messageType {
		case types.MessageTypeCall:
			cm.handleCall(chargePointID, message)
		case types.MessageTypeCallResult:
			cm.handleCallResult(chargePointID, message)
		case types.MessageTypeCallError:
			cm.handleCallError(chargePointID, message)
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
	connectorID, transactionID, details := extractPayloadContext(action, payload)
	cm.db.AddEvent(
		"ocpp",
		"info",
		action,
		"Incoming CALL from charge point",
		chargePointID,
		connectorID,
		transactionID,
		details,
	)

	var response interface{}
	var err error
	shouldSyncConnectors := false

	// Marshal payload to JSON and unmarshal to specific type
	payloadJSON, _ := json.Marshal(payload)

	switch action {
	case types.ActionBootNotification:
		var req types.BootNotificationRequest
		json.Unmarshal(payloadJSON, &req)
		response = cm.handlers.HandleBootNotification(chargePointID, req)
		shouldSyncConnectors = true

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
		sendErr := cm.sendCallError(chargePointID, messageID, "NotImplemented", err.Error())
		cm.db.AddEvent(
			"ocpp",
			"warn",
			action,
			"Action not implemented",
			chargePointID,
			connectorID,
			transactionID,
			map[string]interface{}{"error": err.Error()},
		)
		if sendErr != nil {
			cm.db.AddEvent(
				"ocpp",
				"error",
				"CallErrorSendFailed",
				sendErr.Error(),
				chargePointID,
				connectorID,
				transactionID,
				nil,
			)
		}
	} else {
		sendErr := cm.sendCallResult(chargePointID, messageID, response)
		if sendErr != nil {
			cm.db.AddEvent(
				"ocpp",
				"error",
				"CallResultSendFailed",
				sendErr.Error(),
				chargePointID,
				connectorID,
				transactionID,
				nil,
			)
			return
		}
		cm.db.AddEvent(
			"ocpp",
			"info",
			action+"Result",
			"Sent CALLRESULT to charge point",
			chargePointID,
			connectorID,
			transactionID,
			nil,
		)
		if shouldSyncConnectors {
			go cm.syncConnectorCountFromChargePoint(chargePointID)
		}
	}
}

// handleCallResult processes a CALLRESULT (response to our request)
func (cm *ConnectionManager) handleCallResult(chargePointID string, message types.OCPPMessage) {
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
		cm.db.AddEvent(
			"ocpp",
			"info",
			pending.action+"Response",
			"Received CALLRESULT for outbound request",
			pending.chargePointID,
			pending.connectorID,
			pending.transactionID,
			map[string]interface{}{"messageId": messageID},
		)
		return
	}
	cm.db.AddEvent(
		"ocpp",
		"warn",
		"UnexpectedCallResult",
		"Received CALLRESULT without pending request",
		chargePointID,
		nil,
		nil,
		map[string]interface{}{"messageId": messageID},
	)
}

// handleCallError processes a CALLERROR (error response to our request)
func (cm *ConnectionManager) handleCallError(chargePointID string, message types.OCPPMessage) {
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
		cm.db.AddEvent(
			"ocpp",
			"error",
			pending.action+"Error",
			errorCode+": "+errorDescription,
			pending.chargePointID,
			pending.connectorID,
			pending.transactionID,
			map[string]interface{}{"messageId": messageID},
		)
		return
	}
	cm.db.AddEvent(
		"ocpp",
		"warn",
		"UnexpectedCallError",
		errorCode+": "+errorDescription,
		chargePointID,
		nil,
		nil,
		map[string]interface{}{"messageId": messageID},
	)
}

// handleDisconnect handles charge point disconnection
func (cm *ConnectionManager) handleDisconnect(chargePointID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.connections, chargePointID)
	cm.db.UpdateChargePointStatus(chargePointID, "offline")
	cm.db.AddEvent(
		"ocpp",
		"warn",
		"ConnectionClosed",
		"WebSocket connection closed",
		chargePointID,
		nil,
		nil,
		nil,
	)

	log.Printf("[ConnectionManager] %s disconnected", chargePointID)
	log.Printf("[ConnectionManager] Total connections: %d", len(cm.connections))
}

// sendCallResult sends a CALLRESULT to charge point
func (cm *ConnectionManager) sendCallResult(chargePointID, messageID string, payload interface{}) error {
	message := types.OCPPMessage{
		types.MessageTypeCallResult,
		messageID,
		payload,
	}
	return cm.sendMessage(chargePointID, message)
}

// sendCallError sends a CALLERROR to charge point
func (cm *ConnectionManager) sendCallError(chargePointID, messageID, errorCode, errorDescription string) error {
	message := types.OCPPMessage{
		types.MessageTypeCallError,
		messageID,
		errorCode,
		errorDescription,
		map[string]interface{}{},
	}
	return cm.sendMessage(chargePointID, message)
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
	connectorID, transactionID, details := extractPayloadContext(action, payload)

	// Create pending request
	pending := &pendingRequest{
		responseChan:  make(chan interface{}, 1),
		errorChan:     make(chan error, 1),
		action:        action,
		chargePointID: chargePointID,
		connectorID:   connectorID,
		transactionID: transactionID,
		sentAt:        time.Now(),
	}

	cm.mu.Lock()
	cm.pendingRequests[messageID] = pending
	cm.mu.Unlock()

	// Send message
	log.Printf("[ConnectionManager] -> %s: %s", chargePointID, action)
	cm.db.AddEvent(
		"ocpp",
		"info",
		action,
		"Sending outbound CALL to charge point",
		chargePointID,
		connectorID,
		transactionID,
		details,
	)
	if err := cm.sendMessage(chargePointID, message); err != nil {
		cm.mu.Lock()
		delete(cm.pendingRequests, messageID)
		cm.mu.Unlock()
		cm.db.AddEvent(
			"ocpp",
			"error",
			action+"SendFailed",
			err.Error(),
			chargePointID,
			connectorID,
			transactionID,
			nil,
		)
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-pending.responseChan:
		cm.db.AddEvent(
			"ocpp",
			"info",
			action+"Completed",
			"Outbound CALL completed",
			chargePointID,
			connectorID,
			transactionID,
			map[string]interface{}{
				"durationMs": time.Since(pending.sentAt).Milliseconds(),
			},
		)
		return response, nil
	case err := <-pending.errorChan:
		return nil, err
	case <-time.After(requestTimeout):
		cm.mu.Lock()
		delete(cm.pendingRequests, messageID)
		cm.mu.Unlock()
		cm.db.AddEvent(
			"ocpp",
			"error",
			action+"Timeout",
			"Timed out waiting for CALLRESULT",
			chargePointID,
			connectorID,
			transactionID,
			map[string]interface{}{
				"timeoutSec": int(requestTimeout.Seconds()),
			},
		)
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
		"totalConnections":      len(cm.connections),
		"pendingRequests":       len(cm.pendingRequests),
		"connectedChargePoints": cm.GetConnectedChargePoints(),
	}
}

// GetConnectorCountFromChargePoint asks the station for NumberOfConnectors.
func (cm *ConnectionManager) GetConnectorCountFromChargePoint(chargePointID string) (int, error) {
	req := types.GetConfigurationRequest{
		Key: []string{"NumberOfConnectors"},
	}

	response, err := cm.SendCall(chargePointID, types.ActionGetConfiguration, req)
	if err != nil {
		return 0, err
	}

	responseJSON, _ := json.Marshal(response)
	var parsed types.GetConfigurationResponse
	if err := json.Unmarshal(responseJSON, &parsed); err != nil {
		return 0, fmt.Errorf("failed to parse GetConfiguration response: %w", err)
	}

	for _, cfg := range parsed.ConfigurationKey {
		if cfg.Key == "NumberOfConnectors" && cfg.Value != nil {
			n, convErr := strconv.Atoi(*cfg.Value)
			if convErr != nil {
				return 0, fmt.Errorf("invalid NumberOfConnectors value %q", *cfg.Value)
			}
			return n, nil
		}
	}

	return 0, fmt.Errorf("NumberOfConnectors key not returned by charge point")
}

func extractPayloadContext(
	action string,
	payload interface{},
) (*int, *int, map[string]interface{}) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &raw); err != nil {
		return nil, nil, nil
	}

	connectorID := parseIntField(raw["connectorId"])
	transactionID := parseIntField(raw["transactionId"])
	details := map[string]interface{}{}

	switch action {
	case types.ActionBootNotification:
		if vendor, ok := raw["chargePointVendor"]; ok {
			details["vendor"] = vendor
		}
		if model, ok := raw["chargePointModel"]; ok {
			details["model"] = model
		}
	case types.ActionStatusNotification:
		if status, ok := raw["status"]; ok {
			details["status"] = status
		}
		if errorCode, ok := raw["errorCode"]; ok {
			details["errorCode"] = errorCode
		}
	case types.ActionStartTransaction:
		if idTag, ok := raw["idTag"]; ok {
			details["idTag"] = idTag
		}
	case types.ActionStopTransaction:
		if reason, ok := raw["reason"]; ok {
			details["reason"] = reason
		}
	case types.ActionAuthorize:
		if idTag, ok := raw["idTag"]; ok {
			details["idTag"] = idTag
		}
	case types.ActionMeterValues:
		if meterValues, ok := raw["meterValue"].([]interface{}); ok {
			details["meterValueBatches"] = len(meterValues)
		}
		if transactionID == nil {
			transactionID = parseIntField(raw["transactionID"])
		}
	case types.ActionRemoteStartTransaction:
		if idTag, ok := raw["idTag"]; ok {
			details["idTag"] = idTag
		}
	case types.ActionRemoteStopTransaction:
		// no-op
	case types.ActionGetConfiguration:
		if keys, ok := raw["key"]; ok {
			details["keys"] = keys
		}
	}

	if len(details) == 0 {
		details = nil
	}

	return connectorID, transactionID, details
}

func parseIntField(v interface{}) *int {
	switch n := v.(type) {
	case float64:
		parsed := int(n)
		return &parsed
	case int:
		parsed := n
		return &parsed
	case int32:
		parsed := int(n)
		return &parsed
	case int64:
		parsed := int(n)
		return &parsed
	default:
		return nil
	}
}

func (cm *ConnectionManager) syncConnectorCountFromChargePoint(chargePointID string) {
	count, err := cm.GetConnectorCountFromChargePoint(chargePointID)
	if err != nil {
		cm.db.AddEvent(
			"ocpp",
			"warn",
			"ConnectorSyncAutoFailed",
			"Automatic connector sync failed: "+err.Error(),
			chargePointID,
			nil,
			nil,
			nil,
		)
		return
	}

	created, removed, syncErr := cm.db.SyncConnectorCount(chargePointID, count)
	if syncErr != nil {
		cm.db.AddEvent(
			"ocpp",
			"warn",
			"ConnectorSyncAutoFailed",
			"Automatic connector sync failed: "+syncErr.Error(),
			chargePointID,
			nil,
			nil,
			nil,
		)
		return
	}

	cm.db.AddEvent(
		"ocpp",
		"info",
		"ConnectorSyncAutoSuccess",
		"Automatic connector sync completed after BootNotification",
		chargePointID,
		nil,
		nil,
		map[string]interface{}{
			"numberOfConnectors": count,
			"created":            created,
			"removed":            removed,
		},
	)
}

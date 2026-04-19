package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/errors"
	"github.com/ADEXITUM/ocpp_management_system/pkg/ocpp"
	"github.com/ADEXITUM/ocpp_management_system/pkg/service"
)

// Server provides HTTP REST API for charging operations
type Server struct {
	port              int
	chargingService   *service.ChargingService
	db                *database.MockDatabase
	connectionManager *ocpp.ConnectionManager
}

// NewServer creates a new API server
func NewServer(
	port int,
	chargingService *service.ChargingService,
	db *database.MockDatabase,
	connectionManager *ocpp.ConnectionManager,
) *Server {
	return &Server{
		port:              port,
		chargingService:   chargingService,
		db:                db,
		connectionManager: connectionManager,
	}
}

// StartSessionRequest represents the request to start a charging session
type StartSessionRequest struct {
	ChargePointID string  `json:"chargePointId"`
	ConnectorID   int     `json:"connectorId"`
	UserID        string  `json:"userId"`
	AmountPaid    float64 `json:"amountPaid"` // Мок оплаты
}

// StartSessionResponse represents the response from starting a session
type StartSessionResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TransactionID int    `json:"transactionId,omitempty"`
	ChargePointID string `json:"chargePointId,omitempty"`
	ConnectorID   int    `json:"connectorId,omitempty"`
	UserID        string `json:"userId,omitempty"`
	PaymentStatus string `json:"paymentStatus"` // Мок статуса оплаты
}

// EnergyResponse represents energy consumption data
type EnergyResponse struct {
	Success         bool    `json:"success"`
	TransactionID   int     `json:"transactionId"`
	ChargePointID   string  `json:"chargePointId"`
	ConnectorID     int     `json:"connectorId"`
	UserID          string  `json:"userId"`
	EnergyWh        float64 `json:"energyWh"`
	EnergyKwh       float64 `json:"energyKwh"`
	DurationSeconds int     `json:"durationSeconds"`
	DurationMinutes int     `json:"durationMinutes"`
	Status          string  `json:"status"`
	CostEstimate    float64 `json:"costEstimate"` // $0.30 per kWh
}

// StopSessionResponse represents the response from stopping a session
type StopSessionResponse struct {
	Success         bool    `json:"success"`
	Message         string  `json:"message"`
	TransactionID   int     `json:"transactionId"`
	ChargePointID   string  `json:"chargePointId"`
	EnergyConsumed  float64 `json:"energyConsumed"`
	EnergyKwh       float64 `json:"energyKwh"`
	DurationSeconds int     `json:"durationSeconds"`
	DurationMinutes int     `json:"durationMinutes"`
	FinalCost       float64 `json:"finalCost"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error"`
	ErrorCode   string `json:"errorCode,omitempty"`
	UserMessage string `json:"userMessage"`
}

// Start starts the HTTP API server
func (s *Server) Start() error {
	http.HandleFunc("/sessions/start", s.handleStartSession)
	http.HandleFunc("/sessions/", s.handleSessionOperations)
	http.HandleFunc("/charge-points/", s.handleChargePointOperations)
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/dashboard", s.handleDashboard)
	http.HandleFunc("/state", s.handleState)
	http.HandleFunc("/events", s.handleEvents)

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("🌐 REST API listening on http://%s", addr)
	log.Printf("   POST   http://%s/sessions/start", addr)
	log.Printf("   GET    http://%s/sessions/{transactionId}/energy", addr)
	log.Printf("   POST   http://%s/sessions/{transactionId}/stop", addr)
	log.Printf("   GET    http://%s/health", addr)
	log.Printf("   GET    http://%s/dashboard", addr)
	log.Printf("   GET    http://%s/state", addr)
	log.Println()

	return http.ListenAndServe(addr, s.corsMiddleware(http.DefaultServeMux))
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleStartSession handles POST /sessions/start
func (s *Server) handleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.recordEvent(
			"warn",
			"StartSessionInvalidBody",
			"Failed to decode start session payload",
			"",
			nil,
			nil,
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}
	s.recordEvent(
		"info",
		"StartSessionRequest",
		"Received API request to start session",
		req.ChargePointID,
		ptrInt(req.ConnectorID),
		nil,
		map[string]interface{}{
			"userId":     req.UserID,
			"amountPaid": req.AmountPaid,
		},
	)

	// Валидация
	if req.ChargePointID == "" {
		s.recordEvent(
			"warn",
			"StartSessionValidationFailed",
			"chargePointId is required",
			"",
			ptrInt(req.ConnectorID),
			nil,
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "chargePointId is required", "")
		return
	}
	if req.ConnectorID == 0 {
		s.recordEvent(
			"warn",
			"StartSessionValidationFailed",
			"connectorId is required",
			req.ChargePointID,
			nil,
			nil,
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "connectorId is required", "")
		return
	}
	if req.UserID == "" {
		s.recordEvent(
			"warn",
			"StartSessionValidationFailed",
			"userId is required",
			req.ChargePointID,
			ptrInt(req.ConnectorID),
			nil,
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "userId is required", "")
		return
	}

	log.Printf("💳 [API] Payment approved for user %s: $%.2f", req.UserID, req.AmountPaid)
	log.Printf("       Starting session on %s connector %d", req.ChargePointID, req.ConnectorID)

	// Мок успешной оплаты
	paymentStatus := "ОПЛАТА ПРОШЛА УСПЕШНО ✓"

	// Запуск сессии
	result, err := s.chargingService.TurnOn(req.ChargePointID, req.ConnectorID, req.UserID)
	if err != nil {
		s.recordEvent(
			"warn",
			"StartSessionFailed",
			err.Error(),
			req.ChargePointID,
			ptrInt(req.ConnectorID),
			nil,
			nil,
		)
		s.sendOCPPError(w, err)
		return
	}
	s.recordEvent(
		"info",
		"StartSessionSuccess",
		"Charging session started",
		result.ChargePointID,
		ptrInt(result.ConnectorID),
		ptrInt(result.TransactionID),
		map[string]interface{}{"userId": result.UserID},
	)

	log.Printf("✅ [API] Session started: Transaction ID %d", result.TransactionID)

	response := StartSessionResponse{
		Success:       true,
		Message:       result.Message,
		TransactionID: result.TransactionID,
		ChargePointID: result.ChargePointID,
		ConnectorID:   result.ConnectorID,
		UserID:        result.UserID,
		PaymentStatus: paymentStatus,
	}

	s.sendJSON(w, http.StatusOK, response)
}

// handleSessionOperations handles /sessions/{transactionId}/...
func (s *Server) handleSessionOperations(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /sessions/{transactionId}/{operation}
	path := strings.TrimPrefix(r.URL.Path, "/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		s.sendError(w, http.StatusBadRequest, "Invalid URL format", "")
		return
	}

	transactionID, err := strconv.Atoi(parts[0])
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid transaction ID", "")
		return
	}

	operation := parts[1]

	switch operation {
	case "energy":
		if r.Method != http.MethodGet {
			s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}
		s.handleGetEnergy(w, r, transactionID)
	case "stop":
		if r.Method != http.MethodPost {
			s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}
		s.handleStopSession(w, r, transactionID)
	default:
		s.sendError(w, http.StatusNotFound, "Unknown operation", "")
	}
}

// handleChargePointOperations handles /charge-points/{chargePointId}/...
func (s *Server) handleChargePointOperations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/charge-points/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		s.sendError(w, http.StatusBadRequest, "Invalid URL format", "")
		return
	}

	chargePointID := strings.TrimSpace(parts[0])
	operation := strings.TrimSpace(parts[1])
	if chargePointID == "" {
		s.sendError(w, http.StatusBadRequest, "chargePointId is required", "")
		return
	}

	switch operation {
	case "sync-connectors":
		if r.Method != http.MethodPost {
			s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
			return
		}
		s.handleSyncConnectors(w, r, chargePointID)
	default:
		s.sendError(w, http.StatusNotFound, "Unknown operation", "")
	}
}

func (s *Server) handleSyncConnectors(w http.ResponseWriter, r *http.Request, chargePointID string) {
	s.recordEvent(
		"info",
		"SyncConnectorsRequest",
		"Sync connector count from charge point configuration",
		chargePointID,
		nil,
		nil,
		nil,
	)

	count, err := s.connectionManager.GetConnectorCountFromChargePoint(chargePointID)
	if err != nil {
		s.recordEvent(
			"warn",
			"SyncConnectorsFailed",
			err.Error(),
			chargePointID,
			nil,
			nil,
			nil,
		)
		s.sendJSON(w, http.StatusBadGateway, map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"userMessage": "Failed to read NumberOfConnectors from charger",
		})
		return
	}

	created, removed, syncErr := s.db.SyncConnectorCount(chargePointID, count)
	if syncErr != nil {
		s.recordEvent(
			"warn",
			"SyncConnectorsFailed",
			syncErr.Error(),
			chargePointID,
			nil,
			nil,
			nil,
		)
		s.sendError(w, http.StatusNotFound, syncErr.Error(), "CHARGE_POINT_NOT_FOUND")
		return
	}

	s.recordEvent(
		"info",
		"SyncConnectorsSuccess",
		"Connector list synchronized from charger",
		chargePointID,
		nil,
		nil,
		map[string]interface{}{
			"numberOfConnectors": count,
			"created":            created,
			"removed":            removed,
		},
	)

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":            true,
		"chargePointId":      chargePointID,
		"numberOfConnectors": count,
		"created":            created,
		"removed":            removed,
		"message":            "Connector list synchronized from charger configuration",
	})
}

// handleGetEnergy handles GET /sessions/{transactionId}/energy
func (s *Server) handleGetEnergy(w http.ResponseWriter, r *http.Request, transactionID int) {
	// Get chargePointId from query param
	chargePointID := r.URL.Query().Get("chargePointId")
	if chargePointID == "" {
		s.recordEvent(
			"warn",
			"GetEnergyValidationFailed",
			"chargePointId query parameter is required",
			"",
			nil,
			ptrInt(transactionID),
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "chargePointId query parameter is required", "")
		return
	}
	s.recordEvent(
		"info",
		"GetEnergyRequest",
		"Received API request for energy snapshot",
		chargePointID,
		nil,
		ptrInt(transactionID),
		nil,
	)

	consumption, err := s.chargingService.GetEnergyConsumption(chargePointID, transactionID)
	if err != nil {
		s.recordEvent(
			"warn",
			"GetEnergyFailed",
			err.Error(),
			chargePointID,
			nil,
			ptrInt(transactionID),
			nil,
		)
		s.sendOCPPError(w, err)
		return
	}

	energyKwh := consumption.CurrentEnergyWh / 1000
	durationMinutes := consumption.DurationSeconds / 60
	costEstimate := energyKwh * 0.30 // $0.30 per kWh

	response := EnergyResponse{
		Success:         true,
		TransactionID:   consumption.TransactionID,
		ChargePointID:   consumption.ChargePointID,
		ConnectorID:     consumption.ConnectorID,
		UserID:          consumption.UserID,
		EnergyWh:        consumption.CurrentEnergyWh,
		EnergyKwh:       energyKwh,
		DurationSeconds: consumption.DurationSeconds,
		DurationMinutes: durationMinutes,
		Status:          consumption.Status,
		CostEstimate:    costEstimate,
	}
	s.recordEvent(
		"info",
		"GetEnergySuccess",
		"Energy snapshot generated",
		chargePointID,
		ptrInt(consumption.ConnectorID),
		ptrInt(transactionID),
		map[string]interface{}{
			"energyWh":  consumption.CurrentEnergyWh,
			"status":    consumption.Status,
			"durationS": consumption.DurationSeconds,
		},
	)

	s.sendJSON(w, http.StatusOK, response)
}

// handleStopSession handles POST /sessions/{transactionId}/stop
func (s *Server) handleStopSession(w http.ResponseWriter, r *http.Request, transactionID int) {
	// Get chargePointId from query param
	chargePointID := r.URL.Query().Get("chargePointId")
	if chargePointID == "" {
		s.recordEvent(
			"warn",
			"StopSessionValidationFailed",
			"chargePointId query parameter is required",
			"",
			nil,
			ptrInt(transactionID),
			nil,
		)
		s.sendError(w, http.StatusBadRequest, "chargePointId query parameter is required", "")
		return
	}
	s.recordEvent(
		"info",
		"StopSessionRequest",
		"Received API request to stop session",
		chargePointID,
		nil,
		ptrInt(transactionID),
		nil,
	)

	log.Printf("🛑 [API] Stopping session %d on %s", transactionID, chargePointID)

	result, err := s.chargingService.TurnOff(chargePointID, transactionID)
	if err != nil {
		s.recordEvent(
			"warn",
			"StopSessionFailed",
			err.Error(),
			chargePointID,
			nil,
			ptrInt(transactionID),
			nil,
		)
		s.sendOCPPError(w, err)
		return
	}

	energyKwh := result.EnergyConsumed / 1000
	durationMinutes := result.Duration / 60
	finalCost := energyKwh * 0.30 // $0.30 per kWh

	log.Printf("✅ [API] Session stopped: %.2f kWh consumed, $%.2f cost", energyKwh, finalCost)

	response := StopSessionResponse{
		Success:         true,
		Message:         result.Message,
		TransactionID:   result.TransactionID,
		ChargePointID:   result.ChargePointID,
		EnergyConsumed:  result.EnergyConsumed,
		EnergyKwh:       energyKwh,
		DurationSeconds: result.Duration,
		DurationMinutes: durationMinutes,
		FinalCost:       finalCost,
	}
	s.recordEvent(
		"info",
		"StopSessionSuccess",
		"Charging session stopped",
		result.ChargePointID,
		nil,
		ptrInt(result.TransactionID),
		map[string]interface{}{
			"energyWh":  result.EnergyConsumed,
			"durationS": result.Duration,
		},
	)

	s.sendJSON(w, http.StatusOK, response)
}

// handleHealth handles GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "ocpp-management-system",
	})
}

// Helper methods

func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, status int, message string, errorCode string) {
	response := ErrorResponse{
		Success:     false,
		Error:       message,
		ErrorCode:   errorCode,
		UserMessage: message,
	}
	s.sendJSON(w, status, response)
}

func (s *Server) sendOCPPError(w http.ResponseWriter, err error) {
	if ocppErr, ok := err.(*errors.OCPPError); ok {
		response := ErrorResponse{
			Success:     false,
			Error:       ocppErr.Error(),
			ErrorCode:   ocppErr.Code,
			UserMessage: ocppErr.UserMessage,
		}

		// Выбираем HTTP статус на основе типа ошибки
		status := http.StatusBadRequest
		switch ocppErr.Code {
		case "CHARGE_POINT_NOT_FOUND", "SESSION_NOT_FOUND", "CONNECTOR_NOT_FOUND":
			status = http.StatusNotFound
		case "CHARGE_POINT_OFFLINE", "CONNECTOR_UNAVAILABLE", "SESSION_ALREADY_ACTIVE":
			status = http.StatusConflict
		case "OPERATION_TIMEOUT":
			status = http.StatusRequestTimeout
		}

		s.sendJSON(w, status, response)
	} else {
		s.sendError(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
	}
}

func (s *Server) recordEvent(
	level, action, message, chargePointID string,
	connectorID, transactionID *int,
	details map[string]interface{},
) {
	s.db.AddEvent("api", level, action, message, chargePointID, connectorID, transactionID, details)
}

func ptrInt(v int) *int {
	if v == 0 {
		return nil
	}
	cp := v
	return &cp
}

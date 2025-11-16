package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ADEXITUM/ocpp_management_system/pkg/errors"
	"github.com/ADEXITUM/ocpp_management_system/pkg/service"
)

// Server provides HTTP REST API for charging operations
type Server struct {
	port            int
	chargingService *service.ChargingService
}

// NewServer creates a new API server
func NewServer(port int, chargingService *service.ChargingService) *Server {
	return &Server{
		port:            port,
		chargingService: chargingService,
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
	Success      bool   `json:"success"`
	Error        string `json:"error"`
	ErrorCode    string `json:"errorCode,omitempty"`
	UserMessage  string `json:"userMessage"`
}

// Start starts the HTTP API server
func (s *Server) Start() error {
	http.HandleFunc("/sessions/start", s.handleStartSession)
	http.HandleFunc("/sessions/", s.handleSessionOperations)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("🌐 REST API listening on http://%s", addr)
	log.Printf("   POST   http://%s/sessions/start", addr)
	log.Printf("   GET    http://%s/sessions/{transactionId}/energy", addr)
	log.Printf("   POST   http://%s/sessions/{transactionId}/stop", addr)
	log.Printf("   GET    http://%s/health", addr)
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
		s.sendError(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	// Валидация
	if req.ChargePointID == "" {
		s.sendError(w, http.StatusBadRequest, "chargePointId is required", "")
		return
	}
	if req.ConnectorID == 0 {
		s.sendError(w, http.StatusBadRequest, "connectorId is required", "")
		return
	}
	if req.UserID == "" {
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
		s.sendOCPPError(w, err)
		return
	}

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

// handleGetEnergy handles GET /sessions/{transactionId}/energy
func (s *Server) handleGetEnergy(w http.ResponseWriter, r *http.Request, transactionID int) {
	// Get chargePointId from query param
	chargePointID := r.URL.Query().Get("chargePointId")
	if chargePointID == "" {
		s.sendError(w, http.StatusBadRequest, "chargePointId query parameter is required", "")
		return
	}

	consumption, err := s.chargingService.GetEnergyConsumption(chargePointID, transactionID)
	if err != nil {
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

	s.sendJSON(w, http.StatusOK, response)
}

// handleStopSession handles POST /sessions/{transactionId}/stop
func (s *Server) handleStopSession(w http.ResponseWriter, r *http.Request, transactionID int) {
	// Get chargePointId from query param
	chargePointID := r.URL.Query().Get("chargePointId")
	if chargePointID == "" {
		s.sendError(w, http.StatusBadRequest, "chargePointId query parameter is required", "")
		return
	}

	log.Printf("🛑 [API] Stopping session %d on %s", transactionID, chargePointID)

	result, err := s.chargingService.TurnOff(chargePointID, transactionID)
	if err != nil {
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

	s.sendJSON(w, http.StatusOK, response)
}

// handleHealth handles GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
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

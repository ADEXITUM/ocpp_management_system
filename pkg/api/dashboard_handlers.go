package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/ADEXITUM/ocpp_management_system/pkg/types"
)

var connectorStatusTransitions = map[string][]string{
	types.StatusAvailable: {
		types.StatusPreparing,
		types.StatusReserved,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusPreparing: {
		types.StatusCharging,
		types.StatusSuspendedEV,
		types.StatusAvailable,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusCharging: {
		types.StatusSuspendedEV,
		types.StatusFinishing,
		types.StatusAvailable,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusSuspendedEV: {
		types.StatusCharging,
		types.StatusFinishing,
		types.StatusAvailable,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusFinishing: {
		types.StatusAvailable,
		types.StatusCharging,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusReserved: {
		types.StatusPreparing,
		types.StatusAvailable,
		types.StatusUnavailable,
		types.StatusFaulted,
	},
	types.StatusUnavailable: {
		types.StatusAvailable,
		types.StatusFaulted,
	},
	types.StatusFaulted: {
		types.StatusAvailable,
		types.StatusUnavailable,
	},
}

var allConnectorStatuses = []string{
	types.StatusAvailable,
	types.StatusPreparing,
	types.StatusCharging,
	types.StatusSuspendedEV,
	types.StatusFinishing,
	types.StatusReserved,
	types.StatusUnavailable,
	types.StatusFaulted,
}

type connectorState struct {
	ConnectorID         int      `json:"connectorId"`
	Status              string   `json:"status"`
	CurrentTransaction  *int     `json:"currentTransaction,omitempty"`
	LastStatusUpdate    string   `json:"lastStatusUpdate"`
	AllowedNextStatuses []string `json:"allowedNextStatuses"`
}

type chargePointState struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Vendor             string           `json:"vendor"`
	Model              string           `json:"model"`
	Status             string           `json:"status"`
	RegistrationStatus string           `json:"registrationStatus"`
	LastSeen           string           `json:"lastSeen"`
	Connected          bool             `json:"connected"`
	Connectors         []connectorState `json:"connectors"`
}

type activeSessionState struct {
	TransactionID   int     `json:"transactionId"`
	ChargePointID   string  `json:"chargePointId"`
	ConnectorID     int     `json:"connectorId"`
	UserID          string  `json:"userId"`
	CurrentEnergyWh float64 `json:"currentEnergyWh"`
	DurationSeconds int     `json:"durationSeconds"`
	Status          string  `json:"status"`
	LastUpdate      string  `json:"lastUpdate"`
}

type stateResponse struct {
	Success                bool                     `json:"success"`
	Timestamp              string                   `json:"timestamp"`
	Stats                  interface{}              `json:"stats"`
	ConnectedChargePoints  []string                 `json:"connectedChargePoints"`
	ValidConnectorStatuses []string                 `json:"validConnectorStatuses"`
	StatusTransitions      map[string][]string      `json:"statusTransitions"`
	ChargePoints           []chargePointState       `json:"chargePoints"`
	ActiveSessions         []activeSessionState     `json:"activeSessions"`
	RecentEvents           []database.TimelineEvent `json:"recentEvents"`
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	connected := s.connectionManager.GetConnectedChargePoints()
	sort.Strings(connected)

	connectedSet := make(map[string]bool, len(connected))
	for _, cpID := range connected {
		connectedSet[cpID] = true
	}

	chargePointsRaw := s.db.GetAllChargePoints()
	sort.Slice(chargePointsRaw, func(i, j int) bool {
		return chargePointsRaw[i].ID < chargePointsRaw[j].ID
	})

	chargePoints := make([]chargePointState, 0, len(chargePointsRaw))
	seenChargePoints := make(map[string]bool, len(chargePointsRaw))
	for _, cp := range chargePointsRaw {
		seenChargePoints[cp.ID] = true
		connectorsRaw := s.db.GetConnectorsByChargePoint(cp.ID)
		sort.Slice(connectorsRaw, func(i, j int) bool {
			return connectorsRaw[i].ConnectorID < connectorsRaw[j].ConnectorID
		})

		connectors := make([]connectorState, 0, len(connectorsRaw))
		for _, conn := range connectorsRaw {
			connectors = append(connectors, connectorState{
				ConnectorID:         conn.ConnectorID,
				Status:              conn.Status,
				CurrentTransaction:  conn.CurrentTransaction,
				LastStatusUpdate:    conn.LastStatusUpdate.Format(time.RFC3339),
				AllowedNextStatuses: allowedNextStatuses(conn.Status),
			})
		}

		chargePoints = append(chargePoints, chargePointState{
			ID:                 cp.ID,
			Name:               cp.Name,
			Vendor:             cp.Vendor,
			Model:              cp.Model,
			Status:             cp.Status,
			RegistrationStatus: cp.RegistrationStatus,
			LastSeen:           cp.LastSeen.Format(time.RFC3339),
			Connected:          connectedSet[cp.ID],
			Connectors:         connectors,
		})
	}

	// Include currently connected charge points even if they are not present
	// in the mock database configuration.
	for _, cpID := range connected {
		if seenChargePoints[cpID] {
			continue
		}

		connectorsRaw := s.db.GetConnectorsByChargePoint(cpID)
		sort.Slice(connectorsRaw, func(i, j int) bool {
			return connectorsRaw[i].ConnectorID < connectorsRaw[j].ConnectorID
		})

		connectors := make([]connectorState, 0, len(connectorsRaw))
		for _, conn := range connectorsRaw {
			connectors = append(connectors, connectorState{
				ConnectorID:         conn.ConnectorID,
				Status:              conn.Status,
				CurrentTransaction:  conn.CurrentTransaction,
				LastStatusUpdate:    conn.LastStatusUpdate.Format(time.RFC3339),
				AllowedNextStatuses: allowedNextStatuses(conn.Status),
			})
		}

		chargePoints = append(chargePoints, chargePointState{
			ID:                 cpID,
			Name:               "Connected (not in mock DB)",
			Vendor:             "-",
			Model:              "-",
			Status:             "online",
			RegistrationStatus: "unknown",
			LastSeen:           "",
			Connected:          true,
			Connectors:         connectors,
		})
	}

	sort.Slice(chargePoints, func(i, j int) bool {
		return chargePoints[i].ID < chargePoints[j].ID
	})

	activeSessionsRaw := s.chargingService.GetActiveSessions()
	sort.Slice(activeSessionsRaw, func(i, j int) bool {
		return activeSessionsRaw[i].TransactionID < activeSessionsRaw[j].TransactionID
	})

	activeSessions := make([]activeSessionState, 0, len(activeSessionsRaw))
	for _, session := range activeSessionsRaw {
		activeSessions = append(activeSessions, activeSessionState{
			TransactionID:   session.TransactionID,
			ChargePointID:   session.ChargePointID,
			ConnectorID:     session.ConnectorID,
			UserID:          session.UserID,
			CurrentEnergyWh: session.CurrentEnergyWh,
			DurationSeconds: session.DurationSeconds,
			Status:          session.Status,
			LastUpdate:      session.LastUpdate.Format(time.RFC3339),
		})
	}

	statusTransitions := make(map[string][]string, len(connectorStatusTransitions))
	for current, next := range connectorStatusTransitions {
		statusTransitions[current] = append([]string(nil), next...)
	}

	s.sendJSON(w, http.StatusOK, stateResponse{
		Success:                true,
		Timestamp:              time.Now().UTC().Format(time.RFC3339),
		Stats:                  s.db.GetStats(),
		ConnectedChargePoints:  connected,
		ValidConnectorStatuses: append([]string(nil), allConnectorStatuses...),
		StatusTransitions:      statusTransitions,
		ChargePoints:           chargePoints,
		ActiveSessions:         activeSessions,
		RecentEvents:           s.db.GetRecentEvents(200),
	})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	limit := 200
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 1000 {
		limit = 1000
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"limit":     limit,
		"events":    s.db.GetRecentEvents(limit),
	})
}

func allowedNextStatuses(current string) []string {
	for known, next := range connectorStatusTransitions {
		if strings.EqualFold(known, current) {
			return append([]string(nil), next...)
		}
	}
	return append([]string(nil), allConnectorStatuses...)
}

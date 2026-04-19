package ocpp

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ADEXITUM/ocpp_management_system/pkg/database"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
	Subprotocols: []string{"ocpp1.6", "ocpp1.6j", "ocpp16", "ocpp2.0", "ocpp2.0.1"},
}

// Server is the OCPP WebSocket server
type Server struct {
	port              int
	connectionManager *ConnectionManager
}

// NewServer creates a new OCPP server
func NewServer(port int, db *database.MockDatabase) *Server {
	return &Server{
		port:              port,
		connectionManager: NewConnectionManager(db),
	}
}

// GetConnectionManager returns the connection manager
func (s *Server) GetConnectionManager() *ConnectionManager {
	return s.connectionManager
}

func (s *Server) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebSocket)
	return mux
}

// Start starts the OCPP WebSocket server (plain WS)
func (s *Server) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("🔌 OCPP Server listening on ws://%s", addr)
	log.Printf("   Waiting for charge points to connect...")

	return http.ListenAndServe(addr, s.newMux())
}

// StartTLS starts the OCPP WebSocket server over TLS (WSS)
func (s *Server) StartTLS(port int, certFile, keyFile string) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("🔐 OCPP TLS server listening on wss://%s", addr)
	log.Printf("   Certificate: %s", certFile)
	log.Printf("   Key: %s", keyFile)
	log.Printf("   Waiting for charge points to connect (TLS)...")

	return http.ListenAndServeTLS(addr, certFile, keyFile, s.newMux())
}

// handleWebSocket handles WebSocket connections from charge points
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract charge point ID from URL path
	chargePointID := s.extractChargePointID(r.URL.Path)

	if chargePointID == "" {
		log.Printf("[OCPPServer] Connection rejected: No charge point ID in URL")
		http.Error(w, "Charge point ID required in URL path (e.g., /CP001)", http.StatusBadRequest)
		return
	}

	log.Printf("[OCPPServer] New connection from %s", chargePointID)
	log.Printf("            Remote address: %s", r.RemoteAddr)
	if requested := r.Header.Get("Sec-WebSocket-Protocol"); requested != "" {
		log.Printf("            Requested subprotocol(s): %s", requested)
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[OCPPServer] Failed to upgrade connection: %v", err)
		return
	}

	// Log the negotiated subprotocol
	if conn.Subprotocol() != "" {
		log.Printf("            Subprotocol: %s", conn.Subprotocol())
	}

	// Register the connection
	s.connectionManager.RegisterConnection(chargePointID, conn)
}

// extractChargePointID extracts charge point ID from URL path
func (s *Server) extractChargePointID(path string) string {
	// Handle formats like:
	// /CP001
	// /ocpp16/CP001
	// /CP001/
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 {
		return ""
	}

	// Return the last non-empty part
	return parts[len(parts)-1]
}

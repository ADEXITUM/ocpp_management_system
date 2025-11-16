package types

import "time"

// OCPP Message Types
const (
	MessageTypeCall       = 2
	MessageTypeCallResult = 3
	MessageTypeCallError  = 4
)

// OCPPMessage represents a generic OCPP message
// Format: [MessageTypeId, UniqueId, Action, Payload]
type OCPPMessage []interface{}

// Charge Point Status
const (
	StatusAvailable   = "Available"
	StatusPreparing   = "Preparing"
	StatusCharging    = "Charging"
	StatusSuspendedEV = "SuspendedEV"
	StatusFinishing   = "Finishing"
	StatusReserved    = "Reserved"
	StatusUnavailable = "Unavailable"
	StatusFaulted     = "Faulted"
)

// Registration Status
const (
	RegistrationAccepted = "Accepted"
	RegistrationPending  = "Pending"
	RegistrationRejected = "Rejected"
)

// Authorization Status
const (
	AuthStatusAccepted     = "Accepted"
	AuthStatusBlocked      = "Blocked"
	AuthStatusExpired      = "Expired"
	AuthStatusInvalid      = "Invalid"
	AuthStatusConcurrentTx = "ConcurrentTx"
)

// Remote Start/Stop Status
const (
	RemoteStartStopAccepted = "Accepted"
	RemoteStartStopRejected = "Rejected"
)

// OCPP Actions
const (
	ActionBootNotification      = "BootNotification"
	ActionHeartbeat             = "Heartbeat"
	ActionStatusNotification    = "StatusNotification"
	ActionMeterValues           = "MeterValues"
	ActionStartTransaction      = "StartTransaction"
	ActionStopTransaction       = "StopTransaction"
	ActionAuthorize             = "Authorize"
	ActionRemoteStartTransaction = "RemoteStartTransaction"
	ActionRemoteStopTransaction  = "RemoteStopTransaction"
)

// BootNotification Request/Response
type BootNotificationRequest struct {
	ChargePointVendor       string  `json:"chargePointVendor"`
	ChargePointModel        string  `json:"chargePointModel"`
	ChargePointSerialNumber *string `json:"chargePointSerialNumber,omitempty"`
	ChargeBoxSerialNumber   *string `json:"chargeBoxSerialNumber,omitempty"`
	FirmwareVersion         *string `json:"firmwareVersion,omitempty"`
	Iccid                   *string `json:"iccid,omitempty"`
	Imsi                    *string `json:"imsi,omitempty"`
	MeterType               *string `json:"meterType,omitempty"`
	MeterSerialNumber       *string `json:"meterSerialNumber,omitempty"`
}

type BootNotificationResponse struct {
	Status      string `json:"status"`
	CurrentTime string `json:"currentTime"`
	Interval    int    `json:"interval"`
}

// Heartbeat Request/Response
type HeartbeatRequest struct{}

type HeartbeatResponse struct {
	CurrentTime string `json:"currentTime"`
}

// StatusNotification Request/Response
type StatusNotificationRequest struct {
	ConnectorID     int     `json:"connectorId"`
	ErrorCode       string  `json:"errorCode"`
	Status          string  `json:"status"`
	Timestamp       *string `json:"timestamp,omitempty"`
	Info            *string `json:"info,omitempty"`
	VendorID        *string `json:"vendorId,omitempty"`
	VendorErrorCode *string `json:"vendorErrorCode,omitempty"`
}

type StatusNotificationResponse struct{}

// MeterValues Request/Response
type SampledValue struct {
	Value     string  `json:"value"`
	Context   *string `json:"context,omitempty"`
	Format    *string `json:"format,omitempty"`
	Measurand *string `json:"measurand,omitempty"`
	Phase     *string `json:"phase,omitempty"`
	Location  *string `json:"location,omitempty"`
	Unit      *string `json:"unit,omitempty"`
}

type MeterValue struct {
	Timestamp    string         `json:"timestamp"`
	SampledValue []SampledValue `json:"sampledValue"`
}

type MeterValuesRequest struct {
	ConnectorID   int          `json:"connectorId"`
	TransactionID *int         `json:"transactionId,omitempty"`
	MeterValue    []MeterValue `json:"meterValue"`
}

type MeterValuesResponse struct{}

// StartTransaction Request/Response
type StartTransactionRequest struct {
	ConnectorID   int     `json:"connectorId"`
	IDTag         string  `json:"idTag"`
	MeterStart    int     `json:"meterStart"`
	Timestamp     string  `json:"timestamp"`
	ReservationID *int    `json:"reservationId,omitempty"`
}

type IDTagInfo struct {
	Status       string  `json:"status"`
	ExpiryDate   *string `json:"expiryDate,omitempty"`
	ParentIDTag  *string `json:"parentIdTag,omitempty"`
}

type StartTransactionResponse struct {
	TransactionID int       `json:"transactionId"`
	IDTagInfo     IDTagInfo `json:"idTagInfo"`
}

// StopTransaction Request/Response
type StopTransactionRequest struct {
	TransactionID   int           `json:"transactionId"`
	IDTag           *string       `json:"idTag,omitempty"`
	MeterStop       int           `json:"meterStop"`
	Timestamp       string        `json:"timestamp"`
	Reason          *string       `json:"reason,omitempty"`
	TransactionData []MeterValue  `json:"transactionData,omitempty"`
}

type StopTransactionResponse struct {
	IDTagInfo *IDTagInfo `json:"idTagInfo,omitempty"`
}

// Authorize Request/Response
type AuthorizeRequest struct {
	IDTag string `json:"idTag"`
}

type AuthorizeResponse struct {
	IDTagInfo IDTagInfo `json:"idTagInfo"`
}

// RemoteStartTransaction Request/Response
type RemoteStartTransactionRequest struct {
	IDTag          string      `json:"idTag"`
	ConnectorID    *int        `json:"connectorId,omitempty"`
	ChargingProfile interface{} `json:"chargingProfile,omitempty"`
}

type RemoteStartTransactionResponse struct {
	Status string `json:"status"`
}

// RemoteStopTransaction Request/Response
type RemoteStopTransactionRequest struct {
	TransactionID int `json:"transactionId"`
}

type RemoteStopTransactionResponse struct {
	Status string `json:"status"`
}

// Domain Types

type ChargePoint struct {
	ID                 string
	Name               string
	Vendor             string
	Model              string
	SerialNumber       string
	FirmwareVersion    string
	NumberOfConnectors int
	Status             string // "online" or "offline"
	RegistrationStatus string // "accepted", "pending", "rejected"
	LastSeen           time.Time
	CreatedAt          time.Time
}

type Connector struct {
	ChargePointID      string
	ConnectorID        int
	Status             string
	CurrentTransaction *int
	LastStatusUpdate   time.Time
}

type ChargingSession struct {
	TransactionID     int
	ChargePointID     string
	ConnectorID       int
	UserID            string
	StartTime         time.Time
	StartMeterValue   int
	CurrentMeterValue int
	EndTime           *time.Time
	EndMeterValue     *int
	Status            string // "active", "completed", "failed"
	StopReason        string
}

type EnergyReading struct {
	Timestamp     time.Time
	TransactionID int
	EnergyWh      float64
	PowerW        *float64
	CurrentA      *float64
	VoltageV      *float64
}

// Service Result Types

type TurnOnResult struct {
	Success       bool
	TransactionID int
	ChargePointID string
	ConnectorID   int
	UserID        string
	Message       string
}

type TurnOffResult struct {
	Success       bool
	TransactionID int
	ChargePointID string
	EnergyConsumed float64
	Duration      int
	Message       string
}

type EnergyConsumptionResult struct {
	TransactionID   int
	ChargePointID   string
	ConnectorID     int
	UserID          string
	StartTime       time.Time
	CurrentEnergyWh float64
	DurationSeconds int
	Status          string
	LastUpdate      time.Time
}

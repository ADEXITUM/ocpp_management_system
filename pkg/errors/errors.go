package errors

import "fmt"

// OCPPError is the base error type for OCPP-related errors
type OCPPError struct {
	Code        string
	Message     string
	UserMessage string
}

func (e *OCPPError) Error() string {
	return e.Message
}

// Error constructors

func NewChargePointNotFoundError(chargePointID string) *OCPPError {
	return &OCPPError{
		Code:    "CHARGE_POINT_NOT_FOUND",
		Message: fmt.Sprintf("Charge point '%s' not found in database", chargePointID),
		UserMessage: "This charging station is not configured in the system. " +
			"Please contact support to register this charge point.",
	}
}

func NewChargePointOfflineError(chargePointID string) *OCPPError {
	return &OCPPError{
		Code:    "CHARGE_POINT_OFFLINE",
		Message: fmt.Sprintf("Charge point '%s' is not connected", chargePointID),
		UserMessage: "This charging station is currently offline. " +
			"Please try another station or check back later.",
	}
}

func NewConnectorNotFoundError(chargePointID string, connectorID int) *OCPPError {
	return &OCPPError{
		Code:    "CONNECTOR_NOT_FOUND",
		Message: fmt.Sprintf("Connector %d not found on charge point '%s'", connectorID, chargePointID),
		UserMessage: fmt.Sprintf("Connector %d does not exist on this charging station.", connectorID),
	}
}

func NewConnectorUnavailableError(chargePointID string, connectorID int, status string) *OCPPError {
	return &OCPPError{
		Code:    "CONNECTOR_UNAVAILABLE",
		Message: fmt.Sprintf("Connector %d on '%s' is %s", connectorID, chargePointID, status),
		UserMessage: fmt.Sprintf("This connector is currently %s. "+
			"Please try another connector or wait until it becomes available.", status),
	}
}

func NewSessionNotFoundError(transactionID int) *OCPPError {
	return &OCPPError{
		Code:    "SESSION_NOT_FOUND",
		Message: fmt.Sprintf("Charging session %d not found", transactionID),
		UserMessage: "The charging session was not found. " +
			"It may have already been completed or cancelled.",
	}
}

func NewSessionAlreadyActiveError(chargePointID string, connectorID int) *OCPPError {
	return &OCPPError{
		Code:    "SESSION_ALREADY_ACTIVE",
		Message: fmt.Sprintf("Connector %d on '%s' already has an active session", connectorID, chargePointID),
		UserMessage: "This connector is already in use. " +
			"Please wait for the current session to complete.",
	}
}

func NewRemoteOperationFailedError(operation, chargePointID, reason string) *OCPPError {
	detailMessage := ""
	if reason != "" {
		detailMessage = fmt.Sprintf(" Reason: %s", reason)
	}
	return &OCPPError{
		Code:    "REMOTE_OPERATION_FAILED",
		Message: fmt.Sprintf("Remote %s failed for '%s'.%s", operation, chargePointID, detailMessage),
		UserMessage: fmt.Sprintf("Unable to %s the charging session. "+
			"The charging station may be experiencing issues. "+
			"Please try again or contact support.", operation),
	}
}

func NewOperationTimeoutError(operation, chargePointID string) *OCPPError {
	return &OCPPError{
		Code:    "OPERATION_TIMEOUT",
		Message: fmt.Sprintf("%s timed out for charge point '%s'", operation, chargePointID),
		UserMessage: "The operation took too long to complete. " +
			"The charging station may be experiencing connectivity issues. " +
			"Please try again.",
	}
}

func NewInvalidUserIDError(userID string) *OCPPError {
	return &OCPPError{
		Code:    "INVALID_USER_ID",
		Message: fmt.Sprintf("User ID '%s' is invalid or blocked", userID),
		UserMessage: "Your user account is not authorized to start charging sessions. " +
			"Please contact support.",
	}
}

func NewChargePointNotConfiguredError(chargePointID string) *OCPPError {
	return &OCPPError{
		Code:    "CHARGE_POINT_NOT_CONFIGURED",
		Message: fmt.Sprintf("Charge point '%s' connected but not configured in database", chargePointID),
		UserMessage: fmt.Sprintf("⚠️  WARNING: Charge point '%s' is not configured. "+
			"Please configure this charge point in the database before use!", chargePointID),
	}
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	if ocppErr, ok := err.(*OCPPError); ok {
		return ocppErr.UserMessage
	}
	return "An unexpected error occurred. Please try again or contact support."
}

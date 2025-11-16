/**
 * Custom Error Classes for OCPP Management System
 * These provide clear, user-friendly error messages for calling services
 */

export class OCPPError extends Error {
  public readonly code: string;
  public readonly userMessage: string;

  constructor(code: string, message: string, userMessage: string) {
    super(message);
    this.name = this.constructor.name;
    this.code = code;
    this.userMessage = userMessage;
    Error.captureStackTrace(this, this.constructor);
  }
}

export class ChargePointNotFoundError extends OCPPError {
  constructor(chargePointId: string) {
    super(
      'CHARGE_POINT_NOT_FOUND',
      `Charge point '${chargePointId}' not found in database`,
      'This charging station is not configured in the system. Please contact support to register this charge point.'
    );
  }
}

export class ChargePointOfflineError extends OCPPError {
  constructor(chargePointId: string) {
    super(
      'CHARGE_POINT_OFFLINE',
      `Charge point '${chargePointId}' is not connected`,
      'This charging station is currently offline. Please try another station or check back later.'
    );
  }
}

export class ConnectorNotFoundError extends OCPPError {
  constructor(chargePointId: string, connectorId: number) {
    super(
      'CONNECTOR_NOT_FOUND',
      `Connector ${connectorId} not found on charge point '${chargePointId}'`,
      `Connector ${connectorId} does not exist on this charging station.`
    );
  }
}

export class ConnectorUnavailableError extends OCPPError {
  constructor(chargePointId: string, connectorId: number, status: string) {
    super(
      'CONNECTOR_UNAVAILABLE',
      `Connector ${connectorId} on '${chargePointId}' is ${status}`,
      `This connector is currently ${status.toLowerCase()}. Please try another connector or wait until it becomes available.`
    );
  }
}

export class SessionNotFoundError extends OCPPError {
  constructor(transactionId: number) {
    super(
      'SESSION_NOT_FOUND',
      `Charging session ${transactionId} not found`,
      'The charging session was not found. It may have already been completed or cancelled.'
    );
  }
}

export class SessionAlreadyActiveError extends OCPPError {
  constructor(chargePointId: string, connectorId: number) {
    super(
      'SESSION_ALREADY_ACTIVE',
      `Connector ${connectorId} on '${chargePointId}' already has an active session`,
      'This connector is already in use. Please wait for the current session to complete.'
    );
  }
}

export class RemoteOperationFailedError extends OCPPError {
  constructor(operation: string, chargePointId: string, reason?: string) {
    const detailMessage = reason ? ` Reason: ${reason}` : '';
    super(
      'REMOTE_OPERATION_FAILED',
      `Remote ${operation} failed for '${chargePointId}'.${detailMessage}`,
      `Unable to ${operation.toLowerCase()} the charging session. The charging station may be experiencing issues. Please try again or contact support.`
    );
  }
}

export class OperationTimeoutError extends OCPPError {
  constructor(operation: string, chargePointId: string) {
    super(
      'OPERATION_TIMEOUT',
      `${operation} timed out for charge point '${chargePointId}'`,
      'The operation took too long to complete. The charging station may be experiencing connectivity issues. Please try again.'
    );
  }
}

export class InvalidUserIdError extends OCPPError {
  constructor(userId: string) {
    super(
      'INVALID_USER_ID',
      `User ID '${userId}' is invalid or blocked`,
      'Your user account is not authorized to start charging sessions. Please contact support.'
    );
  }
}

export class ChargePointNotConfiguredError extends OCPPError {
  constructor(chargePointId: string) {
    super(
      'CHARGE_POINT_NOT_CONFIGURED',
      `Charge point '${chargePointId}' connected but not configured in database`,
      `⚠️  WARNING: Charge point '${chargePointId}' is not configured. Please configure this charge point in the database before use!`
    );
  }
}

/**
 * Helper function to convert errors to user-friendly messages
 */
export function getUserFriendlyErrorMessage(error: Error): string {
  if (error instanceof OCPPError) {
    return error.userMessage;
  }

  // Fallback for unexpected errors
  return 'An unexpected error occurred. Please try again or contact support.';
}

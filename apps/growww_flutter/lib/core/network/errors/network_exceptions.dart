/// Network exception hierarchy for Growww Flutter Client (Prompt 525)
abstract class NetworkException implements Exception {
  final String message;
  final int? statusCode;
  final String? errorCode;

  const NetworkException(this.message, {this.statusCode, this.errorCode});

  @override
  String toString() => '[$runtimeType]: $message (code: $errorCode, status: $statusCode)';
}

class ApiException extends NetworkException {
  const ApiException(super.message, {super.statusCode, super.errorCode});
}

class UnauthorizedException extends NetworkException {
  const UnauthorizedException(super.message, {super.statusCode = 401, super.errorCode = 'UNAUTHORIZED'});
}

class NetworkConnectionException extends NetworkException {
  const NetworkConnectionException(super.message, {super.statusCode, super.errorCode = 'CONNECTION_FAILED'});
}

class RateLimitException extends NetworkException {
  final int? retryAfterSeconds;

  const RateLimitException(
    super.message, {
    super.statusCode = 429,
    super.errorCode = 'RATE_LIMIT_EXCEEDED',
    this.retryAfterSeconds,
  });
}

class ServerUnavailableException extends NetworkException {
  const ServerUnavailableException(
    super.message, {
    super.statusCode = 503,
    super.errorCode = 'SERVER_UNAVAILABLE',
  });
}

class RiskValidationException extends NetworkException {
  final Map<String, dynamic>? riskDetails;

  const RiskValidationException(
    super.message, {
    super.statusCode = 422,
    super.errorCode = 'RISK_VALIDATION_FAILED',
    this.riskDetails,
  });
}

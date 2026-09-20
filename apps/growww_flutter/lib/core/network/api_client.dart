import 'dart:async';
import 'errors/network_exceptions.dart';
import 'interceptors/auth_token_interceptor.dart';
import 'interceptors/idempotency_interceptor.dart';
import 'interceptors/retry_interceptor.dart';
import 'security/secure_storage_service.dart';

/// Standard response wrapper for API requests
class ApiResponse<T> {
  final int statusCode;
  final T data;
  final Map<String, String> headers;

  const ApiResponse({
    required this.statusCode,
    required this.data,
    this.headers = const {},
  });

  bool get isSuccess => statusCode >= 200 && statusCode < 300;
}

/// Abstract transport adapter allowing mock injection or standard HTTP clients
abstract class HttpTransportAdapter {
  Future<ApiResponse<dynamic>> sendRequest({
    required String method,
    required String path,
    Map<String, String>? headers,
    dynamic body,
  });
}

/// Production API Client for Growww Sovereign Financial Workstation (Prompt 525)
class ApiClient {
  final String baseUrl;
  final HttpTransportAdapter transport;
  final AuthTokenInterceptor authInterceptor;
  final IdempotencyInterceptor idempotencyInterceptor;
  final RetryInterceptor retryInterceptor;

  ApiClient({
    required this.baseUrl,
    required this.transport,
    required ISecureStorageService secureStorage,
    TokenRefreshHandler? refreshHandler,
    IdempotencyInterceptor? idempotency,
    RetryInterceptor? retry,
  })  : authInterceptor = AuthTokenInterceptor(
          secureStorage: secureStorage,
          refreshHandler: refreshHandler,
        ),
        idempotencyInterceptor = idempotency ?? IdempotencyInterceptor(),
        retryInterceptor = retry ?? const RetryInterceptor();

  /// Executes request pipeline: headers processing -> auth token injection -> idempotency -> retries -> error mapping
  Future<ApiResponse<T>> request<T>({
    required String method,
    required String path,
    Map<String, String>? headers,
    dynamic body,
  }) async {
    int attempts = 0;
    while (true) {
      attempts++;
      // 1. Ingest existing and apply idempotency headers
      Map<String, String> processedHeaders = idempotencyInterceptor.processHeaders(method, headers);

      // 2. Attach Authorization Bearer token
      processedHeaders = await authInterceptor.onRequest(processedHeaders);

      try {
        final response = await transport.sendRequest(
          method: method,
          path: path,
          headers: processedHeaders,
          body: body,
        );

        // Handle 401 Unauthorized with atomic mutex token refresh
        if (response.statusCode == 401) {
          final attemptedToken = processedHeaders['Authorization']?.replaceFirst('Bearer ', '');
          final newToken = await authInterceptor.handleUnauthorizedError(attemptedToken: attemptedToken);

          // Retry request with newly refreshed token
          processedHeaders['Authorization'] = 'Bearer $newToken';
          final retriedResponse = await transport.sendRequest(
            method: method,
            path: path,
            headers: processedHeaders,
            body: body,
          );
          return _castResponse<T>(retriedResponse);
        }

        // Handle transient 5xx server errors with retry interceptor
        if (retryInterceptor.isRetryableStatusCode(response.statusCode)) {
          if (attempts <= retryInterceptor.maxRetries) {
            final delay = retryInterceptor.calculateDelay(attempts - 1);
            await Future.delayed(delay);
            continue;
          }
          throw ServerUnavailableException(
            'Server temporarily unavailable after $attempts attempts',
            statusCode: response.statusCode,
          );
        }

        // Check for client-side and business error status codes
        if (response.statusCode >= 400) {
          throw _mapHttpError(response.statusCode, response.data);
        }

        return _castResponse<T>(response);
      } catch (e) {
        if (e is NetworkException) rethrow;
        throw NetworkConnectionException('Network request failed: $e');
      }
    }
  }

  Future<ApiResponse<T>> get<T>(String path, {Map<String, String>? headers}) {
    return request<T>(method: 'GET', path: path, headers: headers);
  }

  Future<ApiResponse<T>> post<T>(String path, {Map<String, String>? headers, dynamic body}) {
    return request<T>(method: 'POST', path: path, headers: headers, body: body);
  }

  Future<ApiResponse<T>> put<T>(String path, {Map<String, String>? headers, dynamic body}) {
    return request<T>(method: 'PUT', path: path, headers: headers, body: body);
  }

  Future<ApiResponse<T>> delete<T>(String path, {Map<String, String>? headers}) {
    return request<T>(method: 'DELETE', path: path, headers: headers);
  }

  ApiResponse<T> _castResponse<T>(ApiResponse<dynamic> resp) {
    return ApiResponse<T>(
      statusCode: resp.statusCode,
      data: resp.data as T,
      headers: resp.headers,
    );
  }

  NetworkException _mapHttpError(int statusCode, dynamic data) {
    String msg = 'Request failed with status $statusCode';
    String? code;
    if (data is Map) {
      msg = data['message']?.toString() ?? msg;
      code = data['error_code']?.toString();
      if (code == 'INSUFFICIENT_MARGIN' || code == 'CIRCUIT_LIMIT_BREACH') {
        return RiskValidationException(msg, statusCode: statusCode, errorCode: code, riskDetails: Map<String, dynamic>.from(data));
      }
    }

    if (statusCode == 401) return UnauthorizedException(msg, statusCode: statusCode, errorCode: code);
    if (statusCode == 429) return RateLimitException(msg, statusCode: statusCode, errorCode: code);
    if (statusCode >= 500) return ServerUnavailableException(msg, statusCode: statusCode, errorCode: code);
    return ApiException(msg, statusCode: statusCode, errorCode: code);
  }
}

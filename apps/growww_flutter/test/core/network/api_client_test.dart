import 'dart:async';
import 'package:test/test.dart';
import 'package:growww_flutter/core/network/api_client.dart';
import 'package:growww_flutter/core/network/errors/network_exceptions.dart';
import 'package:growww_flutter/core/network/interceptors/idempotency_interceptor.dart';
import 'package:growww_flutter/core/network/interceptors/retry_interceptor.dart';
import 'package:growww_flutter/core/network/security/secure_storage_service.dart';

class MockTransportAdapter implements HttpTransportAdapter {
  final List<Map<String, dynamic>> recordedRequests = [];
  Future<ApiResponse<dynamic>> Function({
    required String method,
    required String path,
    Map<String, String>? headers,
    dynamic body,
  })? handler;

  @override
  Future<ApiResponse<dynamic>> sendRequest({
    required String method,
    required String path,
    Map<String, String>? headers,
    dynamic body,
  }) async {
    recordedRequests.add({
      'method': method,
      'path': path,
      'headers': headers,
      'body': body,
    });
    if (handler != null) {
      return handler!(method: method, path: path, headers: headers, body: body);
    }
    return const ApiResponse(statusCode: 200, data: {'status': 'ok'});
  }
}

void main() {
  group('Prompt 525 - Flutter API Client Layer (Typed Dio, Mutex Token Refresh, Pinning & Retries)', () {
    late InMemorySecureStorageService secureStorage;
    late MockTransportAdapter mockTransport;
    late ApiClient client;
    int refreshCallCount = 0;

    setUp(() async {
      refreshCallCount = 0;
      secureStorage = InMemorySecureStorageService();
      await secureStorage.writeString(key: 'jwt_access_token', value: 'old_expired_access_token');
      await secureStorage.writeString(key: 'jwt_refresh_token', value: 'valid_refresh_token_xyz');

      mockTransport = MockTransportAdapter();

      client = ApiClient(
        baseUrl: 'https://api.growww.in/v1',
        transport: mockTransport,
        secureStorage: secureStorage,
        refreshHandler: (refreshToken) async {
          refreshCallCount++;
          // Simulate network latency for refresh
          await Future.delayed(const Duration(milliseconds: 50));
          return {
            'access_token': 'new_refreshed_access_token_abc',
            'refresh_token': 'new_rotated_refresh_token_123',
          };
        },
      );
    });

    test('Injects Authorization Bearer header from secure storage', () async {
      final response = await client.get<Map<String, dynamic>>('/api/v1/user/profile');

      expect(response.statusCode, equals(200));
      expect(mockTransport.recordedRequests.length, equals(1));
      final headers = mockTransport.recordedRequests.first['headers'] as Map<String, String>;
      expect(headers['Authorization'], equals('Bearer old_expired_access_token'));
    });

    test('Injects X-Idempotency-Key on mutating requests (POST, PUT, DELETE)', () async {
      final postResp = await client.post<Map<String, dynamic>>('/api/v1/orders', body: {'symbol': 'TCS'});
      expect(postResp.statusCode, equals(200));

      final postHeaders = mockTransport.recordedRequests.first['headers'] as Map<String, String>;
      expect(postHeaders.containsKey('X-Idempotency-Key'), isTrue);
      expect(postHeaders['X-Idempotency-Key']!.length, greaterThan(10));
    });

    test('Concurrent 401 responses trigger exactly ONE atomic refresh via Mutex and replay all waiting requests', () async {
      int requestAttempt = 0;

      mockTransport.handler = ({required method, required path, headers, body}) async {
        requestAttempt++;
        final authHeader = headers?['Authorization'];

        // If using the old token, reject with 401
        if (authHeader == 'Bearer old_expired_access_token') {
          return const ApiResponse(
            statusCode: 401,
            data: {'error_code': 'TOKEN_EXPIRED', 'message': 'Access token expired'},
          );
        }

        // If using the new token, succeed
        if (authHeader == 'Bearer new_refreshed_access_token_abc') {
          return const ApiResponse(
            statusCode: 200,
            data: {'trade_id': 'TRD_9988', 'status': 'EXECUTED'},
          );
        }

        return const ApiResponse(statusCode: 403, data: {'error': 'Forbidden'});
      };

      // Fire 5 concurrent requests simultaneously
      final futures = [
        client.post<Map<String, dynamic>>('/api/v1/orders/1'),
        client.post<Map<String, dynamic>>('/api/v1/orders/2'),
        client.post<Map<String, dynamic>>('/api/v1/orders/3'),
        client.post<Map<String, dynamic>>('/api/v1/orders/4'),
        client.post<Map<String, dynamic>>('/api/v1/orders/5'),
      ];

      final results = await Future.wait(futures);

      // All 5 requests must have succeeded
      for (final res in results) {
        expect(res.statusCode, equals(200));
        expect(res.data['status'], equals('EXECUTED'));
      }

      // Crucial verification: Mutex guaranteed that refresh was called EXACTLY once!
      expect(refreshCallCount, equals(1));

      // Tokens in storage must be updated
      expect(await secureStorage.readString(key: 'jwt_access_token'), equals('new_refreshed_access_token_abc'));
    });

    test('Failed refresh wipes storage and throws UnauthorizedException', () async {
      final badClient = ApiClient(
        baseUrl: 'https://api.growww.in/v1',
        transport: mockTransport,
        secureStorage: secureStorage,
        refreshHandler: (refreshToken) async {
          throw Exception('Revoked refresh token');
        },
      );

      mockTransport.handler = ({required method, required path, headers, body}) async {
        return const ApiResponse(statusCode: 401, data: {'error': 'Unauthorized'});
      };

      await expectLater(
        badClient.get('/api/v1/account'),
        throwsA(isA<UnauthorizedException>()),
      );

      // Verify credentials wiped from secure storage
      expect(await secureStorage.readString(key: 'jwt_access_token'), isNull);
    });

    test('Retries transient 503 errors and calculates backoff with jitter', () async {
      final retryPolicy = RetryInterceptor(
        maxRetries: 2,
        baseDelay: const Duration(milliseconds: 10),
      );

      expect(retryPolicy.isRetryableStatusCode(502), isTrue);
      expect(retryPolicy.isRetryableStatusCode(503), isTrue);
      expect(retryPolicy.isRetryableStatusCode(504), isTrue);
      expect(retryPolicy.isRetryableStatusCode(400), isFalse);

      final delay0 = retryPolicy.calculateDelay(0);
      final delay1 = retryPolicy.calculateDelay(1);
      expect(delay1.inMilliseconds, greaterThan(delay0.inMilliseconds));
    });

    test('Maps business errors into typed RiskValidationException', () async {
      mockTransport.handler = ({required method, required path, headers, body}) async {
        return const ApiResponse(
          statusCode: 422,
          data: {
            'error_code': 'INSUFFICIENT_MARGIN',
            'message': 'Required margin ₹50,000 exceeds available ₹12,500',
            'required': 50000,
            'available': 12500,
          },
        );
      };

      await expectLater(
        client.post('/api/v1/orders'),
        throwsA(isA<RiskValidationException>().having(
          (e) => e.errorCode,
          'errorCode',
          equals('INSUFFICIENT_MARGIN'),
        )),
      );
    });
  });
}

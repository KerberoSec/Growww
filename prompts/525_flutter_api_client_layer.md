# 525 - Flutter to Backend API Client Layer (Typed Dio, Mutex Token Refresh, Pinning & Retries)

## Purpose
The API client layer is the central circulatory network connecting the Flutter client across all 5 operating systems (Android, iOS, Windows, Linux, macOS) to backend microservices, the API Gateway / BFF (Prompt 219), and market data relays. In high-frequency equity trading, network failures, expired OAuth2 access tokens, and microservice rate limits must be handled deterministically. 

If multiple concurrent network requests encounter an expired access token (HTTP 401), the client must lock a mutual exclusion (Mutex) queue, execute a single atomic refresh token call, replay queued requests with the new token, and prevent cascading session invalidations. Furthermore, SEBI cybersecurity guidelines demand TLS certificate pinning and dynamic idempotency key injections to prevent man-in-the-middle attacks and duplicate trade executions.

This prompt defines the enterprise API Client Layer for the Growww Flutter client. It delivers a typed HTTP/REST client using **Dio** and code-generated **Retrofit** services, integrated mutex-locked token refresh interceptors, exponential backoff retries with full jitter, TLS 1.3 certificate pinning, client-side idempotency injection, and unified domain exception mapping.

## What You Are Building
A production-grade networking layer in `lib/core/network/`:
- `ApiClient`: Configured Dio instance with timeout policies (5s connect, 10s receive), base URL resolution per environment flavor, and custom HTTP adapter options.
- `AuthTokenInterceptor`: Stateful Dio interceptor that injects Bearer JWTs, catches 401 Unauthorized responses, acquires an async Mutex lock, executes `POST /api/v1/auth/refresh`, updates secure storage, and replays all blocked requests.
- `IdempotencyInterceptor`: Request interceptor automatically injecting a unique `X-Idempotency-Key` header on all mutating HTTP requests (`POST`, `PUT`, `DELETE`).
- `RetryInterceptor`: Configurable exponential backoff retry handler with jitter for transient HTTP 502/503/504 errors and socket connection resets.
- `CertificatePinningAdapter`: Multi-platform TLS certificate and public key pinning validator enforcing trusted SHA-256 SPKI hashes for backend domains.
- `ApiExceptionMapper`: Unified mapper transforming raw HTTP status codes, JSON error schemas (`GrowwwErrorResponse`), and socket errors into typed Dart domain exceptions.
- `RetrofitClients`: Code-generated typed API clients for `UserApi`, `OrderApi`, `PortfolioApi`, `WalletApi`, and `MarketDataApi`.

## Scope Boundaries
- **In Scope:**
 - Complete Dio networking setup and configuration.
 - Asynchronous Mutex-locked token refresh lifecycle.
 - TLS certificate pinning and security configurations across mobile and desktop.
 - Request retry policies and timeout orchestration.
 - Idempotency key generation and injection.
 - Type-safe exception mapping and Retrofit service definitions.
- **Out of Scope / Handled Elsewhere:**
 - Backend API Gateway implementation (Prompt 219).
 - WebSockets market data streaming (Prompt 507).
 - Secure storage implementation (Prompt 521).
 - Offline local queuing database (Prompt 515).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `dio` (v5.4+), `retrofit` (v4.1+), and `mutex` (v3.1+).
- **Justification:** Dio provides interceptor chaining, custom connection adapters (`IOClient` / native Darwin/Android engines), and fine-grained cancellation tokens. Retrofit automates boilerplate-free type-safe HTTP client generation.
- **Dependencies:**
 - `dio: ^5.4.3+1`
 - `retrofit: ^4.1.0`
 - `retrofit_generator: ^8.1.0` (dev)
 - `mutex: ^3.1.0`
 - `uuid: ^4.4.0`
 - `flutter_riverpod: ^2.5.1`

## Backend / Infra Touchpoints
- **API Gateway / BFF (Prompt 219):** Unified entrypoint `https://api.growww.in/v1/`.
- **Auth Microservice (Prompt 201 / 105):** `POST /api/v1/auth/refresh`, `POST /api/v1/auth/revoke`.
- **Order Service (Prompt 204):** `POST /api/v1/orders` requiring `X-Idempotency-Key`.

## Blockchain Interaction
- **Ledger RPC Relay Handshake:** Communicates with the secure JSON-RPC relay proxy (Prompt 219) querying block confirmations, Merkle proof branches, and DvP settlement transaction states.

## Step-by-Step Build Instructions
1. Scaffold directory `lib/core/network/` with `interceptors/`, `errors/`, `pinning/`, `api/`, and `models/`.
2. Define domain network exceptions in `errors/network_exceptions.dart`: `ApiException`, `UnauthorizedException`, `NetworkConnectionException`, `RateLimitException`, `ServerUnavailableException`, and `RiskValidationException`.
3. Create `CertificatePinningManager` configuring SPKI SHA-256 public key hashes for `*.growww.in` using `SecurityContext` on `HttpClient`.
4. Initialize `Dio` instance configuring `BaseOptions(connectTimeout: Duration(seconds: 5), receiveTimeout: Duration(seconds: 10), headers: {'Accept': 'application/json', 'X-Client-Platform': Platform.operatingSystem})`.
5. Implement `IdempotencyInterceptor`: check if request is `POST` / `PUT` / `PATCH`; if present, attach `X-Idempotency-Key: uuid.v4()` unless already specified by calling repository.
6. Implement `AuthTokenInterceptor` with `Mutex`:
 - `onRequest`: Retrieve active access token from `ISecureStorageService` (Prompt 521) and attach `Authorization: Bearer <token>`.
 - `onError`: If `response?.statusCode == 401`, intercept error, acquire `_tokenRefreshMutex`, check if token was already refreshed by a peer request; if not, invoke refresh token API. If refresh succeeds, update storage and retry the failed request using `dio.fetch()`. If refresh fails, clear auth state and emit logout event.
7. Implement `RetryInterceptor` with exponential backoff algorithm (`delay = initialDelay * (2 ^ attempt) + randomJitter`) capped at 3 retries for idempotent read requests and specific 5xx status codes.
8. Implement `ApiExceptionMapper` parsing standard JSON error responses: `{"error_code": "INSUFFICIENT_MARGIN", "message": "Required ₹50,000, available ₹12,500", "timestamp": "..."}` into strongly typed exceptions.
9. Define Retrofit interfaces (`@RestApi()`) for core backend services: `OrderRestClient`, `PortfolioRestClient`, `WalletRestClient`, `MarketRestClient`.
10. Run `dart run build_runner build --delete-conflicting-outputs` to generate Retrofit client implementations.
11. Expose `dioProvider` and typed API client providers via Riverpod with proper dependency injection.
12. Write unit tests for `AuthTokenInterceptor` simulating 10 concurrent requests receiving 401 and verifying exactly 1 refresh token call is executed while all 10 requests succeed.
13. Write unit tests for `RetryInterceptor` verifying exponential backoff intervals and jitter distribution.
14. Perform end-to-end integration tests with mock HTTP server verifying SSL pinning and header assertions.

## Interfaces / Contracts

```dart
// lib/core/network/interceptors/auth_token_interceptor.dart
import 'package:dio/dio.dart';
import 'package:mutex/mutex.dart';
import '../../security/storage/domain/secure_storage_service.dart';

class AuthTokenInterceptor extends QueuedInterceptor {
  final ISecureStorageService _secureStorage;
  final Dio _dio;
  final Mutex _refreshMutex = Mutex();

  AuthTokenInterceptor(this._secureStorage, this._dio);

  @override
  Future<void> onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    final token = await _secureStorage.readString(key: 'jwt_access_token');
    if (token != null && !options.headers.containsKey('Authorization')) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401) {
      await _refreshMutex.acquire();
      try {
        final currentToken = await _secureStorage.readString(key: 'jwt_access_token');
        final requestToken = err.requestOptions.headers['Authorization']?.toString().replaceFirst('Bearer ', '');

        // If another thread already refreshed the token, retry immediately
        if (currentToken != null && currentToken != requestToken) {
          err.requestOptions.headers['Authorization'] = 'Bearer $currentToken';
          final response = await _dio.fetch(err.requestOptions);
          return handler.resolve(response);
        }

        // Execute atomic refresh
        final refreshToken = await _secureStorage.readString(key: 'jwt_refresh_token');
        if (refreshToken == null) {
          return handler.next(err);
        }

        final refreshResponse = await _dio.post(
          '/api/v1/auth/refresh',
          data: {'refresh_token': refreshToken},
          options: Options(headers: {'Authorization': null}),
        );

        final newAccessToken = refreshResponse.data['access_token'] as String;
        final newRefreshToken = refreshResponse.data['refresh_token'] as String?;

        await _secureStorage.writeString(key: 'jwt_access_token', value: newAccessToken);
        if (newRefreshToken != null) {
          await _secureStorage.writeString(key: 'jwt_refresh_token', value: newRefreshToken);
        }

        // Retry original failed request
        err.requestOptions.headers['Authorization'] = 'Bearer $newAccessToken';
        final retryResponse = await _dio.fetch(err.requestOptions);
        return handler.resolve(retryResponse);
      } catch (refreshErr) {
        // Refresh failed: wipe credentials and bubble up error
        await _secureStorage.clearAll();
        return handler.next(err);
      } finally {
        _refreshMutex.release();
      }
    }
    handler.next(err);
  }
}
```

```dart
// lib/core/network/api/order_rest_client.dart
import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

part 'order_rest_client.g.dart';

@RestApi()
abstract class OrderRestClient {
  factory OrderRestClient(Dio dio, {String baseUrl}) = _OrderRestClient;

  @POST('/api/v1/orders')
  Future<Map<String, dynamic>> submitOrder(
    @Body() Map<String, dynamic> orderPayload,
    @Header('X-Idempotency-Key') String? idempotencyKey,
  );

  @GET('/api/v1/orders/{orderId}')
  Future<Map<String, dynamic>> getOrderStatus(@Path('orderId') String orderId);

  @DELETE('/api/v1/orders/{orderId}')
  Future<void> cancelOrder(@Path('orderId') String orderId);
}
```

## Security & Compliance Notes
- **TLS Certificate Pinning:** Pinning against SHA-256 Subject Public Key Info (SPKI) prevents MITM proxy attacks on hostile networks. Include primary and backup root/intermediate CA pins to avoid lockouts during certificate rotation.
- **Atomic Token Storage:** Access and refresh tokens are stored exclusively in hardware-backed secure storage (Prompt 521) and never cached in plaintext memory beyond short-lived request scopes.
- **Idempotency Standards:** All order submissions and financial withdrawals MUST generate a non-reusable `X-Idempotency-Key` to prevent duplicate ledger transactions upon network timeouts.

## Acceptance Criteria
- [ ] All network requests seamlessly attach Bearer JWT authorization headers.
- [ ] Concurrent 401 Unauthorized responses trigger exactly one token refresh call via Mutex locking and successfully replay all waiting requests.
- [ ] Failed refresh attempts trigger secure credential cleanup and redirect to login.
- [ ] Mutating HTTP requests automatically receive a unique `X-Idempotency-Key`.
- [ ] Transient 5xx network errors retry with exponential backoff and jitter.
- [ ] Certificate pinning validates authentic server certificates and rejects fraudulent or MITM proxy certificates.
- [ ] Unit and mock tests achieve >90% code coverage across all interceptor pathways.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 219 (API Gateway & BFF), Prompt 103 (API Design Standards).
- **Parallel Tasks:** Prompt 509 (Order Placement Flow), Prompt 515 (Offline Queued Orders), Prompt 526 (Deep Linking).

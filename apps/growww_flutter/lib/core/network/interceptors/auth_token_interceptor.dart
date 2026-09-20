import 'dart:async';
import '../security/secure_storage_service.dart';
import '../errors/network_exceptions.dart';

/// Lightweight asynchronous Mutex lock for Flutter network pipelines
class AsyncMutex {
  Completer<void>? _currentLock;

  Future<void> acquire() async {
    while (_currentLock != null) {
      await _currentLock!.future;
    }
    _currentLock = Completer<void>();
  }

  void release() {
    final lock = _currentLock;
    _currentLock = null;
    lock?.complete();
  }

  Future<T> synchronized<T>(Future<T> Function() action) async {
    await acquire();
    try {
      return await action();
    } finally {
      release();
    }
  }
}

typedef TokenRefreshHandler = Future<Map<String, String>> Function(String refreshToken);

/// Thread-safe JWT interceptor with atomic mutex token refresh
class AuthTokenInterceptor {
  final ISecureStorageService _secureStorage;
  final TokenRefreshHandler? _refreshHandler;
  final AsyncMutex _mutex = AsyncMutex();

  static const String accessTokenKey = 'jwt_access_token';
  static const String refreshTokenKey = 'jwt_refresh_token';

  AuthTokenInterceptor({
    required ISecureStorageService secureStorage,
    TokenRefreshHandler? refreshHandler,
  })  : _secureStorage = secureStorage,
        _refreshHandler = refreshHandler;

  /// Attaches Bearer JWT to outgoing headers if token is present
  Future<Map<String, String>> onRequest(Map<String, String>? headers) async {
    final map = Map<String, String>.from(headers ?? {});
    if (!map.containsKey('Authorization')) {
      final token = await _secureStorage.readString(key: accessTokenKey);
      if (token != null && token.isNotEmpty) {
        map['Authorization'] = 'Bearer $token';
      }
    }
    return map;
  }

  /// Handles 401 Unauthorized by locking mutex, executing atomic refresh, and returning new access token
  Future<String> handleUnauthorizedError({required String? attemptedToken}) async {
    return await _mutex.synchronized<String>(() async {
      final currentToken = await _secureStorage.readString(key: accessTokenKey);

      // If another concurrent request already refreshed the token, return the newly refreshed token
      if (currentToken != null && currentToken != attemptedToken) {
        return currentToken;
      }

      final refreshToken = await _secureStorage.readString(key: refreshTokenKey);
      if (refreshToken == null || refreshToken.isEmpty) {
        await _secureStorage.clearAll();
        throw const UnauthorizedException('Refresh token missing from secure storage');
      }

      if (_refreshHandler == null) {
        await _secureStorage.clearAll();
        throw const UnauthorizedException('No token refresh handler configured');
      }

      try {
        final refreshResult = await _refreshHandler!(refreshToken);
        final newAccessToken = refreshResult['access_token'];
        final newRefreshToken = refreshResult['refresh_token'];

        if (newAccessToken == null) {
          throw const UnauthorizedException('Token refresh response did not contain access_token');
        }

        await _secureStorage.writeString(key: accessTokenKey, value: newAccessToken);
        if (newRefreshToken != null) {
          await _secureStorage.writeString(key: refreshTokenKey, value: newRefreshToken);
        }

        return newAccessToken;
      } catch (err) {
        await _secureStorage.clearAll();
        if (err is NetworkException) rethrow;
        throw UnauthorizedException('Token refresh failed: $err');
      }
    });
  }
}

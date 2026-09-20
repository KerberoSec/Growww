import 'dart:math';

/// Interceptor that attaches unique idempotency keys to mutating requests
class IdempotencyInterceptor {
  static const String idempotencyHeader = 'X-Idempotency-Key';
  final Random _random;

  IdempotencyInterceptor([Random? random]) : _random = random ?? Random.secure();

  /// Generates a RFC-4122 v4 UUID formatted string using secure random bytes
  String generateIdempotencyKey() {
    final bytes = List<int>.generate(16, (_) => _random.nextInt(256));
    // Set version to 4
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    // Set variant to RFC 4122
    bytes[8] = (bytes[8] & 0x3f) | 0x80;

    final hex = bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
    return '${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20, 32)}';
  }

  /// Injects idempotency key into headers if HTTP method is mutating and header not already set
  Map<String, String> processHeaders(String method, Map<String, String>? existingHeaders) {
    final headers = Map<String, String>.from(existingHeaders ?? {});
    final upperMethod = method.toUpperCase();

    if (upperMethod == 'POST' || upperMethod == 'PUT' || upperMethod == 'PATCH' || upperMethod == 'DELETE') {
      if (!headers.containsKey(idempotencyHeader)) {
        headers[idempotencyHeader] = generateIdempotencyKey();
      }
    }

    return headers;
  }
}

import 'dart:math';

/// Retry policy with exponential backoff and randomized jitter (Prompt 525)
class RetryInterceptor {
  final int maxRetries;
  final Duration baseDelay;
  final Random _random;

  const RetryInterceptor({
    this.maxRetries = 3,
    this.baseDelay = const Duration(milliseconds: 200),
    Random? random,
  }) : _random = random ?? const _DefaultRandom();

  /// Determines if an HTTP status code is transient and eligible for retry
  bool isRetryableStatusCode(int statusCode) {
    return statusCode == 502 || statusCode == 503 || statusCode == 504 || statusCode == 408;
  }

  /// Calculates backoff delay for the given attempt index (0-indexed)
  Duration calculateDelay(int attempt) {
    final exponentialFactor = 1 << attempt; // 2^attempt
    final baseMs = baseDelay.inMilliseconds * exponentialFactor;
    // Add jitter between 0% and 30% of baseMs
    final jitterMs = (_random.nextDouble() * 0.3 * baseMs).round();
    return Duration(milliseconds: baseMs + jitterMs);
  }
}

class _DefaultRandom implements Random {
  const _DefaultRandom();

  @override
  bool nextBool() => false;

  @override
  double nextDouble() => 0.15;

  @override
  int nextInt(int max) => max ~/ 2;
}

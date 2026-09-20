import 'dart:convert';
import 'package:crypto/crypto.dart';
import '../models/trade_confirmation_models.dart';

abstract class HapticDriver {
  void trigger(HapticPattern pattern);
}

class MockHapticDriver implements HapticDriver {
  final List<HapticPattern> history = [];

  @override
  void trigger(HapticPattern pattern) {
    history.add(pattern);
  }

  void clear() => history.clear();
}

abstract class LocalBiometricAuthenticator {
  Future<bool> authenticate({required String reason});
}

class MockBiometricAuthenticator implements LocalBiometricAuthenticator {
  bool shouldSucceed = true;

  @override
  Future<bool> authenticate({required String reason}) async {
    return shouldSucceed;
  }
}

class HapticTradeConfirmationService {
  final HapticDriver hapticDriver;
  final LocalBiometricAuthenticator biometricAuthenticator;
  final double swipeThreshold; // e.g. 0.85 (85% of slider width)

  HapticTradeConfirmationService({
    HapticDriver? hapticDriver,
    LocalBiometricAuthenticator? biometricAuthenticator,
    this.swipeThreshold = 0.85,
  })  : hapticDriver = hapticDriver ?? MockHapticDriver(),
        biometricAuthenticator = biometricAuthenticator ?? MockBiometricAuthenticator();

  bool isSwipeThresholdReached(double progress) {
    return progress >= swipeThreshold;
  }

  void onSliderProgressChanged(double previousProgress, double currentProgress) {
    // Light tactile tick every 20% progress
    final prevBucket = (previousProgress * 5).floor();
    final curBucket = (currentProgress * 5).floor();
    if (curBucket > prevBucket && currentProgress < swipeThreshold) {
      hapticDriver.trigger(HapticPattern.lightTap);
    } else if (currentProgress >= swipeThreshold && previousProgress < swipeThreshold) {
      // Threshold reached - medium impact
      hapticDriver.trigger(HapticPattern.mediumImpact);
    }
  }

  Future<String?> confirmAndSignOrder({
    required TradeConfirmationIntent intent,
    required double sliderProgress,
  }) async {
    if (!isSwipeThresholdReached(sliderProgress)) {
      hapticDriver.trigger(HapticPattern.warningBuzz);
      return null;
    }

    if (intent.requireBiometric) {
      final authenticated = await biometricAuthenticator.authenticate(
        reason: 'Confirm ${intent.side.name.toUpperCase()} order for ${intent.symbol}',
      );
      if (!authenticated) {
        hapticDriver.trigger(HapticPattern.warningBuzz);
        return null;
      }
    }

    // Heavy pulse on cryptographic signature finalization
    hapticDriver.trigger(HapticPattern.heavyPulse);

    final digest = intent.computeOrderDigest();
    final signature = hmacSha256(digest, intent.mpcKeyShardId ?? 'local_mpc_shard_default');
    return signature;
  }

  String hmacSha256(String data, String key) {
    final hmac = Hmac(sha256, utf8.encode(key));
    return hmac.convert(utf8.encode(data)).toString();
  }
}

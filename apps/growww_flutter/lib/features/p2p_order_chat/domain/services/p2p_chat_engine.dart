import 'dart:convert';
import 'package:crypto/crypto.dart';
import '../models/p2p_chat_enums.dart';

/// Business logic, anti-scam surveillance, and escrow state machine for P2P trading.
class P2PChatEngine {
  // Prohibited keywords indicating off-platform fraud or escrow bypass attempts
  static const List<String> _suspiciousKeywords = [
    'telegram',
    'whatsapp',
    'call me',
    'paytm wallet direct',
    'direct transfer without escrow',
    'release first',
    'release before payment',
    'bypass escrow',
    'deal outside',
    'insta',
    'discord',
  ];

  /// Validates standard 12-character Indian banking UTR (Unique Transaction Reference).
  static bool validateUtr(String input) {
    final clean = input.trim();
    if (clean.length != 12) return false;
    return RegExp(r'^[0-9A-Za-z]{12}$').hasMatch(clean);
  }

  /// Scans message text for anti-scam keywords.
  /// Returns warning message if flagged, or null if clean.
  static String? detectAntiScamWarning(String message) {
    final lower = message.toLowerCase();
    for (final kw in _suspiciousKeywords) {
      if (lower.contains(kw)) {
        return 'SECURITY ALERT: Message contains flagged term "$kw". Never transact off-platform or release crypto before bank account verification!';
      }
    }
    return null;
  }

  /// Validates if an escrow state transition is legally permissible.
  static bool isValidTransition({
    required P2PEscrowState from,
    required P2PEscrowState to,
    required P2PUserRole role,
  }) {
    if (from.isTerminal) return false;

    switch (to) {
      case P2PEscrowState.paidMarked:
        // Only buyer can mark as paid when payment is pending
        return role == P2PUserRole.buyer && from == P2PEscrowState.paymentPending;

      case P2PEscrowState.released:
        // Seller or Arbitrator can release crypto once marked paid or disputed
        if (role == P2PUserRole.seller || role == P2PUserRole.arbitrator) {
          return from == P2PEscrowState.paidMarked || from == P2PEscrowState.disputed;
        }
        return false;

      case P2PEscrowState.disputed:
        // Buyer, Seller, or Arbitrator can dispute if not terminal
        return from == P2PEscrowState.paidMarked || from == P2PEscrowState.paymentPending;

      case P2PEscrowState.cancelled:
        // Buyer or Arbitrator can cancel if payment is pending (not yet marked paid)
        return (role == P2PUserRole.buyer || role == P2PUserRole.arbitrator) &&
            from == P2PEscrowState.paymentPending;

      default:
        return false;
    }
  }

  /// Generates SHA-256 cryptographic hash of UTR payment proof.
  static String computeProofHash({
    required String orderId,
    required String utr,
    required double amount,
    required DateTime timestamp,
  }) {
    final raw = '$orderId|$utr|$amount|${timestamp.toIso8601String()}';
    final bytes = utf8.encode(raw);
    return sha256.convert(bytes).toString();
  }

  /// Computes remaining seconds until payment window expires.
  static int getRemainingWindowSeconds(DateTime deadline) {
    final now = DateTime.now();
    if (now.isAfter(deadline)) return 0;
    return deadline.difference(now).inSeconds;
  }
}

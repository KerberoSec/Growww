import 'dart:convert';
import 'package:crypto/crypto.dart';

class DeliveredBarDetail {
  final String barSerialNumber;
  final double grossWeightGrams;
  final double fineness;
  final String refineryName;
  final String bisHallmarkNumber;
  final String assayCertificateUrl;

  const DeliveredBarDetail({
    required this.barSerialNumber,
    required this.grossWeightGrams,
    required this.fineness,
    required this.refineryName,
    required this.bisHallmarkNumber,
    required this.assayCertificateUrl,
  });

  Map<String, dynamic> toJson() => {
        'bar_serial_number': barSerialNumber,
        'gross_weight_grams': grossWeightGrams,
        'fineness': fineness,
        'refinery_name': refineryName,
        'bis_hallmark_number': bisHallmarkNumber,
        'assay_certificate_url': assayCertificateUrl,
      };

  factory DeliveredBarDetail.fromJson(Map<String, dynamic> json) {
    return DeliveredBarDetail(
      barSerialNumber: json['bar_serial_number'] as String,
      grossWeightGrams: (json['gross_weight_grams'] as num).toDouble(),
      fineness: (json['fineness'] as num).toDouble(),
      refineryName: json['refinery_name'] as String,
      bisHallmarkNumber: json['bis_hallmark_number'] as String,
      assayCertificateUrl: json['assay_certificate_url'] as String,
    );
  }
}

class DeliveryReceipt {
  final String receiptId;
  final String orderId;
  final String enwrNumber;
  final String repositoryName;
  final String onChainBurnTxHash;
  final List<DeliveredBarDetail> deliveredBars;
  final double totalFineGramsDelivered;
  final String recipientName;
  final String verificationSignature;
  final String deliveryQrPayload;
  final DateTime completedAt;

  const DeliveryReceipt({
    required this.receiptId,
    required this.orderId,
    required this.enwrNumber,
    required this.repositoryName,
    required this.onChainBurnTxHash,
    required this.deliveredBars,
    required this.totalFineGramsDelivered,
    required this.recipientName,
    required this.verificationSignature,
    required this.deliveryQrPayload,
    required this.completedAt,
  });

  /// Computes cryptographic SHA-256 checksum of receipt invariants
  String computeChecksum() {
    final payload = '$receiptId:$orderId:$enwrNumber:$onChainBurnTxHash:$totalFineGramsDelivered:${completedAt.toIso8601String()}';
    return sha256.convert(utf8.encode(payload)).toString();
  }

  bool verifyIntegrity() {
    final checksum = computeChecksum();
    return verificationSignature.startsWith(checksum.substring(0, 16));
  }

  Map<String, dynamic> toJson() => {
        'receipt_id': receiptId,
        'order_id': orderId,
        'enwr_number': enwrNumber,
        'repository_name': repositoryName,
        'on_chain_burn_tx_hash': onChainBurnTxHash,
        'delivered_bars': deliveredBars.map((e) => e.toJson()).toList(),
        'total_fine_grams_delivered': totalFineGramsDelivered,
        'recipient_name': recipientName,
        'verification_signature': verificationSignature,
        'delivery_qr_payload': deliveryQrPayload,
        'completed_at': completedAt.toIso8601String(),
      };

  factory DeliveryReceipt.fromJson(Map<String, dynamic> json) {
    return DeliveryReceipt(
      receiptId: json['receipt_id'] as String,
      orderId: json['order_id'] as String,
      enwrNumber: json['enwr_number'] as String,
      repositoryName: json['repository_name'] as String,
      onChainBurnTxHash: json['on_chain_burn_tx_hash'] as String,
      deliveredBars: (json['delivered_bars'] as List<dynamic>)
          .map((e) => DeliveredBarDetail.fromJson(e as Map<String, dynamic>))
          .toList(),
      totalFineGramsDelivered: (json['total_fine_grams_delivered'] as num).toDouble(),
      recipientName: json['recipient_name'] as String,
      verificationSignature: json['verification_signature'] as String,
      deliveryQrPayload: json['delivery_qr_payload'] as String,
      completedAt: DateTime.parse(json['completed_at'] as String),
    );
  }
}

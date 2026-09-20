import 'p2p_chat_enums.dart';

/// Cryptographic and banking proof for verified Indian Rupee settlement.
class P2PPaymentProof {
  final String proofId;
  final String orderId;
  final String utrNumber; // 12-digit Indian banking UTR (Unique Transaction Reference)
  final P2PPaymentMethod paymentMethod;
  final double paidAmount;
  final String? receiptImageFileName;
  final String proofHash; // SHA-256 integrity hash
  final DateTime submittedAt;
  final bool isConfirmedBySeller;
  final DateTime? confirmedAt;

  const P2PPaymentProof({
    required this.proofId,
    required this.orderId,
    required this.utrNumber,
    required this.paymentMethod,
    required this.paidAmount,
    this.receiptImageFileName,
    required this.proofHash,
    required this.submittedAt,
    this.isConfirmedBySeller = false,
    this.confirmedAt,
  });

  P2PPaymentProof copyWith({
    String? proofId,
    String? orderId,
    String? utrNumber,
    P2PPaymentMethod? paymentMethod,
    double? paidAmount,
    String? receiptImageFileName,
    String? proofHash,
    DateTime? submittedAt,
    bool? isConfirmedBySeller,
    DateTime? confirmedAt,
  }) {
    return P2PPaymentProof(
      proofId: proofId ?? this.proofId,
      orderId: orderId ?? this.orderId,
      utrNumber: utrNumber ?? this.utrNumber,
      paymentMethod: paymentMethod ?? this.paymentMethod,
      paidAmount: paidAmount ?? this.paidAmount,
      receiptImageFileName: receiptImageFileName ?? this.receiptImageFileName,
      proofHash: proofHash ?? this.proofHash,
      submittedAt: submittedAt ?? this.submittedAt,
      isConfirmedBySeller: isConfirmedBySeller ?? this.isConfirmedBySeller,
      confirmedAt: confirmedAt ?? this.confirmedAt,
    );
  }
}

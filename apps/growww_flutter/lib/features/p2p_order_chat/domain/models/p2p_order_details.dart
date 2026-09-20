import 'p2p_chat_enums.dart';

/// Metadata for a P2P trade under Growww Escrow custody.
class P2POrderDetails {
  final String orderId;
  final String cryptoSymbol; // e.g. USDT
  final double cryptoAmount;
  final double fiatUnitPrice; // e.g. 89.50 INR
  final String fiatCurrency; // INR
  final double fiatTotalAmount;
  final P2PEscrowState escrowState;
  final String buyerId;
  final String buyerName;
  final String sellerId;
  final String sellerName;
  final String sellerUpiId;
  final String sellerAccountNumber;
  final String sellerIfsc;
  final String sellerBankName;
  final DateTime createdAt;
  final int paymentWindowMinutes;
  final DateTime paymentDeadline;

  const P2POrderDetails({
    required this.orderId,
    this.cryptoSymbol = 'USDT',
    required this.cryptoAmount,
    required this.fiatUnitPrice,
    this.fiatCurrency = 'INR',
    required this.fiatTotalAmount,
    this.escrowState = P2PEscrowState.paymentPending,
    required this.buyerId,
    required this.buyerName,
    required this.sellerId,
    required this.sellerName,
    required this.sellerUpiId,
    required this.sellerAccountNumber,
    required this.sellerIfsc,
    required this.sellerBankName,
    required this.createdAt,
    this.paymentWindowMinutes = 15,
    required this.paymentDeadline,
  });

  P2POrderDetails copyWith({
    String? orderId,
    String? cryptoSymbol,
    double? cryptoAmount,
    double? fiatUnitPrice,
    String? fiatCurrency,
    double? fiatTotalAmount,
    P2PEscrowState? escrowState,
    String? buyerId,
    String? buyerName,
    String? sellerId,
    String? sellerName,
    String? sellerUpiId,
    String? sellerAccountNumber,
    String? sellerIfsc,
    String? sellerBankName,
    DateTime? createdAt,
    int? paymentWindowMinutes,
    DateTime? paymentDeadline,
  }) {
    return P2POrderDetails(
      orderId: orderId ?? this.orderId,
      cryptoSymbol: cryptoSymbol ?? this.cryptoSymbol,
      cryptoAmount: cryptoAmount ?? this.cryptoAmount,
      fiatUnitPrice: fiatUnitPrice ?? this.fiatUnitPrice,
      fiatCurrency: fiatCurrency ?? this.fiatCurrency,
      fiatTotalAmount: fiatTotalAmount ?? this.fiatTotalAmount,
      escrowState: escrowState ?? this.escrowState,
      buyerId: buyerId ?? this.buyerId,
      buyerName: buyerName ?? this.buyerName,
      sellerId: sellerId ?? this.sellerId,
      sellerName: sellerName ?? this.sellerName,
      sellerUpiId: sellerUpiId ?? this.sellerUpiId,
      sellerAccountNumber: sellerAccountNumber ?? this.sellerAccountNumber,
      sellerIfsc: sellerIfsc ?? this.sellerIfsc,
      sellerBankName: sellerBankName ?? this.sellerBankName,
      createdAt: createdAt ?? this.createdAt,
      paymentWindowMinutes: paymentWindowMinutes ?? this.paymentWindowMinutes,
      paymentDeadline: paymentDeadline ?? this.paymentDeadline,
    );
  }
}

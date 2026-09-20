/// P2P order chat, payment methods, and escrow lifecycle enums.
enum P2PUserRole {
  buyer,
  seller,
  arbitrator;

  String get displayName {
    switch (this) {
      case P2PUserRole.buyer:
        return 'Buyer';
      case P2PUserRole.seller:
        return 'Seller';
      case P2PUserRole.arbitrator:
        return 'Growww Sovereign Arbitrator';
    }
  }
}

enum P2PEscrowState {
  escrowLocked,
  paymentPending,
  paidMarked,
  disputed,
  released,
  cancelled;

  String get displayName {
    switch (this) {
      case P2PEscrowState.escrowLocked:
        return 'Escrow Locked (Secured)';
      case P2PEscrowState.paymentPending:
        return 'Awaiting Buyer Payment';
      case P2PEscrowState.paidMarked:
        return 'Payment Marked by Buyer';
      case P2PEscrowState.disputed:
        return 'In Arbitration / Disputed';
      case P2PEscrowState.released:
        return 'Crypto Released (Completed)';
      case P2PEscrowState.cancelled:
        return 'Order Cancelled';
    }
  }

  bool get isTerminal =>
      this == P2PEscrowState.released || this == P2PEscrowState.cancelled;
}

enum P2PMessageType {
  text,
  systemNotice,
  bankDetails,
  paymentProofReceipt,
  disputeAlert;

  bool get isSystem =>
      this == P2PMessageType.systemNotice || this == P2PMessageType.disputeAlert;
}

enum P2PPaymentMethod {
  upi,
  imps,
  bankTransfer,
  rtgs;

  String get displayName {
    switch (this) {
      case P2PPaymentMethod.upi:
        return 'UPI (Instant)';
      case P2PPaymentMethod.imps:
        return 'IMPS Immediate Payment';
      case P2PPaymentMethod.bankTransfer:
        return 'NEFT / NetBanking';
      case P2PPaymentMethod.rtgs:
        return 'RTGS Real-Time Gross Settlement';
    }
  }
}

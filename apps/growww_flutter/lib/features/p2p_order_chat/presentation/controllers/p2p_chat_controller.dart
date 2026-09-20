import 'dart:async';
import '../../domain/models/p2p_chat_enums.dart';
import '../../domain/models/p2p_chat_message.dart';
import '../../domain/models/p2p_order_details.dart';
import '../../domain/models/p2p_payment_proof.dart';
import '../../domain/services/p2p_chat_engine.dart';

class P2PChatState {
  final P2POrderDetails orderDetails;
  final List<P2PChatMessage> messages;
  final P2PUserRole currentUserRole;
  final String currentUserId;
  final String currentUserName;
  final bool isPaymentDrawerOpen;
  final String utrInput;
  final P2PPaymentMethod selectedPaymentMethod;
  final P2PPaymentProof? paymentProof;
  final bool isSubmitting;
  final int remainingSeconds;
  final String? toastMessage;
  final String? errorMessage;

  const P2PChatState({
    required this.orderDetails,
    this.messages = const [],
    this.currentUserRole = P2PUserRole.buyer,
    this.currentUserId = 'user_buyer_1',
    this.currentUserName = 'Aarav (You)',
    this.isPaymentDrawerOpen = false,
    this.utrInput = '',
    this.selectedPaymentMethod = P2PPaymentMethod.upi,
    this.paymentProof,
    this.isSubmitting = false,
    this.remainingSeconds = 900,
    this.toastMessage,
    this.errorMessage,
  });

  bool get isBuyer => currentUserRole == P2PUserRole.buyer;
  bool get isSeller => currentUserRole == P2PUserRole.seller;

  P2PChatState copyWith({
    P2POrderDetails? orderDetails,
    List<P2PChatMessage>? messages,
    P2PUserRole? currentUserRole,
    String? currentUserId,
    String? currentUserName,
    bool? isPaymentDrawerOpen,
    String? utrInput,
    P2PPaymentMethod? selectedPaymentMethod,
    P2PPaymentProof? paymentProof,
    bool? isSubmitting,
    int? remainingSeconds,
    String? toastMessage,
    String? errorMessage,
  }) {
    return P2PChatState(
      orderDetails: orderDetails ?? this.orderDetails,
      messages: messages ?? this.messages,
      currentUserRole: currentUserRole ?? this.currentUserRole,
      currentUserId: currentUserId ?? this.currentUserId,
      currentUserName: currentUserName ?? this.currentUserName,
      isPaymentDrawerOpen: isPaymentDrawerOpen ?? this.isPaymentDrawerOpen,
      utrInput: utrInput ?? this.utrInput,
      selectedPaymentMethod: selectedPaymentMethod ?? this.selectedPaymentMethod,
      paymentProof: paymentProof ?? this.paymentProof,
      isSubmitting: isSubmitting ?? this.isSubmitting,
      remainingSeconds: remainingSeconds ?? this.remainingSeconds,
      toastMessage: toastMessage,
      errorMessage: errorMessage,
    );
  }
}

class P2PChatController {
  late P2PChatState _state;
  final _stateController = StreamController<P2PChatState>.broadcast();
  Timer? _countdownTimer;

  P2PChatState get state => _state;
  Stream<P2PChatState> get stream => _stateController.stream;

  P2PChatController({
    P2POrderDetails? initialOrder,
    P2PUserRole role = P2PUserRole.buyer,
  }) {
    final order = initialOrder ??
        P2POrderDetails(
          orderId: 'P2P-ORD-94821',
          cryptoSymbol: 'USDT',
          cryptoAmount: 500.0,
          fiatUnitPrice: 89.40,
          fiatCurrency: 'INR',
          fiatTotalAmount: 44700.0,
          escrowState: P2PEscrowState.paymentPending,
          buyerId: 'user_buyer_1',
          buyerName: 'Aarav S.',
          sellerId: 'user_seller_2',
          sellerName: 'Vikas Trading Co.',
          sellerUpiId: 'vikas.traders@okaxis',
          sellerAccountNumber: '918273645012',
          sellerIfsc: 'UTIB0001234',
          sellerBankName: 'Axis Bank Ltd',
          createdAt: DateTime.now(),
          paymentWindowMinutes: 15,
          paymentDeadline: DateTime.now().add(const Duration(minutes: 15)),
        );

    final initialMessages = [
      P2PChatMessage(
        messageId: 'MSG-INIT-1',
        orderId: order.orderId,
        senderId: 'SYSTEM',
        senderName: 'Growww Escrow Bot',
        senderRole: P2PUserRole.arbitrator,
        messageType: P2PMessageType.systemNotice,
        text: 'Escrow secured: 500.00 USDT locked in sovereign smart vault. Pay ₹44,700.00 to seller via verified banking details below.',
        timestamp: DateTime.now().subtract(const Duration(minutes: 1)),
      ),
    ];

    _state = P2PChatState(
      orderDetails: order,
      messages: initialMessages,
      currentUserRole: role,
      currentUserId: role == P2PUserRole.buyer ? order.buyerId : order.sellerId,
      currentUserName: role == P2PUserRole.buyer ? 'Aarav (You)' : 'Vikas (You)',
      remainingSeconds: P2PChatEngine.getRemainingWindowSeconds(order.paymentDeadline),
    );

    _emit(_state);
    _startTimer();
  }

  void _emit(P2PChatState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void _startTimer() {
    _countdownTimer?.cancel();
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (_state.remainingSeconds > 0) {
        _emit(_state.copyWith(remainingSeconds: _state.remainingSeconds - 1));
      } else {
        t.cancel();
      }
    });
  }

  /// Sends a text message with anti-scam surveillance.
  void sendMessage(String text) {
    final clean = text.trim();
    if (clean.isEmpty) return;

    final newMsg = P2PChatMessage(
      messageId: 'MSG-${DateTime.now().millisecondsSinceEpoch}',
      orderId: _state.orderDetails.orderId,
      senderId: _state.currentUserId,
      senderName: _state.currentUserName,
      senderRole: _state.currentUserRole,
      messageType: P2PMessageType.text,
      text: clean,
      timestamp: DateTime.now(),
    );

    final updatedList = List<P2PChatMessage>.from(_state.messages)..add(newMsg);

    // Run anti-scam filter
    final warning = P2PChatEngine.detectAntiScamWarning(clean);
    if (warning != null) {
      final alertMsg = P2PChatMessage(
        messageId: 'MSG-ALERT-${DateTime.now().millisecondsSinceEpoch}',
        orderId: _state.orderDetails.orderId,
        senderId: 'SYSTEM',
        senderName: 'Anti-Scam Sentinel',
        senderRole: P2PUserRole.arbitrator,
        messageType: P2PMessageType.disputeAlert,
        text: warning,
        timestamp: DateTime.now(),
      );
      updatedList.add(alertMsg);
    }

    _emit(_state.copyWith(messages: updatedList, errorMessage: null));
  }

  void openPaymentDrawer() {
    _emit(_state.copyWith(isPaymentDrawerOpen: true));
  }

  void closePaymentDrawer() {
    _emit(_state.copyWith(isPaymentDrawerOpen: false));
  }

  void setUtrInput(String utr) {
    _emit(_state.copyWith(utrInput: utr.trim().toUpperCase(), errorMessage: null));
  }

  void setPaymentMethod(P2PPaymentMethod method) {
    _emit(_state.copyWith(selectedPaymentMethod: method));
  }

  /// Buyer marks payment as complete with valid Indian Banking 12-digit UTR.
  bool submitPaymentProof({String? receiptFileName}) {
    final utr = _state.utrInput.trim();
    if (!P2PChatEngine.validateUtr(utr)) {
      _emit(_state.copyWith(
        errorMessage: 'Invalid UTR reference. Must be exactly 12 alphanumeric characters.',
      ));
      return false;
    }

    final isValid = P2PChatEngine.isValidTransition(
      from: _state.orderDetails.escrowState,
      to: P2PEscrowState.paidMarked,
      role: _state.currentUserRole,
    );

    if (!isValid) {
      _emit(_state.copyWith(errorMessage: 'Invalid action for current escrow state.'));
      return false;
    }

    final now = DateTime.now();
    final hash = P2PChatEngine.computeProofHash(
      orderId: _state.orderDetails.orderId,
      utr: utr,
      amount: _state.orderDetails.fiatTotalAmount,
      timestamp: now,
    );

    final proof = P2PPaymentProof(
      proofId: 'PROOF-${now.millisecondsSinceEpoch}',
      orderId: _state.orderDetails.orderId,
      utrNumber: utr,
      paymentMethod: _state.selectedPaymentMethod,
      paidAmount: _state.orderDetails.fiatTotalAmount,
      receiptImageFileName: receiptFileName ?? 'upi_payment_receipt.png',
      proofHash: hash,
      submittedAt: now,
    );

    final updatedOrder = _state.orderDetails.copyWith(
      escrowState: P2PEscrowState.paidMarked,
    );

    final proofMsg = P2PChatMessage(
      messageId: 'MSG-PROOF-${now.millisecondsSinceEpoch}',
      orderId: updatedOrder.orderId,
      senderId: _state.currentUserId,
      senderName: _state.currentUserName,
      senderRole: _state.currentUserRole,
      messageType: P2PMessageType.paymentProofReceipt,
      text: 'Payment marked! UTR: $utr via ${_state.selectedPaymentMethod.displayName}. Amount: ₹${updatedOrder.fiatTotalAmount}',
      utrNumber: utr,
      attachmentName: proof.receiptImageFileName,
      timestamp: now,
    );

    final updatedMessages = List<P2PChatMessage>.from(_state.messages)..add(proofMsg);

    _emit(_state.copyWith(
      orderDetails: updatedOrder,
      paymentProof: proof,
      messages: updatedMessages,
      isPaymentDrawerOpen: false,
      toastMessage: 'Payment verified and marked. Waiting for seller to release crypto.',
      errorMessage: null,
    ));

    return true;
  }

  /// Seller releases crypto from escrow custody.
  bool confirmReleaseCrypto() {
    final isValid = P2PChatEngine.isValidTransition(
      from: _state.orderDetails.escrowState,
      to: P2PEscrowState.released,
      role: _state.currentUserRole,
    );

    if (!isValid) {
      _emit(_state.copyWith(
        errorMessage: 'Cannot release crypto until payment is marked and verified.',
      ));
      return false;
    }

    final updatedOrder = _state.orderDetails.copyWith(
      escrowState: P2PEscrowState.released,
    );

    final releaseMsg = P2PChatMessage(
      messageId: 'MSG-REL-${DateTime.now().millisecondsSinceEpoch}',
      orderId: updatedOrder.orderId,
      senderId: 'SYSTEM',
      senderName: 'Growww Settlement Vault',
      senderRole: P2PUserRole.arbitrator,
      messageType: P2PMessageType.systemNotice,
      text: 'Order Completed! 500.00 USDT has been released from escrow into Buyer wallet.',
      timestamp: DateTime.now(),
    );

    final updatedMessages = List<P2PChatMessage>.from(_state.messages)..add(releaseMsg);

    _emit(_state.copyWith(
      orderDetails: updatedOrder,
      messages: updatedMessages,
      toastMessage: 'Crypto released successfully to buyer.',
      errorMessage: null,
    ));

    return true;
  }

  /// Dispatches arbitration dispute if counterparty acts fraudulently.
  void raiseDispute(String reason) {
    final updatedOrder = _state.orderDetails.copyWith(
      escrowState: P2PEscrowState.disputed,
    );

    final disputeMsg = P2PChatMessage(
      messageId: 'MSG-DISP-${DateTime.now().millisecondsSinceEpoch}',
      orderId: updatedOrder.orderId,
      senderId: _state.currentUserId,
      senderName: _state.currentUserName,
      senderRole: _state.currentUserRole,
      messageType: P2PMessageType.disputeAlert,
      text: 'DISPUTE RAISED: "$reason". Growww Sovereign Arbitration team notified with order evidence.',
      timestamp: DateTime.now(),
    );

    final updatedMessages = List<P2PChatMessage>.from(_state.messages)..add(disputeMsg);

    _emit(_state.copyWith(
      orderDetails: updatedOrder,
      messages: updatedMessages,
      toastMessage: 'Dispute opened. Arbitrator assigned.',
    ));
  }

  void dispose() {
    _countdownTimer?.cancel();
    _stateController.close();
  }
}

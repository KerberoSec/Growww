import 'package:test/test.dart';
import 'package:growww_flutter/features/p2p_order_chat/domain/models/p2p_chat_enums.dart';
import 'package:growww_flutter/features/p2p_order_chat/domain/models/p2p_order_details.dart';
import 'package:growww_flutter/features/p2p_order_chat/domain/services/p2p_chat_engine.dart';
import 'package:growww_flutter/features/p2p_order_chat/presentation/controllers/p2p_chat_controller.dart';

void main() {
  group('Prompt 543 - P2P Order Chat & Payment Verification Drawer', () {
    late P2POrderDetails defaultOrder;
    late P2PChatController buyerController;
    late P2PChatController sellerController;

    setUp(() {
      defaultOrder = P2POrderDetails(
        orderId: 'P2P-TEST-100',
        cryptoSymbol: 'USDT',
        cryptoAmount: 500.0,
        fiatUnitPrice: 89.40,
        fiatCurrency: 'INR',
        fiatTotalAmount: 44700.0,
        escrowState: P2PEscrowState.paymentPending,
        buyerId: 'buyer_1',
        buyerName: 'Aarav S.',
        sellerId: 'seller_2',
        sellerName: 'Vikas Trading',
        sellerUpiId: 'vikas.traders@okaxis',
        sellerAccountNumber: '918273645012',
        sellerIfsc: 'UTIB0001234',
        sellerBankName: 'Axis Bank',
        createdAt: DateTime.now(),
        paymentWindowMinutes: 15,
        paymentDeadline: DateTime.now().add(const Duration(minutes: 15)),
      );

      buyerController = P2PChatController(
        initialOrder: defaultOrder,
        role: P2PUserRole.buyer,
      );

      sellerController = P2PChatController(
        initialOrder: defaultOrder,
        role: P2PUserRole.seller,
      );
    });

    tearDown(() {
      buyerController.dispose();
      sellerController.dispose();
    });

    test('P2PChatEngine validates 12-character UTR and flags off-platform scam keywords', () {
      // Valid UTR
      expect(P2PChatEngine.validateUtr('427189012345'), isTrue);
      expect(P2PChatEngine.validateUtr('AXIS12345678'), isTrue);

      // Invalid UTR
      expect(P2PChatEngine.validateUtr('12345'), isFalse);
      expect(P2PChatEngine.validateUtr('1234567890123'), isFalse);
      expect(P2PChatEngine.validateUtr('1234 5678 90'), isFalse);

      // Anti-scam surveillance filter
      expect(P2PChatEngine.detectAntiScamWarning('Hello, I have made UPI transfer'), isNull);
      expect(
        P2PChatEngine.detectAntiScamWarning('Message me on Telegram @fast_crypto'),
        contains('SECURITY ALERT'),
      );
      expect(
        P2PChatEngine.detectAntiScamWarning('Send on WhatsApp for direct transfer without escrow'),
        contains('SECURITY ALERT'),
      );
      expect(
        P2PChatEngine.detectAntiScamWarning('Please release first bro'),
        contains('SECURITY ALERT'),
      );
    });

    test('P2PChatEngine enforces strict escrow state machine transitions', () {
      // Buyer marking paid from paymentPending -> Allowed
      expect(
        P2PChatEngine.isValidTransition(
          from: P2PEscrowState.paymentPending,
          to: P2PEscrowState.paidMarked,
          role: P2PUserRole.buyer,
        ),
        isTrue,
      );

      // Seller attempting to mark paid -> Disallowed
      expect(
        P2PChatEngine.isValidTransition(
          from: P2PEscrowState.paymentPending,
          to: P2PEscrowState.paidMarked,
          role: P2PUserRole.seller,
        ),
        isFalse,
      );

      // Seller releasing crypto when paymentPending (unpaid) -> Disallowed
      expect(
        P2PChatEngine.isValidTransition(
          from: P2PEscrowState.paymentPending,
          to: P2PEscrowState.released,
          role: P2PUserRole.seller,
        ),
        isFalse,
      );

      // Seller releasing crypto after paidMarked -> Allowed
      expect(
        P2PChatEngine.isValidTransition(
          from: P2PEscrowState.paidMarked,
          to: P2PEscrowState.released,
          role: P2PUserRole.seller,
        ),
        isTrue,
      );

      // Transitioning out of terminal state released -> Disallowed
      expect(
        P2PChatEngine.isValidTransition(
          from: P2PEscrowState.released,
          to: P2PEscrowState.cancelled,
          role: P2PUserRole.arbitrator,
        ),
        isFalse,
      );
    });

    test('P2PChatController handles messaging and automatically dispatches scam alert notices', () {
      final initialCount = buyerController.state.messages.length;

      // Clean message
      buyerController.sendMessage('I am sending ₹44,700 via UPI now');
      expect(buyerController.state.messages.length, equals(initialCount + 1));
      expect(buyerController.state.messages.last.text, contains('I am sending ₹44,700'));

      // Suspicious message with flagged keyword
      buyerController.sendMessage('Contact me on WhatsApp: +919876543210');
      // Should add both the message AND an anti-scam alert message
      expect(buyerController.state.messages.length, equals(initialCount + 3));
      expect(buyerController.state.messages.last.messageType, equals(P2PMessageType.disputeAlert));
      expect(buyerController.state.messages.last.text, contains('SECURITY ALERT'));
    });

    test('P2PChatController manages payment proof submission and seller release lifecycle', () {
      expect(buyerController.state.orderDetails.escrowState, equals(P2PEscrowState.paymentPending));

      // Open and populate payment drawer
      buyerController.openPaymentDrawer();
      expect(buyerController.state.isPaymentDrawerOpen, isTrue);

      // Invalid UTR rejection
      buyerController.setUtrInput('12345');
      final failSubmit = buyerController.submitPaymentProof();
      expect(failSubmit, isFalse);
      expect(buyerController.state.errorMessage, contains('Invalid UTR'));

      // Valid 12-digit UTR submission
      buyerController.setUtrInput('427189012345');
      final okSubmit = buyerController.submitPaymentProof();
      expect(okSubmit, isTrue);
      expect(buyerController.state.isPaymentDrawerOpen, isFalse);
      expect(buyerController.state.orderDetails.escrowState, equals(P2PEscrowState.paidMarked));
      expect(buyerController.state.paymentProof, isNotNull);
      expect(buyerController.state.paymentProof?.utrNumber, equals('427189012345'));
      expect(buyerController.state.messages.last.messageType, equals(P2PMessageType.paymentProofReceipt));

      // Synchronize order to seller controller
      final markedOrder = buyerController.state.orderDetails;
      sellerController = P2PChatController(initialOrder: markedOrder, role: P2PUserRole.seller);

      // Seller confirms bank credit and releases crypto
      final releaseOk = sellerController.confirmReleaseCrypto();
      expect(releaseOk, isTrue);
      expect(sellerController.state.orderDetails.escrowState, equals(P2PEscrowState.released));
      expect(sellerController.state.messages.last.text, contains('500.00 USDT has been released'));
    });

    test('P2PChatController supports opening arbitration dispute', () {
      buyerController.raiseDispute('Seller claims payment not received after bank debit');
      expect(buyerController.state.orderDetails.escrowState, equals(P2PEscrowState.disputed));
      expect(buyerController.state.messages.last.messageType, equals(P2PMessageType.disputeAlert));
      expect(buyerController.state.messages.last.text, contains('DISPUTE RAISED'));
    });
  });
}

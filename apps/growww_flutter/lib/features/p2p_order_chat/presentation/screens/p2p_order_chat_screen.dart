import 'package:flutter/material.dart';
import '../../domain/models/p2p_chat_enums.dart';
import '../../domain/models/p2p_chat_message.dart';
import '../controllers/p2p_chat_controller.dart';

/// Flutter P2P Order Chat & Payment Verification Drawer Screen
/// Institutional sovereign P2P trading workstation with encrypted escrow messaging,
/// automated anti-scam surveillance, 12-digit UTR validation, and sliding settlement drawer.
class P2POrderChatScreen extends StatefulWidget {
  final P2PChatController controller;

  const P2POrderChatScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<P2POrderChatScreen> createState() => _P2POrderChatScreenState();
}

class _P2POrderChatScreenState extends State<P2POrderChatScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonCyan = Color(0xFF00E5FF);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color neonYellow = Color(0xFFFFD166);
  static const Color textMuted = Color(0xFF8B949E);

  final TextEditingController _msgInputController = TextEditingController();
  final TextEditingController _utrInputController = TextEditingController();

  @override
  void dispose() {
    _msgInputController.dispose();
    _utrInputController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<P2PChatState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        final order = state.orderDetails;

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            leading: IconButton(
              icon: const Icon(Icons.arrow_back_ios, color: Colors.white, size: 18),
              onPressed: () => Navigator.of(context).maybePop(),
            ),
            title: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      order.orderId,
                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                    ),
                    const SizedBox(width: 8),
                    _buildEscrowBadge(order.escrowState),
                  ],
                ),
                Text(
                  'Counterparty: ${state.isBuyer ? order.sellerName : order.buyerName}',
                  style: const TextStyle(color: textMuted, fontSize: 11),
                ),
              ],
            ),
            actions: [
              Container(
                alignment: Alignment.center,
                padding: const EdgeInsets.only(right: 16),
                child: Text(
                  _formatCountdown(state.remainingSeconds),
                  style: const TextStyle(
                    color: neonYellow,
                    fontFamily: 'JetBrains Mono',
                    fontWeight: FontWeight.bold,
                    fontSize: 13,
                  ),
                ),
              ),
            ],
          ),
          body: Stack(
            children: [
              Column(
                children: [
                  _buildOrderBanner(state),
                  Expanded(
                    child: ListView.builder(
                      padding: const EdgeInsets.all(12),
                      itemCount: state.messages.length,
                      itemBuilder: (context, index) {
                        final msg = state.messages[index];
                        return _buildChatMessageBubble(msg, state);
                      },
                    ),
                  ),
                  if (state.errorMessage != null)
                    _buildErrorNotification(state.errorMessage!),
                  _buildChatBottomBar(state),
                ],
              ),
              if (state.isPaymentDrawerOpen)
                _buildPaymentVerificationDrawer(state),
            ],
          ),
        );
      },
    );
  }

  Widget _buildEscrowBadge(P2PEscrowState state) {
    Color bg;
    Color fg;

    switch (state) {
      case P2PEscrowState.escrowLocked:
      case P2PEscrowState.paymentPending:
        bg = neonCyan.withOpacity(0.15);
        fg = neonCyan;
        break;
      case P2PEscrowState.paidMarked:
        bg = neonYellow.withOpacity(0.15);
        fg = neonYellow;
        break;
      case P2PEscrowState.released:
        bg = neonGreen.withOpacity(0.15);
        fg = neonGreen;
        break;
      case P2PEscrowState.disputed:
      case P2PEscrowState.cancelled:
        bg = neonRed.withOpacity(0.15);
        fg = neonRed;
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        state.displayName,
        style: TextStyle(color: fg, fontSize: 10, fontWeight: FontWeight.bold),
      ),
    );
  }

  Widget _buildOrderBanner(P2PChatState state) {
    final o = state.orderDetails;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      color: surfaceCard,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Total Fiat to Pay / Receive', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '₹${o.fiatTotalAmount.toStringAsFixed(2)}',
                style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 16),
              ),
            ],
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              const Text('Secured Crypto', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '${o.cryptoAmount.toStringAsFixed(2)} ${o.cryptoSymbol}',
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildChatMessageBubble(P2PChatMessage msg, P2PChatState state) {
    final isMe = msg.isSentBy(state.currentUserId);

    if (msg.messageType.isSystem) {
      final isDispute = msg.messageType == P2PMessageType.disputeAlert;
      return Container(
        margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isDispute ? neonRed.withOpacity(0.15) : neonCyan.withOpacity(0.1),
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: isDispute ? neonRed : neonCyan.withOpacity(0.4)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(isDispute ? Icons.warning_amber_rounded : Icons.shield_outlined,
                color: isDispute ? neonRed : neonCyan, size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(msg.senderName, style: TextStyle(color: isDispute ? neonRed : neonCyan, fontWeight: FontWeight.bold, fontSize: 11)),
                  const SizedBox(height: 2),
                  Text(msg.text, style: const TextStyle(color: Colors.white, fontSize: 12)),
                ],
              ),
            ),
          ],
        ),
      );
    }

    if (msg.messageType == P2PMessageType.paymentProofReceipt) {
      return Container(
        margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: neonGreen.withOpacity(0.12),
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: neonGreen.withOpacity(0.5)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: const [
                Icon(Icons.receipt_long, color: neonGreen, size: 18),
                SizedBox(width: 6),
                Text('PAYMENT RECEIPT MARKED', style: TextStyle(color: neonGreen, fontWeight: FontWeight.bold, fontSize: 11)),
              ],
            ),
            const SizedBox(height: 6),
            Text('UTR Ref: ${msg.utrNumber}', style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13)),
            Text(msg.text, style: const TextStyle(color: textMuted, fontSize: 11)),
          ],
        ),
      );
    }

    return Align(
      alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        constraints: const BoxConstraints(maxWidth: 280),
        decoration: BoxDecoration(
          color: isMe ? const Color(0xFF1B382B) : surfaceCard,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: isMe ? neonGreen.withOpacity(0.3) : Colors.white10),
        ),
        child: Column(
          crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            Text(
              msg.senderName,
              style: TextStyle(color: isMe ? neonGreen : textMuted, fontSize: 10, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 4),
            Text(msg.text, style: const TextStyle(color: Colors.white, fontSize: 13)),
          ],
        ),
      ),
    );
  }

  Widget _buildChatBottomBar(P2PChatState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      color: surfaceCard,
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _msgInputController,
                  style: const TextStyle(color: Colors.white, fontSize: 13),
                  decoration: InputDecoration(
                    hintText: 'Type message to ${state.isBuyer ? 'Seller' : 'Buyer'}...',
                    hintStyle: const TextStyle(color: textMuted),
                    filled: true,
                    fillColor: obsidianBackground,
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(6), borderSide: BorderSide.none),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.send, color: neonGreen, size: 20),
                onPressed: () {
                  widget.controller.sendMessage(_msgInputController.text);
                  _msgInputController.clear();
                },
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              if (state.isBuyer && state.orderDetails.escrowState == P2PEscrowState.paymentPending)
                Expanded(
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: neonGreen,
                      foregroundColor: Colors.black,
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                    ),
                    icon: const Icon(Icons.check_circle_outline, size: 16),
                    label: const Text('I Have Paid / Submit UTR', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                    onPressed: () => widget.controller.openPaymentDrawer(),
                  ),
                ),
              if (state.isSeller && state.orderDetails.escrowState == P2PEscrowState.paidMarked)
                Expanded(
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: neonGreen,
                      foregroundColor: Colors.black,
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                    ),
                    icon: const Icon(Icons.lock_open, size: 16),
                    label: const Text('Confirm Receipt & Release Crypto', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                    onPressed: () => widget.controller.confirmReleaseCrypto(),
                  ),
                ),
              const SizedBox(width: 8),
              OutlinedButton(
                style: OutlinedButton.styleFrom(
                  foregroundColor: neonRed,
                  side: const BorderSide(color: neonRed),
                  padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                ),
                onPressed: () => widget.controller.raiseDispute('Payment dispute or non-responsive counterparty'),
                child: const Text('Dispute', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentVerificationDrawer(P2PChatState state) {
    final order = state.orderDetails;

    return Container(
      color: Colors.black87,
      child: Align(
        alignment: Alignment.bottomCenter,
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: const BoxDecoration(
            color: surfaceCard,
            borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
            boxShadow: [BoxShadow(color: Colors.black, blurRadius: 20)],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Payment Verification Drawer', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 15)),
                  IconButton(
                    icon: const Icon(Icons.close, color: textMuted),
                    onPressed: () => widget.controller.closePaymentDrawer(),
                  ),
                ],
              ),
              const Divider(color: Colors.white12),
              _buildBankingDetailRow('Seller UPI ID', order.sellerUpiId),
              _buildBankingDetailRow('Account No.', order.sellerAccountNumber),
              _buildBankingDetailRow('Bank / IFSC', '${order.sellerBankName} (${order.sellerIfsc})'),
              const SizedBox(height: 14),
              const Text('Enter 12-Digit Indian Banking UTR Number', style: TextStyle(color: textMuted, fontSize: 12)),
              const SizedBox(height: 6),
              TextField(
                controller: _utrInputController,
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 15, letterSpacing: 2),
                maxLength: 12,
                decoration: InputDecoration(
                  hintText: '427189012345',
                  hintStyle: const TextStyle(color: Colors.white24),
                  counterText: '',
                  filled: true,
                  fillColor: obsidianBackground,
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(6), borderSide: BorderSide.none),
                ),
                onChanged: (val) => widget.controller.setUtrInput(val),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: neonGreen,
                  foregroundColor: Colors.black,
                  minimumSize: const Size.fromHeight(46),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                ),
                onPressed: () => widget.controller.submitPaymentProof(),
                child: const Text('Confirm Payment & Notify Seller', style: TextStyle(fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBankingDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: textMuted, fontSize: 12)),
          Text(
            value,
            style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 12),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorNotification(String error) {
    return Container(
      padding: const EdgeInsets.all(8),
      color: neonRed.withOpacity(0.2),
      child: Center(
        child: Text(error, style: const TextStyle(color: neonRed, fontSize: 11, fontWeight: FontWeight.bold)),
      ),
    );
  }

  String _formatCountdown(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
  }
}

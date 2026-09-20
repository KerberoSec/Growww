import 'p2p_chat_enums.dart';

/// Single encrypted message in the P2P order chat stream.
class P2PChatMessage {
  final String messageId;
  final String orderId;
  final String senderId;
  final String senderName;
  final P2PUserRole senderRole;
  final P2PMessageType messageType;
  final String text;
  final String? utrNumber;
  final String? attachmentName;
  final DateTime timestamp;

  const P2PChatMessage({
    required this.messageId,
    required this.orderId,
    required this.senderId,
    required this.senderName,
    required this.senderRole,
    required this.messageType,
    required this.text,
    this.utrNumber,
    this.attachmentName,
    required this.timestamp,
  });

  bool isSentBy(String userId) => senderId == userId;

  P2PChatMessage copyWith({
    String? messageId,
    String? orderId,
    String? senderId,
    String? senderName,
    P2PUserRole? senderRole,
    P2PMessageType? messageType,
    String? text,
    String? utrNumber,
    String? attachmentName,
    DateTime? timestamp,
  }) {
    return P2PChatMessage(
      messageId: messageId ?? this.messageId,
      orderId: orderId ?? this.orderId,
      senderId: senderId ?? this.senderId,
      senderName: senderName ?? this.senderName,
      senderRole: senderRole ?? this.senderRole,
      messageType: messageType ?? this.messageType,
      text: text ?? this.text,
      utrNumber: utrNumber ?? this.utrNumber,
      attachmentName: attachmentName ?? this.attachmentName,
      timestamp: timestamp ?? this.timestamp,
    );
  }
}

import 'commodity_enums.dart';

class LogisticsTrackingEvent {
  final String eventId;
  final String orderId;
  final LogisticsCarrier carrier;
  final String trackingNumber;
  final RedemptionStatus status;
  final String statusDescription;
  final String locationCity;
  final double? currentLatitude;
  final double? currentLongitude;
  final String securityBagSealNumber;
  final String? armoredEscortId;
  final DateTime timestamp;

  const LogisticsTrackingEvent({
    required this.eventId,
    required this.orderId,
    required this.carrier,
    required this.trackingNumber,
    required this.status,
    required this.statusDescription,
    required this.locationCity,
    this.currentLatitude,
    this.currentLongitude,
    required this.securityBagSealNumber,
    this.armoredEscortId,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
        'event_id': eventId,
        'order_id': orderId,
        'carrier': carrier.name,
        'tracking_number': trackingNumber,
        'status': status.name,
        'status_description': statusDescription,
        'location_city': locationCity,
        'current_latitude': currentLatitude,
        'current_longitude': currentLongitude,
        'security_bag_seal_number': securityBagSealNumber,
        'armored_escort_id': armoredEscortId,
        'timestamp': timestamp.toIso8601String(),
      };

  factory LogisticsTrackingEvent.fromJson(Map<String, dynamic> json) {
    return LogisticsTrackingEvent(
      eventId: json['event_id'] as String,
      orderId: json['order_id'] as String,
      carrier: LogisticsCarrier.values.byName(json['carrier'] as String),
      trackingNumber: json['tracking_number'] as String,
      status: RedemptionStatus.values.byName(json['status'] as String),
      statusDescription: json['status_description'] as String,
      locationCity: json['location_city'] as String,
      currentLatitude: json['current_latitude'] != null ? (json['current_latitude'] as num).toDouble() : null,
      currentLongitude: json['current_longitude'] != null ? (json['current_longitude'] as num).toDouble() : null,
      securityBagSealNumber: json['security_bag_seal_number'] as String,
      armoredEscortId: json['armored_escort_id'] as String?,
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }
}

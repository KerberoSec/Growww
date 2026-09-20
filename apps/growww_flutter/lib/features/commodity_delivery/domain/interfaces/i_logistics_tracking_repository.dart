import '../models/logistics_tracking_event.dart';

abstract class ILogisticsTrackingRepository {
  Future<List<LogisticsTrackingEvent>> fetchTrackingHistory({
    required String orderId,
  });

  Stream<LogisticsTrackingEvent> streamLiveTrackingEvents({
    required String orderId,
  });

  Future<String> generateTimeLockedDeliveryOtp({
    required String orderId,
  });
}

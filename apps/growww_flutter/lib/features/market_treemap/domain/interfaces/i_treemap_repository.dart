import '../models/treemap_enums.dart';
import '../models/treemap_node.dart';
import '../models/treemap_tick_update.dart';

abstract class ITreemapRepository {
  Future<SectorGroupNode> fetchTreemapSnapshot({
    required String indexCode,
    required TreemapMetricType metricType,
    required TreemapTimeHorizon timeHorizon,
  });

  Stream<List<TreemapTickUpdate>> streamTreemapTicks({
    required String indexCode,
  });

  Future<List<String>> fetchAvailableIndices();
}

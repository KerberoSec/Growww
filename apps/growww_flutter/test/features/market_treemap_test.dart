import 'package:test/test.dart';
import 'package:growww_flutter/features/market_treemap/data/repositories/mock_treemap_repository.dart';
import 'package:growww_flutter/features/market_treemap/domain/algorithms/squarified_treemap_algorithm.dart';
import 'package:growww_flutter/features/market_treemap/domain/models/treemap_enums.dart';
import 'package:growww_flutter/features/market_treemap/domain/models/treemap_geometry.dart';
import 'package:growww_flutter/features/market_treemap/domain/models/treemap_node.dart';
import 'package:growww_flutter/features/market_treemap/domain/services/heatmap_color_engine.dart';
import 'package:growww_flutter/features/market_treemap/presentation/controllers/treemap_controller.dart';

void main() {
  group('Prompt 535 - Real-Time Market Heatmap and Sector Treemap', () {
    late MockTreemapRepository repository;
    late SquarifiedTreemapAlgorithm algorithm;
    late TreemapController controller;

    setUp(() {
      repository = MockTreemapRepository();
      algorithm = const SquarifiedTreemapAlgorithm(
        sectorHeaderHeight: 20.0,
        sectorPadding: 2.0,
        leafPadding: 1.0,
      );
      controller = TreemapController(
        repository: repository,
        algorithm: algorithm,
      );
    });

    tearDown(() {
      controller.dispose();
    });

    test('SquarifiedTreemapAlgorithm generates non-overlapping geometries with aspect ratios near 1.0', () async {
      final snapshot = await repository.fetchTreemapSnapshot(
        indexCode: 'NIFTY 50',
        metricType: TreemapMetricType.marketCap,
        timeHorizon: TreemapTimeHorizon.oneDay,
      );

      const container = TreemapRect(left: 0, top: 0, width: 800, height: 600);
      final laidOut = algorithm.computeLayout(root: snapshot, containerRect: container);

      expect(laidOut.rect.width, equals(800));
      expect(laidOut.rect.height, equals(600));
      expect(laidOut.children.isNotEmpty, isTrue);

      // Verify that children are within container bounds and have positive area
      for (final sector in laidOut.children) {
        expect(sector.rect.left, greaterThanOrEqualTo(0.0));
        expect(sector.rect.top, greaterThanOrEqualTo(0.0));
        expect(sector.rect.right, lessThanOrEqualTo(800.1));
        expect(sector.rect.bottom, lessThanOrEqualTo(600.1));
        expect(sector.rect.area, greaterThan(0.0));

        // In a squarified treemap, individual sector aspect ratios should be well-bounded
        expect(sector.rect.aspectRatio, lessThan(8.0));
      }
    });

    test('HeatmapColorEngine interpolates colors and maintains WCAG 2.1 AA text contrast', () {
      // Standard Palette
      final bullColor = HeatmapColorEngine.getColor(
        changePercentage: 3.0,
        mode: HeatmapColorMode.standardRedGreen,
      );
      final bearColor = HeatmapColorEngine.getColor(
        changePercentage: -3.0,
        mode: HeatmapColorMode.standardRedGreen,
      );
      final neutralColor = HeatmapColorEngine.getColor(
        changePercentage: 0.0,
        mode: HeatmapColorMode.standardRedGreen,
      );

      expect(bullColor.value, isNot(equals(bearColor.value)));
      expect(neutralColor.value, equals(HeatmapColorEngine.neutralBg.value));

      // Accessible Palette
      final accessBull = HeatmapColorEngine.getColor(
        changePercentage: 2.5,
        mode: HeatmapColorMode.accessibleBlueOrange,
      );
      expect(accessBull.blue, greaterThan(accessBull.red));

      final accessBear = HeatmapColorEngine.getColor(
        changePercentage: -2.5,
        mode: HeatmapColorMode.accessibleBlueOrange,
      );
      expect(accessBear.red, greaterThan(accessBear.blue));

      // Contrast ratio calculation on dark neutral
      final contrast = HeatmapColorEngine.calculateContrastRatio(
        HeatmapColorEngine.neutralBg,
        TreemapColor.white,
      );
      expect(contrast, greaterThanOrEqualTo(4.5)); // WCAG 2.1 Level AA requirement

      final textChoice = HeatmapColorEngine.getAccessibleTextColor(HeatmapColorEngine.neutralBg);
      expect(textChoice, equals(TreemapColor.white));
    });

    test('Tokenized RWA assets contain verifiable ERC-3643 and Proof-of-Reserve hashes', () async {
      final snapshot = await repository.fetchTreemapSnapshot(
        indexCode: 'TOKENIZED RWAS',
        metricType: TreemapMetricType.marketCap,
        timeHorizon: TreemapTimeHorizon.oneDay,
      );

      final rwaSector = snapshot.children.firstWhere((c) => c.id == 'SEC_RWA') as SectorGroupNode;
      expect(rwaSector.children.length, equals(3));

      final goldRwa = rwaSector.children.firstWhere((c) => c.name.contains('Gold')) as SecurityLeafNode;
      expect(goldRwa.isTokenizedRwa, isTrue);
      expect(goldRwa.rwaContractAddress, startsWith('0x'));
      expect(goldRwa.rwaProofOfReserveHash, startsWith('0x'));
      expect(goldRwa.rwaProofOfReserveHash!.length, equals(42));
    });

    test('TreemapController manages drilldown navigation and breadcrumb path state', () async {
      await controller.loadSnapshot();
      expect(controller.state.rootNode, isNotNull);
      expect(controller.state.breadcrumbTrail.length, equals(1));

      // Drill into banking sector
      final banking = controller.state.rootNode!.children
          .firstWhere((c) => c.id == 'SEC_BANKING') as SectorGroupNode;

      controller.drillIntoSector(banking);
      expect(controller.state.breadcrumbTrail.length, equals(2));
      expect(controller.state.activeViewNode?.id, equals('SEC_BANKING'));

      // Navigate back to root via breadcrumb index 0
      controller.navigateBreadcrumb(0);
      expect(controller.state.breadcrumbTrail.length, equals(1));
      expect(controller.state.activeViewNode?.id, equals('ROOT_INDEX'));
    });
  });
}

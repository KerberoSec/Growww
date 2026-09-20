import 'dart:async';
import '../../domain/algorithms/squarified_treemap_algorithm.dart';
import '../../domain/interfaces/i_treemap_repository.dart';
import '../../domain/models/treemap_enums.dart';
import '../../domain/models/treemap_geometry.dart';
import '../../domain/models/treemap_node.dart';
import '../../domain/models/treemap_tick_update.dart';

class TreemapState {
  final String selectedIndex;
  final TreemapMetricType selectedMetric;
  final TreemapTimeHorizon selectedHorizon;
  final HeatmapColorMode colorMode;
  final SectorGroupNode? rootNode;
  final SectorGroupNode? activeViewNode;
  final List<SectorGroupNode> breadcrumbTrail;
  final SecurityLeafNode? selectedSecurity;
  final bool isLoading;
  final String? errorMessage;
  final TreemapRect viewportRect;

  const TreemapState({
    this.selectedIndex = 'NIFTY 50',
    this.selectedMetric = TreemapMetricType.marketCap,
    this.selectedHorizon = TreemapTimeHorizon.oneDay,
    this.colorMode = HeatmapColorMode.standardRedGreen,
    this.rootNode,
    this.activeViewNode,
    this.breadcrumbTrail = const [],
    this.selectedSecurity,
    this.isLoading = false,
    this.errorMessage,
    this.viewportRect = const TreemapRect(left: 0, top: 0, width: 800, height: 600),
  });

  TreemapState copyWith({
    String? selectedIndex,
    TreemapMetricType? selectedMetric,
    TreemapTimeHorizon? selectedHorizon,
    HeatmapColorMode? colorMode,
    SectorGroupNode? rootNode,
    SectorGroupNode? activeViewNode,
    List<SectorGroupNode>? breadcrumbTrail,
    SecurityLeafNode? selectedSecurity,
    bool? isLoading,
    String? errorMessage,
    TreemapRect? viewportRect,
  }) {
    return TreemapState(
      selectedIndex: selectedIndex ?? this.selectedIndex,
      selectedMetric: selectedMetric ?? this.selectedMetric,
      selectedHorizon: selectedHorizon ?? this.selectedHorizon,
      colorMode: colorMode ?? this.colorMode,
      rootNode: rootNode ?? this.rootNode,
      activeViewNode: activeViewNode ?? this.activeViewNode,
      breadcrumbTrail: breadcrumbTrail ?? this.breadcrumbTrail,
      selectedSecurity: selectedSecurity,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      viewportRect: viewportRect ?? this.viewportRect,
    );
  }
}

class TreemapController {
  final ITreemapRepository _repository;
  final SquarifiedTreemapAlgorithm _algorithm;

  TreemapState _state = const TreemapState();
  final _stateController = StreamController<TreemapState>.broadcast();
  StreamSubscription<List<TreemapTickUpdate>>? _tickSubscription;

  TreemapState get state => _state;
  Stream<TreemapState> get stream => _stateController.stream;

  TreemapController({
    required ITreemapRepository repository,
    SquarifiedTreemapAlgorithm algorithm = const SquarifiedTreemapAlgorithm(),
  })  : _repository = repository,
        _algorithm = algorithm {
    loadSnapshot();
    _subscribeTicks();
  }

  void _emit(TreemapState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void updateViewportDimensions(double width, double height) {
    if (width <= 0 || height <= 0) return;
    final newRect = TreemapRect(left: 0, top: 0, width: width, height: height);
    if (_state.rootNode != null) {
      final laidOutRoot = _algorithm.computeLayout(
        root: _state.rootNode!,
        containerRect: newRect,
      );
      final laidOutActive = _findNodeById(laidOutRoot, _state.activeViewNode?.id ?? laidOutRoot.id) as SectorGroupNode? ?? laidOutRoot;
      _emit(_state.copyWith(
        viewportRect: newRect,
        rootNode: laidOutRoot,
        activeViewNode: laidOutActive,
      ));
    } else {
      _emit(_state.copyWith(viewportRect: newRect));
    }
  }

  Future<void> loadSnapshot() async {
    _emit(_state.copyWith(isLoading: true));
    try {
      final rawRoot = await _repository.fetchTreemapSnapshot(
        indexCode: _state.selectedIndex,
        metricType: _state.selectedMetric,
        timeHorizon: _state.selectedHorizon,
      );

      final laidOutRoot = _algorithm.computeLayout(
        root: rawRoot,
        containerRect: _state.viewportRect,
      );

      _emit(_state.copyWith(
        isLoading: false,
        rootNode: laidOutRoot,
        activeViewNode: laidOutRoot,
        breadcrumbTrail: [laidOutRoot],
      ));
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  void setMetric(TreemapMetricType metric) {
    _emit(_state.copyWith(selectedMetric: metric));
    loadSnapshot();
  }

  void setTimeHorizon(TreemapTimeHorizon horizon) {
    _emit(_state.copyWith(selectedHorizon: horizon));
    loadSnapshot();
  }

  void setColorMode(HeatmapColorMode mode) {
    _emit(_state.copyWith(colorMode: mode));
  }

  void drillIntoSector(SectorGroupNode sector) {
    // Re-layout sector in full container dimensions
    final laidOutSector = _algorithm.computeLayout(
      root: sector,
      containerRect: _state.viewportRect,
    );

    final newTrail = List<SectorGroupNode>.from(_state.breadcrumbTrail)..add(laidOutSector);
    _emit(_state.copyWith(
      activeViewNode: laidOutSector,
      breadcrumbTrail: newTrail,
    ));
  }

  void navigateBreadcrumb(int index) {
    if (index < 0 || index >= _state.breadcrumbTrail.length) return;
    final target = _state.breadcrumbTrail[index];
    final laidOutTarget = _algorithm.computeLayout(
      root: target,
      containerRect: _state.viewportRect,
    );
    final newTrail = _state.breadcrumbTrail.sublist(0, index + 1);
    newTrail[index] = laidOutTarget;

    _emit(_state.copyWith(
      activeViewNode: laidOutTarget,
      breadcrumbTrail: newTrail,
    ));
  }

  void selectSecurity(SecurityLeafNode security) {
    _emit(_state.copyWith(selectedSecurity: security));
  }

  void clearSelectedSecurity() {
    _emit(_state.copyWith(selectedSecurity: null));
  }

  void _subscribeTicks() {
    _tickSubscription?.cancel();
    _tickSubscription = _repository.streamTreemapTicks(indexCode: _state.selectedIndex).listen((ticks) {
      if (_state.rootNode == null) return;
      final updatedRoot = _applyTicksToNode(_state.rootNode!, ticks) as SectorGroupNode;
      final updatedActive = _findNodeById(updatedRoot, _state.activeViewNode?.id ?? updatedRoot.id) as SectorGroupNode? ?? updatedRoot;

      _emit(_state.copyWith(
        rootNode: updatedRoot,
        activeViewNode: updatedActive,
      ));
    });
  }

  TreemapNode _applyTicksToNode(TreemapNode node, List<TreemapTickUpdate> ticks) {
    if (node is SecurityLeafNode) {
      final match = ticks.firstWhere(
        (t) => t.symbol == node.symbol,
        orElse: () => TreemapTickUpdate(
          symbol: '',
          ltp: node.ltp,
          changeInr: node.changeInr,
          changePercentage: node.changePercentage,
          turnoverInr: node.turnoverInr,
          totalBidVolume: node.marketDepthBidVolume,
          totalAskVolume: node.marketDepthAskVolume,
          timestamp: DateTime.now(),
        ),
      );
      if (match.symbol.isNotEmpty) {
        return node.copyWithTick(
          newLtp: match.ltp,
          newChangeInr: match.changeInr,
          newChangePercentage: match.changePercentage,
          newTurnoverInr: match.turnoverInr,
          newBidVolume: match.totalBidVolume,
          newAskVolume: match.totalAskVolume,
        );
      }
      return node;
    } else if (node is SectorGroupNode) {
      final updatedChildren = node.children.map((c) => _applyTicksToNode(c, ticks)).toList();
      return node.copyWithChildren(updatedChildren);
    }
    return node;
  }

  TreemapNode? _findNodeById(TreemapNode node, String id) {
    if (node.id == id) return node;
    if (node is SectorGroupNode) {
      for (final child in node.children) {
        final found = _findNodeById(child, id);
        if (found != null) return found;
      }
    }
    return null;
  }

  void dispose() {
    _tickSubscription?.cancel();
    _stateController.close();
  }
}

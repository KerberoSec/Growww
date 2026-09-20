import 'treemap_geometry.dart';

abstract class TreemapNode {
  final String id;
  final String name;
  final double value; // Sizing weight
  final TreemapRect rect;

  const TreemapNode({
    required this.id,
    required this.name,
    required this.value,
    this.rect = const TreemapRect(left: 0, top: 0, width: 0, height: 0),
  });

  TreemapNode copyWithRect(TreemapRect newRect);
}

class SectorGroupNode extends TreemapNode {
  final List<TreemapNode> children;
  final double cumulativeChangePercentage;
  final int totalConstituents;

  const SectorGroupNode({
    required super.id,
    required super.name,
    required super.value,
    super.rect,
    required this.children,
    required this.cumulativeChangePercentage,
    required this.totalConstituents,
  });

  @override
  SectorGroupNode copyWithRect(TreemapRect newRect) {
    return SectorGroupNode(
      id: id,
      name: name,
      value: value,
      rect: newRect,
      children: children,
      cumulativeChangePercentage: cumulativeChangePercentage,
      totalConstituents: totalConstituents,
    );
  }

  SectorGroupNode copyWithChildren(List<TreemapNode> newChildren) {
    return SectorGroupNode(
      id: id,
      name: name,
      value: value,
      rect: rect,
      children: newChildren,
      cumulativeChangePercentage: cumulativeChangePercentage,
      totalConstituents: totalConstituents,
    );
  }
}

class SecurityLeafNode extends TreemapNode {
  final String symbol;
  final String isin;
  final String sectorId;
  final double ltp;
  final double previousClose;
  final double changeInr;
  final double changePercentage;
  final double turnoverInr;
  final double marketDepthBidVolume;
  final double marketDepthAskVolume;
  final bool isTokenizedRwa;
  final String? rwaContractAddress;
  final String? rwaProofOfReserveHash;

  const SecurityLeafNode({
    required super.id,
    required super.name,
    required super.value,
    super.rect,
    required this.symbol,
    required this.isin,
    required this.sectorId,
    required this.ltp,
    required this.previousClose,
    required this.changeInr,
    required this.changePercentage,
    required this.turnoverInr,
    required this.marketDepthBidVolume,
    required this.marketDepthAskVolume,
    this.isTokenizedRwa = false,
    this.rwaContractAddress,
    this.rwaProofOfReserveHash,
  });

  @override
  SecurityLeafNode copyWithRect(TreemapRect newRect) {
    return SecurityLeafNode(
      id: id,
      name: name,
      value: value,
      rect: newRect,
      symbol: symbol,
      isin: isin,
      sectorId: sectorId,
      ltp: ltp,
      previousClose: previousClose,
      changeInr: changeInr,
      changePercentage: changePercentage,
      turnoverInr: turnoverInr,
      marketDepthBidVolume: marketDepthBidVolume,
      marketDepthAskVolume: marketDepthAskVolume,
      isTokenizedRwa: isTokenizedRwa,
      rwaContractAddress: rwaContractAddress,
      rwaProofOfReserveHash: rwaProofOfReserveHash,
    );
  }

  SecurityLeafNode copyWithTick({
    required double newLtp,
    required double newChangeInr,
    required double newChangePercentage,
    required double newTurnoverInr,
    double? newBidVolume,
    double? newAskVolume,
  }) {
    return SecurityLeafNode(
      id: id,
      name: name,
      value: value,
      rect: rect,
      symbol: symbol,
      isin: isin,
      sectorId: sectorId,
      ltp: newLtp,
      previousClose: previousClose,
      changeInr: newChangeInr,
      changePercentage: newChangePercentage,
      turnoverInr: newTurnoverInr,
      marketDepthBidVolume: newBidVolume ?? marketDepthBidVolume,
      marketDepthAskVolume: newAskVolume ?? marketDepthAskVolume,
      isTokenizedRwa: isTokenizedRwa,
      rwaContractAddress: rwaContractAddress,
      rwaProofOfReserveHash: rwaProofOfReserveHash,
    );
  }
}

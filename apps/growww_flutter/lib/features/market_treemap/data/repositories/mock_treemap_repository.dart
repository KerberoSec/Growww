import 'dart:async';
import '../../domain/interfaces/i_treemap_repository.dart';
import '../../domain/models/treemap_enums.dart';
import '../../domain/models/treemap_node.dart';
import '../../domain/models/treemap_tick_update.dart';

class MockTreemapRepository implements ITreemapRepository {
  @override
  Future<List<String>> fetchAvailableIndices() async {
    return ['NIFTY 50', 'NIFTY BANK', 'NIFTY IT', 'TOKENIZED RWAS', 'ALL MARKETS'];
  }

  @override
  Future<SectorGroupNode> fetchTreemapSnapshot({
    required String indexCode,
    required TreemapMetricType metricType,
    required TreemapTimeHorizon timeHorizon,
  }) async {
    // Banking Sector
    final bankingChildren = <TreemapNode>[
      const SecurityLeafNode(
        id: 'HDFCBANK',
        name: 'HDFC Bank Ltd',
        value: 1250000.0,
        symbol: 'HDFCBANK',
        isin: 'INE040A01034',
        sectorId: 'BANKING',
        ltp: 1642.50,
        previousClose: 1618.00,
        changeInr: 24.50,
        changePercentage: 1.51,
        turnoverInr: 3450000000.0,
        marketDepthBidVolume: 850000.0,
        marketDepthAskVolume: 620000.0,
      ),
      const SecurityLeafNode(
        id: 'ICICIBANK',
        name: 'ICICI Bank Ltd',
        value: 890000.0,
        symbol: 'ICICIBANK',
        isin: 'INE090A01021',
        sectorId: 'BANKING',
        ltp: 1184.20,
        previousClose: 1155.00,
        changeInr: 29.20,
        changePercentage: 2.53,
        turnoverInr: 2800000000.0,
        marketDepthBidVolume: 650000.0,
        marketDepthAskVolume: 410000.0,
      ),
      const SecurityLeafNode(
        id: 'SBIN',
        name: 'State Bank of India',
        value: 720000.0,
        symbol: 'SBIN',
        isin: 'INE062A01020',
        sectorId: 'BANKING',
        ltp: 785.40,
        previousClose: 792.00,
        changeInr: -6.60,
        changePercentage: -0.83,
        turnoverInr: 1950000000.0,
        marketDepthBidVolume: 450000.0,
        marketDepthAskVolume: 580000.0,
      ),
      const SecurityLeafNode(
        id: 'KOTAKBANK',
        name: 'Kotak Mahindra Bank',
        value: 410000.0,
        symbol: 'KOTAKBANK',
        isin: 'INE237A01028',
        sectorId: 'BANKING',
        ltp: 1780.00,
        previousClose: 1810.00,
        changeInr: -30.00,
        changePercentage: -1.66,
        turnoverInr: 1200000000.0,
        marketDepthBidVolume: 220000.0,
        marketDepthAskVolume: 310000.0,
      ),
    ];

    final bankingSector = SectorGroupNode(
      id: 'SEC_BANKING',
      name: 'Financial Services & Banking',
      value: 3270000.0,
      children: bankingChildren,
      cumulativeChangePercentage: 1.15,
      totalConstituents: 4,
    );

    // IT Sector
    final itChildren = <TreemapNode>[
      const SecurityLeafNode(
        id: 'TCS',
        name: 'Tata Consultancy Services',
        value: 1420000.0,
        symbol: 'TCS',
        isin: 'INE467B01029',
        sectorId: 'IT',
        ltp: 4180.00,
        previousClose: 4220.00,
        changeInr: -40.00,
        changePercentage: -0.95,
        turnoverInr: 2100000000.0,
        marketDepthBidVolume: 320000.0,
        marketDepthAskVolume: 490000.0,
      ),
      const SecurityLeafNode(
        id: 'INFY',
        name: 'Infosys Ltd',
        value: 780000.0,
        symbol: 'INFY',
        isin: 'INE009A01021',
        sectorId: 'IT',
        ltp: 1890.50,
        previousClose: 1845.00,
        changeInr: 45.50,
        changePercentage: 2.47,
        turnoverInr: 3100000000.0,
        marketDepthBidVolume: 780000.0,
        marketDepthAskVolume: 510000.0,
      ),
      const SecurityLeafNode(
        id: 'WIPRO',
        name: 'Wipro Ltd',
        value: 290000.0,
        symbol: 'WIPRO',
        isin: 'INE075A01022',
        sectorId: 'IT',
        ltp: 540.20,
        previousClose: 555.00,
        changeInr: -14.80,
        changePercentage: -2.67,
        turnoverInr: 890000000.0,
        marketDepthBidVolume: 180000.0,
        marketDepthAskVolume: 390000.0,
      ),
    ];

    final itSector = SectorGroupNode(
      id: 'SEC_IT',
      name: 'Information Technology',
      value: 2490000.0,
      children: itChildren,
      cumulativeChangePercentage: 0.35,
      totalConstituents: 3,
    );

    // Energy & Oil Sector
    final energyChildren = <TreemapNode>[
      const SecurityLeafNode(
        id: 'RELIANCE',
        name: 'Reliance Industries Ltd',
        value: 1980000.0,
        symbol: 'RELIANCE',
        isin: 'INE002A01018',
        sectorId: 'ENERGY',
        ltp: 2980.00,
        previousClose: 2910.00,
        changeInr: 70.00,
        changePercentage: 2.41,
        turnoverInr: 4800000000.0,
        marketDepthBidVolume: 920000.0,
        marketDepthAskVolume: 540000.0,
      ),
      const SecurityLeafNode(
        id: 'ONGC',
        name: 'Oil & Natural Gas Corp',
        value: 360000.0,
        symbol: 'ONGC',
        isin: 'INE213A01029',
        sectorId: 'ENERGY',
        ltp: 288.50,
        previousClose: 292.00,
        changeInr: -3.50,
        changePercentage: -1.20,
        turnoverInr: 950000000.0,
        marketDepthBidVolume: 210000.0,
        marketDepthAskVolume: 340000.0,
      ),
    ];

    final energySector = SectorGroupNode(
      id: 'SEC_ENERGY',
      name: 'Energy, Oil & Gas',
      value: 2340000.0,
      children: energyChildren,
      cumulativeChangePercentage: 1.85,
      totalConstituents: 2,
    );

    // Tokenized RWAs Sector
    final rwaChildren = <TreemapNode>[
      const SecurityLeafNode(
        id: 'gGOLD',
        name: 'Tokenized Fine Gold 999',
        value: 950000.0,
        symbol: 'gGOLD',
        isin: 'INRWA000GOLD1',
        sectorId: 'RWA',
        ltp: 7245.00,
        previousClose: 7180.00,
        changeInr: 65.00,
        changePercentage: 0.91,
        turnoverInr: 1850000000.0,
        marketDepthBidVolume: 420000.0,
        marketDepthAskVolume: 310000.0,
        isTokenizedRwa: true,
        rwaContractAddress: '0x3a4b5c6d7e8f901234567890abcdef1234567890',
        rwaProofOfReserveHash: '0x9924ab8c7d1e4f3a2b1c0e9d8f7a6b5c4d3e2f1a',
      ),
      const SecurityLeafNode(
        id: 'gSILVER',
        name: 'Tokenized Fine Silver 999',
        value: 420000.0,
        symbol: 'gSILVER',
        isin: 'INRWA000SLV01',
        sectorId: 'RWA',
        ltp: 86.40,
        previousClose: 84.20,
        changeInr: 2.20,
        changePercentage: 2.61,
        turnoverInr: 980000000.0,
        marketDepthBidVolume: 610000.0,
        marketDepthAskVolume: 450000.0,
        isTokenizedRwa: true,
        rwaContractAddress: '0x4b5c6d7e8f901234567890abcdef12345678901a',
        rwaProofOfReserveHash: '0x8813bc7d1e4f3a2b1c0e9d8f7a6b5c4d3e2f1b2c',
      ),
      const SecurityLeafNode(
        id: 'gGSEC',
        name: '7.18% GS 2033 Sovereign Bond',
        value: 680000.0,
        symbol: 'gGSEC',
        isin: 'IN0020230085',
        sectorId: 'RWA',
        ltp: 101.45,
        previousClose: 101.30,
        changeInr: 0.15,
        changePercentage: 0.15,
        turnoverInr: 1250000000.0,
        marketDepthBidVolume: 890000.0,
        marketDepthAskVolume: 760000.0,
        isTokenizedRwa: true,
        rwaContractAddress: '0x5c6d7e8f901234567890abcdef12345678901a2b',
        rwaProofOfReserveHash: '0x7702cd7d1e4f3a2b1c0e9d8f7a6b5c4d3e2f1c3d',
      ),
    ];

    final rwaSector = SectorGroupNode(
      id: 'SEC_RWA',
      name: 'Tokenized Real-World Assets (RWA)',
      value: 2050000.0,
      children: rwaChildren,
      cumulativeChangePercentage: 1.01,
      totalConstituents: 3,
    );

    final rootChildren = <TreemapNode>[
      bankingSector,
      itSector,
      energySector,
      rwaSector,
    ];

    return SectorGroupNode(
      id: 'ROOT_INDEX',
      name: indexCode,
      value: 10150000.0,
      children: rootChildren,
      cumulativeChangePercentage: 1.12,
      totalConstituents: 12,
    );
  }

  @override
  Stream<List<TreemapTickUpdate>> streamTreemapTicks({
    required String indexCode,
  }) {
    return Stream.periodic(const Duration(seconds: 3), (count) {
      final now = DateTime.now();
      final isUp = (count % 2) == 0;
      return [
        TreemapTickUpdate(
          symbol: 'HDFCBANK',
          ltp: 1642.50 + (isUp ? 2.5 : -1.8),
          changeInr: 24.50 + (isUp ? 2.5 : -1.8),
          changePercentage: 1.51 + (isUp ? 0.15 : -0.11),
          turnoverInr: 3450000000.0 + (count * 1000000),
          totalBidVolume: 850000.0,
          totalAskVolume: 620000.0,
          timestamp: now,
        ),
        TreemapTickUpdate(
          symbol: 'gGOLD',
          ltp: 7245.00 + (isUp ? 10.0 : -5.0),
          changeInr: 65.00 + (isUp ? 10.0 : -5.0),
          changePercentage: 0.91 + (isUp ? 0.14 : -0.07),
          turnoverInr: 1850000000.0 + (count * 500000),
          totalBidVolume: 420000.0,
          totalAskVolume: 310000.0,
          timestamp: now,
        ),
      ];
    });
  }
}

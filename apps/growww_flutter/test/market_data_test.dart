import 'package:test/test.dart';
import '../lib/models/market_data.dart';

void main() {
  group('Market Data & Depth Model Tests', () {
    test('DepthEntry JSON serialization and properties', () {
      const entry = DepthEntry(
        price: 64250.0,
        quantity: 1.5,
        orderCount: 3,
        cumulativeQuantity: 4.5,
      );

      final json = entry.toJson();
      expect(json['price'], equals(64250.0));
      expect(json['quantity'], equals(1.5));
      expect(json['order_count'], equals(3));
      expect(json['cumulative_quantity'], equals(4.5));

      final roundtrip = DepthEntry.fromJson(json);
      expect(roundtrip.price, equals(64250.0));
      expect(roundtrip.quantity, equals(1.5));
      expect(roundtrip.cumulativeQuantity, equals(4.5));

      // Test array format deserialization [price, qty, count]
      final fromArray = DepthEntry.fromJson([64100.0, 2.0, 5]);
      expect(fromArray.price, equals(64100.0));
      expect(fromArray.quantity, equals(2.0));
      expect(fromArray.orderCount, equals(5));
    });

    test('OrderBookDepth spread, midPrice, and cumulative calculation', () {
      final depth = OrderBookDepth.withCumulatives(
        symbol: 'BTC/USDT',
        sequence: 500,
        timestamp: DateTime.utc(2026, 9, 20, 10, 0, 0),
        rawBids: [
          const DepthEntry(price: 64200.0, quantity: 1.0),
          const DepthEntry(price: 64150.0, quantity: 2.0),
        ],
        rawAsks: [
          const DepthEntry(price: 64250.0, quantity: 0.5),
          const DepthEntry(price: 64300.0, quantity: 1.5),
        ],
      );

      expect(depth.bestBid?.price, equals(64200.0));
      expect(depth.bestAsk?.price, equals(64250.0));
      expect(depth.spread, equals(50.0));
      expect(depth.midPrice, equals(64225.0));
      expect(depth.spreadPercentage, closeTo(0.0778, 0.001));

      // Check cumulative quantities
      expect(depth.bids[0].cumulativeQuantity, equals(1.0));
      expect(depth.bids[1].cumulativeQuantity, equals(3.0));
      expect(depth.asks[0].cumulativeQuantity, equals(0.5));
      expect(depth.asks[1].cumulativeQuantity, equals(2.0));
    });

    test('OrderBookDepth applies incremental diff correctly', () {
      final initial = OrderBookDepth.withCumulatives(
        symbol: 'BTC/USDT',
        sequence: 100,
        timestamp: DateTime.now(),
        rawBids: [
          const DepthEntry(price: 64000.0, quantity: 1.0),
          const DepthEntry(price: 63900.0, quantity: 2.0),
        ],
        rawAsks: [
          const DepthEntry(price: 64100.0, quantity: 1.0),
          const DepthEntry(price: 64200.0, quantity: 3.0),
        ],
      );

      // Diff that:
      // 1. Adds a new best bid at 64050
      // 2. Removes ask at 64100 (quantity = 0)
      // 3. Updates ask at 64200 (quantity = 5.0)
      final diff = OrderBookDiff(
        symbol: 'BTC/USDT',
        firstSequence: 101,
        lastSequence: 101,
        bids: [const DepthEntry(price: 64050.0, quantity: 1.5)],
        asks: [
          const DepthEntry(price: 64100.0, quantity: 0.0), // Removal
          const DepthEntry(price: 64200.0, quantity: 5.0), // Update
        ],
        timestamp: DateTime.now(),
      );

      final updated = initial.applyDiff(diff);
      expect(updated.sequence, equals(101));
      expect(updated.bestBid?.price, equals(64050.0));
      expect(updated.bestBid?.quantity, equals(1.5));
      expect(updated.bestAsk?.price, equals(64200.0));
      expect(updated.bestAsk?.quantity, equals(5.0));
      expect(updated.asks.any((a) => a.price == 64100.0), isFalse);
    });

    test('Ticker model bullish/bearish detection and serialization', () {
      final bullishTicker = Ticker(
        symbol: 'BTC/USDT',
        lastPrice: 64500.0,
        priceChange24h: 1500.0,
        priceChangePercent24h: 2.38,
        highPrice24h: 65000.0,
        lowPrice24h: 62000.0,
        baseVolume24h: 1250.0,
        quoteVolume24h: 80000000.0,
        openPrice24h: 63000.0,
        bidPrice: 64490.0,
        askPrice: 64510.0,
        timestamp: DateTime.now(),
      );

      expect(bullishTicker.isBullish, isTrue);

      final json = bullishTicker.toJson();
      final reconstituted = Ticker.fromJson(json);
      expect(reconstituted.symbol, equals('BTC/USDT'));
      expect(reconstituted.lastPrice, equals(64500.0));
      expect(reconstituted.priceChangePercent24h, equals(2.38));

      final bearishTicker = Ticker(
        symbol: 'ETH/USDT',
        lastPrice: 3200.0,
        priceChange24h: -100.0,
        priceChangePercent24h: -3.03,
        highPrice24h: 3350.0,
        lowPrice24h: 3180.0,
        baseVolume24h: 5000.0,
        quoteVolume24h: 16000000.0,
        openPrice24h: 3300.0,
        bidPrice: 3195.0,
        askPrice: 3205.0,
        timestamp: DateTime.now(),
      );

      expect(bearishTicker.isBullish, isFalse);
    });

    test('CandlestickData parsing and bullish calculation', () {
      final bullCandle = CandlestickData(
        timestamp: DateTime.now(),
        open: 100.0,
        high: 110.0,
        low: 95.0,
        close: 105.0,
        volume: 50.0,
      );
      expect(bullCandle.isBullish, isTrue);

      final bearCandle = CandlestickData(
        timestamp: DateTime.now(),
        open: 100.0,
        high: 102.0,
        low: 88.0,
        close: 92.0,
        volume: 40.0,
      );
      expect(bearCandle.isBullish, isFalse);

      // Parse array format [timestampMs, open, high, low, close, volume]
      final fromArr = CandlestickData.fromJson([1700000000000, 200.0, 215.0, 195.0, 210.0, 1000.0]);
      expect(fromArr.open, equals(200.0));
      expect(fromArr.high, equals(215.0));
      expect(fromArr.low, equals(195.0));
      expect(fromArr.close, equals(210.0));
      expect(fromArr.volume, equals(1000.0));
      expect(fromArr.isBullish, isTrue);
    });

    test('PortfolioAsset P&L and Section 115BBH 30% tax estimation', () {
      const asset = PortfolioAsset(
        symbol: 'BTC/INR',
        assetName: 'Bitcoin',
        assetClass: AssetClass.cryptoVda,
        balance: 1.0,
        averageCostBasis: 5000000.0,
        currentPrice: 6000000.0,
        tdsAccrued: 60000.0,
      );

      expect(asset.totalBalance, equals(1.0));
      expect(asset.investedValue, equals(5000000.0));
      expect(asset.currentValueInr, equals(6000000.0));
      expect(asset.unrealizedPnL, equals(1000000.0));
      expect(asset.unrealizedPnLPercent, equals(20.0));

      // Section 115BBH 30% tax on ₹ 10,00,000 gain = ₹ 3,00,000
      expect(asset.estimatedVdaGainTax, equals(300000.0));
    });

    test('PortfolioSummary consolidation and asset allocations', () {
      final summary = PortfolioSummary(
        unallocatedCashInr: 100000.0,
        assets: const [
          // Equity: ₹ 5,00,000
          PortfolioAsset(
            symbol: 'RELIANCE.BSE',
            assetName: 'Reliance',
            assetClass: AssetClass.equityInr,
            balance: 200.0,
            averageCostBasis: 2000.0, // Invested: 4,00,000
            currentPrice: 2500.0,     // Current: 5,00,000
          ),
          // Crypto VDA: ₹ 4,00,000
          PortfolioAsset(
            symbol: 'ETH/INR',
            assetName: 'Ethereum',
            assetClass: AssetClass.cryptoVda,
            balance: 2.0,
            averageCostBasis: 150000.0, // Invested: 3,00,000
            currentPrice: 200000.0,     // Current: 4,00,000
            tdsAccrued: 4000.0,
          ),
        ],
        lastUpdated: DateTime.now(),
      );

      // Equity = 5,00,000
      // Crypto = 4,00,000
      // Cash = 1,00,000
      // Total Net Worth = 10,00,000
      expect(summary.totalEquityInr, equals(500000.0));
      expect(summary.totalCryptoVdaInr, equals(400000.0));
      expect(summary.totalNetWorthInr, equals(1000000.0));

      // Invested = 4,00,000 + 3,00,000 = 7,00,000
      // Unrealized PnL = 9,00,000 - 7,00,000 = 2,00,000
      expect(summary.totalInvestedInr, equals(700000.0));
      expect(summary.totalUnrealizedPnLInr, equals(200000.0));
      expect(summary.totalUnrealizedPnLPercent, closeTo(28.57, 0.01));

      // Allocation percentages
      expect(summary.equityAllocationPercent, equals(50.0));
      expect(summary.cryptoAllocationPercent, equals(40.0));
      expect(summary.cashAllocationPercent, equals(10.0));

      // TDS accrued
      expect(summary.totalTdsAccruedInr, equals(4000.0));
      // 115BBH tax on ETH gain (1,00,000 * 30% = 30,000)
      expect(summary.totalSection115bbhTaxEstimate, equals(30000.0));
    });

    test('SequenceGapException formats diagnostic message', () {
      final ex = SequenceGapException(
        symbol: 'BTC/USDT',
        expectedSequence: 501,
        receivedSequence: 504,
      );

      expect(ex.symbol, equals('BTC/USDT'));
      expect(ex.expectedSequence, equals(501));
      expect(ex.receivedSequence, equals(504));
      expect(ex.toString(), contains('expected 501, received 504'));
      expect(ex.toString(), contains('REST snapshot resync'));
    });
  });
}

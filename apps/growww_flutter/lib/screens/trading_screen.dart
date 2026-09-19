import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/market_data.dart';
import '../models/order.dart';
import '../services/biometric_auth_service.dart';
import '../services/websocket_service.dart';
import 'trading/chart_view.dart';
import 'demo/mode_switcher.dart';

/// Main Trading Workstation Screen for Spot and Derivatives Trading.
/// Features:
/// - Real-time header ticker with 24h metrics and Besu DvP settlement badge
/// - Interactive Candlestick Chart with timeframe selectors
/// - Level 2 Orderbook Depth Ladder with cumulative liquidity bars and tap-to-fill
/// - Buy/Sell Order Entry panel with quick chips (25%, 50%, 75%, 100%)
/// - Tactile haptic feedback on gestures and submission
/// - High-value order biometric verification (Face ID / Fingerprint)
/// - Zero-Fee Presentation Invariant (0.00% Maker / 0.00% Taker / 0 Gas)
class TradingScreen extends StatefulWidget {
  final String initialSymbol;
  final WebSocketService? webSocketService;
  final BiometricAuthService? biometricAuthService;

  const TradingScreen({
    Key? key,
    this.initialSymbol = 'BTC/USDT',
    this.webSocketService,
    this.biometricAuthService,
  }) : super(key: key);

  @override
  State<TradingScreen> createState() => _TradingScreenState();
}

class _TradingScreenState extends State<TradingScreen> with SingleTickerProviderStateMixin {
  late String _currentSymbol;
  late final BiometricAuthService _biometricService;
  final ModeSwitcherNotifier _modeNotifier = ModeSwitcherNotifier();

  // Selected Order Parameters
  OrderSide _selectedSide = OrderSide.buy;
  OrderType _selectedType = OrderType.limit;
  TimeInForce _selectedTif = TimeInForce.gtc;
  bool _isPostOnly = false;

  // Controllers
  final TextEditingController _priceController = TextEditingController();
  final TextEditingController _quantityController = TextEditingController();
  final TextEditingController _stopPriceController = TextEditingController();

  // Active open orders
  final List<Order> _openOrders = [];
  final List<TradeExecution> _recentTrades = [];

  // Tab controller for blotter (Orderbook / Recent Trades / Open Orders)
  late TabController _tabController;

  // Mock initial market data
  late Ticker _currentTicker;
  late OrderBookDepth _orderBook;
  late List<CandlestickPoint> _candles;

  @override
  void initState() {
    super.initState();
    _currentSymbol = widget.initialSymbol;
    _biometricService = widget.biometricAuthService ?? BiometricAuthService();
    _tabController = TabController(length: 3, vsync: this);

    _initMockMarketData();
    _priceController.text = _currentTicker.lastPrice.toStringAsFixed(2);
    _quantityController.text = '0.050';
  }

  void _initMockMarketData() {
    _currentTicker = Ticker(
      symbol: _currentSymbol,
      lastPrice: 64250.00,
      priceChange24h: 2145.50,
      priceChangePercent24h: 3.45,
      highPrice24h: 65100.00,
      lowPrice24h: 62800.00,
      baseVolume24h: 4215.85,
      quoteVolume24h: 270868362.50,
      openPrice24h: 62104.50,
      bidPrice: 64249.50,
      askPrice: 64250.50,
      timestamp: DateTime.now(),
    );

    // Initial L2 Depth Book with 6 asks and 6 bids
    _orderBook = OrderBookDepth.withCumulatives(
      symbol: _currentSymbol,
      sequence: 1001,
      timestamp: DateTime.now(),
      rawBids: [
        const DepthEntry(price: 64249.50, quantity: 1.150),
        const DepthEntry(price: 64248.00, quantity: 2.420),
        const DepthEntry(price: 64246.50, quantity: 0.910),
        const DepthEntry(price: 64245.00, quantity: 3.100),
        const DepthEntry(price: 64243.00, quantity: 1.840),
        const DepthEntry(price: 64240.00, quantity: 5.250),
      ],
      rawAsks: [
        const DepthEntry(price: 64250.50, quantity: 0.852),
        const DepthEntry(price: 64252.00, quantity: 1.240),
        const DepthEntry(price: 64253.50, quantity: 0.510),
        const DepthEntry(price: 64255.00, quantity: 2.219),
        const DepthEntry(price: 64258.00, quantity: 1.630),
        const DepthEntry(price: 64260.00, quantity: 4.800),
      ],
    );

    // Seed initial candlestick points
    final nowMs = DateTime.now().millisecondsSinceEpoch;
    _candles = List.generate(24, (i) {
      final t = nowMs - (24 - i) * 15 * 60 * 1000;
      final base = 63500.0 + (i * 35.0) + (math.sin(i.toDouble()) * 200.0);
      return CandlestickPoint(
        timestampMs: t,
        open: base,
        high: base + 85.0,
        low: base - 60.0,
        close: base + 40.0,
        volume: 12.5 + (i * 1.5),
      );
    });

    // Seed mock recent trades
    _recentTrades.addAll([
      TradeExecution(
        id: 'tx_101',
        orderId: 'ord_1',
        clientOrderId: 'cl_1',
        symbol: _currentSymbol,
        side: OrderSide.buy,
        price: 64250.50,
        quantity: 0.125,
        quoteAmount: 8031.31,
        timestamp: DateTime.now().subtract(const Duration(seconds: 4)),
      ),
      TradeExecution(
        id: 'tx_102',
        orderId: 'ord_2',
        clientOrderId: 'cl_2',
        symbol: _currentSymbol,
        side: OrderSide.sell,
        price: 64249.50,
        quantity: 0.450,
        quoteAmount: 28912.27,
        timestamp: DateTime.now().subtract(const Duration(seconds: 12)),
      ),
    ]);
  }

  @override
  void dispose() {
    _tabController.dispose();
    _priceController.dispose();
    _quantityController.dispose();
    _stopPriceController.dispose();
    _modeNotifier.dispose();
    super.dispose();
  }

  /// Interactive Tap-to-Fill: User taps an orderbook level to autofill price.
  void _onTapDepthLevel(double price, double size) {
    HapticFeedback.lightImpact();
    setState(() {
      _priceController.text = price.toStringAsFixed(2);
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Price updated to \$${price.toStringAsFixed(2)}'),
        duration: const Duration(milliseconds: 700),
        backgroundColor: const Color(0xFF1E222D),
      ),
    );
  }

  /// Quick quantity percentage selector.
  void _onSelectPercentageChip(double factor) {
    HapticFeedback.selectionClick();
    // Simulate available balance = 1.0 BTC or 50,000 USDT
    final maxQty = _selectedSide == OrderSide.buy ? 0.75 : 1.20;
    final calculated = maxQty * factor;
    setState(() {
      _quantityController.text = calculated.toStringAsFixed(3);
    });
  }

  /// Submits order with validation, haptics, and high-value biometric gate.
  Future<void> _submitOrder() async {
    final price = double.tryParse(_priceController.text) ?? 0.0;
    final quantity = double.tryParse(_quantityController.text) ?? 0.0;
    final stopPrice = double.tryParse(_stopPriceController.text);

    if (quantity <= 0) {
      HapticFeedback.heavyImpact();
      _showErrorDialog('Invalid Quantity', 'Please specify a quantity greater than zero.');
      return;
    }

    if (_selectedType == OrderType.limit && price <= 0) {
      HapticFeedback.heavyImpact();
      _showErrorDialog('Invalid Price', 'Limit orders require an execution price above zero.');
      return;
    }

    final notionalValue = (_selectedType == OrderType.market ? _currentTicker.lastPrice : price) * quantity;
    final currency = _currentSymbol.contains('/') ? _currentSymbol.split('/').last : 'USDT';

    // 1. High-Value Biometric Gate Check
    String? biometricProof;
    if (_biometricService.requiresBiometricAuth(notionalValue: notionalValue, currency: currency)) {
      HapticFeedback.mediumImpact();
      final bioResult = await _biometricService.authenticateForOrder(
        clientOrderId: 'cl_${DateTime.now().millisecondsSinceEpoch}',
        symbol: _currentSymbol,
        notionalValue: notionalValue,
        currency: currency,
      );

      if (!bioResult.success) {
        HapticFeedback.heavyImpact();
        _showErrorDialog('Biometric Verification Failed', bioResult.errorMessage ?? 'Authorization rejected.');
        return;
      }
      biometricProof = bioResult.signature;
    }

    // 2. Success Haptic Pulse
    HapticFeedback.mediumImpact();

    // 3. Create domain Order
    final newOrder = Order(
      id: 'ord_${DateTime.now().millisecondsSinceEpoch}',
      clientOrderId: 'cl_${DateTime.now().millisecondsSinceEpoch}',
      symbol: _currentSymbol,
      side: _selectedSide,
      type: _selectedType,
      timeInForce: _selectedTif,
      price: _selectedType == OrderType.market ? _currentTicker.lastPrice : price,
      quantity: quantity,
      status: OrderStatus.newOrder,
      stopPrice: stopPrice,
      isBiometricVerified: biometricProof != null,
      biometricSignature: biometricProof,
      settlementMode: SettlementMode.instantDvP,
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    );

    setState(() {
      _openOrders.insert(0, newOrder);
    });

    _showOrderPlacedModal(newOrder);
  }

  void _cancelOrder(Order order) {
    HapticFeedback.mediumImpact();
    setState(() {
      _openOrders.removeWhere((o) => o.id == order.id);
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Order ${order.id} cancelled.'),
        backgroundColor: const Color(0xFF222630),
      ),
    );
  }

  void _showErrorDialog(String title, String message) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF1E222D),
        title: Text(title, style: const TextStyle(color: Color(0xFFFF1744), fontWeight: FontWeight.bold)),
        content: Text(message, style: const TextStyle(color: Color(0xFFB2B5BE))),
        actions: [
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF222630)),
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('OK', style: TextStyle(color: Colors.white)),
          )
        ],
      ),
    );
  }

  void _showOrderPlacedModal(Order order) {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF12141A),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) {
        return Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.check_circle, color: Color(0xFF00E676), size: 28),
                  const SizedBox(width: 10),
                  Text(
                    'Order Placed Successfully',
                    style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              _buildModalRow('Symbol', order.symbol),
              _buildModalRow('Side', order.side.label, color: order.side == OrderSide.buy ? const Color(0xFF00E676) : const Color(0xFFFF1744)),
              _buildModalRow('Type', order.type.displayName),
              _buildModalRow('Price', '\$${order.price.toStringAsFixed(2)}'),
              _buildModalRow('Quantity', order.quantity.toString()),
              _buildModalRow('Total Notional', '\$${order.notionalValue.toStringAsFixed(2)}'),
              _buildModalRow('Exchange Fee', '0.00 USDT (0.00% Zero-Fee Guarantee)', color: const Color(0xFF00E676)),
              if (order.isBiometricVerified)
                _buildModalRow('Biometric Sign-off', 'Verified via Secure Enclave', color: const Color(0xFF7C4DFF)),
              _buildModalRow('Settlement', 'Hyperledger Besu QBFT Instant DvP', color: const Color(0xFF00E5FF)),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF00E676),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  onPressed: () => Navigator.of(ctx).pop(),
                  child: const Text('View Open Orders', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold)),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildModalRow(String label, String value, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Color(0xFF9096A2), fontSize: 13)),
          Text(value, style: TextStyle(color: color ?? Colors.white, fontSize: 13, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF08090C),
      appBar: _buildAppBar(),
      body: SafeArea(
        child: Column(
          children: [
            // Top Metrics & 24h Bar
            _buildTickerMetricsBar(),

            // Flexible middle section with Chart & Depth Ladder
            Expanded(
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                child: Column(
                  children: [
                    // Candlestick Chart View
                    SizedBox(
                      height: 240,
                      child: TradingViewChartWidget(
                        symbol: _currentSymbol,
                        candles: _candles,
                      ),
                    ),

                    // Interactive Level 2 Orderbook Ladder
                    _buildOrderBookSection(),

                    // Order Placement Entry Panel
                    _buildOrderEntryPanel(),

                    // Blotter Tabs (Open Orders, Recent Trades)
                    _buildBlotterTabs(),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    final isBull = _currentTicker.isBullish;
    return AppBar(
      backgroundColor: const Color(0xFF12141A),
      elevation: 0,
      title: Row(
        children: [
          Text(
            _currentSymbol,
            style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 17),
          ),
          const SizedBox(width: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: isBull ? const Color(0xFF1B382B) : const Color(0xFF381B22),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              '${isBull ? '+' : ''}${_currentTicker.priceChangePercent24h.toStringAsFixed(2)}%',
              style: TextStyle(
                color: isBull ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                fontSize: 11,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
      actions: [
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 10.0),
          child: ModeSwitcherWidget(notifier: _modeNotifier),
        ),
        IconButton(
          icon: const Icon(Icons.notifications_none, color: Color(0xFF9096A2)),
          onPressed: () {},
        ),
      ],
    );
  }

  Widget _buildTickerMetricsBar() {
    return Container(
      color: const Color(0xFF12141A),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '\$${_currentTicker.lastPrice.toStringAsFixed(2)}',
                style: TextStyle(
                  color: _currentTicker.isBullish ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  fontFamily: 'monospace',
                ),
              ),
              const SizedBox(height: 2),
              const Row(
                children: [
                  Icon(Icons.shield_outlined, color: Color(0xFF00E5FF), size: 12),
                  SizedBox(width: 4),
                  Text('Besu QBFT DvP', style: TextStyle(color: Color(0xFF00E5FF), fontSize: 10)),
                ],
              ),
            ],
          ),
          _buildMetricColumn('24h High', _currentTicker.highPrice24h.toStringAsFixed(2)),
          _buildMetricColumn('24h Low', _currentTicker.lowPrice24h.toStringAsFixed(2)),
          _buildMetricColumn('24h Vol', '${_currentTicker.baseVolume24h.toStringAsFixed(1)} BTC'),
        ],
      ),
    );
  }

  Widget _buildMetricColumn(String label, String val) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Text(label, style: const TextStyle(color: Color(0xFF5A606D), fontSize: 10)),
        const SizedBox(height: 2),
        Text(val, style: const TextStyle(color: Color(0xFF9096A2), fontSize: 11, fontFamily: 'monospace')),
      ],
    );
  }

  Widget _buildOrderBookSection() {
    double maxCum = 1.0;
    if (_orderBook.bids.isNotEmpty) {
      maxCum = math.max(maxCum, _orderBook.bids.last.cumulativeQuantity);
    }
    if (_orderBook.asks.isNotEmpty) {
      maxCum = math.max(maxCum, _orderBook.asks.last.cumulativeQuantity);
    }

    return Container(
      color: const Color(0xFF08090C),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('LEVEL 2 DEPTH LADDER', style: TextStyle(color: Color(0xFF9096A2), fontSize: 11, fontWeight: FontWeight.bold)),
              Text(
                'Spread: \$${_orderBook.spread.toStringAsFixed(2)} (${_orderBook.spreadPercentage.toStringAsFixed(3)}%)',
                style: const TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontFamily: 'monospace'),
              ),
            ],
          ),
          const SizedBox(height: 6),
          const Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('PRICE (USDT)', style: TextStyle(color: Color(0xFF5A606D), fontSize: 10)),
              Text('SIZE', style: TextStyle(color: Color(0xFF5A606D), fontSize: 10)),
              Text('TOTAL', style: TextStyle(color: Color(0xFF5A606D), fontSize: 10)),
            ],
          ),
          const SizedBox(height: 4),

          // Top Asks (Red)
          for (final ask in _orderBook.asks.take(4))
            _buildDepthRow(ask, isAsk: true, maxCum: maxCum),

          // Mid Price Line
          Container(
            padding: const EdgeInsets.symmetric(vertical: 4),
            alignment: Alignment.center,
            decoration: const BoxDecoration(
              border: Border.symmetric(horizontal: BorderSide(color: Color(0xFF232732))),
            ),
            child: Text(
              'Spread Mid: \$${_orderBook.midPrice.toStringAsFixed(2)}',
              style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold, fontFamily: 'monospace'),
            ),
          ),

          // Top Bids (Green)
          for (final bid in _orderBook.bids.take(4))
            _buildDepthRow(bid, isAsk: false, maxCum: maxCum),
        ],
      ),
    );
  }

  Widget _buildDepthRow(DepthEntry entry, {required bool isAsk, required double maxCum}) {
    final ratio = (entry.cumulativeQuantity / maxCum).clamp(0.05, 1.0);
    final color = isAsk ? const Color(0xFFFF1744) : const Color(0xFF00E676);

    return InkWell(
      onTap: () => _onTapDepthLevel(entry.price, entry.quantity),
      child: Stack(
        children: [
          Positioned.fill(
            child: Align(
              alignment: isAsk ? Alignment.centerLeft : Alignment.centerRight,
              child: FractionallySizedBox(
                widthFactor: ratio,
                child: Container(color: color.withOpacity(0.12)),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 3.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(entry.price.toStringAsFixed(2), style: TextStyle(color: color, fontSize: 12, fontFamily: 'monospace', fontWeight: FontWeight.w600)),
                Text(entry.quantity.toStringAsFixed(3), style: const TextStyle(color: Colors.white, fontSize: 12, fontFamily: 'monospace')),
                Text(entry.cumulativeQuantity.toStringAsFixed(3), style: const TextStyle(color: Color(0xFF9096A2), fontSize: 12, fontFamily: 'monospace')),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrderEntryPanel() {
    final isBuy = _selectedSide == OrderSide.buy;
    final primaryColor = isBuy ? const Color(0xFF00E676) : const Color(0xFFFF1744);

    return Container(
      margin: const EdgeInsets.all(12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF12141A),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF232732)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Buy / Sell Selector
          Row(
            children: [
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: isBuy ? const Color(0xFF00E676) : const Color(0xFF1A1D24),
                    foregroundColor: isBuy ? Colors.black : Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                  onPressed: () {
                    HapticFeedback.selectionClick();
                    setState(() => _selectedSide = OrderSide.buy);
                  },
                  child: const Text('BUY BTC', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: !isBuy ? const Color(0xFFFF1744) : const Color(0xFF1A1D24),
                    foregroundColor: !isBuy ? Colors.black : Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                  onPressed: () {
                    HapticFeedback.selectionClick();
                    setState(() => _selectedSide = OrderSide.sell);
                  },
                  child: const Text('SELL BTC', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Order Type Chips
          Row(
            children: [
              for (final t in [OrderType.limit, OrderType.market, OrderType.stopLimit])
                Padding(
                  padding: const EdgeInsets.only(right: 6.0),
                  child: ChoiceChip(
                    label: Text(t.displayName),
                    selected: _selectedType == t,
                    selectedColor: const Color(0xFF222630),
                    backgroundColor: const Color(0xFF1A1D24),
                    labelStyle: TextStyle(
                      color: _selectedType == t ? primaryColor : const Color(0xFF9096A2),
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                    ),
                    onSelected: (val) {
                      if (val) {
                        HapticFeedback.selectionClick();
                        setState(() => _selectedType = t);
                      }
                    },
                  ),
                ),
            ],
          ),
          const SizedBox(height: 12),

          // Price Input
          if (_selectedType != OrderType.market)
            _buildInputField(
              controller: _priceController,
              label: 'PRICE',
              suffix: 'USDT',
              onDecrement: () {
                final cur = double.tryParse(_priceController.text) ?? _currentTicker.lastPrice;
                _priceController.text = (cur - 1.0).toStringAsFixed(2);
              },
              onIncrement: () {
                final cur = double.tryParse(_priceController.text) ?? _currentTicker.lastPrice;
                _priceController.text = (cur + 1.0).toStringAsFixed(2);
              },
            ),

          const SizedBox(height: 8),

          // Quantity Input
          _buildInputField(
            controller: _quantityController,
            label: 'AMOUNT',
            suffix: 'BTC',
            onDecrement: () {
              final cur = double.tryParse(_quantityController.text) ?? 0.05;
              _quantityController.text = math.max(0.001, cur - 0.01).toStringAsFixed(3);
            },
            onIncrement: () {
              final cur = double.tryParse(_quantityController.text) ?? 0.05;
              _quantityController.text = (cur + 0.01).toStringAsFixed(3);
            },
          ),

          const SizedBox(height: 8),

          // Percentage Quick Chips (25%, 50%, 75%, 100%)
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              for (final pct in [0.25, 0.50, 0.75, 1.00])
                InkWell(
                  onTap: () => _onSelectPercentageChip(pct),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                    decoration: BoxDecoration(
                      color: const Color(0xFF1A1D24),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: const Color(0xFF232732)),
                    ),
                    child: Text(
                      '${(pct * 100).toInt()}%',
                      style: const TextStyle(color: Color(0xFF9096A2), fontSize: 11, fontWeight: FontWeight.bold),
                    ),
                  ),
                ),
            ],
          ),

          const SizedBox(height: 12),

          // Post-Only Toggle & Zero-Fee Presentation
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Checkbox(
                    value: _isPostOnly,
                    activeColor: primaryColor,
                    onChanged: (v) => setState(() => _isPostOnly = v ?? false),
                  ),
                  const Text('Post-Only (Maker)', style: TextStyle(color: Color(0xFF9096A2), fontSize: 12)),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFF1B382B),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: const Text('0.00% ZERO FEE', style: TextStyle(color: Color(0xFF00E676), fontSize: 10, fontWeight: FontWeight.bold)),
              ),
            ],
          ),

          const SizedBox(height: 14),

          // Submit Order Button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryColor,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              ),
              onPressed: _submitOrder,
              child: Text(
                '${isBuy ? 'BUY' : 'SELL'} BTC NOW',
                style: const TextStyle(color: Colors.black, fontSize: 15, fontWeight: FontWeight.bold),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInputField({
    required TextEditingController controller,
    required String label,
    required String suffix,
    required VoidCallback onDecrement,
    required VoidCallback onIncrement,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFF1A1D24),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFF232732)),
      ),
      child: Row(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 10),
            child: Text(label, style: const TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontWeight: FontWeight.bold)),
          ),
          Expanded(
            child: TextField(
              controller: controller,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              style: const TextStyle(color: Colors.white, fontSize: 13, fontFamily: 'monospace'),
              decoration: const InputDecoration(
                border: InputBorder.none,
                isDense: true,
              ),
            ),
          ),
          Text(suffix, style: const TextStyle(color: Color(0xFF9096A2), fontSize: 11)),
          IconButton(
            icon: const Icon(Icons.remove, color: Color(0xFF9096A2), size: 16),
            onPressed: onDecrement,
          ),
          IconButton(
            icon: const Icon(Icons.add, color: Color(0xFF9096A2), size: 16),
            onPressed: onIncrement,
          ),
        ],
      ),
    );
  }

  Widget _buildBlotterTabs() {
    return Container(
      color: const Color(0xFF12141A),
      child: Column(
        children: [
          TabBar(
            controller: _tabController,
            indicatorColor: const Color(0xFF00E676),
            labelColor: Colors.white,
            unselectedLabelColor: const Color(0xFF9096A2),
            tabs: [
              Tab(text: 'Open Orders (${_openOrders.length})'),
              const Tab(text: 'Recent Trades'),
              const Tab(text: 'Assets / Info'),
            ],
          ),
          SizedBox(
            height: 180,
            child: TabBarView(
              controller: _tabController,
              children: [
                // Open Orders List
                _openOrders.isEmpty
                    ? const Center(child: Text('No resting open orders', style: TextStyle(color: Color(0xFF5A606D))))
                    : ListView.builder(
                        itemCount: _openOrders.length,
                        itemBuilder: (ctx, i) {
                          final o = _openOrders[i];
                          return ListTile(
                            dense: true,
                            title: Row(
                              children: [
                                Text(
                                  o.side.label,
                                  style: TextStyle(
                                    color: o.side == OrderSide.buy ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Text('\$${o.price.toStringAsFixed(2)}', style: const TextStyle(color: Colors.white, fontFamily: 'monospace')),
                                const SizedBox(width: 8),
                                Text('${o.quantity} BTC', style: const TextStyle(color: Color(0xFF9096A2), fontFamily: 'monospace')),
                              ],
                            ),
                            subtitle: Text('Status: ${o.status.label} • DvP', style: const TextStyle(color: Color(0xFF5A606D), fontSize: 11)),
                            trailing: IconButton(
                              icon: const Icon(Icons.close, color: Color(0xFFFF1744), size: 16),
                              onPressed: () => _cancelOrder(o),
                            ),
                          );
                        },
                      ),

                // Recent Trades Tape
                ListView.builder(
                  itemCount: _recentTrades.length,
                  itemBuilder: (ctx, i) {
                    final tr = _recentTrades[i];
                    return Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '\$${tr.price.toStringAsFixed(2)}',
                            style: TextStyle(
                              color: tr.side == OrderSide.buy ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                              fontFamily: 'monospace',
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          Text('${tr.quantity.toStringAsFixed(3)} BTC', style: const TextStyle(color: Colors.white, fontFamily: 'monospace')),
                          Text(
                            '${tr.timestamp.hour.toString().padLeft(2, '0')}:${tr.timestamp.minute.toString().padLeft(2, '0')}:${tr.timestamp.second.toString().padLeft(2, '0')}',
                            style: const TextStyle(color: Color(0xFF5A606D), fontSize: 11),
                          ),
                        ],
                      ),
                    );
                  },
                ),

                // Assets & Info
                const Padding(
                  padding: EdgeInsets.all(16.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Sovereign Clearing Guarantee', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                      SizedBox(height: 6),
                      Text(
                        'Orders execute with atomic DvP settlement on the Hyperledger Besu consortium network with zero counterparty risk.',
                        style: TextStyle(color: Color(0xFF9096A2), fontSize: 12),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

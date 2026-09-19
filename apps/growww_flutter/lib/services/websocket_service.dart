import 'dart:async';
import 'dart:convert';
import 'dart:math' as math;
import '../models/market_data.dart';
import '../models/order.dart';

/// Connection status states for the market data WebSocket.
enum WebSocketStatus {
  disconnected,
  connecting,
  connected,
  reconnecting,
}

/// Abstract socket transport layer to enable both real network WebSockets
/// and deterministic mock testing.
abstract class ISocketTransport {
  Stream<dynamic> get stream;
  void send(String data);
  Future<void> close([int? code, String? reason]);
}

/// Simulated mock socket transport for unit testing and offline development.
class MockSocketTransport implements ISocketTransport {
  final StreamController<dynamic> _controller = StreamController<dynamic>.broadcast();
  final List<String> sentMessages = [];
  bool isClosed = false;

  @override
  Stream<dynamic> get stream => _controller.stream;

  @override
  void send(String data) {
    if (!isClosed) {
      sentMessages.add(data);
    }
  }

  void emitFromServer(String message) {
    if (!_controller.isClosed) {
      _controller.add(message);
    }
  }

  void emitError(dynamic error) {
    if (!_controller.isClosed) {
      _controller.addError(error);
    }
  }

  @override
  Future<void> close([int? code, String? reason]) async {
    isClosed = true;
    await _controller.close();
  }
}

/// Resilient, high-throughput WebSocket client featuring:
/// - Multiplexed channel subscriptions with auto-resubscription on reconnect
/// - Monotonic sequence gap detection for Level 2 orderbook diffs
/// - Automatic REST snapshot resync on packet drop
/// - Exponential backoff with jitter
/// - Periodic ping/pong heartbeat monitoring
class WebSocketService {
  final String url;
  final Duration heartbeatInterval;
  final Duration heartbeatTimeout;
  final Duration minReconnectDelay;
  final Duration maxReconnectDelay;
  final Future<ISocketTransport> Function(String url)? transportFactory;

  ISocketTransport? _transport;
  WebSocketStatus _status = WebSocketStatus.disconnected;
  int _reconnectAttempts = 0;
  Timer? _reconnectTimer;
  Timer? _heartbeatTimer;
  Timer? _heartbeatTimeoutTimer;
  bool _isDisposed = false;

  // Tracked channel subscriptions (e.g., 'btc_usdt@depth20', 'btc_usdt@ticker')
  final Set<String> _subscriptions = <String>{};

  // Monotonic sequence tracking per symbol for gap detection
  final Map<String, int> _lastSequenceBySymbol = <String, int>{};

  // Orderbook depth caches per symbol
  final Map<String, OrderBookDepth> _orderbooks = <String, OrderBookDepth>{};

  // Broadcast stream controllers
  final StreamController<WebSocketStatus> _statusController =
      StreamController<WebSocketStatus>.broadcast();
  final StreamController<Ticker> _tickerController =
      StreamController<Ticker>.broadcast();
  final StreamController<OrderBookDepth> _depthController =
      StreamController<OrderBookDepth>.broadcast();
  final StreamController<TradeExecution> _tradeController =
      StreamController<TradeExecution>.broadcast();
  final StreamController<Order> _orderController =
      StreamController<Order>.broadcast();
  final StreamController<SequenceGapException> _sequenceGapController =
      StreamController<SequenceGapException>.broadcast();

  // Callback to fetch REST snapshot when a sequence gap occurs
  Future<OrderBookDepth?> Function(String symbol)? onFetchSnapshotRequest;

  WebSocketService({
    this.url = 'wss://ws.growww.in/v1/market',
    this.heartbeatInterval = const Duration(seconds: 15),
    this.heartbeatTimeout = const Duration(seconds: 5),
    this.minReconnectDelay = const Duration(milliseconds: 500),
    this.maxReconnectDelay = const Duration(seconds: 10),
    this.transportFactory,
    this.onFetchSnapshotRequest,
  });

  // Public Streams
  Stream<WebSocketStatus> get statusStream => _statusController.stream;
  Stream<Ticker> get tickerStream => _tickerController.stream;
  Stream<OrderBookDepth> get depthStream => _depthController.stream;
  Stream<TradeExecution> get tradeStream => _tradeController.stream;
  Stream<Order> get orderStream => _orderController.stream;
  Stream<SequenceGapException> get sequenceGapStream => _sequenceGapController.stream;

  WebSocketStatus get status => _status;
  int get reconnectAttempts => _reconnectAttempts;
  Set<String> get activeSubscriptions => Set.unmodifiable(_subscriptions);
  int? getLastSequence(String symbol) => _lastSequenceBySymbol[symbol.toLowerCase()];
  OrderBookDepth? getOrderBook(String symbol) => _orderbooks[symbol.toLowerCase()];

  /// Establishes connection to the streaming gateway.
  Future<void> connect() async {
    if (_isDisposed) return;
    if (_status == WebSocketStatus.connected || _status == WebSocketStatus.connecting) {
      return;
    }

    _setStatus(_reconnectAttempts > 0 ? WebSocketStatus.reconnecting : WebSocketStatus.connecting);

    try {
      if (transportFactory != null) {
        _transport = await transportFactory!(url);
      } else {
        // Fallback default: mock transport for offline/test environments
        _transport = MockSocketTransport();
      }

      _setStatus(WebSocketStatus.connected);
      _reconnectAttempts = 0;
      _startHeartbeat();

      // Auto-resubscribe to all active channels
      _resubscribeAll();

      _transport!.stream.listen(
        _handleIncomingMessage,
        onError: _handleSocketError,
        onDone: _handleSocketClosed,
        cancelOnError: true,
      );
    } catch (e) {
      _scheduleReconnect();
    }
  }

  /// Subscribe to Level 2 orderbook depth stream for a symbol.
  void subscribeDepth(String symbol) {
    final channel = '${symbol.toLowerCase()}@depth';
    _subscriptions.add(channel);
    _sendSubscriptionMessage('subscribe', channel);
  }

  /// Subscribe to ticker stream for a symbol.
  void subscribeTicker(String symbol) {
    final channel = '${symbol.toLowerCase()}@ticker';
    _subscriptions.add(channel);
    _sendSubscriptionMessage('subscribe', channel);
  }

  /// Subscribe to real-time executed trade tape.
  void subscribeTrades(String symbol) {
    final channel = '${symbol.toLowerCase()}@trade';
    _subscriptions.add(channel);
    _sendSubscriptionMessage('subscribe', channel);
  }

  /// Subscribe to authenticated user order fill updates.
  void subscribeUserOrders(String userToken) {
    final channel = 'user@orders';
    _subscriptions.add(channel);
    _sendSubscriptionMessage('subscribe', channel, token: userToken);
  }

  /// Unsubscribe from a channel.
  void unsubscribe(String channel) {
    _subscriptions.remove(channel);
    _sendSubscriptionMessage('unsubscribe', channel);
  }

  /// Handles incoming raw WebSocket frames.
  void _handleIncomingMessage(dynamic raw) {
    try {
      final messageStr = raw is String ? raw : utf8.decode(raw as List<int>);
      final decoded = json.decode(messageStr) as Map<String, dynamic>;

      final type = decoded['type'] as String? ?? decoded['event'] as String? ?? '';

      // Heartbeat pong response
      if (type == 'pong') {
        _heartbeatTimeoutTimer?.cancel();
        return;
      }

      // Orderbook full snapshot
      if (type == 'depth_snapshot' || decoded.containsKey('bids') && decoded.containsKey('sequence')) {
        _processDepthSnapshot(decoded);
        return;
      }

      // Orderbook incremental diff
      if (type == 'depth_update' || decoded.containsKey('first_sequence')) {
        _processDepthDiff(decoded);
        return;
      }

      // Market 24h Ticker update
      if (type == 'ticker' || decoded.containsKey('last_price')) {
        final ticker = Ticker.fromJson(decoded);
        _tickerController.add(ticker);
        return;
      }

      // Trade execution event
      if (type == 'trade' || decoded.containsKey('quote_amount')) {
        final trade = TradeExecution.fromJson(decoded);
        _tradeController.add(trade);
        return;
      }

      // User Order status update
      if (type == 'order_update' || decoded.containsKey('client_order_id')) {
        final order = Order.fromJson(decoded);
        _orderController.add(order);
        return;
      }
    } catch (e) {
      // Discard malformed frames safely without crashing stream
    }
  }

  /// Ingests full snapshot and resets sequence tracking.
  void _processDepthSnapshot(Map<String, dynamic> data) {
    final depth = OrderBookDepth.fromJson(data);
    final sym = depth.symbol.toLowerCase();
    _orderbooks[sym] = depth;
    _lastSequenceBySymbol[sym] = depth.sequence;
    _depthController.add(depth);
  }

  /// Ingests L2 diff with monotonic sequence gap detection.
  void _processDepthDiff(Map<String, dynamic> data) {
    final diff = OrderBookDiff.fromJson(data);
    final sym = diff.symbol.toLowerCase();
    final lastSeq = _lastSequenceBySymbol[sym];

    if (lastSeq == null) {
      // Initial sequence not yet seeded; request snapshot
      _triggerSnapshotResync(diff.symbol);
      return;
    }

    // Strict Sequence Gap Detection Invariant:
    // If incoming firstSequence > lastSequence + 1, packet loss occurred!
    if (diff.firstSequence > lastSeq + 1) {
      final gap = SequenceGapException(
        symbol: diff.symbol,
        expectedSequence: lastSeq + 1,
        receivedSequence: diff.firstSequence,
      );
      _sequenceGapController.add(gap);
      _triggerSnapshotResync(diff.symbol);
      return;
    }

    // Ignore outdated or duplicate diffs
    if (diff.lastSequence <= lastSeq) {
      return;
    }

    // Apply diff to current orderbook
    final currentBook = _orderbooks[sym];
    if (currentBook != null) {
      final updatedBook = currentBook.applyDiff(diff);
      _orderbooks[sym] = updatedBook;
      _lastSequenceBySymbol[sym] = diff.lastSequence;
      _depthController.add(updatedBook);
    } else {
      _triggerSnapshotResync(diff.symbol);
    }
  }

  /// Resyncs orderbook snapshot after a detected packet drop.
  Future<void> _triggerSnapshotResync(String symbol) async {
    if (onFetchSnapshotRequest != null) {
      try {
        final snapshot = await onFetchSnapshotRequest!(symbol);
        if (snapshot != null) {
          final sym = symbol.toLowerCase();
          _orderbooks[sym] = snapshot;
          _lastSequenceBySymbol[sym] = snapshot.sequence;
          _depthController.add(snapshot);
        }
      } catch (e) {
        // Fallback: request via socket
        _sendSubscriptionMessage('request_snapshot', '${symbol.toLowerCase()}@depth');
      }
    } else {
      _sendSubscriptionMessage('request_snapshot', '${symbol.toLowerCase()}@depth');
    }
  }

  /// Seeds or manually resets an orderbook (used by snapshot providers).
  void seedOrderBook(OrderBookDepth snapshot) {
    final sym = snapshot.symbol.toLowerCase();
    _orderbooks[sym] = snapshot;
    _lastSequenceBySymbol[sym] = snapshot.sequence;
    _depthController.add(snapshot);
  }

  void _sendSubscriptionMessage(String action, String channel, {String? token}) {
    if (_transport != null && _status == WebSocketStatus.connected) {
      final payload = json.encode({
        'action': action,
        'channel': channel,
        if (token != null) 'token': token,
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      });
      _transport!.send(payload);
    }
  }

  void _resubscribeAll() {
    for (final channel in _subscriptions) {
      _sendSubscriptionMessage('subscribe', channel);
    }
  }

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimeoutTimer?.cancel();

    _heartbeatTimer = Timer.periodic(heartbeatInterval, (_) {
      if (_status == WebSocketStatus.connected && _transport != null) {
        final pingMsg = json.encode({
          'action': 'ping',
          'timestamp': DateTime.now().millisecondsSinceEpoch,
        });
        _transport!.send(pingMsg);

        // Expect pong within heartbeatTimeout
        _heartbeatTimeoutTimer?.cancel();
        _heartbeatTimeoutTimer = Timer(heartbeatTimeout, () {
          // Pong timed out; force reconnect
          _handleSocketClosed();
        });
      }
    });
  }

  void _handleSocketError(dynamic error) {
    _scheduleReconnect();
  }

  void _handleSocketClosed() {
    _heartbeatTimer?.cancel();
    _heartbeatTimeoutTimer?.cancel();
    _transport = null;
    _scheduleReconnect();
  }

  /// Exponential backoff with random jitter up to maxReconnectDelay.
  Duration calculateReconnectDelay() {
    final base = minReconnectDelay.inMilliseconds.toDouble();
    final maxMs = maxReconnectDelay.inMilliseconds.toDouble();
    final expDelay = base * math.pow(1.5, _reconnectAttempts);
    final cappedDelay = math.min(expDelay, maxMs);
    final jitter = (math.Random().nextDouble() * 0.25 + 0.875); // 87.5% - 112.5%
    return Duration(milliseconds: (cappedDelay * jitter).toInt());
  }

  void _scheduleReconnect() {
    if (_isDisposed) return;
    _setStatus(WebSocketStatus.reconnecting);
    _reconnectTimer?.cancel();

    final delay = calculateReconnectDelay();
    _reconnectAttempts++;

    _reconnectTimer = Timer(delay, () {
      connect();
    });
  }

  void _setStatus(WebSocketStatus newStatus) {
    if (_status != newStatus) {
      _status = newStatus;
      _statusController.add(_status);
    }
  }

  /// Disconnects client gracefully.
  Future<void> disconnect() async {
    _reconnectTimer?.cancel();
    _heartbeatTimer?.cancel();
    _heartbeatTimeoutTimer?.cancel();
    if (_transport != null) {
      await _transport!.close(1000, 'User initiated disconnect');
      _transport = null;
    }
    _setStatus(WebSocketStatus.disconnected);
  }

  /// Disposes client and closes all broadcast streams.
  Future<void> dispose() async {
    _isDisposed = true;
    await disconnect();
    await _statusController.close();
    await _tickerController.close();
    await _depthController.close();
    await _tradeController.close();
    await _orderController.close();
    await _sequenceGapController.close();
  }
}

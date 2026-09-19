use demo_matching_execution_engine::{
    clob::OrderBook,
    demo_analytics::DemoAnalyticsEngine,
    demo_engine::DemoMatchingEngine,
    liquidity_bot::{BotConfig, LiquidityBot},
    types::{Order, OrderStatus, OrderType, Side},
};

#[test]
fn test_order_book_fifo_price_time_priority() {
    let mut book = OrderBook::new("BTC/USDT".to_string());

    // Place two resting limit sell orders at the same price: 60,000 USDT (60,000 * 1e8)
    let order1 = Order::new(
        "sell-1".to_string(),
        "seller-A".to_string(),
        "BTC/USDT".to_string(),
        Side::Sell,
        OrderType::Limit,
        60_000_00000000,
        1_00000000, // 1 BTC
        1000,
        false,
    );
    let order2 = Order::new(
        "sell-2".to_string(),
        "seller-B".to_string(),
        "BTC/USDT".to_string(),
        Side::Sell,
        OrderType::Limit,
        60_000_00000000,
        2_00000000, // 2 BTC
        1001,
        false,
    );

    let (_, rest1) = book.place_limit_order(order1);
    assert!(rest1.is_some());
    let (_, rest2) = book.place_limit_order(order2);
    assert!(rest2.is_some());

    // Incoming buy order for 1.5 BTC at 60,000
    let buy_order = Order::new(
        "buy-1".to_string(),
        "buyer-1".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Limit,
        60_000_00000000,
        1_50000000, // 1.5 BTC
        1002,
        false,
    );

    let (matches, resting_buy) = book.place_limit_order(buy_order);
    assert!(resting_buy.is_none());
    assert_eq!(matches.len(), 2);

    // FIFO verification: first match must be against order1 (seller-A) for 1 BTC
    assert_eq!(matches[0].seller_id, "seller-A");
    assert_eq!(matches[0].quantity_e8, 1_00000000);
    assert_eq!(matches[0].price_e8, 60_000_00000000);

    // Second match against order2 (seller-B) for 0.5 BTC
    assert_eq!(matches[1].seller_id, "seller-B");
    assert_eq!(matches[1].quantity_e8, 50000000);

    // Remaining resting order2 should have 1.5 BTC left
    let depth = book.get_l2_depth(10, 1003);
    assert_eq!(depth.asks.len(), 1);
    assert_eq!(depth.asks[0].quantity_e8, 1_50000000);
}

#[test]
fn test_price_priority_matching() {
    let mut book = OrderBook::new("BTC/USDT".to_string());

    // Two asks: 61,000 and 60,000
    let ask_high = Order::new(
        "ask-high".to_string(),
        "seller-high".to_string(),
        "BTC/USDT".to_string(),
        Side::Sell,
        OrderType::Limit,
        61_000_00000000,
        1_00000000,
        1000,
        false,
    );
    let ask_low = Order::new(
        "ask-low".to_string(),
        "seller-low".to_string(),
        "BTC/USDT".to_string(),
        Side::Sell,
        OrderType::Limit,
        60_000_00000000,
        1_00000000,
        1001,
        false,
    );

    book.place_limit_order(ask_high);
    book.place_limit_order(ask_low);

    // Incoming buy at 62,000 should match ask_low (60,000) first due to price priority
    let buy = Order::new(
        "buy-market".to_string(),
        "buyer-1".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Limit,
        62_000_00000000,
        1_00000000,
        1002,
        false,
    );

    let (matches, _) = book.place_limit_order(buy);
    assert_eq!(matches.len(), 1);
    assert_eq!(matches[0].seller_id, "seller-low");
    assert_eq!(matches[0].price_e8, 60_000_00000000);
}

#[test]
fn test_order_cancellation() {
    let mut book = OrderBook::new("BTC/USDT".to_string());

    let order = Order::new(
        "order-cancel-me".to_string(),
        "user-1".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Limit,
        59_000_00000000,
        2_00000000,
        1000,
        false,
    );

    book.place_limit_order(order);
    assert_eq!(book.total_orders(), 1);

    let cancelled = book.cancel_order("order-cancel-me").unwrap();
    assert_eq!(cancelled.status, OrderStatus::Cancelled);
    assert_eq!(book.total_orders(), 0);

    // Second cancel should fail
    let err = book.cancel_order("order-cancel-me");
    assert!(err.is_err());
}

#[test]
fn test_self_trade_prevention() {
    let mut book = OrderBook::new("BTC/USDT".to_string());

    let resting_sell = Order::new(
        "sell-stp".to_string(),
        "trader-sam".to_string(),
        "BTC/USDT".to_string(),
        Side::Sell,
        OrderType::Limit,
        60_000_00000000,
        1_00000000,
        1000,
        false,
    );
    book.place_limit_order(resting_sell);

    // Same user places buy order: STP must prevent match
    let buy_stp = Order::new(
        "buy-stp".to_string(),
        "trader-sam".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Limit,
        60_000_00000000,
        1_00000000,
        1001,
        false,
    );

    let (matches, resting_buy) = book.place_limit_order(buy_stp);
    assert!(matches.is_empty(), "STP should prevent self matches");
    assert!(resting_buy.is_some(), "Order rests in book rather than self-matching");
}

#[test]
fn test_market_order_walks_book() {
    let mut book = OrderBook::new("BTC/USDT".to_string());

    // Add 3 levels of asks
    book.place_limit_order(Order::new(
        "ask-1".to_string(), "s1".to_string(), "BTC/USDT".to_string(),
        Side::Sell, OrderType::Limit, 50_000_00000000, 1_00000000, 1000, false
    ));
    book.place_limit_order(Order::new(
        "ask-2".to_string(), "s2".to_string(), "BTC/USDT".to_string(),
        Side::Sell, OrderType::Limit, 51_000_00000000, 2_00000000, 1001, false
    ));
    book.place_limit_order(Order::new(
        "ask-3".to_string(), "s3".to_string(), "BTC/USDT".to_string(),
        Side::Sell, OrderType::Limit, 52_000_00000000, 3_00000000, 1002, false
    ));

    // Buy market order for 2.5 BTC
    let market_buy = Order::new(
        "mkt-1".to_string(), "buyer-mkt".to_string(), "BTC/USDT".to_string(),
        Side::Buy, OrderType::Market, 0, 2_50000000, 1003, false
    );

    let (matches, unfilled) = book.place_market_order(market_buy);
    assert_eq!(unfilled, 0);
    assert_eq!(matches.len(), 2);
    assert_eq!(matches[0].price_e8, 50_000_00000000);
    assert_eq!(matches[0].quantity_e8, 1_00000000);
    assert_eq!(matches[1].price_e8, 51_000_00000000);
    assert_eq!(matches[1].quantity_e8, 1_50000000);
}

#[test]
fn test_simulated_liquidity_bot() {
    let mut book = OrderBook::new("BTC/USDT".to_string());
    let mut bot = LiquidityBot::new(BotConfig {
        bot_id: "bot-test".to_string(),
        symbol: "BTC/USDT".to_string(),
        num_levels: 3,
        spread_bps: 10,
        level_step_bps: 5,
        base_qty_e8: 1_00000000,
        qty_multiplier: 1.0,
    });

    let mark_price = 50_000_00000000; // 50,000
    bot.refresh_quotes(&mut book, mark_price, 1000);

    let depth = book.get_l2_depth(5, 1001);
    assert_eq!(depth.bids.len(), 3);
    assert_eq!(depth.asks.len(), 3);

    // Check spread
    assert!(book.best_ask().unwrap() > mark_price);
    assert!(book.best_bid().unwrap() < mark_price);

    // Retail user can immediately execute market buy against bot quotes
    let retail_buy = Order::new(
        "retail-buy".to_string(), "retail-user".to_string(), "BTC/USDT".to_string(),
        Side::Buy, OrderType::Market, 0, 50000000, 1002, false
    );
    let (matches, unfilled) = book.place_market_order(retail_buy);
    assert_eq!(unfilled, 0);
    assert!(!matches.is_empty());
    assert_eq!(matches[0].seller_id, "bot-test");
}

#[test]
fn test_demo_matching_engine_staleness_and_analytics() {
    let mut engine = DemoMatchingEngine::new();
    engine.init_symbol("BTC/USDT", 60_000_00000000);

    // Check depth exists thanks to liquidity bot
    let depth = engine.get_l2_depth("BTC/USDT", 5).unwrap();
    assert!(!depth.bids.is_empty());
    assert!(!depth.asks.is_empty());

    // Submit retail market order
    let user_order = Order::new(
        "order-100".to_string(),
        "retail-1".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Market,
        0,
        50000000, // 0.5 BTC
        demo_matching_execution_engine::demo_engine::current_time_ms(),
        false,
    );

    let matches = engine.submit_order(user_order).unwrap();
    assert!(!matches.is_empty());

    // Test Staleness breaker by setting last mark update timestamp to 10 seconds ago
    let now = demo_matching_execution_engine::demo_engine::current_time_ms();
    engine.last_mark_update_ms.insert("BTC/USDT".to_string(), now.saturating_sub(10_000));
    let stale_order = Order::new(
        "order-stale".to_string(),
        "retail-1".to_string(),
        "BTC/USDT".to_string(),
        Side::Buy,
        OrderType::Market,
        0,
        10000000,
        demo_matching_execution_engine::demo_engine::current_time_ms(),
        false,
    );
    let res = engine.submit_order(stale_order);
    assert!(res.is_err());
    assert!(res.unwrap_err().contains("stale"));

    // Test Demo Analytics Engine
    let mut analytics = DemoAnalyticsEngine::new();
    analytics.record_closed_trade("trader-1", 1250.50);
    analytics.record_closed_trade("trader-1", -300.00);
    analytics.record_closed_trade("trader-2", 2500.00);

    let leaderboard = analytics.get_leaderboard(5);
    assert_eq!(leaderboard.len(), 2);
    assert_eq!(leaderboard[0].user_id, "trader-2");
    assert_eq!(leaderboard[0].leaderboard_rank, 1);
    assert_eq!(leaderboard[1].user_id, "trader-1");
    assert_eq!(leaderboard[1].winning_trades, 1);
    assert_eq!(leaderboard[1].total_trades, 2);
    assert_eq!(leaderboard[1].win_rate_pct, 50.0);
}

#[test]
fn test_empty_book_and_invalid_orders() {
    let mut book = OrderBook::new("ETH/USDT".to_string());

    // Market order against empty book
    let mkt_order = Order::new(
        "mkt-empty".to_string(), "user-1".to_string(), "ETH/USDT".to_string(),
        Side::Buy, OrderType::Market, 0, 10_00000000, 1000, false
    );
    let (matches, unfilled) = book.place_market_order(mkt_order);
    assert!(matches.is_empty());
    assert_eq!(unfilled, 10_00000000);

    // Limit order with 0 quantity
    let zero_qty_order = Order::new(
        "zero-qty".to_string(), "user-1".to_string(), "ETH/USDT".to_string(),
        Side::Buy, OrderType::Limit, 3000_00000000, 0, 1001, false
    );
    let (matches, resting) = book.place_limit_order(zero_qty_order);
    assert!(matches.is_empty());
    assert_eq!(resting.unwrap().status, OrderStatus::Rejected);
}

#[test]
fn test_bot_quote_replacement_on_price_shift() {
    let mut book = OrderBook::new("BTC/USDT".to_string());
    let mut bot = LiquidityBot::new(BotConfig {
        bot_id: "bot-mm".to_string(),
        symbol: "BTC/USDT".to_string(),
        num_levels: 3,
        spread_bps: 20,
        level_step_bps: 10,
        base_qty_e8: 1_00000000,
        qty_multiplier: 1.0,
    });

    // Initial quote at 60k
    bot.refresh_quotes(&mut book, 60_000_00000000, 1000);
    assert_eq!(book.total_orders(), 6); // 3 bids, 3 asks

    // Price moves to 65k
    bot.refresh_quotes(&mut book, 65_000_00000000, 2000);
    assert_eq!(book.total_orders(), 6); // Prior quotes purged, exactly 6 fresh quotes remain
    assert!(book.best_bid().unwrap() > 60_000_00000000);
}


use order_matching_engine::*;

#[test]
fn test_strict_fifo_price_time_priority() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Place two sell orders at identical price 50,000
    // Maker 1 arrives at t=100
    let maker1 = Order::new(
        1,
        "alice",
        "BTC-USDT",
        Side::Sell,
        OrderType::Limit,
        STPMode::None,
        50_000_00000000,
        2_00000000, // 2 BTC
        100,
    );
    // Maker 2 arrives at t=200
    let maker2 = Order::new(
        2,
        "bob",
        "BTC-USDT",
        Side::Sell,
        OrderType::Limit,
        STPMode::None,
        50_000_00000000,
        3_00000000, // 3 BTC
        200,
    );

    book.process_order(maker1);
    book.process_order(maker2);

    // Taker buys 3 BTC at 50,000
    let taker = Order::new(
        3,
        "charlie",
        "BTC-USDT",
        Side::Buy,
        OrderType::Limit,
        STPMode::None,
        50_000_00000000,
        3_00000000, // 3 BTC
        300,
    );

    let res = book.process_order(taker);
    assert_eq!(res.status, OrderStatus::Filled);
    assert_eq!(res.trades.len(), 2);

    // First trade must fill Alice (FIFO priority)
    assert_eq!(res.trades[0].maker_order_id, 1);
    assert_eq!(res.trades[0].quantity_e8, 2_00000000);

    // Second trade partially fills Bob (1 BTC remaining of 3 BTC)
    assert_eq!(res.trades[1].maker_order_id, 2);
    assert_eq!(res.trades[1].quantity_e8, 1_00000000);

    let (bids, asks) = book.get_l2_depth(5);
    assert_eq!(bids.len(), 0);
    assert_eq!(asks.len(), 1);
    assert_eq!(asks[0], (50_000_00000000, 2_00000000)); // Bob has 2 BTC left
}

#[test]
fn test_multi_level_sweep_and_price_improvement() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Three ask levels: 50,000 (1 BTC), 50,100 (2 BTC), 50,200 (3 BTC)
    book.process_order(Order::new(1, "s1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 1_00000000, 1));
    book.process_order(Order::new(2, "s2", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_100_00000000, 2_00000000, 2));
    book.process_order(Order::new(3, "s3", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_200_00000000, 3_00000000, 3));

    // Aggressive taker buys 4 BTC with limit price 50,300 (sweeps level 1 and 2, partially level 3)
    let taker = Order::new(4, "buyer", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_300_00000000, 4_00000000, 4);
    let res = book.process_order(taker);

    assert_eq!(res.status, OrderStatus::Filled);
    assert_eq!(res.trades.len(), 3);
    assert_eq!(res.trades[0].price_e8, 50_000_00000000);
    assert_eq!(res.trades[0].quantity_e8, 1_00000000);
    assert_eq!(res.trades[1].price_e8, 50_100_00000000);
    assert_eq!(res.trades[1].quantity_e8, 2_00000000);
    assert_eq!(res.trades[2].price_e8, 50_200_00000000);
    assert_eq!(res.trades[2].quantity_e8, 1_00000000);

    let (_, asks) = book.get_l2_depth(5);
    assert_eq!(asks.len(), 1);
    assert_eq!(asks[0], (50_200_00000000, 2_00000000));
}

#[test]
fn test_self_trade_prevention_cancel_newest() {
    let mut book = OrderBook::new("ETH-USDT".to_string());

    // Alice places resting sell order
    let maker = Order::new(1, "alice", "ETH-USDT", Side::Sell, OrderType::Limit, STPMode::CancelNewest, 3_000_00000000, 5_00000000, 10);
    book.process_order(maker);

    // Alice inadvertently submits matching buy order with CancelNewest
    let taker = Order::new(2, "alice", "ETH-USDT", Side::Buy, OrderType::Limit, STPMode::CancelNewest, 3_000_00000000, 2_00000000, 20);
    let res = book.process_order(taker);

    // Zero trades executed, taker cancelled via STP
    assert_eq!(res.trades.len(), 0);
    assert_eq!(res.status, OrderStatus::Canceled);
    assert!(res.resting_order.is_none());

    // Maker order must still remain on book intact
    let (_, asks) = book.get_l2_depth(5);
    assert_eq!(asks.len(), 1);
    assert_eq!(asks[0], (3_000_00000000, 5_00000000));
}

#[test]
fn test_self_trade_prevention_cancel_oldest() {
    let mut book = OrderBook::new("ETH-USDT".to_string());

    // Alice places resting sell order at 3,000
    let maker_alice = Order::new(1, "alice", "ETH-USDT", Side::Sell, OrderType::Limit, STPMode::CancelOldest, 3_000_00000000, 2_00000000, 10);
    // Bob places resting sell order at 3,000
    let maker_bob = Order::new(2, "bob", "ETH-USDT", Side::Sell, OrderType::Limit, STPMode::CancelOldest, 3_000_00000000, 2_00000000, 11);
    book.process_order(maker_alice);
    book.process_order(maker_bob);

    // Alice sends aggressive buy order for 2 BTC with CancelOldest
    let taker_alice = Order::new(3, "alice", "ETH-USDT", Side::Buy, OrderType::Limit, STPMode::CancelOldest, 3_000_00000000, 2_00000000, 20);
    let res = book.process_order(taker_alice);

    // Alice's resting maker (order 1) must be cancelled via STP
    assert_eq!(res.stp_canceled_maker_ids, vec![1]);
    // Alice's taker continues matching and fills Bob!
    assert_eq!(res.trades.len(), 1);
    assert_eq!(res.trades[0].maker_order_id, 2);
    assert_eq!(res.trades[0].quantity_e8, 2_00000000);
}

#[test]
fn test_self_trade_prevention_decrement_and_cancel() {
    let mut book = OrderBook::new("SOL-USDT".to_string());

    // Alice resting sell 5.0 SOL
    let maker = Order::new(1, "alice", "SOL-USDT", Side::Sell, OrderType::Limit, STPMode::DecrementAndCancel, 150_00000000, 5_00000000, 1);
    book.process_order(maker);

    // Alice incoming buy 3.0 SOL with DecrementAndCancel
    let taker = Order::new(2, "alice", "SOL-USDT", Side::Buy, OrderType::Limit, STPMode::DecrementAndCancel, 150_00000000, 3_00000000, 2);
    let res = book.process_order(taker);

    assert_eq!(res.trades.len(), 0);
    // Resting maker should now have 2.0 SOL remaining
    let (_, asks) = book.get_l2_depth(5);
    assert_eq!(asks.len(), 1);
    assert_eq!(asks[0], (150_00000000, 2_00000000));
}

#[test]
fn test_order_types_ioc_immediate_or_cancel() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Resting sell 2.0 BTC at 60,000
    book.process_order(Order::new(1, "s1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 60_000_00000000, 2_00000000, 1));

    // IOC Buy for 5.0 BTC at 60,000
    let ioc_order = Order::new(2, "buyer", "BTC-USDT", Side::Buy, OrderType::ImmediateOrCancel, STPMode::None, 60_000_00000000, 5_00000000, 2);
    let res = book.process_order(ioc_order);

    // Fills 2.0 BTC, cancels remainder (3.0 BTC), does not rest in bids!
    assert_eq!(res.trades.len(), 1);
    assert_eq!(res.trades[0].quantity_e8, 2_00000000);
    assert_eq!(res.status, OrderStatus::PartiallyFilled);
    assert!(res.resting_order.is_none());

    let (bids, asks) = book.get_l2_depth(5);
    assert_eq!(bids.len(), 0);
    assert_eq!(asks.len(), 0);
}

#[test]
fn test_order_types_fok_fill_or_kill() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Resting sell 2.0 BTC
    book.process_order(Order::new(1, "s1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 60_000_00000000, 2_00000000, 1));

    // FOK Buy for 3.0 BTC (exceeds available 2.0 BTC) -> must kill completely with 0 fills
    let fok_fail = Order::new(2, "buyer", "BTC-USDT", Side::Buy, OrderType::FillOrKill, STPMode::None, 60_000_00000000, 3_00000000, 2);
    let res_fail = book.process_order(fok_fail);
    assert_eq!(res_fail.trades.len(), 0);
    assert_eq!(res_fail.status, OrderStatus::Canceled);

    // Book remains unchanged
    let (_, asks) = book.get_l2_depth(5);
    assert_eq!(asks[0], (60_000_00000000, 2_00000000));

    // FOK Buy for exactly 2.0 BTC -> fills completely
    let fok_ok = Order::new(3, "buyer2", "BTC-USDT", Side::Buy, OrderType::FillOrKill, STPMode::None, 60_000_00000000, 2_00000000, 3);
    let res_ok = book.process_order(fok_ok);
    assert_eq!(res_ok.status, OrderStatus::Filled);
    assert_eq!(res_ok.trades.len(), 1);
}

#[test]
fn test_order_types_post_only() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Resting sell at 60,000
    book.process_order(Order::new(1, "s1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 60_000_00000000, 1_00000000, 1));

    // Post-Only Buy at 61,000 would cross -> rejected immediately
    let bad_post = Order::new(2, "b1", "BTC-USDT", Side::Buy, OrderType::PostOnly, STPMode::None, 61_000_00000000, 1_00000000, 2);
    let res = book.process_order(bad_post);
    assert!(res.rejected);
    assert_eq!(res.status, OrderStatus::Rejected);

    // Post-Only Buy at 59,000 does not cross -> rests on book
    let good_post = Order::new(3, "b2", "BTC-USDT", Side::Buy, OrderType::PostOnly, STPMode::None, 59_000_00000000, 1_00000000, 3);
    let res_good = book.process_order(good_post);
    assert!(!res_good.rejected);

    let (bids, _) = book.get_l2_depth(5);
    assert_eq!(bids.len(), 1);
    assert_eq!(bids[0], (59_000_00000000, 1_00000000));
}

#[test]
fn test_cancel_replace_priority_retention() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Order 1 arrives at t=100 (5 BTC at 50,000)
    let o1 = Order::new(1, "u1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 5_00000000, 100);
    // Order 2 arrives at t=200 (5 BTC at 50,000)
    let o2 = Order::new(2, "u2", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 5_00000000, 200);

    book.process_order(o1);
    book.process_order(o2);

    // Order 1 is amended: quantity reduced from 5 BTC to 3 BTC at same price!
    // Priority retention rule: Order 1 MUST keep its front-of-queue priority!
    let (mod_order, _) = book.cancel_replace_order(1, 50_000_00000000, 3_00000000, 300).unwrap();
    assert_eq!(mod_order.unwrap().remaining_e8(), 3_00000000);

    // Incoming sell for 4 BTC at 50,000
    let seller = Order::new(3, "seller", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 4_00000000, 400);
    let res = book.process_order(seller);

    // Order 1 must be filled first for 3 BTC, then Order 2 for 1 BTC!
    assert_eq!(res.trades.len(), 2);
    assert_eq!(res.trades[0].maker_order_id, 1);
    assert_eq!(res.trades[0].quantity_e8, 3_00000000);
    assert_eq!(res.trades[1].maker_order_id, 2);
    assert_eq!(res.trades[1].quantity_e8, 1_00000000);
}

#[test]
fn test_cancel_replace_priority_loss_on_size_increase() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Order 1 arrives at t=100 (2 BTC at 50,000)
    let o1 = Order::new(1, "u1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 2_00000000, 100);
    // Order 2 arrives at t=200 (5 BTC at 50,000)
    let o2 = Order::new(2, "u2", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 5_00000000, 200);

    book.process_order(o1);
    book.process_order(o2);

    // Order 1 increases quantity from 2 BTC to 4 BTC!
    // Priority loss rule: Order 1 moves to back of queue behind Order 2!
    book.cancel_replace_order(1, 50_000_00000000, 4_00000000, 300).unwrap();

    // Incoming sell for 5 BTC at 50,000
    let seller = Order::new(3, "seller", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 5_00000000, 400);
    let res = book.process_order(seller);

    // Order 2 must now be filled first for 5 BTC!
    assert_eq!(res.trades.len(), 1);
    assert_eq!(res.trades[0].maker_order_id, 2);
    assert_eq!(res.trades[0].quantity_e8, 5_00000000);
}

#[test]
fn test_price_collar_risk_controls() {
    // Reference price 50,000 USDT, max deviation 1000 bps (10%)
    // Allowed band: 45,000 to 55,000
    let mut book = OrderBook::with_config("BTC-USDT".to_string(), MatchingAlgorithm::Fifo, 50_000_00000000, 1000);

    // Buy at 56,000 (> 55,000 upper bound) -> REJECTED
    let bad_buy = Order::new(1, "b1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 56_000_00000000, 1_00000000, 1);
    let res_buy = book.process_order(bad_buy);
    assert!(res_buy.rejected);

    // Sell at 44,000 (< 45,000 lower bound) -> REJECTED
    let bad_sell = Order::new(2, "s1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 44_000_00000000, 1_00000000, 2);
    let res_sell = book.process_order(bad_sell);
    assert!(res_sell.rejected);

    // Buy at 54,000 (within 10%) -> ACCEPTED
    let good_buy = Order::new(3, "b2", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 54_000_00000000, 1_00000000, 3);
    let res_good = book.process_order(good_buy);
    assert!(!res_good.rejected);
}

#[test]
fn test_dynamic_fee_hot_reloading() {
    let mut book = OrderBook::new("BTC-USDT".to_string());

    // Initially fees are 0 bps per ARCHITECTURE.md
    FeeEngine::set_maker_fee_bps(0);
    FeeEngine::set_taker_fee_bps(0);

    book.process_order(Order::new(1, "m1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 1_00000000, 1));
    let t1 = book.process_order(Order::new(2, "t1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 1_00000000, 2));

    assert_eq!(t1.trades[0].maker_fee_e8, 0);
    assert_eq!(t1.trades[0].taker_fee_e8, 0);
    assert_eq!(t1.trades[0].quote_amount_e8, 50_000_00000000);

    // Hot-reload governance fees: 10 bps maker (0.10%), 20 bps taker (0.20%)
    MatchingEngine::hot_reload_fees(10, 20);

    book.process_order(Order::new(3, "m2", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 1_00000000, 3));
    let t2 = book.process_order(Order::new(4, "t2", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 1_00000000, 4));

    // Quote = 50,000 USDT (50_000_00000000)
    // Maker fee = 50,000 * 10 / 10,000 = 50 USDT (50_00000000)
    // Taker fee = 50,000 * 20 / 10,000 = 100 USDT (100_00000000)
    assert_eq!(t2.trades[0].maker_fee_e8, 50_00000000);
    assert_eq!(t2.trades[0].taker_fee_e8, 100_00000000);

    // Reset back to 0
    MatchingEngine::hot_reload_fees(0, 0);
}

#[test]
fn test_pro_rata_queue_allocation() {
    let mut book = OrderBook::with_config("BTC-USDT".to_string(), MatchingAlgorithm::ProRata, 0, 0);

    // Two makers at price 50,000:
    // Maker 1 has 10 BTC
    // Maker 2 has 30 BTC
    // Total available = 40 BTC (ratio 1 : 3)
    book.process_order(Order::new(1, "m1", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 10_00000000, 1));
    book.process_order(Order::new(2, "m2", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 50_000_00000000, 30_00000000, 2));

    // Taker buys 20 BTC
    let taker = Order::new(3, "taker", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 50_000_00000000, 20_00000000, 3);
    let res = book.process_order(taker);

    assert_eq!(res.status, OrderStatus::Filled);
    assert_eq!(res.trades.len(), 2);

    // Maker 1 allocated: 20 * 10 / 40 = 5 BTC
    // Maker 2 allocated: 20 * 30 / 40 = 15 BTC
    let trade_m1 = res.trades.iter().find(|t| t.maker_order_id == 1).unwrap();
    let trade_m2 = res.trades.iter().find(|t| t.maker_order_id == 2).unwrap();

    assert_eq!(trade_m1.quantity_e8, 5_00000000);
    assert_eq!(trade_m2.quantity_e8, 15_00000000);

    // Remaining on book: Maker 1 has 5 BTC, Maker 2 has 15 BTC (total 20 BTC)
    let (_, asks) = book.get_l2_depth(5);
    assert_eq!(asks[0], (50_000_00000000, 20_00000000));
}

#[test]
fn test_snapshot_and_delta_client_synchronization() {
    let mut book = OrderBook::new("BTC-USDT".to_string());
    let mut client = ClientBookSync::new("BTC-USDT");

    // Place initial orders
    book.process_order(Order::new(1, "m1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 49_000_00000000, 2_00000000, 1));
    book.process_order(Order::new(2, "m2", "BTC-USDT", Side::Sell, OrderType::Limit, STPMode::None, 51_000_00000000, 3_00000000, 2));

    let deltas_1 = book.drain_deltas();
    for d in deltas_1 {
        client.buffer_delta(d);
    }

    // Client fetches snapshot
    let snapshot = book.generate_l2_snapshot(10);
    client.apply_snapshot(snapshot).unwrap();
    assert_eq!(client.status, SyncStatus::Synchronized);

    // New order arrives
    book.process_order(Order::new(3, "m3", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 49_500_00000000, 1_50000000, 3));
    let deltas_2 = book.drain_deltas();
    for d in &deltas_2 {
        client.apply_delta(d).unwrap();
    }

    let (c_bids, _) = client.get_top_levels(5);
    assert_eq!(c_bids[0].price_e8, 49_500_00000000);
    assert_eq!(c_bids[0].total_quantity_e8, 1_50000000);

    // Gap detection: simulate missing sequence packet
    let bad_delta = BookDelta {
        symbol: "BTC-USDT".to_string(),
        first_update_id: client.last_update_id + 5, // Gap!
        last_update_id: client.last_update_id + 5,
        timestamp_ns: 999,
        bids: vec![],
        asks: vec![],
    };

    let sync_res = client.apply_delta(&bad_delta);
    assert!(sync_res.is_err());
    assert_eq!(client.status, SyncStatus::Desynchronized);
}

#[test]
fn test_zero_ghost_liquidity_on_cancel() {
    let mut book = OrderBook::new("BTC-USDT".to_string());
    let mut client = ClientBookSync::new("BTC-USDT");

    book.process_order(Order::new(1, "m1", "BTC-USDT", Side::Buy, OrderType::Limit, STPMode::None, 48_000_00000000, 1_00000000, 1));
    let snapshot = book.generate_l2_snapshot(10);
    let _ = book.drain_deltas();
    client.apply_snapshot(snapshot).unwrap();

    // Cancel order -> delta emitted with quantity 0
    book.cancel_order(1);
    let deltas = book.drain_deltas();
    assert_eq!(deltas.len(), 1);
    assert_eq!(deltas[0].bids, vec![(48_000_00000000, 0)]);

    client.apply_delta(&deltas[0]).unwrap();
    let (c_bids, _) = client.get_top_levels(5);
    assert_eq!(c_bids.len(), 0); // Zero ghost liquidity confirmed!
}

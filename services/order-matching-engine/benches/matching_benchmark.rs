use order_matching_engine::*;
use std::time::Instant;

fn main() {
    println!("============================================================");
    println!("Growww / NBSE Matching Engine - Microsecond Latency Benchmark");
    println!("============================================================");

    benchmark_order_matching_latency();
    benchmark_high_throughput_burst();
    benchmark_order_cancellation_latency();
    benchmark_snapshot_generation();
    benchmark_client_delta_synchronization();

    println!("============================================================");
    println!("All matching engine benchmarks completed successfully!");
    println!("============================================================");
}

fn benchmark_order_matching_latency() {
    let mut engine = MatchingEngine::new(4);
    engine.register_symbol("BTC-USDT", MatchingAlgorithm::Fifo, 60_000_00000000, 1000);

    let num_orders = 50_000;
    let mut latencies_ns: Vec<u64> = Vec::with_capacity(num_orders);

    // Warm up book with resting maker sell orders
    for i in 0..1_000 {
        let maker = Order::new(
            i as u64,
            format!("maker_{}", i),
            "BTC-USDT",
            Side::Sell,
            OrderType::Limit,
            STPMode::CancelNewest,
            60_000_00000000 + ((i % 100) as u64) * 1_00000000,
            1_00000000,
            0,
        );
        engine.submit_order(maker);
    }

    // Measure aggressive taker matching latency
    for i in 0..num_orders {
        let taker = Order::new(
            1_000_000 + i as u64,
            format!("taker_{}", i),
            "BTC-USDT",
            Side::Buy,
            OrderType::Limit,
            STPMode::CancelNewest,
            60_000_00000000 + ((i % 100) as u64) * 1_00000000,
            50000000, // 0.5 BTC
            0,
        );

        let start = Instant::now();
        let _res = engine.submit_order(taker);
        let elapsed = start.elapsed().as_nanos() as u64;
        latencies_ns.push(elapsed);
    }

    latencies_ns.sort_unstable();
    let p50 = latencies_ns[(num_orders as f64 * 0.50) as usize];
    let p90 = latencies_ns[(num_orders as f64 * 0.90) as usize];
    let p95 = latencies_ns[(num_orders as f64 * 0.95) as usize];
    let p99 = latencies_ns[(num_orders as f64 * 0.99) as usize];
    let p999 = latencies_ns[(num_orders as f64 * 0.999) as usize];
    let min = latencies_ns[0];
    let max = latencies_ns[num_orders - 1];

    println!("[1] Aggressive Order Matching Latency ({} iterations):", num_orders);
    println!("    Min:   {:>7.2} µs ({} ns)", min as f64 / 1000.0, min);
    println!("    p50:   {:>7.2} µs ({} ns)", p50 as f64 / 1000.0, p50);
    println!("    p90:   {:>7.2} µs ({} ns)", p90 as f64 / 1000.0, p90);
    println!("    p95:   {:>7.2} µs ({} ns)", p95 as f64 / 1000.0, p95);
    println!("    p99:   {:>7.2} µs ({} ns)", p99 as f64 / 1000.0, p99);
    println!("    p99.9: {:>7.2} µs ({} ns)", p999 as f64 / 1000.0, p999);
    println!("    Max:   {:>7.2} µs ({} ns)", max as f64 / 1000.0, max);
    assert!(p99 < 100_000, "p99 latency target breached: {} ns", p99);
}

fn benchmark_high_throughput_burst() {
    let mut engine = MatchingEngine::new(4);
    let total_orders = 200_000;

    let start = Instant::now();
    for i in 0..total_orders {
        let side = if i % 2 == 0 { Side::Buy } else { Side::Sell };
        let price = if side == Side::Buy {
            59_000_00000000 + ((i % 50) as u64) * 10_000000
        } else {
            61_000_00000000 + ((i % 50) as u64) * 10_000000
        };

        let order = Order::new(
            i as u64,
            format!("user_{}", i % 500),
            "ETH-USDT",
            side,
            OrderType::Limit,
            STPMode::CancelNewest,
            price,
            1_00000000,
            0,
        );
        engine.submit_order(order);
    }
    let duration = start.elapsed();
    let ops_per_sec = (total_orders as f64) / duration.as_secs_f64();

    println!("[2] High-Throughput Limit Order Insertion Burst:");
    println!("    Total Orders:  {}", total_orders);
    println!("    Duration:      {:.3} s", duration.as_secs_f64());
    println!("    Throughput:    {:.0} orders/sec", ops_per_sec);
    assert!(ops_per_sec > 30_000.0, "Throughput below expected baseline: {}", ops_per_sec);
}

fn benchmark_order_cancellation_latency() {
    let mut engine = MatchingEngine::new(4);
    let num_orders = 20_000;

    for i in 0..num_orders {
        let order = Order::new(
            i as u64,
            "cancel_tester",
            "SOL-USDT",
            Side::Buy,
            OrderType::Limit,
            STPMode::None,
            100_00000000 + ((i % 100) as u64) * 10_00000,
            1_00000000,
            0,
        );
        engine.submit_order(order);
    }

    let mut cancel_latencies: Vec<u64> = Vec::with_capacity(num_orders);
    for i in 0..num_orders {
        let start = Instant::now();
        let _res = engine.cancel_order("SOL-USDT", i as u64);
        let elapsed = start.elapsed().as_nanos() as u64;
        cancel_latencies.push(elapsed);
    }

    cancel_latencies.sort_unstable();
    let p50 = cancel_latencies[(num_orders as f64 * 0.50) as usize];
    let p99 = cancel_latencies[(num_orders as f64 * 0.99) as usize];

    println!("[3] Order Cancellation Latency ({} iterations):", num_orders);
    println!("    p50:   {:>7.2} µs ({} ns)", p50 as f64 / 1000.0, p50);
    println!("    p99:   {:>7.2} µs ({} ns)", p99 as f64 / 1000.0, p99);
}

fn benchmark_snapshot_generation() {
    let mut engine = MatchingEngine::new(4);
    // Populate deep book (1,000 price levels)
    for p in 0..1_000 {
        let bid = Order::new(
            p as u64,
            "market_maker",
            "BTC-USDT",
            Side::Buy,
            OrderType::Limit,
            STPMode::None,
            50_000_00000000 + p * 10_00000000,
            10_00000000,
            0,
        );
        let ask = Order::new(
            10_000 + p as u64,
            "market_maker",
            "BTC-USDT",
            Side::Sell,
            OrderType::Limit,
            STPMode::None,
            70_000_00000000 + p * 10_00000000,
            10_00000000,
            0,
        );
        engine.submit_order(bid);
        engine.submit_order(ask);
    }

    let iterations = 10_000;
    let start = Instant::now();
    for _ in 0..iterations {
        let _snap = engine.get_l2_snapshot("BTC-USDT", 100);
    }
    let elapsed = start.elapsed();
    let avg_us = (elapsed.as_nanos() as f64) / (iterations as f64 * 1000.0);

    println!("[4] L2 Orderbook Snapshot Generation (Top 100 levels):");
    println!("    Iterations:    {}", iterations);
    println!("    Total Time:    {:.3} ms", elapsed.as_secs_f64() * 1000.0);
    println!("    Avg Latency:   {:.2} µs/snapshot", avg_us);
}

fn benchmark_client_delta_synchronization() {
    let mut client = ClientBookSync::new("BTC-USDT");
    let mut engine = MatchingEngine::new(4);

    // Populate initial state
    for i in 0..100 {
        let o = Order::new(
            i as u64,
            "init_maker",
            "BTC-USDT",
            Side::Buy,
            OrderType::Limit,
            STPMode::None,
            50_000_00000000 + i * 1_00000000,
            1_00000000,
            0,
        );
        engine.submit_order(o);
    }

    let snapshot = engine.get_l2_snapshot("BTC-USDT", 100).unwrap();
    client.apply_snapshot(snapshot).unwrap();

    let num_deltas = 20_000;
    let mut deltas = Vec::with_capacity(num_deltas);
    for i in 0..num_deltas {
        let o = Order::new(
            1_000 + i as u64,
            "trader",
            "BTC-USDT",
            Side::Buy,
            OrderType::Limit,
            STPMode::None,
            51_000_00000000 + ((i % 50) as u64) * 1_00000000,
            2_00000000,
            0,
        );
        engine.submit_order(o);
        let mut emitted = engine.drain_deltas("BTC-USDT");
        deltas.append(&mut emitted);
    }

    let start = Instant::now();
    for d in &deltas {
        let _ = client.apply_delta(d);
    }
    let elapsed = start.elapsed();
    let throughput = (deltas.len() as f64) / elapsed.as_secs_f64();

    println!("[5] Client Incremental Delta Processing Throughput:");
    println!("    Deltas Synced: {}", deltas.len());
    println!("    Duration:      {:.3} ms", elapsed.as_secs_f64() * 1000.0);
    println!("    Throughput:    {:.0} deltas/sec", throughput);
}

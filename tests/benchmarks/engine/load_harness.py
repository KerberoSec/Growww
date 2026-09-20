"""
High-Throughput Matching Engine Load & Performance Benchmark Harness (Prompt 903).
Measures:
- Sustained matching throughput (orders/sec)
- Latency percentiles in microseconds (p50, p90, p95, p99, p99.9, max)
- Order book invariant consistency under burst volume
- Evaluates NFR criteria (p99 < 1ms / 1000us)
"""

import sys
from pathlib import Path

# Ensure repository root is on sys.path
_repo_root = Path(__file__).resolve().parent.parent.parent.parent
if str(_repo_root) not in sys.path:
    sys.path.insert(0, str(_repo_root))

import json
import math
import os
import time
from dataclasses import asdict, dataclass
from typing import Dict, List, Optional
from tests.benchmarks.engine.market_data_generator import SyntheticMarketGenerator
from tests.benchmarks.engine.matching_engine_simulator import OrderBook


@dataclass
class BenchmarkResult:
    total_orders: int
    matched_trades: int
    elapsed_seconds: float
    throughput_ops: float
    p50_latency_us: float
    p90_latency_us: float
    p95_latency_us: float
    p99_latency_us: float
    p999_latency_us: float
    max_latency_us: float
    min_latency_us: float
    mean_latency_us: float
    sla_p99_passed: bool
    invariants_passed: bool
    error_message: Optional[str] = None


class MatchingEngineBenchmarkHarness:
    def __init__(self, symbols: Optional[List[str]] = None, target_orders: int = 50000):
        self.symbols = symbols or ["RELIANCE", "TCS", "HDFCBANK", "INFY", "ICICIBANK"]
        self.target_orders = target_orders
        self.generators = {sym: SyntheticMarketGenerator(symbol=sym) for sym in self.symbols}
        self.order_books = {sym: OrderBook(symbol=sym) for sym in self.symbols}

    def warmup(self, warmup_orders: int = 5000):
        """Pre-populate order books with resting liquidity."""
        for i in range(warmup_orders):
            sym = self.symbols[i % len(self.symbols)]
            event = self.generators[sym].generate_event()
            self.order_books[sym].process_event(event)

    def run_benchmark(self) -> BenchmarkResult:
        """Run the full benchmark stream and measure latency and throughput."""
        print(f"[*] Pre-generating {self.target_orders} order events across {len(self.symbols)} symbols...")
        events = []
        for i in range(self.target_orders):
            sym = self.symbols[i % len(self.symbols)]
            events.append(self.generators[sym].generate_event())

        print(f"[*] Executing benchmark against OrderBook engines...")
        latencies_ns: List[int] = []
        total_trades = 0

        start_time = time.perf_counter()
        for event in events:
            matches, lat_ns = self.order_books[event.symbol].process_event(event)
            latencies_ns.append(lat_ns)
            total_trades += len(matches)
        end_time = time.perf_counter()

        elapsed_sec = end_time - start_time
        throughput = len(events) / elapsed_sec if elapsed_sec > 0 else 0.0

        # Sort latencies for percentile calculation (convert ns -> us)
        latencies_ns.sort()
        latencies_us = [l / 1000.0 for l in latencies_ns]

        def get_percentile(p: float) -> float:
            idx = int(math.ceil(p * len(latencies_us))) - 1
            return round(latencies_us[max(0, min(idx, len(latencies_us) - 1))], 3)

        p50 = get_percentile(0.50)
        p90 = get_percentile(0.90)
        p95 = get_percentile(0.95)
        p99 = get_percentile(0.99)
        p999 = get_percentile(0.999)
        max_lat = round(latencies_us[-1], 3)
        min_lat = round(latencies_us[0], 3)
        mean_lat = round(sum(latencies_us) / len(latencies_us), 3)

        # Invariant checks across all symbols
        invariants_ok = True
        err_msg = None
        for sym, ob in self.order_books.items():
            valid, msg = ob.verify_invariants()
            if not valid:
                invariants_ok = False
                err_msg = f"{sym}: {msg}"
                break

        # Prompt 903 SLA: p99 latency < 1000 microseconds (1 millisecond)
        sla_passed = p99 < 1000.0

        result = BenchmarkResult(
            total_orders=len(events),
            matched_trades=total_trades,
            elapsed_seconds=round(elapsed_sec, 4),
            throughput_ops=round(throughput, 2),
            p50_latency_us=p50,
            p90_latency_us=p90,
            p95_latency_us=p95,
            p99_latency_us=p99,
            p999_latency_us=p999,
            max_latency_us=max_lat,
            min_latency_us=min_lat,
            mean_latency_us=mean_lat,
            sla_p99_passed=sla_passed,
            invariants_passed=invariants_ok,
            error_message=err_msg,
        )

        return result

    def export_report(self, result: BenchmarkResult, output_path: Path):
        """Export benchmark result to JSON file."""
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(asdict(result), f, indent=2)
        print(f"[+] Exported benchmark report to {output_path}")


def main():
    repo_root = Path(__file__).resolve().parent.parent.parent.parent
    report_file = repo_root / "tests" / "benchmarks" / "reports" / "benchmark_report.json"

    print("============================================================")
    print(" Growww Order Matching Engine Load Benchmark (Prompt 903)")
    print(" Target Volume: 50,000 orders | SLA: p99 < 1.000 ms")
    print("============================================================")

    harness = MatchingEngineBenchmarkHarness(target_orders=50000)
    print("[*] Warming up order books...")
    harness.warmup(5000)

    result = harness.run_benchmark()
    harness.export_report(result, report_file)

    print("\n---------------- Benchmark Metrics ----------------")
    print(f" Total Orders Processed : {result.total_orders:,}")
    print(f" Executed Trades Matched: {result.matched_trades:,}")
    print(f" Execution Time (sec)   : {result.elapsed_seconds:.4f} s")
    print(f" Sustained Throughput   : {result.throughput_ops:,.2f} orders/sec")
    print(f" Latency p50 (median)   : {result.p50_latency_us:.3f} µs")
    print(f" Latency p90            : {result.p90_latency_us:.3f} µs")
    print(f" Latency p95            : {result.p95_latency_us:.3f} µs")
    print(f" Latency p99 (SLA Gate) : {result.p99_latency_us:.3f} µs (Target: < 1000 µs)")
    print(f" Latency p99.9          : {result.p999_latency_us:.3f} µs")
    print(f" Latency Max            : {result.max_latency_us:.3f} µs")
    print("---------------------------------------------------")

    if result.sla_p99_passed and result.invariants_passed:
        print("[SUCCESS] Matching Engine satisfied all performance SLAs and invariants!")
        return 0
    else:
        print(f"[FAILURE] Benchmark did not satisfy requirements! Error: {result.error_message}")
        return 1


if __name__ == "__main__":
    sys.exit(main())

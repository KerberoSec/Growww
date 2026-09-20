use order_matching_engine::*;

fn main() {
    println!("[Growww Order Matching Engine] Starting low-latency matching engine core...");
    let engine = MatchingEngine::new(MatchingAlgorithm::Fifo);
    println!("[Growww Order Matching Engine] Initialized engine ready: algorithm={:?}", engine.algorithm);
}

from collections import defaultdict
from typing import List, Dict, Set, Tuple
import uuid
import time
try:
    from .graph_models import TradeEvent, CollusionRing
except ImportError:
    from graph_models import TradeEvent, CollusionRing

class StreamingTradeCycleDetector:
    """
    Real-time directed cycle detector over temporal trade streams
    identifying multi-party circular wash trading rings (A -> B -> C -> ... -> A).
    """

    def __init__(self, time_window_ms: int = 60_000, min_cycle_len: int = 3, max_cycle_len: int = 8):
        self.time_window_ms = time_window_ms
        self.min_cycle_len = min_cycle_len
        self.max_cycle_len = max_cycle_len
        # adjacency: symbol -> buyer -> {seller: [TradeEvent]}
        self.graph: Dict[str, Dict[str, Dict[str, List[TradeEvent]]]] = defaultdict(lambda: defaultdict(lambda: defaultdict(list)))

    def record_trade(self, trade: TradeEvent):
        # A buyer receives asset from seller -> Directed edge from seller to buyer
        self.graph[trade.symbol][trade.seller_id][trade.buyer_id].append(trade)
        self._prune_stale_trades(trade.symbol, trade.timestamp_ms)

    def _prune_stale_trades(self, symbol: str, current_time_ms: int):
        cutoff = current_time_ms - self.time_window_ms
        symbol_graph = self.graph[symbol]
        sellers_to_delete = []

        for seller, buyers in symbol_graph.items():
            buyers_to_delete = []
            for buyer, trades in buyers.items():
                active_trades = [t for t in trades if t.timestamp_ms >= cutoff]
                if active_trades:
                    buyers[buyer] = active_trades
                else:
                    buyers_to_delete.append(buyer)
            for b in buyers_to_delete:
                del buyers[b]
            if not buyers:
                sellers_to_delete.append(seller)

        for s in sellers_to_delete:
            del symbol_graph[s]

    def detect_collusion_rings(self, symbol: str) -> List[CollusionRing]:
        """
        Executes bounded DFS cycle traversal to find all simple directed cycles
        with lengths between min_cycle_len and max_cycle_len.
        """
        symbol_graph = self.graph.get(symbol, {})
        visited_nodes: Set[str] = set()
        detected_rings: List[CollusionRing] = []
        found_cycles_signatures: Set[str] = set()

        for start_node in list(symbol_graph.keys()):
            self._find_cycles_from_node(
                symbol=symbol,
                start_node=start_node,
                current_node=start_node,
                path=[start_node],
                visited=set([start_node]),
                found_cycles=found_cycles_signatures,
                detected_rings=detected_rings
            )

        return detected_rings

    def _find_cycles_from_node(
        self,
        symbol: str,
        start_node: str,
        current_node: str,
        path: List[str],
        visited: Set[str],
        found_cycles: Set[str],
        detected_rings: List[CollusionRing]
    ):
        if len(path) > self.max_cycle_len:
            return

        neighbors = self.graph[symbol].get(current_node, {})

        for next_node in neighbors.keys():
            if next_node == start_node and len(path) >= self.min_cycle_len:
                # Cycle found!
                # Canonical canonicalization: rotate cycle so smallest element is first
                cycle = path[:]
                min_idx = cycle.index(min(cycle))
                canonical_cycle = tuple(cycle[min_idx:] + cycle[:min_idx])
                sig = f"{symbol}:{','.join(canonical_cycle)}"

                if sig not in found_cycles:
                    found_cycles.add(sig)
                    ring = self._build_collusion_ring(symbol, canonical_cycle)
                    detected_rings.append(ring)

            elif next_node not in visited and len(path) < self.max_cycle_len:
                visited.add(next_node)
                path.append(next_node)
                self._find_cycles_from_node(
                    symbol, start_node, next_node, path, visited, found_cycles, detected_rings
                )
                path.pop()
                visited.remove(next_node)

    def _build_collusion_ring(self, symbol: str, cycle: Tuple[str, ...]) -> CollusionRing:
        total_volume = 0.0
        weighted_price_sum = 0.0
        n = len(cycle)

        for i in range(n):
            u = cycle[i]
            v = cycle[(i + 1) % n]
            trades = self.graph[symbol][u][v]
            if trades:
                latest = trades[-1]
                total_volume += latest.quantity
                weighted_price_sum += (latest.price * latest.quantity)

        avg_price = (weighted_price_sum / total_volume) if total_volume > 0 else 0.0

        # Confidence score based on volume symmetry and cycle tightness
        confidence = min(0.99, 0.70 + (0.05 * len(cycle)))

        return CollusionRing(
            ring_id=f"RING-{uuid.uuid4().hex[:8].upper()}",
            participant_ids=list(cycle),
            cycle_length=len(cycle),
            symbol=symbol,
            total_volume=total_volume,
            avg_price=avg_price,
            confidence_score=confidence
        )

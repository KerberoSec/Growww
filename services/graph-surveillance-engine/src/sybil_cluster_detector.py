from collections import defaultdict
from typing import List, Dict, Set
import uuid
try:
    from .graph_models import TraderNode, SybilCluster
except ImportError:
    from graph_models import TraderNode, SybilCluster

class SybilClusterDetector:
    """
    Multi-layer identity graph engine clustering accounts by shared
    device UUIDs, IP subnets, bank accounts, and wallet seeds.
    """

    def __init__(self, min_cluster_size: int = 2):
        self.min_cluster_size = min_cluster_size
        self.device_map: Dict[str, Set[str]] = defaultdict(set)
        self.ip_map: Dict[str, Set[str]] = defaultdict(set)
        self.bank_map: Dict[str, Set[str]] = defaultdict(set)
        self.wallet_map: Dict[str, Set[str]] = defaultdict(set)

    def register_trader(self, trader: TraderNode):
        if trader.device_uuid:
            self.device_map[trader.device_uuid].add(trader.trader_id)
        if trader.ip_subnet:
            self.ip_map[trader.ip_subnet].add(trader.trader_id)
        if trader.bank_account_hash:
            self.bank_map[trader.bank_account_hash].add(trader.trader_id)
        if trader.wallet_address:
            self.wallet_map[trader.wallet_address].add(trader.trader_id)

    def detect_clusters(self) -> List[SybilCluster]:
        clusters: List[SybilCluster] = []

        # High risk: Shared Bank Account
        for bank_hash, traders in self.bank_map.items():
            if len(traders) >= self.min_cluster_size:
                clusters.append(SybilCluster(
                    cluster_id=f"SYBIL-BANK-{uuid.uuid4().hex[:6].upper()}",
                    shared_attribute_type="BANK_ACCOUNT_HASH",
                    shared_attribute_value=bank_hash,
                    trader_ids=sorted(list(traders)),
                    risk_score=95.0
                ))

        # High risk: Shared Device UUID
        for dev_id, traders in self.device_map.items():
            if len(traders) >= self.min_cluster_size:
                clusters.append(SybilCluster(
                    cluster_id=f"SYBIL-DEV-{uuid.uuid4().hex[:6].upper()}",
                    shared_attribute_type="DEVICE_UUID",
                    shared_attribute_value=dev_id,
                    trader_ids=sorted(list(traders)),
                    risk_score=85.0
                ))

        # Medium risk: Shared IP Subnet
        for subnet, traders in self.ip_map.items():
            if len(traders) >= max(3, self.min_cluster_size):
                clusters.append(SybilCluster(
                    cluster_id=f"SYBIL-IP-{uuid.uuid4().hex[:6].upper()}",
                    shared_attribute_type="IP_SUBNET",
                    shared_attribute_value=subnet,
                    trader_ids=sorted(list(traders)),
                    risk_score=60.0
                ))

        return clusters

from dataclasses import dataclass, field
from typing import List, Dict, Optional
import time

@dataclass
class TraderNode:
    trader_id: str
    pan_hash: Optional[str] = None
    device_uuid: Optional[str] = None
    ip_subnet: Optional[str] = None
    bank_account_hash: Optional[str] = None
    wallet_address: Optional[str] = None

@dataclass
class TradeEvent:
    trade_id: str
    symbol: str
    buyer_id: str
    seller_id: str
    quantity: float
    price: float
    timestamp_ms: int = field(default_factory=lambda: int(time.time() * 1000))

    @property
    def notional_value(self) -> float:
        return self.quantity * self.price

@dataclass
class CollusionRing:
    ring_id: str
    participant_ids: List[str]
    cycle_length: int
    symbol: str
    total_volume: float
    avg_price: float
    confidence_score: float # 0.0 to 1.0
    detected_at_ms: int = field(default_factory=lambda: int(time.time() * 1000))

@dataclass
class SybilCluster:
    cluster_id: str
    shared_attribute_type: str # e.g. "DEVICE_UUID", "IP_SUBNET", "BANK_HASH"
    shared_attribute_value: str
    trader_ids: List[str]
    risk_score: float # 0.0 to 100.0

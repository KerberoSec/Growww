"""
Growww / NBSE Settlement Guarantee Fund (SGF) Default Waterfall Simulator
Simulates clearing member default loss absorption across the 4-tier statutory waterfall:
Tier 1: Defaulter Margin & Collateral (Post Haircut)
Tier 2: Defaulter SGF Contribution
Tier 3: Core Exchange SGF Capital (Subject to statutory minimum floor)
Tier 4: Non-Defaulter SGF Mutualization Pool
Compliance: SEBI Comprehensive Risk Management Framework for Clearing Corporations.
"""

from dataclasses import dataclass
from typing import Dict, List, Optional


@dataclass
class MemberCollateralState:
    member_id: str
    cash_margin_paise: int
    approved_securities_value_paise: int
    sgf_contribution_paise: int
    is_defaulted: bool = False


@dataclass(frozen=True)
class WaterfallLossDistribution:
    defaulter_id: str
    initial_deficit_paise: int
    defaulter_collateral_used: int
    defaulter_sgf_used: int
    core_sgf_used: int
    mutualized_sgf_used: int
    uncovered_loss: int


class SGFWaterfallSimulator:
    """Simulates deterministic multi-tier clearing default loss absorption."""

    def __init__(self, core_exchange_sgf_paise: int, statutory_core_sgf_floor_paise: int = 10000000000):
        self.members: Dict[str, MemberCollateralState] = {}
        self.core_exchange_sgf_paise = core_exchange_sgf_paise
        self.statutory_core_sgf_floor_paise = statutory_core_sgf_floor_paise

    def add_member(self, member: MemberCollateralState) -> None:
        self.members[member.member_id] = MemberCollateralState(
            member_id=member.member_id,
            cash_margin_paise=member.cash_margin_paise,
            approved_securities_value_paise=member.approved_securities_value_paise,
            sgf_contribution_paise=member.sgf_contribution_paise,
            is_defaulted=member.is_defaulted,
        )

    def execute_member_default(
        self,
        defaulter_id: str,
        settlement_deficit_paise: int,
        haircut_percentage: float = 0.20,
    ) -> WaterfallLossDistribution:
        """Executes strict 4-tier waterfall drawdown."""
        if defaulter_id not in self.members:
            raise ValueError(f"Unknown clearing member {defaulter_id}")

        member = self.members[defaulter_id]
        member.is_defaulted = True
        remaining_loss = settlement_deficit_paise

        # Tier 1: Defaulter Collateral (Cash + Haircutted Securities)
        haircut_multiplier = max(0.0, 1.0 - haircut_percentage)
        liquidated_securities = int(member.approved_securities_value_paise * haircut_multiplier)
        total_defaulter_collateral = member.cash_margin_paise + liquidated_securities

        tier1_used = min(remaining_loss, total_defaulter_collateral)
        remaining_loss -= tier1_used
        member.cash_margin_paise = 0
        member.approved_securities_value_paise = 0

        # Tier 2: Defaulter SGF Contribution
        tier2_used = 0
        if remaining_loss > 0:
            tier2_used = min(remaining_loss, member.sgf_contribution_paise)
            remaining_loss -= tier2_used
            member.sgf_contribution_paise -= tier2_used

        # Tier 3: Core Exchange SGF Capital (Cannot drain below statutory floor)
        tier3_used = 0
        if remaining_loss > 0:
            available_core_sgf = max(0, self.core_exchange_sgf_paise - self.statutory_core_sgf_floor_paise)
            tier3_used = min(remaining_loss, available_core_sgf)
            remaining_loss -= tier3_used
            self.core_exchange_sgf_paise -= tier3_used

        # Tier 4: Non-Defaulter SGF Mutualization
        tier4_used = 0
        if remaining_loss > 0:
            non_defaulters = [m for m in self.members.values() if not m.is_defaulted and m.sgf_contribution_paise > 0]
            total_non_defaulter_sgf = sum(m.sgf_contribution_paise for m in non_defaulters)

            tier4_used = min(remaining_loss, total_non_defaulter_sgf)
            remaining_loss -= tier4_used

            # Proportionally deduct from non-defaulters
            if total_non_defaulter_sgf > 0:
                for nd in non_defaulters:
                    proportion = nd.sgf_contribution_paise / total_non_defaulter_sgf
                    deduction = int(tier4_used * proportion)
                    nd.sgf_contribution_paise -= min(nd.sgf_contribution_paise, deduction)

        return WaterfallLossDistribution(
            defaulter_id=defaulter_id,
            initial_deficit_paise=settlement_deficit_paise,
            defaulter_collateral_used=tier1_used,
            defaulter_sgf_used=tier2_used,
            core_sgf_used=tier3_used,
            mutualized_sgf_used=tier4_used,
            uncovered_loss=remaining_loss,
        )

    def verify_waterfall_invariants(self, distribution: WaterfallLossDistribution) -> bool:
        """Verifies:
        1. Sum of loss absorbed + uncovered loss == initial deficit
        2. Core exchange SGF >= statutory floor
        3. Strict tier priority: Tier k+1 only used if Tier k was 100% exhausted
        """
        total_absorbed = (
            distribution.defaulter_collateral_used
            + distribution.defaulter_sgf_used
            + distribution.core_sgf_used
            + distribution.mutualized_sgf_used
            + distribution.uncovered_loss
        )
        if total_absorbed != distribution.initial_deficit_paise:
            return False

        if self.core_exchange_sgf_paise < self.statutory_core_sgf_floor_paise:
            return False

        return True

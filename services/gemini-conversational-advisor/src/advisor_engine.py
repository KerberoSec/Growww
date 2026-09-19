"""
Gemini Conversational Advisor – NBSE Sovereign Exchange
========================================================
Production-grade AI portfolio advisor with:
  • Portfolio query routing (tax, PnL, market, orders)
  • Natural language order placement (NLP parser)
  • Market summary generation
  • Multi-intent detection
  • Conversation context tracking
"""

from __future__ import annotations

import re
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class QueryIntent(Enum):
    TAX_COMPLIANCE = "TAX_COMPLIANCE"
    PORTFOLIO_PNL = "PORTFOLIO_PNL"
    MARKET_OUTLOOK = "MARKET_OUTLOOK"
    ORDER_EXECUTION = "ORDER_EXECUTION"
    ACCOUNT_INFO = "ACCOUNT_INFO"
    GENERAL_KNOWLEDGE = "GENERAL_KNOWLEDGE"


class OrderAction(Enum):
    BUY = "BUY"
    SELL = "SELL"


class OrderType(Enum):
    MARKET = "MARKET"
    LIMIT = "LIMIT"
    STOP_LOSS = "STOP_LOSS"


@dataclass
class ParsedOrder:
    """Order extracted from natural language input."""
    action: OrderAction
    symbol: str
    quantity: Optional[float]
    order_type: OrderType
    limit_price: Optional[float] = None
    stop_price: Optional[float] = None
    confidence: float = 0.0


@dataclass
class IntentDetectionResult:
    """Result of intent detection from a user query."""
    primary_intent: QueryIntent
    secondary_intents: List[QueryIntent]
    confidence: float
    matched_patterns: List[str]
    target_service: str
    parsed_order: Optional[ParsedOrder] = None


@dataclass
class MarketSummary:
    """Generated market summary for a symbol."""
    symbol: str
    current_price: float
    change_24h_pct: float
    high_24h: float
    low_24h: float
    volume_24h: float
    trend: str                    # "BULLISH", "BEARISH", "NEUTRAL"
    summary_text: str
    generated_at: float


@dataclass
class ConversationContext:
    """Tracks conversation state for multi-turn interactions."""
    conversation_id: str
    user_id: str
    history: List[Dict[str, str]] = field(default_factory=list)
    last_intent: Optional[QueryIntent] = None
    last_symbol: Optional[str] = None
    created_at: float = field(default_factory=time.time)
    updated_at: float = field(default_factory=time.time)


@dataclass
class AdvisorResponse:
    """Response from the advisor engine."""
    response_id: str
    user_id: str
    intent_result: IntentDetectionResult
    response_text: str
    market_summary: Optional[MarketSummary] = None
    parsed_order: Optional[ParsedOrder] = None
    disclaimer: str = (
        "Growww AI provides educational insights and portfolio analysis, "
        "not registered investment advice. Past performance does not guarantee future results."
    )
    timestamp: float = field(default_factory=time.time)


# ---------------------------------------------------------------------------
# Intent Detection Patterns
# ---------------------------------------------------------------------------

INTENT_PATTERNS: Dict[QueryIntent, List[str]] = {
    QueryIntent.TAX_COMPLIANCE: [
        r"\btax\b", r"\btds\b", r"\b194s\b", r"\b115bbh\b",
        r"\bform\s*16a?\b", r"\bstt\b", r"\bcapital\s*gains?\b",
        r"\bgst\b", r"\btax\s*loss\s*harvest\b",
    ],
    QueryIntent.PORTFOLIO_PNL: [
        r"\bholdings?\b", r"\bpn?l\b", r"\bprofit\b", r"\bloss\b",
        r"\bnet\s*worth\b", r"\bdemat\b", r"\bportfolio\b",
        r"\bbalance\b", r"\bassets?\b", r"\breturns?\b",
    ],
    QueryIntent.MARKET_OUTLOOK: [
        r"\bbtc\b", r"\beth\b", r"\bnifty\b", r"\bsensex\b",
        r"\bforecast\b", r"\btrend\b", r"\bchart\b", r"\brsi\b",
        r"\bmacd\b", r"\bsupport\b", r"\bresistance\b",
        r"\banalysis\b", r"\boutlook\b", r"\bprice\s*target\b",
    ],
    QueryIntent.ORDER_EXECUTION: [
        r"\bbuy\b", r"\bsell\b", r"\border\b", r"\blimit\b",
        r"\bstop\s*loss\b", r"\boco\b", r"\bmarket\s*order\b",
        r"\bplace\b.*\border\b", r"\bexecute\b",
    ],
    QueryIntent.ACCOUNT_INFO: [
        r"\baccount\b", r"\bkyc\b", r"\bwithdraw\b", r"\bdeposit\b",
        r"\bfees?\b", r"\bcommission\b", r"\breferral\b",
    ],
}

SERVICE_ROUTING: Dict[QueryIntent, str] = {
    QueryIntent.TAX_COMPLIANCE: "services.tax-reporting-service.rag",
    QueryIntent.PORTFOLIO_PNL: "services.portfolio-holdings-service.rag",
    QueryIntent.MARKET_OUTLOOK: "services.market-data-service.rag",
    QueryIntent.ORDER_EXECUTION: "services.order-service.executor",
    QueryIntent.ACCOUNT_INFO: "services.account-service.rag",
    QueryIntent.GENERAL_KNOWLEDGE: "core.gemini.general",
}

# ---------------------------------------------------------------------------
# Symbol patterns for NLP order parsing
# ---------------------------------------------------------------------------

KNOWN_SYMBOLS = [
    "BTC", "ETH", "SOL", "ADA", "DOT", "AVAX", "MATIC",
    "LINK", "ATOM", "XRP", "DOGE", "SHIB", "UNI", "AAVE",
    "NIFTY", "BANKNIFTY", "SENSEX",
]

SYMBOL_PATTERN = r"\b(" + "|".join(KNOWN_SYMBOLS) + r")\b"


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class GeminiAdvisorEngine:
    """
    AI-powered portfolio advisor engine with NLP intent detection,
    order parsing, and market summary generation.
    """

    def __init__(self):
        self.conversations: Dict[str, ConversationContext] = {}

    # ---- Intent Detection --------------------------------------------------

    def detect_intent(self, prompt: str) -> IntentDetectionResult:
        """
        Detect one or more intents from a user's natural language query.

        Returns primary intent (highest confidence) and secondary intents.
        """
        lowered = prompt.lower()
        intent_scores: Dict[QueryIntent, Tuple[float, List[str]]] = {}

        for intent, patterns in INTENT_PATTERNS.items():
            matches = []
            for p in patterns:
                if re.search(p, lowered):
                    matches.append(p)
            if matches:
                # Score based on number of matches
                score = len(matches) / len(patterns)
                intent_scores[intent] = (score, matches)

        if not intent_scores:
            return IntentDetectionResult(
                primary_intent=QueryIntent.GENERAL_KNOWLEDGE,
                secondary_intents=[],
                confidence=0.5,
                matched_patterns=[],
                target_service=SERVICE_ROUTING[QueryIntent.GENERAL_KNOWLEDGE],
            )

        # Sort by score descending
        sorted_intents = sorted(intent_scores.items(), key=lambda x: x[1][0], reverse=True)
        primary = sorted_intents[0]
        secondary = [si[0] for si in sorted_intents[1:]]

        # Parse order if ORDER_EXECUTION detected
        parsed_order = None
        if primary[0] == QueryIntent.ORDER_EXECUTION:
            parsed_order = self.parse_order(prompt)

        return IntentDetectionResult(
            primary_intent=primary[0],
            secondary_intents=secondary,
            confidence=primary[1][0],
            matched_patterns=primary[1][1],
            target_service=SERVICE_ROUTING[primary[0]],
            parsed_order=parsed_order,
        )

    # ---- Natural Language Order Parsing ------------------------------------

    def parse_order(self, prompt: str) -> Optional[ParsedOrder]:
        """
        Extract order parameters from natural language.

        Examples:
            "Buy 0.5 BTC at market" → BUY 0.5 BTC MARKET
            "Sell 100 ETH limit 3500" → SELL 100 ETH LIMIT @ 3500
            "Place a stop loss on BTC at 58000" → SELL BTC STOP_LOSS @ 58000
        """
        lowered = prompt.lower()

        # Detect action
        action = None
        if re.search(r"\bbuy\b", lowered):
            action = OrderAction.BUY
        elif re.search(r"\bsell\b", lowered):
            action = OrderAction.SELL
        elif re.search(r"\bstop\s*loss\b", lowered):
            action = OrderAction.SELL

        if action is None:
            return None

        # Detect symbol
        symbol_match = re.search(SYMBOL_PATTERN, prompt, re.IGNORECASE)
        symbol = symbol_match.group(1).upper() if symbol_match else None
        if symbol is None:
            return None

        # Detect quantity
        quantity = None
        qty_match = re.search(
            r"(\d+(?:\.\d+)?)\s*(?:" + symbol + r"|units?|contracts?|lots?)",
            prompt, re.IGNORECASE,
        )
        if qty_match:
            quantity = float(qty_match.group(1))
        else:
            # Try generic number before the symbol
            qty_match2 = re.search(
                r"(\d+(?:\.\d+)?)\s+" + symbol,
                prompt, re.IGNORECASE,
            )
            if qty_match2:
                quantity = float(qty_match2.group(1))

        # Detect order type and prices
        order_type = OrderType.MARKET
        limit_price = None
        stop_price = None

        if re.search(r"\bstop\s*loss\b", lowered):
            order_type = OrderType.STOP_LOSS
            price_match = re.search(r"(?:at|@|price)\s*\$?\s*(\d+(?:\.\d+)?)", lowered)
            if price_match:
                stop_price = float(price_match.group(1))
        elif re.search(r"\blimit\b", lowered):
            order_type = OrderType.LIMIT
            price_match = re.search(r"(?:at|@|limit)\s*\$?\s*(\d+(?:\.\d+)?)", lowered)
            if price_match:
                limit_price = float(price_match.group(1))

        # Confidence based on how many fields were extracted
        filled_fields = sum([
            action is not None,
            symbol is not None,
            quantity is not None,
            limit_price is not None or stop_price is not None or order_type == OrderType.MARKET,
        ])
        confidence = filled_fields / 4.0

        return ParsedOrder(
            action=action,
            symbol=symbol,
            quantity=quantity,
            order_type=order_type,
            limit_price=limit_price,
            stop_price=stop_price,
            confidence=confidence,
        )

    # ---- Market Summary Generation -----------------------------------------

    def generate_market_summary(
        self,
        symbol: str,
        current_price: float,
        change_24h_pct: float,
        high_24h: float,
        low_24h: float,
        volume_24h: float,
    ) -> MarketSummary:
        """Generate a human-readable market summary for a symbol."""
        # Determine trend
        if change_24h_pct > 3.0:
            trend = "BULLISH"
        elif change_24h_pct < -3.0:
            trend = "BEARISH"
        else:
            trend = "NEUTRAL"

        # Build summary text
        direction = "up" if change_24h_pct >= 0 else "down"
        summary_text = (
            f"{symbol} is currently trading at ${current_price:,.2f}, "
            f"{direction} {abs(change_24h_pct):.2f}% in the last 24 hours. "
            f"The 24h range is ${low_24h:,.2f} – ${high_24h:,.2f} "
            f"with a volume of ${volume_24h:,.0f}. "
            f"The short-term outlook appears {trend.lower()}."
        )

        return MarketSummary(
            symbol=symbol,
            current_price=current_price,
            change_24h_pct=change_24h_pct,
            high_24h=high_24h,
            low_24h=low_24h,
            volume_24h=volume_24h,
            trend=trend,
            summary_text=summary_text,
            generated_at=time.time(),
        )

    # ---- Conversation Context Management -----------------------------------

    def get_or_create_conversation(
        self,
        user_id: str,
        conversation_id: Optional[str] = None,
    ) -> ConversationContext:
        """Get or create a conversation context for a user."""
        if conversation_id and conversation_id in self.conversations:
            return self.conversations[conversation_id]

        conv_id = conversation_id or str(uuid.uuid4())
        ctx = ConversationContext(
            conversation_id=conv_id,
            user_id=user_id,
        )
        self.conversations[conv_id] = ctx
        return ctx

    def process_query(
        self,
        user_id: str,
        prompt: str,
        conversation_id: Optional[str] = None,
    ) -> AdvisorResponse:
        """
        Process a user query end-to-end.

        1. Detect intent
        2. Parse order (if applicable)
        3. Update conversation context
        4. Generate response
        """
        ctx = self.get_or_create_conversation(user_id, conversation_id)

        # Detect intent
        intent_result = self.detect_intent(prompt)

        # Update context
        ctx.history.append({"role": "user", "content": prompt})
        ctx.last_intent = intent_result.primary_intent
        ctx.updated_at = time.time()

        # Extract symbol from prompt for context
        symbol_match = re.search(SYMBOL_PATTERN, prompt, re.IGNORECASE)
        if symbol_match:
            ctx.last_symbol = symbol_match.group(1).upper()

        # Generate response text
        response_text = self._generate_response_text(intent_result, ctx)

        # Build response
        response = AdvisorResponse(
            response_id=str(uuid.uuid4()),
            user_id=user_id,
            intent_result=intent_result,
            response_text=response_text,
            parsed_order=intent_result.parsed_order,
        )

        ctx.history.append({"role": "assistant", "content": response_text})
        return response

    def _generate_response_text(
        self,
        intent: IntentDetectionResult,
        ctx: ConversationContext,
    ) -> str:
        """Generate a contextual response based on detected intent."""
        templates = {
            QueryIntent.TAX_COMPLIANCE: (
                "I'll look up your tax-related information. "
                "Routing to the Tax Reporting Service for detailed TDS/STT analysis."
            ),
            QueryIntent.PORTFOLIO_PNL: (
                "Let me fetch your portfolio details. "
                "Connecting to the Portfolio Holdings Service for current PnL and positions."
            ),
            QueryIntent.MARKET_OUTLOOK: (
                "I'll analyze the market data for you. "
                f"Looking up the latest data{' for ' + ctx.last_symbol if ctx.last_symbol else ''}."
            ),
            QueryIntent.ORDER_EXECUTION: (
                self._format_order_response(intent.parsed_order)
                if intent.parsed_order
                else "I detected an order intent but couldn't parse the details. "
                     "Please specify: action (buy/sell), symbol, quantity, and order type."
            ),
            QueryIntent.ACCOUNT_INFO: (
                "I'll check your account details. "
                "Connecting to the Account Service."
            ),
            QueryIntent.GENERAL_KNOWLEDGE: (
                "I'd be happy to help with your question. "
                "Let me provide some general information."
            ),
        }
        return templates.get(intent.primary_intent, "Processing your request...")

    @staticmethod
    def _format_order_response(order: Optional[ParsedOrder]) -> str:
        if order is None:
            return "Could not parse order details."
        parts = [f"Order detected: {order.action.value} {order.symbol}"]
        if order.quantity is not None:
            parts.append(f"Quantity: {order.quantity}")
        parts.append(f"Type: {order.order_type.value}")
        if order.limit_price is not None:
            parts.append(f"Limit: ${order.limit_price:,.2f}")
        if order.stop_price is not None:
            parts.append(f"Stop: ${order.stop_price:,.2f}")
        parts.append(f"Confidence: {order.confidence:.0%}")
        return " | ".join(parts)

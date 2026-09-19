"""
Gemini AI Portfolio Advisor Query Router (Prompt 299)
Routes conversational wealth queries to specialized RAG vector retrieval, tax optimization, and market outlook agents.
"""

from typing import Dict, Any, List
import re

class AdvisorQueryIntent:
    TAX_COMPLIANCE = "TAX_COMPLIANCE"
    PORTFOLIO_PNL = "PORTFOLIO_PNL"
    MARKET_OUTLOOK = "MARKET_OUTLOOK"
    ORDER_EXECUTION = "ORDER_EXECUTION"
    GENERAL_KNOWLEDGE = "GENERAL_KNOWLEDGE"

class GeminiQueryRouter:
    def __init__(self):
        self.intent_patterns = {
            AdvisorQueryIntent.TAX_COMPLIANCE: [r"\btax\b", r"\btds\b", r"\b194s\b", r"\b115bbh\b", r"\bform 16a\b", r"\bstt\b"],
            AdvisorQueryIntent.PORTFOLIO_PNL: [r"\bholdings\b", r"\bpnl\b", r"\bprofit\b", r"\bloss\b", r"\bnet worth\b", r"\bdemat\b"],
            AdvisorQueryIntent.MARKET_OUTLOOK: [r"\bbtc\b", r"\bnifty\b", r"\bforecast\b", r"\btrend\b", r"\bchart\b", r"\brsi\b"],
            AdvisorQueryIntent.ORDER_EXECUTION: [r"\bbuy\b", r"\bsell\b", r"\border\b", r"\blimit\b", r"\bstop loss\b", r"\boco\b"],
        }

    def route_query(self, user_id: str, prompt: str) -> Dict[str, Any]:
        lowered = prompt.lower()
        detected_intent = AdvisorQueryIntent.GENERAL_KNOWLEDGE

        for intent, patterns in self.intent_patterns.items():
            for p in patterns:
                if re.search(p, lowered):
                    detected_intent = intent
                    break
            if detected_intent != AdvisorQueryIntent.GENERAL_KNOWLEDGE:
                break

        # Generate routing target & prompt enhancement
        route_mapping = {
            AdvisorQueryIntent.TAX_COMPLIANCE: "services.tax-reporting-service.rag",
            AdvisorQueryIntent.PORTFOLIO_PNL: "services.portfolio-holdings-service.rag",
            AdvisorQueryIntent.MARKET_OUTLOOK: "services.market-data-service.rag",
            AdvisorQueryIntent.ORDER_EXECUTION: "services.order-service.rag",
            AdvisorQueryIntent.GENERAL_KNOWLEDGE: "core.gemini.general"
        }

        return {
            "user_id": user_id,
            "raw_prompt": prompt,
            "intent": detected_intent,
            "target_service": route_mapping[detected_intent],
            "disclaimer_required": True,
            "statutory_disclaimer": "Growww AI provides educational insights and portfolio analysis, not registered investment advice."
        }

if __name__ == "__main__":
    router = GeminiQueryRouter()
    res = router.route_query("usr-99", "How much TDS under Section 194S was deducted from my BTC trades?")
    print("Routed Query:", res)

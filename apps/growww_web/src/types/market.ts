/**
 * Growww NBSE Sovereign Exchange - Domain Market & Trading Types
 */

export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'LIMIT' | 'MARKET' | 'STOP_LOSS' | 'ICEBERG' | 'TWAP';
export type TimeInForce = 'GTC' | 'IOC' | 'FOK' | 'POST_ONLY';
export type TradingEnvironment = 'DEMO_TESTNET' | 'REAL_MAINNET';

export interface PriceLevel {
  price: number;
  quantity: number;
  total: number;
  depthPercent: number; // 0 - 100 for depth bar visualization
}

export interface OrderBookState {
  symbol: string;
  lastUpdateId: number;
  bids: PriceLevel[];
  asks: PriceLevel[];
  spread: number;
  spreadPercent: number;
}

export interface OrderSubmission {
  symbol: string;
  side: OrderSide;
  orderType: OrderType;
  price?: number;
  quantity: number;
  displayQuantity?: number; // for Iceberg
  twapDurationMinutes?: number; // for TWAP
  stopPrice?: number;
  timeInForce: TimeInForce;
  environment: TradingEnvironment;
}

export interface HoldingItem {
  assetId: string;
  symbol: string;
  assetType: 'EQUITY_DEMAT' | 'VDA_CRYPTO' | 'CBDC_INR';
  totalQuantity: number;
  lockedQuantity: number;
  availableQuantity: number;
  avgCostPriceINR: number;
  currentPriceINR: number;
  unrealizedGainINR: number;
  tax115BBHEstimateINR: number; // 31.2% flat tax on positive gains
  besuOnChainProofStatus: 'VERIFIED_1_TO_1' | 'PENDING_BATCH';
}

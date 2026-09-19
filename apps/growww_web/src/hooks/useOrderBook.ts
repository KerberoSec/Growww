import { useState, useCallback } from 'react';
import { OrderBookState, PriceLevel } from '../types/market';

export interface OrderBookDelta {
  firstUpdateId: number;
  lastUpdateId: number;
  bids: [number, number][]; // [price, quantity]
  asks: [number, number][];
}

export function useOrderBook(initialSymbol = 'BTC/USDT') {
  const [book, setBook] = useState<OrderBookState>({
    symbol: initialSymbol,
    lastUpdateId: 0,
    bids: [],
    asks: [],
    spread: 0,
    spreadPercent: 0,
  });

  const computeDepthLevels = (rawLevels: Map<number, number>, isBid: boolean): PriceLevel[] => {
    const sortedPrices = Array.from(rawLevels.keys()).sort((a, b) => isBid ? b - a : a - b);
    let cumulative = 0;
    const maxLevels = 20;
    const levels: { price: number; quantity: number; total: number }[] = [];

    for (const price of sortedPrices.slice(0, maxLevels)) {
      const qty = rawLevels.get(price) || 0;
      if (qty > 0) {
        cumulative += qty;
        levels.push({ price, quantity: qty, total: cumulative });
      }
    }

    const maxTotal = levels.length > 0 ? levels[levels.length - 1].total : 1;
    return levels.map(lvl => ({
      ...lvl,
      depthPercent: Math.min(100, (lvl.total / maxTotal) * 100),
    }));
  };

  const applySnapshot = useCallback((snapshot: {
    symbol: string;
    lastUpdateId: number;
    bids: [number, number][];
    asks: [number, number][];
  }) => {
    const bidMap = new Map<number, number>();
    const askMap = new Map<number, number>();

    for (const [p, q] of snapshot.bids) {
      if (q > 0) bidMap.set(p, q);
    }
    for (const [p, q] of snapshot.asks) {
      if (q > 0) askMap.set(p, q);
    }

    const bids = computeDepthLevels(bidMap, true);
    const asks = computeDepthLevels(askMap, false);

    const bestBid = bids.length > 0 ? bids[0].price : 0;
    const bestAsk = asks.length > 0 ? asks[0].price : 0;
    const spread = bestAsk > 0 && bestBid > 0 ? bestAsk - bestBid : 0;
    const spreadPercent = bestBid > 0 ? (spread / bestBid) * 100 : 0;

    setBook({
      symbol: snapshot.symbol,
      lastUpdateId: snapshot.lastUpdateId,
      bids,
      asks,
      spread,
      spreadPercent,
    });
  }, []);

  const applyDelta = useCallback((delta: OrderBookDelta) => {
    setBook(prev => {
      // Monotonic sequence verification: trip if gap detected
      if (delta.firstUpdateId > prev.lastUpdateId + 1) {
        console.warn(`Orderbook sequence gap detected: expected ${prev.lastUpdateId + 1}, got ${delta.firstUpdateId}. Requesting resync.`);
        return prev;
      }

      const bidMap = new Map<number, number>();
      for (const b of prev.bids) bidMap.set(b.price, b.quantity);
      for (const [p, q] of delta.bids) {
        if (q <= 0) {
          bidMap.delete(p); // Zero-ghost liquidity eviction
        } else {
          bidMap.set(p, q);
        }
      }

      const askMap = new Map<number, number>();
      for (const a of prev.asks) askMap.set(a.price, a.quantity);
      for (const [p, q] of delta.asks) {
        if (q <= 0) {
          askMap.delete(p); // Zero-ghost liquidity eviction
        } else {
          askMap.set(p, q);
        }
      }

      const bids = computeDepthLevels(bidMap, true);
      const asks = computeDepthLevels(askMap, false);
      const bestBid = bids.length > 0 ? bids[0].price : 0;
      const bestAsk = asks.length > 0 ? asks[0].price : 0;
      const spread = bestAsk > 0 && bestBid > 0 ? bestAsk - bestBid : 0;
      const spreadPercent = bestBid > 0 ? (spread / bestBid) * 100 : 0;

      return {
        ...prev,
        lastUpdateId: delta.lastUpdateId,
        bids,
        asks,
        spread,
        spreadPercent,
      };
    });
  }, []);

  return { book, applySnapshot, applyDelta };
}

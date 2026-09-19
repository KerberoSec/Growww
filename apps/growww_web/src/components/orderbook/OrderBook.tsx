import React from 'react';
import { OrderBookState } from '../../types/market';

interface OrderBookProps {
  book: OrderBookState;
  onPriceSelect?: (price: number) => void;
}

export const OrderBook: React.FC<OrderBookProps> = ({ book, onPriceSelect }) => {
  return (
    <div className="flex flex-col bg-slate-900 border border-slate-800 rounded-lg p-3 text-xs font-mono select-none w-full max-w-sm">
      {/* Header */}
      <div className="flex justify-between items-center pb-2 mb-2 border-b border-slate-800">
        <span className="font-semibold text-slate-200">Order Book ({book.symbol})</span>
        <span className="text-[10px] text-emerald-400 bg-emerald-950/40 px-1.5 py-0.5 rounded border border-emerald-800/50">
          0.00% Zero Fee
        </span>
      </div>

      {/* Columns */}
      <div className="grid grid-cols-3 text-slate-400 font-medium pb-1">
        <span>Price (USDT)</span>
        <span className="text-right">Size</span>
        <span className="text-right">Total</span>
      </div>

      {/* Asks (Sells) - Reversed so highest ask is at top */}
      <div className="flex flex-col gap-0.5 py-1">
        {book.asks.slice(0, 10).reverse().map((ask) => (
          <div
            key={ask.price}
            onClick={() => onPriceSelect?.(ask.price)}
            className="relative grid grid-cols-3 py-0.5 px-1 cursor-pointer hover:bg-rose-950/30 rounded"
          >
            <div
              className="absolute right-0 top-0 bottom-0 bg-rose-500/10 pointer-events-none rounded-r"
              style={{ width: `${ask.depthPercent}%` }}
            />
            <span className="text-rose-400 z-10">{ask.price.toFixed(2)}</span>
            <span className="text-right text-slate-300 z-10">{ask.quantity.toFixed(4)}</span>
            <span className="text-right text-slate-400 z-10">{ask.total.toFixed(4)}</span>
          </div>
        ))}
      </div>

      {/* Spread Bar */}
      <div className="my-1.5 py-1 px-2 bg-slate-800/60 border-y border-slate-750 flex justify-between items-center text-[11px]">
        <div className="flex items-center gap-1.5">
          <span className="text-slate-400">Spread:</span>
          <span className="text-amber-400 font-semibold">{book.spread.toFixed(2)}</span>
        </div>
        <span className="text-slate-400">({book.spreadPercent.toFixed(3)}%)</span>
        <span className="text-[10px] text-slate-400">ID: #{book.lastUpdateId}</span>
      </div>

      {/* Bids (Buys) */}
      <div className="flex flex-col gap-0.5 py-1">
        {book.bids.slice(0, 10).map((bid) => (
          <div
            key={bid.price}
            onClick={() => onPriceSelect?.(bid.price)}
            className="relative grid grid-cols-3 py-0.5 px-1 cursor-pointer hover:bg-emerald-950/30 rounded"
          >
            <div
              className="absolute right-0 top-0 bottom-0 bg-emerald-500/10 pointer-events-none rounded-r"
              style={{ width: `${bid.depthPercent}%` }}
            />
            <span className="text-emerald-400 z-10">{bid.price.toFixed(2)}</span>
            <span className="text-right text-slate-300 z-10">{bid.quantity.toFixed(4)}</span>
            <span className="text-right text-slate-400 z-10">{bid.total.toFixed(4)}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

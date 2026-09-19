import React from 'react';

export interface DOMPriceLevel {
  price: number;
  size: number;
  total: number;
  ordersCount: number;
}

interface DepthLadderDOMProps {
  bids: DOMPriceLevel[];
  asks: DOMPriceLevel[];
  currentPrice: number;
  onPriceClick?: (price: number, side: 'BUY' | 'SELL') => void;
}

export const DepthLadderDOM: React.FC<DepthLadderDOMProps> = ({
  bids,
  asks,
  currentPrice,
  onPriceClick,
}) => {
  const maxTotal = Math.max(
    ...bids.map(b => b.total),
    ...asks.map(a => a.total),
    1
  );

  return (
    <div className="bg-[#0B0E14] text-white rounded-lg border border-gray-800 p-4 font-mono text-xs">
      <div className="flex justify-between items-center pb-2 mb-2 border-b border-gray-800 text-gray-400 uppercase font-semibold">
        <span>Bids Size</span>
        <span>Price (USDT)</span>
        <span>Asks Size</span>
      </div>

      {/* Asks Ladder (Reverse Order) */}
      <div className="space-y-0.5 mb-2">
        {asks.slice(0, 10).reverse().map((ask, idx) => (
          <div
            key={`ask-${idx}`}
            onClick={() => onPriceClick && onPriceClick(ask.price, 'SELL')}
            className="flex justify-between items-center py-0.5 px-2 hover:bg-red-950/40 cursor-pointer rounded transition relative overflow-hidden"
          >
            <div
              className="absolute right-0 top-0 bottom-0 bg-red-900/20 z-0"
              style={{ width: `${(ask.total / maxTotal) * 100}%` }}
            />
            <span className="text-gray-500 z-10">{ask.ordersCount}</span>
            <span className="text-red-400 font-bold z-10">{ask.price.toFixed(2)}</span>
            <span className="text-gray-200 z-10">{ask.size.toFixed(4)}</span>
          </div>
        ))}
      </div>

      {/* Current Mid Market Price */}
      <div className="py-2 my-1 bg-[#141824] rounded text-center font-bold text-sm text-emerald-400 border border-emerald-900/50">
        ₹{currentPrice.toLocaleString('en-IN', { minimumFractionDigits: 2 })} (NBBO Consolidated)
      </div>

      {/* Bids Ladder */}
      <div className="space-y-0.5 mt-2">
        {bids.slice(0, 10).map((bid, idx) => (
          <div
            key={`bid-${idx}`}
            onClick={() => onPriceClick && onPriceClick(bid.price, 'BUY')}
            className="flex justify-between items-center py-0.5 px-2 hover:bg-emerald-950/40 cursor-pointer rounded transition relative overflow-hidden"
          >
            <div
              className="absolute left-0 top-0 bottom-0 bg-emerald-900/20 z-0"
              style={{ width: `${(bid.total / maxTotal) * 100}%` }}
            />
            <span className="text-gray-200 z-10">{bid.size.toFixed(4)}</span>
            <span className="text-emerald-400 font-bold z-10">{bid.price.toFixed(2)}</span>
            <span className="text-gray-500 z-10">{bid.ordersCount}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

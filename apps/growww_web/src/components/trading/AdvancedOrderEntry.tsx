import React, { useState } from 'react';

export type AdvancedOrderType = 'LIMIT' | 'MARKET' | 'STOP_LOSS_LIMIT' | 'OCO' | 'TRAILING_STOP' | 'ICEBERG' | 'TWAP';

interface AdvancedOrderEntryProps {
  symbol: string;
  currentPrice: number;
  availableBalanceINR: number;
  onSubmitOrder: (order: any) => void;
}

export const AdvancedOrderEntry: React.FC<AdvancedOrderEntryProps> = ({
  symbol,
  currentPrice,
  availableBalanceINR,
  onSubmitOrder,
}) => {
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');
  const [orderType, setOrderType] = useState<AdvancedOrderType>('LIMIT');
  const [price, setPrice] = useState<number>(currentPrice);
  const [stopPrice, setStopPrice] = useState<number>(currentPrice * 0.95);
  const [limitPriceOCO, setLimitPriceOCO] = useState<number>(currentPrice * 1.05);
  const [quantity, setQuantity] = useState<number>(0.1);
  const [icebergDisplaySize, setIcebergDisplaySize] = useState<number>(0.02);
  const [twapDurationMinutes, setTwapDurationMinutes] = useState<number>(60);
  const [postOnly, setPostOnly] = useState<boolean>(false);
  const [slippageToleranceBps, setSlippageToleranceBps] = useState<number>(20);

  const notionalINR = price * quantity * 85.0; // Simulated USD/INR 85
  const section194sTDS = side === 'SELL' ? notionalINR * 0.01 : 0; // 1% TDS on VDA sales

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmitOrder({
      symbol,
      side,
      orderType,
      price,
      quantity,
      stopPrice: orderType === 'STOP_LOSS_LIMIT' || orderType === 'OCO' ? stopPrice : undefined,
      limitPriceOCO: orderType === 'OCO' ? limitPriceOCO : undefined,
      icebergDisplaySize: orderType === 'ICEBERG' ? icebergDisplaySize : undefined,
      twapDurationMinutes: orderType === 'TWAP' ? twapDurationMinutes : undefined,
      postOnly,
      slippageToleranceBps,
      platformFeeRate: 0.00, // 0.00% Zero-fee sovereign model
      section194sTDS,
    });
  };

  return (
    <div className="bg-[#0B0E14] text-white p-5 rounded-xl border border-gray-800 font-sans text-xs">
      <div className="flex justify-between items-center mb-4 pb-3 border-b border-gray-800">
        <h3 className="text-sm font-bold text-gray-200">Advanced Order Terminal</h3>
        <span className="bg-emerald-950 text-emerald-400 border border-emerald-800 px-2 py-0.5 rounded text-[10px] font-mono font-bold">
          0.00% Platform Fee
        </span>
      </div>

      {/* Buy / Sell Tabs */}
      <div className="grid grid-cols-2 gap-2 mb-4">
        <button
          type="button"
          onClick={() => setSide('BUY')}
          className={`py-2 rounded-lg font-bold text-xs transition ${
            side === 'BUY' ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-900/30' : 'bg-gray-850 text-gray-400 hover:bg-gray-800'
          }`}
        >
          BUY {symbol.split('/')[0]}
        </button>
        <button
          type="button"
          onClick={() => setSide('SELL')}
          className={`py-2 rounded-lg font-bold text-xs transition ${
            side === 'SELL' ? 'bg-red-600 text-white shadow-lg shadow-red-900/30' : 'bg-gray-850 text-gray-400 hover:bg-gray-800'
          }`}
        >
          SELL {symbol.split('/')[0]}
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-3">
        {/* Order Type Selector */}
        <div>
          <label className="block text-gray-400 mb-1">Execution Strategy</label>
          <select
            value={orderType}
            onChange={e => setOrderType(e.target.value as AdvancedOrderType)}
            className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
          >
            <option value="LIMIT">Limit Order</option>
            <option value="MARKET">Market Order</option>
            <option value="STOP_LOSS_LIMIT">Stop-Loss Limit</option>
            <option value="OCO">OCO (One-Cancels-the-Other)</option>
            <option value="ICEBERG">Iceberg (Hidden Liquidity)</option>
            <option value="TWAP">TWAP (Time-Weighted Slicing)</option>
          </select>
        </div>

        {/* Price Input */}
        {orderType !== 'MARKET' && (
          <div>
            <label className="block text-gray-400 mb-1">Limit Price (USDT)</label>
            <input
              type="number"
              step="0.01"
              value={price}
              onChange={e => setPrice(Number(e.target.value))}
              className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
            />
          </div>
        )}

        {/* Stop Price for Stop-Loss & OCO */}
        {(orderType === 'STOP_LOSS_LIMIT' || orderType === 'OCO') && (
          <div>
            <label className="block text-gray-400 mb-1">Stop Trigger Price (USDT)</label>
            <input
              type="number"
              step="0.01"
              value={stopPrice}
              onChange={e => setStopPrice(Number(e.target.value))}
              className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
            />
          </div>
        )}

        {/* Iceberg Display Size */}
        {orderType === 'ICEBERG' && (
          <div>
            <label className="block text-gray-400 mb-1">Visible Display Size (Clip)</label>
            <input
              type="number"
              step="0.001"
              value={icebergDisplaySize}
              onChange={e => setIcebergDisplaySize(Number(e.target.value))}
              className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
            />
          </div>
        )}

        {/* TWAP Duration */}
        {orderType === 'TWAP' && (
          <div>
            <label className="block text-gray-400 mb-1">TWAP Execution Window (Minutes)</label>
            <input
              type="number"
              value={twapDurationMinutes}
              onChange={e => setTwapDurationMinutes(Number(e.target.value))}
              min={5}
              max={1440}
              className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
            />
          </div>
        )}

        {/* Quantity */}
        <div>
          <label className="block text-gray-400 mb-1">Order Quantity</label>
          <input
            type="number"
            step="0.0001"
            value={quantity}
            onChange={e => setQuantity(Number(e.target.value))}
            className="w-full bg-[#141824] border border-gray-700 rounded-lg p-2 text-white font-mono"
          />
        </div>

        {/* Post-Only & Slippage Checkbox */}
        <div className="flex items-center justify-between pt-1">
          <label className="flex items-center gap-2 text-gray-300 cursor-pointer">
            <input
              type="checkbox"
              checked={postOnly}
              onChange={e => setPostOnly(e.target.checked)}
              className="rounded bg-gray-800 border-gray-700 text-emerald-500"
            />
            <span>Post-Only (Maker)</span>
          </label>
          <span className="text-gray-400 font-mono">Max Slip: {slippageToleranceBps} bps</span>
        </div>

        {/* Statutory Tax & Fee Breakdown */}
        <div className="bg-[#141824] p-3 rounded-lg border border-gray-800 space-y-1 font-mono text-[11px]">
          <div className="flex justify-between text-gray-400">
            <span>Platform Fee:</span>
            <span className="text-emerald-400 font-bold">₹0.00 (0.00%)</span>
          </div>
          {side === 'SELL' && (
            <div className="flex justify-between text-gray-400">
              <span>Sec 194S TDS (1%):</span>
              <span className="text-red-400 font-semibold">₹{section194sTDS.toFixed(2)}</span>
            </div>
          )}
          <div className="flex justify-between text-gray-300 font-bold pt-1 border-t border-gray-800">
            <span>Est. Total INR:</span>
            <span>₹{notionalINR.toFixed(2)}</span>
          </div>
        </div>

        <button
          type="submit"
          className={`w-full py-3 rounded-lg font-bold text-sm transition mt-2 shadow-lg ${
            side === 'BUY'
              ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-950/50'
              : 'bg-red-600 hover:bg-red-500 text-white shadow-red-950/50'
          }`}
        >
          {side} {quantity} {symbol.split('/')[0]}
        </button>
      </form>
    </div>
  );
};

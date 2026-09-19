import React, { useState } from 'react';
import { OrderSide, OrderType, TimeInForce, TradingEnvironment, OrderSubmission } from '../../types/market';

interface OrderEntryFormProps {
  symbol?: string;
  selectedPrice?: number;
  environment: TradingEnvironment;
  onEnvironmentChange: (env: TradingEnvironment) => void;
  onSubmitOrder: (order: OrderSubmission) => void;
}

export const OrderEntryForm: React.FC<OrderEntryFormProps> = ({
  symbol = 'BTC/USDT',
  selectedPrice,
  environment,
  onEnvironmentChange,
  onSubmitOrder,
}) => {
  const [side, setSide] = useState<OrderSide>('BUY');
  const [orderType, setOrderType] = useState<OrderType>('LIMIT');
  const [price, setPrice] = useState<string>(selectedPrice ? selectedPrice.toString() : '65000');
  const [quantity, setQuantity] = useState<string>('0.10');
  const [displayQuantity, setDisplayQuantity] = useState<string>('0.02');
  const [twapMinutes, setTwapMinutes] = useState<string>('30');
  const [stopPrice, setStopPrice] = useState<string>('64000');
  const [tif, setTif] = useState<TimeInForce>('GTC');

  React.useEffect(() => {
    if (selectedPrice) {
      setPrice(selectedPrice.toString());
    }
  }, [selectedPrice]);

  const numPrice = parseFloat(price) || 0;
  const numQty = parseFloat(quantity) || 0;
  const notionalUSDT = numPrice * numQty;
  const notionalINR = notionalUSDT * 83.5; // Approx USD/INR conversion rate

  // Statutory Indian VDA Tax: 1% Section 194S TDS on sell gross consideration
  const tdsEstimateINR = side === 'SELL' ? notionalINR * 0.01 : 0.0;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (numQty <= 0) return;

    onSubmitOrder({
      symbol,
      side,
      orderType,
      price: orderType !== 'MARKET' ? numPrice : undefined,
      quantity: numQty,
      displayQuantity: orderType === 'ICEBERG' ? parseFloat(displayQuantity) : undefined,
      twapDurationMinutes: orderType === 'TWAP' ? parseInt(twapMinutes) : undefined,
      stopPrice: orderType === 'STOP_LOSS' ? parseFloat(stopPrice) : undefined,
      timeInForce: tif,
      environment,
    });
  };

  return (
    <div className="flex flex-col bg-slate-900 border border-slate-800 rounded-lg p-4 font-sans text-xs w-full max-w-sm">
      {/* Environment Switcher */}
      <div className="flex justify-between items-center pb-3 mb-3 border-b border-slate-800">
        <span className="font-semibold text-slate-200">Mode:</span>
        <div className="flex rounded bg-slate-850 p-0.5 border border-slate-700">
          <button
            type="button"
            onClick={() => onEnvironmentChange('DEMO_TESTNET')}
            className={`px-2.5 py-1 rounded font-medium transition ${
              environment === 'DEMO_TESTNET'
                ? 'bg-amber-500/20 text-amber-400 border border-amber-500/40'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Demo Sandbox
          </button>
          <button
            type="button"
            onClick={() => onEnvironmentChange('REAL_MAINNET')}
            className={`px-2.5 py-1 rounded font-medium transition ${
              environment === 'REAL_MAINNET'
                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Real Mainnet
          </button>
        </div>
      </div>

      {/* Buy / Sell Tabs */}
      <div className="grid grid-cols-2 gap-2 mb-3">
        <button
          type="button"
          onClick={() => setSide('BUY')}
          className={`py-2 rounded font-bold transition text-center ${
            side === 'BUY'
              ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-950'
              : 'bg-slate-800 text-slate-400 hover:bg-slate-750'
          }`}
        >
          BUY
        </button>
        <button
          type="button"
          onClick={() => setSide('SELL')}
          className={`py-2 rounded font-bold transition text-center ${
            side === 'SELL'
              ? 'bg-rose-600 text-white shadow-lg shadow-rose-950'
              : 'bg-slate-800 text-slate-400 hover:bg-slate-750'
          }`}
        >
          SELL
        </button>
      </div>

      {/* Order Type Selector */}
      <div className="grid grid-cols-5 gap-1 mb-3 bg-slate-800/80 p-1 rounded">
        {(['LIMIT', 'MARKET', 'STOP_LOSS', 'ICEBERG', 'TWAP'] as OrderType[]).map((type) => (
          <button
            key={type}
            type="button"
            onClick={() => setOrderType(type)}
            className={`py-1 rounded text-[10px] font-medium text-center ${
              orderType === type ? 'bg-slate-700 text-white font-semibold' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            {type === 'STOP_LOSS' ? 'STOP' : type}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-2.5">
        {/* Stop Price if STOP_LOSS */}
        {orderType === 'STOP_LOSS' && (
          <div>
            <label className="block text-slate-400 mb-1">Trigger Stop Price (USDT)</label>
            <input
              type="number"
              step="any"
              value={stopPrice}
              onChange={(e) => setStopPrice(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
            />
          </div>
        )}

        {/* Limit Price */}
        {orderType !== 'MARKET' && (
          <div>
            <label className="block text-slate-400 mb-1">Order Price (USDT)</label>
            <input
              type="number"
              step="any"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
            />
          </div>
        )}

        {/* Total Quantity */}
        <div>
          <label className="block text-slate-400 mb-1">Quantity ({symbol.split('/')[0]})</label>
          <input
            type="number"
            step="any"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
          />
        </div>

        {/* Iceberg Display Size */}
        {orderType === 'ICEBERG' && (
          <div>
            <label className="block text-slate-400 mb-1">Peak Visible Display Clip Size</label>
            <input
              type="number"
              step="any"
              value={displayQuantity}
              onChange={(e) => setDisplayQuantity(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
            />
          </div>
        )}

        {/* TWAP Duration */}
        {orderType === 'TWAP' && (
          <div>
            <label className="block text-slate-400 mb-1">TWAP Horizon (Minutes)</label>
            <input
              type="number"
              value={twapMinutes}
              onChange={(e) => setTwapMinutes(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
            />
          </div>
        )}

        {/* Time In Force */}
        <div>
          <label className="block text-slate-400 mb-1">Time in Force</label>
          <select
            value={tif}
            onChange={(e) => setTif(e.target.value as TimeInForce)}
            className="w-full bg-slate-800 border border-slate-700 rounded px-2 py-1.5 text-slate-100 font-mono focus:outline-none focus:border-emerald-500"
          >
            <option value="GTC">GTC (Good 'Til Cancelled)</option>
            <option value="IOC">IOC (Immediate or Cancel)</option>
            <option value="FOK">FOK (Fill or Kill)</option>
            <option value="POST_ONLY">Post Only (Maker Guaranteed)</option>
          </select>
        </div>

        {/* Cost & Statutory Breakdown */}
        <div className="bg-slate-800/60 p-2 rounded border border-slate-750 flex flex-col gap-1 text-[11px] font-mono">
          <div className="flex justify-between text-slate-400">
            <span>Estimated Notional:</span>
            <span className="text-slate-200">{notionalUSDT.toFixed(2)} USDT</span>
          </div>
          <div className="flex justify-between text-slate-400">
            <span>Platform Trading Fee:</span>
            <span className="text-emerald-400 font-semibold">0.00% (FREE)</span>
          </div>
          {side === 'SELL' && (
            <div className="flex justify-between text-amber-400/90 border-t border-slate-700 pt-1">
              <span>Section 194S (1% TDS):</span>
              <span>₹{tdsEstimateINR.toFixed(2)} INR</span>
            </div>
          )}
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          className={`w-full py-2.5 rounded font-bold text-white transition mt-1 shadow-md ${
            side === 'BUY'
              ? 'bg-emerald-600 hover:bg-emerald-500 shadow-emerald-950'
              : 'bg-rose-600 hover:bg-rose-500 shadow-rose-950'
          }`}
        >
          {side} {symbol} ({orderType})
        </button>
      </form>
    </div>
  );
};

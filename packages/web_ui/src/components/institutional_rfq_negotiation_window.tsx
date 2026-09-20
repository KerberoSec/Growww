import React, { useState, useEffect } from 'react';

export type RfqSide = 'BUY' | 'SELL' | 'TWO_WAY';
export type RfqStatus = 'DRAFT' | 'DISPATCHED' | 'QUOTING_ACTIVE' | 'ACCEPTED' | 'EXPIRED' | 'CANCELLED';

export interface LpQuote {
  quoteId: string;
  lpName: string;
  tier: 'TIER_1_MM' | 'TIER_2_LP';
  bidPrice?: number;
  askPrice?: number;
  size: number;
  spreadBps: number;
  responseTimeMs: number;
  validUntilSec: number;
}

export interface RfqEventLog {
  timestamp: string;
  message: string;
  type: 'INFO' | 'SUCCESS' | 'WARNING';
}

export function InstitutionalRfqNegotiationWindow() {
  const [rfqId, setRfqId] = useState<string>('RFQ-202609-0891');
  const [instrument, setInstrument] = useState<string>('BTC/USDT (Block Spot)');
  const [side, setSide] = useState<RfqSide>('BUY');
  const [quantity, setQuantity] = useState<number>(25.0);
  const [benchmarkPrice] = useState<number>(65240.0);
  const [settlementVenue, setSettlementVenue] = useState<string>('GROWWW_INSTITUTIONAL_CLEARING');
  const [status, setStatus] = useState<RfqStatus>('QUOTING_ACTIVE');
  const [timerSec, setTimerSec] = useState<number>(24);
  const [selectedQuoteId, setSelectedQuoteId] = useState<string | null>(null);
  const [counterPrice, setCounterPrice] = useState<string>('');

  // Selected LPs to invite
  const [invitedLps, setInvitedLps] = useState<string[]>([
    'Wintermute Institutional',
    'Jump Trading Crypto',
    'Flow Traders BV',
    'QCP Capital',
  ]);

  // Streaming LP Quotes
  const [quotes, setQuotes] = useState<LpQuote[]>([
    {
      quoteId: 'Q-WINT-01',
      lpName: 'Wintermute Institutional',
      tier: 'TIER_1_MM',
      askPrice: 65248.5,
      bidPrice: 65231.5,
      size: 25.0,
      spreadBps: 2.6,
      responseTimeMs: 240,
      validUntilSec: 24,
    },
    {
      quoteId: 'Q-JUMP-02',
      lpName: 'Jump Trading Crypto',
      tier: 'TIER_1_MM',
      askPrice: 65245.0, // Best Ask
      bidPrice: 65234.0,
      size: 30.0,
      spreadBps: 1.7,
      responseTimeMs: 180,
      validUntilSec: 24,
    },
    {
      quoteId: 'Q-FLOW-03',
      lpName: 'Flow Traders BV',
      tier: 'TIER_1_MM',
      askPrice: 65252.0,
      bidPrice: 65228.0,
      size: 25.0,
      spreadBps: 3.7,
      responseTimeMs: 310,
      validUntilSec: 24,
    },
    {
      quoteId: 'Q-QCP-04',
      lpName: 'QCP Capital',
      tier: 'TIER_1_MM',
      askPrice: 65250.0,
      bidPrice: 65230.0,
      size: 25.0,
      spreadBps: 3.1,
      responseTimeMs: 420,
      validUntilSec: 24,
    },
  ]);

  const [eventLogs, setEventLogs] = useState<RfqEventLog[]>([
    { timestamp: '11:20:01', message: 'RFQ-202609-0891 dispatched to 4 Tier-1 Market Makers', type: 'INFO' },
    { timestamp: '11:20:02', message: 'Received quote from Jump Trading Crypto: $65,245.00', type: 'SUCCESS' },
    { timestamp: '11:20:03', message: 'Received quote from Wintermute Institutional: $65,248.50', type: 'INFO' },
  ]);

  // Timer decrement effect
  useEffect(() => {
    if (status !== 'QUOTING_ACTIVE') return;
    const interval = setInterval(() => {
      setTimerSec((prev) => {
        if (prev <= 1) {
          setStatus('EXPIRED');
          setEventLogs((logs) => [
            { timestamp: new Date().toLocaleTimeString(), message: 'Quotes expired after TTL timeout', type: 'WARNING' },
            ...logs,
          ]);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [status]);

  const handleDispatchNewRfq = () => {
    const newId = `RFQ-${Date.now().toString().slice(-6)}`;
    setRfqId(newId);
    setStatus('QUOTING_ACTIVE');
    setTimerSec(30);
    setSelectedQuoteId(null);
    setEventLogs((prev) => [
      { timestamp: new Date().toLocaleTimeString(), message: `New ${side} RFQ dispatched for ${quantity} ${instrument}`, type: 'INFO' },
      ...prev,
    ]);
  };

  const handleAcceptQuote = (quote: LpQuote) => {
    setSelectedQuoteId(quote.quoteId);
    setStatus('ACCEPTED');
    const executionPrice = side === 'BUY' ? quote.askPrice : quote.bidPrice;
    setEventLogs((prev) => [
      {
        timestamp: new Date().toLocaleTimeString(),
        message: `Accepted quote from ${quote.lpName} @ $${executionPrice?.toLocaleString()} for ${quantity} ${instrument}. Settlement initiated on ${settlementVenue}.`,
        type: 'SUCCESS',
      },
      ...prev,
    ]);
  };

  const handleSendCounter = () => {
    if (!counterPrice) return;
    setEventLogs((prev) => [
      {
        timestamp: new Date().toLocaleTimeString(),
        message: `Sent counter-offer of $${counterPrice} to all quoting LPs for ${quantity} ${instrument}`,
        type: 'INFO',
      },
      ...prev,
    ]);
    setCounterPrice('');
  };

  const handleCancelRfq = () => {
    setStatus('CANCELLED');
    setEventLogs((prev) => [
      {
        timestamp: new Date().toLocaleTimeString(),
        message: `RFQ ${rfqId} cancelled by institutional desk trader`,
        type: 'WARNING',
      },
      ...prev,
    ]);
  };

  const notionalUsd = quantity * benchmarkPrice;

  return (
    <div className="p-5 rounded-lg border border-slate-800 bg-[#0B0E14] text-white font-mono text-xs space-y-4">
      {/* Header */}
      <div className="flex flex-wrap justify-between items-center border-b border-slate-800 pb-3 gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#00F0A0]" />
            <h2 className="text-sm font-bold text-white tracking-wide">
              Institutional Request-For-Quote (RFQ) Bilateral Negotiation Window
            </h2>
          </div>
          <p className="text-[11px] text-slate-400 mt-0.5">
            OTC Block Execution Desk • Multi-LP Streaming Quotes • Atomic Settlement Integration
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-slate-400 text-[11px]">RFQ ID:</span>
          <span className="px-2 py-0.5 bg-[#181F2C] border border-slate-700 text-[#00F0A0] font-bold rounded">
            {rfqId}
          </span>
          <span
            className={`px-2.5 py-1 rounded text-[10px] font-bold ${
              status === 'QUOTING_ACTIVE'
                ? 'bg-amber-400/10 text-amber-400 border border-amber-400/30 animate-pulse'
                : status === 'ACCEPTED'
                ? 'bg-[#00F0A0]/10 text-[#00F0A0] border border-[#00F0A0]/30'
                : 'bg-slate-800 text-slate-400 border border-slate-700'
            }`}
          >
            {status}
          </span>
        </div>
      </div>

      {/* RFQ Order Specification Panel */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3 bg-[#121722] p-3.5 rounded border border-slate-800">
        <div>
          <span className="text-[10px] text-slate-400 block">INSTRUMENT</span>
          <select
            value={instrument}
            onChange={(e) => setInstrument(e.target.value)}
            disabled={status === 'QUOTING_ACTIVE'}
            className="w-full bg-[#181F2C] border border-slate-700 rounded p-1 text-slate-200 mt-1 font-bold"
          >
            <option value="BTC/USDT (Block Spot)">BTC/USDT (Block Spot)</option>
            <option value="ETH/USDT (Block Spot)">ETH/USDT (Block Spot)</option>
            <option value="SOL/USDT (Block Spot)">SOL/USDT (Block Spot)</option>
            <option value="BTC-28MAR-C-70000">BTC-28MAR-C-70000</option>
            <option value="NIFTY-50-INDEX-BLOCK">NIFTY-50 Large Block</option>
          </select>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block">SIDE</span>
          <div className="flex gap-1 mt-1">
            {(['BUY', 'SELL', 'TWO_WAY'] as const).map((s) => (
              <button
                key={s}
                onClick={() => setSide(s)}
                disabled={status === 'QUOTING_ACTIVE'}
                className={`flex-1 py-1 rounded text-[10px] font-bold ${
                  side === s
                    ? s === 'BUY'
                      ? 'bg-[#00F0A0] text-black'
                      : s === 'SELL'
                      ? 'bg-rose-500 text-white'
                      : 'bg-indigo-500 text-white'
                    : 'bg-[#181F2C] text-slate-400 hover:text-white'
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block">NOTIONAL SIZE</span>
          <div className="flex items-center gap-1 mt-1">
            <input
              type="number"
              value={quantity}
              onChange={(e) => setQuantity(Math.max(1, Number(e.target.value)))}
              disabled={status === 'QUOTING_ACTIVE'}
              className="w-full bg-[#181F2C] border border-slate-700 rounded p-1 text-white font-bold"
            />
          </div>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block">EST. NOTIONAL</span>
          <div className="text-white font-bold mt-2">
            ${Math.round(notionalUsd).toLocaleString()} USD
          </div>
        </div>

        <div>
          <span className="text-[10px] text-slate-400 block">SETTLEMENT VENUE</span>
          <select
            value={settlementVenue}
            onChange={(e) => setSettlementVenue(e.target.value)}
            className="w-full bg-[#181F2C] border border-slate-700 rounded p-1 text-slate-200 mt-1"
          >
            <option value="GROWWW_INSTITUTIONAL_CLEARING">Growww Institutional Clearing</option>
            <option value="PARADIGM_ATOMIC_SETTLEMENT">Paradigm Atomic Settlement</option>
            <option value="FIREBLOCKS_OFF_EXCHANGE">Fireblocks Off-Exchange Custody</option>
          </select>
        </div>
      </div>

      {/* Countdown Timer Bar */}
      {status === 'QUOTING_ACTIVE' && (
        <div className="p-3 rounded bg-[#0E121B] border border-amber-400/20 space-y-1.5">
          <div className="flex justify-between items-center text-[11px]">
            <span className="text-amber-400 font-bold flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
              Quotes Valid For: {timerSec}s
            </span>
            <span className="text-slate-400">Benchmark Index: ${benchmarkPrice.toLocaleString()}</span>
          </div>
          <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-amber-400 transition-all duration-1000"
              style={{ width: `${(timerSec / 30) * 100}%` }}
            />
          </div>
        </div>
      )}

      {/* Streaming LP Quotes Table */}
      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <span className="font-bold text-slate-300">Live Counterparty Quotes ({quotes.length} LPs Responded)</span>
          <div className="flex items-center gap-2">
            {status === 'QUOTING_ACTIVE' && (
              <button
                onClick={handleCancelRfq}
                className="px-2.5 py-1 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30 hover:bg-rose-500/30 text-[11px]"
              >
                Cancel RFQ
              </button>
            )}
            <button
              onClick={handleDispatchNewRfq}
              className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a] text-[11px]"
            >
              🔄 Request New Quotes
            </button>
          </div>
        </div>

        <div className="border border-slate-800 rounded overflow-hidden">
          <table className="w-full text-left text-[11px]">
            <thead className="bg-[#121722] text-slate-400 border-b border-slate-800">
              <tr>
                <th className="p-2.5">Liquidity Provider</th>
                <th className="p-2.5">Tier</th>
                <th className="p-2.5">Bid Price ($)</th>
                <th className="p-2.5">Ask Price ($)</th>
                <th className="p-2.5">Spread (bps)</th>
                <th className="p-2.5">Available Size</th>
                <th className="p-2.5">Latency</th>
                <th className="p-2.5 text-right">Execution</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 bg-[#0E121B]">
              {quotes.map((quote) => {
                const isSelected = selectedQuoteId === quote.quoteId;
                const isBestAsk = quote.askPrice === Math.min(...quotes.map((q) => q.askPrice || Infinity));
                return (
                  <tr
                    key={quote.quoteId}
                    className={`hover:bg-slate-800/40 ${isSelected ? 'bg-[#00F0A0]/10' : ''}`}
                  >
                    <td className="p-2.5 font-bold text-white flex items-center gap-2">
                      <span className="w-1.5 h-1.5 rounded-full bg-[#00F0A0]" />
                      {quote.lpName}
                      {isBestAsk && side === 'BUY' && (
                        <span className="px-1 py-0.2 bg-[#00F0A0]/20 text-[#00F0A0] rounded text-[9px]">
                          BEST ASK
                        </span>
                      )}
                    </td>
                    <td className="p-2.5 text-slate-400">{quote.tier}</td>
                    <td className="p-2.5 text-slate-300 font-mono">
                      ${quote.bidPrice?.toLocaleString()}
                    </td>
                    <td className="p-2.5 text-[#00F0A0] font-bold font-mono">
                      ${quote.askPrice?.toLocaleString()}
                    </td>
                    <td className="p-2.5 text-slate-300">{quote.spreadBps} bps</td>
                    <td className="p-2.5 text-slate-300">{quote.size} BTC</td>
                    <td className="p-2.5 text-slate-400">{quote.responseTimeMs}ms</td>
                    <td className="p-2.5 text-right">
                      {status === 'ACCEPTED' && isSelected ? (
                        <span className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold text-[10px]">
                          ✓ ACCEPTED
                        </span>
                      ) : (
                        <button
                          onClick={() => handleAcceptQuote(quote)}
                          disabled={status !== 'QUOTING_ACTIVE'}
                          className="px-2.5 py-1 rounded bg-[#00F0A0] text-black font-bold hover:bg-[#00d08a] disabled:opacity-30 text-[10px]"
                        >
                          Accept @ ${side === 'BUY' ? quote.askPrice?.toLocaleString() : quote.bidPrice?.toLocaleString()}
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Negotiation & Counter-Offer Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 bg-[#121722] p-3 rounded border border-slate-800">
        <div className="flex items-center gap-2 flex-1 min-w-[280px]">
          <span className="text-slate-400 text-[11px] whitespace-nowrap">Counter-Offer Price:</span>
          <input
            type="number"
            placeholder="e.g. 65240"
            value={counterPrice}
            onChange={(e) => setCounterPrice(e.target.value)}
            disabled={status !== 'QUOTING_ACTIVE'}
            className="bg-[#181F2C] border border-slate-700 rounded px-2 py-1 text-white text-xs w-36"
          />
          <button
            onClick={handleSendCounter}
            disabled={status !== 'QUOTING_ACTIVE' || !counterPrice}
            className="px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white font-bold disabled:opacity-40 text-[11px]"
          >
            Send Counter to LPs
          </button>
        </div>

        <div className="text-slate-400 text-[10px]">
          Institutional Execution SLA: Instant DvP / PvP Settlement Guarantee
        </div>
      </div>

      {/* Audit Log / Event Feed */}
      <div className="space-y-1 bg-[#0E121B] p-3 rounded border border-slate-800">
        <span className="text-[10px] text-slate-400 font-bold block uppercase tracking-wider">
          Negotiation Audit Trail & Execution Log
        </span>
        <div className="max-h-24 overflow-y-auto space-y-1 pt-1">
          {eventLogs.map((log, idx) => (
            <div key={idx} className="flex items-center gap-2 text-[10px]">
              <span className="text-slate-500 font-mono">[{log.timestamp}]</span>
              <span
                className={
                  log.type === 'SUCCESS'
                    ? 'text-[#00F0A0]'
                    : log.type === 'WARNING'
                    ? 'text-amber-400'
                    : 'text-slate-300'
                }
              >
                {log.message}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export const WebInstitutionalRfqNegotiationWindow = InstitutionalRfqNegotiationWindow;

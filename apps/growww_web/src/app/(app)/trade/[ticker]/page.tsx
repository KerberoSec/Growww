import React from 'react';
import { TradingTerminal } from '../../../../components/trading/trading_terminal';

export default function TradePage({ params }: { params: { ticker: string } }) {
  const ticker = params?.ticker ? params.ticker.toUpperCase() : 'BTC-USDT';

  return (
    <div className="flex-1 flex flex-col h-[calc(100vh-3.5rem)]">
      <TradingTerminal ticker={ticker} />
    </div>
  );
}

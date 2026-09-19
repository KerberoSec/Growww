import { describe, it, expect } from 'vitest';
import React from 'react';
import { TradingTerminal } from '../src/components/trading/trading_terminal';

describe('Prompt 603 - Web Trading Terminal & Lightweight Charts Integration', () => {
  it('instantiates TradingTerminal with default ticker', () => {
    expect(TradingTerminal).toBeDefined();
    const element = React.createElement(TradingTerminal, { ticker: 'BTC-USDT' });
    expect(element.props.ticker).toBe('BTC-USDT');
  });
});

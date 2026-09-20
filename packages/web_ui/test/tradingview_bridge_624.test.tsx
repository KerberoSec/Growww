import { describe, it, expect } from 'vitest';
import React from 'react';
import { TradingViewBridge } from '../src/components/tradingview_bridge';

describe('Prompt 624 - Web TradingView Advanced Charts Canvas Bridge', () => {
  it('instantiates TradingViewBridge component', () => {
    expect(TradingViewBridge).toBeDefined();
    const element = React.createElement(TradingViewBridge, { symbol: 'BTC-USDT' });
    expect(element.props.symbol).toBe('BTC-USDT');
  });
});

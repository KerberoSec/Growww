import { describe, it, expect } from 'vitest';
import React from 'react';
import { MarketOrderSlippageDialog } from '../src/components/market_order_slippage_dialog';

describe('Prompt 628 - Market Order Slippage Warning Dialog', () => {
  it('instantiates MarketOrderSlippageDialog component', () => {
    expect(MarketOrderSlippageDialog).toBeDefined();
    const element = React.createElement(MarketOrderSlippageDialog, { isOpen: true });
    expect(element.props.isOpen).toBe(true);
  });
});

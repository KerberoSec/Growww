import { describe, it, expect } from 'vitest';
import React from 'react';
import { CopyTradingPortal } from '../src/components/copy_trading_portal';

describe('Prompt 616 - Pro-Trader Copy Trading Portal', () => {
  it('instantiates CopyTradingPortal component', () => {
    expect(CopyTradingPortal).toBeDefined();
    const element = React.createElement(CopyTradingPortal);
    expect(element.type).toBe(CopyTradingPortal);
  });
});

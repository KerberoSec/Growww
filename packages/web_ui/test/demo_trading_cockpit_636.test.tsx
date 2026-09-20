import { describe, it, expect } from 'vitest';
import React from 'react';
import { DemoTradingCockpit } from '../src/components/demo_trading_cockpit';

describe('Prompt 636 - Demo Paper Trading Cockpit', () => {
  it('instantiates DemoTradingCockpit component', () => {
    expect(DemoTradingCockpit).toBeDefined();
    const element = React.createElement(DemoTradingCockpit);
    expect(element.type).toBe(DemoTradingCockpit);
  });
});

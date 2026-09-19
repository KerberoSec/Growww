import { describe, it, expect } from 'vitest';
import React from 'react';
import { DemoTradingAnalytics } from '../src/components/demo_analytics';

describe('Prompt 621 - Demo Paper Trading Simulator & Analytics', () => {
  it('instantiates DemoTradingAnalytics component', () => {
    expect(DemoTradingAnalytics).toBeDefined();
    const element = React.createElement(DemoTradingAnalytics);
    expect(element.type).toBe(DemoTradingAnalytics);
  });
});

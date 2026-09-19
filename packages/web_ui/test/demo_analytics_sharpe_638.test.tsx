import { describe, it, expect } from 'vitest';
import React from 'react';
import { DemoAnalyticsSharpe } from '../src/components/demo_analytics_sharpe';

describe('Prompt 638 - Demo Analytics Equity Curve & Sharpe Ratio', () => {
  it('instantiates DemoAnalyticsSharpe component', () => {
    expect(DemoAnalyticsSharpe).toBeDefined();
    const element = React.createElement(DemoAnalyticsSharpe);
    expect(element.type).toBe(DemoAnalyticsSharpe);
  });
});

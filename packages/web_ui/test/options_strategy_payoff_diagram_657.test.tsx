import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  OptionsStrategyPayoffVisualizer,
  WebOptionsStrategyPayoffDiagramPnLVisualizer,
} from '../src/components/options_strategy_payoff_diagram';

describe('Prompt 657 - Web options strategy payoff diagram & PnL visualizer', () => {
  it('instantiates OptionsStrategyPayoffVisualizer and alias component', () => {
    expect(OptionsStrategyPayoffVisualizer).toBeDefined();
    expect(WebOptionsStrategyPayoffDiagramPnLVisualizer).toBeDefined();
    const elem = React.createElement(OptionsStrategyPayoffVisualizer);
    expect(elem.type).toBe(OptionsStrategyPayoffVisualizer);
  });

  it('renders to HTML string without throwing and displays strategy payoff details', () => {
    const html = renderToString(React.createElement(OptionsStrategyPayoffVisualizer));
    expect(html).toContain('Options Strategy Payoff Diagram &amp; PnL Visualizer');
    expect(html).toContain('Analytical Black-Scholes Greeks');
    expect(html).toContain('NET PREMIUM');
    expect(html).toContain('MAX PROFIT');
    expect(html).toContain('MAX LOSS');
    expect(html).toContain('Underlying Scrubber:');
    expect(html).toContain('Strategy Legs');
    expect(html).toContain('Days to Expiration (DTE)');
    expect(html).toContain('Implied Volatility (IV)');
  });

  it('renders SVG chart elements with zero axis and spot lines', () => {
    const html = renderToString(React.createElement(OptionsStrategyPayoffVisualizer));
    expect(html).toContain('<svg');
    expect(html).toContain('$0 PnL');
    expect(html).toContain('Spot');
    expect(html).toContain('65000');
    expect(html).toContain('At Expiry Payoff');
    expect(html).toContain('T+0 Curve');
  });
});

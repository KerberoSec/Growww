import { describe, it, expect } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  ResponsibleTradingModal,
  ResponsibleTradingConfig,
} from '../src/components/responsible_trading_modal';

describe('Prompt 660 - Web Responsible Trading & Voluntary Self-Exclusion Controls', () => {
  it('instantiates ResponsibleTradingModal component without crashing', () => {
    const el = <ResponsibleTradingModal isOpen={true} />;
    expect(el).toBeDefined();
  });

  it('renders to HTML string and displays investor protection portal controls', () => {
    const html = renderToString(<ResponsibleTradingModal isOpen={true} />);
    expect(html).toContain('Investor Protection &amp; Responsible Trading Portal');
    expect(html).toContain('Daily Loss Circuit Breaker (INR)');
    expect(html).toContain('Max Per-Order Value (INR)');
    expect(html).toContain('24-Hour Deposit Limit (INR)');
    expect(html).toContain('Voluntary Self-Exclusion &amp; Cooling-Off');
    expect(html).toContain('Save Protection Controls');
  });

  it('displays voluntary self-exclusion banner when active in initial config', () => {
    const activeConfig: ResponsibleTradingConfig = {
      dailyLossLimitInr: 25000,
      maxOrderValueInr: 50000,
      cooldownPeriodMinutes: 120,
      isSelfExclusionActive: true,
      selfExclusionExpiresAtMs: 1774000000000,
      depositLimit24hInr: 100000,
    };

    const html = renderToString(
      <ResponsibleTradingModal isOpen={true} initialConfig={activeConfig} />
    );
    expect(html).toContain('Voluntary Self-Exclusion Active');
    expect(html).toContain('Trading operations and order placement are strictly disabled');
    expect(html).toContain('₹25,000');
  });

  it('renders null when isOpen is false', () => {
    const html = renderToString(<ResponsibleTradingModal isOpen={false} />);
    expect(html).toBe('');
  });
});

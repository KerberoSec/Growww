import { describe, it, expect } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { ApiKeyDeveloperPortal } from '../src/components/api_key_developer_portal';

describe('Prompt 656 - Web API Key Developer Portal: CIDR Whitelisting & Scopes', () => {
  it('instantiates ApiKeyDeveloperPortal component without crashing', () => {
    expect(ApiKeyDeveloperPortal).toBeDefined();
  });

  it('renders to HTML string and displays initial API keys and CIDR whitelists', () => {
    const html = renderToString(<ApiKeyDeveloperPortal />);

    // Verify Title and Button
    expect(html).toContain('Developer API Key Portal');
    expect(html).toContain('Generate New API Key');

    // Verify API Key Labels and Scopes
    expect(html).toContain('Algo Trading Bot Alpha');
    expect(html).toContain('Portfolio Reporting Daemon');
    expect(html).toContain('203.0.113.15/32');
    expect(html).toContain('198.51.100.0/24');
    expect(html).toContain('50 req/s');
    expect(html).toContain('Revoke');
  });

  it('renders custom initial API keys accurately', () => {
    const customKeys = [
      {
        id: 'test_key',
        label: 'Arbitrage Bot Gamma',
        maskedKey: 'gw_live_abcd****************1234',
        scopes: ['trade' as const],
        ipWhitelist: ['172.16.0.5/32'],
        rateLimitPerSec: 100,
        createdAt: '2026-09-20',
        isActive: true,
      },
    ];

    const html = renderToString(<ApiKeyDeveloperPortal initialKeys={customKeys} />);
    expect(html).toContain('Arbitrage Bot Gamma');
    expect(html).toContain('172.16.0.5/32');
    expect(html).toContain('100 req/s');
  });
});

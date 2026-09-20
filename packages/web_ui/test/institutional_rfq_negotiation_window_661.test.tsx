import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  InstitutionalRfqNegotiationWindow,
  WebInstitutionalRfqNegotiationWindow,
} from '../src/components/institutional_rfq_negotiation_window';

describe('Prompt 661 - Web institutional RFQ negotiation window', () => {
  it('instantiates InstitutionalRfqNegotiationWindow and alias component', () => {
    expect(InstitutionalRfqNegotiationWindow).toBeDefined();
    expect(WebInstitutionalRfqNegotiationWindow).toBeDefined();
    const elem = React.createElement(InstitutionalRfqNegotiationWindow);
    expect(elem.type).toBe(InstitutionalRfqNegotiationWindow);
  });

  it('renders to HTML string without throwing and verifies RFQ negotiation blotter', () => {
    const html = renderToString(React.createElement(InstitutionalRfqNegotiationWindow));
    expect(html).toContain('Institutional Request-For-Quote (RFQ) Bilateral Negotiation Window');
    expect(html).toContain('OTC Block Execution Desk');
    expect(html).toContain('INSTRUMENT');
    expect(html).toContain('NOTIONAL SIZE');
    expect(html).toContain('SETTLEMENT VENUE');
    expect(html).toContain('Quotes Valid For:');
    expect(html).toContain('Wintermute Institutional');
    expect(html).toContain('Jump Trading Crypto');
    expect(html).toContain('Negotiation Audit Trail');
    expect(html).toContain('Send Counter to LPs');
  });
});

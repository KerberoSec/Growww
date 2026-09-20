import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  ClearingMemberCapitalAdequacyMonitor,
  FpiLimitMonitorSectoralHeadroomBar,
  WebClearingMemberCapitalAdequacyMonitor,
} from '../src/components/clearing_member_capital_adequacy_monitor';

describe('Prompt 675 - Web Clearing Member Capital Adequacy & FPI Limit Monitor', () => {
  it('instantiates ClearingMemberCapitalAdequacyMonitor and FpiLimitMonitorSectoralHeadroomBar', () => {
    expect(ClearingMemberCapitalAdequacyMonitor).toBeDefined();
    expect(FpiLimitMonitorSectoralHeadroomBar).toBeDefined();
    expect(WebClearingMemberCapitalAdequacyMonitor).toBeDefined();
    const element = React.createElement(ClearingMemberCapitalAdequacyMonitor);
    expect(element.type).toBe(ClearingMemberCapitalAdequacyMonitor);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(ClearingMemberCapitalAdequacyMonitor));
    expect(html).toContain('Clearing Member Capital Adequacy &amp; FPI Sectoral Headroom Monitor');
    expect(html).toContain('GROWWW-CM-0092');
  });
});

import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { FpiLimitMonitorSectoralHeadroomBar } from '../src/components/clearing_member_capital_adequacy_monitor';

describe('Prompt 675 - Web FPI Limit Monitor Sectoral Headroom Bar', () => {
  it('instantiates FpiLimitMonitorSectoralHeadroomBar component', () => {
    expect(FpiLimitMonitorSectoralHeadroomBar).toBeDefined();
    const element = React.createElement(FpiLimitMonitorSectoralHeadroomBar);
    expect(element.type).toBe(FpiLimitMonitorSectoralHeadroomBar);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(FpiLimitMonitorSectoralHeadroomBar));
    expect(html).toContain('FPI Sectoral Investment Cap &amp; Headroom Bar');
    expect(html).toContain('Banking &amp; Financial Services');
  });
});

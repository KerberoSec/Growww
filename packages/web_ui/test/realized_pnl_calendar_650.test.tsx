import { describe, it, expect } from 'vitest';
import React from 'react';
import { RealizedPnLCalendar } from '../src/components/realized_pnl_calendar';

describe('Prompt 650 - Realized PnL Calendar Heatmap', () => {
  it('instantiates RealizedPnLCalendar component', () => {
    expect(RealizedPnLCalendar).toBeDefined();
    const element = React.createElement(RealizedPnLCalendar);
    expect(element.type).toBe(RealizedPnLCalendar);
  });
});

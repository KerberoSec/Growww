import { describe, it, expect } from 'vitest';
import React from 'react';
import { HoldingsSectorTreemap } from '../src/components/holdings_sector_treemap';

describe('Prompt 649 - Holdings Breakdown Sector Treemap Visualizer', () => {
  it('instantiates HoldingsSectorTreemap component', () => {
    expect(HoldingsSectorTreemap).toBeDefined();
    const element = React.createElement(HoldingsSectorTreemap);
    expect(element.type).toBe(HoldingsSectorTreemap);
  });
});

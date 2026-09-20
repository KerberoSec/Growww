import { describe, it, expect } from 'vitest';
import React from 'react';
import { MarketingLandingSite } from '../src/components/landing_site';

describe('Prompt 608 - Public Marketing Landing Site & Disclosures', () => {
  it('instantiates MarketingLandingSite component', () => {
    expect(MarketingLandingSite).toBeDefined();
    const element = React.createElement(MarketingLandingSite);
    expect(element.type).toBe(MarketingLandingSite);
  });
});

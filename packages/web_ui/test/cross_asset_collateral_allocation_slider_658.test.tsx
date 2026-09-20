import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  CrossAssetCollateralSlider,
  WebCrossAssetCollateralAllocationSlider,
} from '../src/components/cross_asset_collateral_slider';

describe('Prompt 658 - Web cross-asset collateral allocation slider', () => {
  it('instantiates CrossAssetCollateralSlider and alias component', () => {
    expect(CrossAssetCollateralSlider).toBeDefined();
    expect(WebCrossAssetCollateralAllocationSlider).toBeDefined();
    const elem = React.createElement(CrossAssetCollateralSlider);
    expect(elem.type).toBe(CrossAssetCollateralSlider);
  });

  it('renders to HTML string without throwing and verifies multi-asset collateral metrics', () => {
    const html = renderToString(React.createElement(CrossAssetCollateralSlider));
    expect(html).toContain('Cross-Asset Collateral Allocation Slider &amp; Margin Optimizer');
    expect(html).toContain('TOTAL NOMINAL ASSETS');
    expect(html).toContain('EFFECTIVE COLLATERAL');
    expect(html).toContain('MARGIN UTILIZATION');
    expect(html).toContain('CAPITAL EFFICIENCY');
    expect(html).toContain('Auto-Balance Allocation');
  });

  it('displays asset haircut table and collateral categories', () => {
    const html = renderToString(React.createElement(CrossAssetCollateralSlider));
    expect(html).toContain('INR_CASH');
    expect(html).toContain('USDC');
    expect(html).toContain('WBTC');
    expect(html).toContain('ST_ETH');
    expect(html).toContain('Pledged Collateral Asset Pool &amp; Haircut Matrix');
    expect(html).toContain('Equity &amp; Index F&amp;O');
  });
});

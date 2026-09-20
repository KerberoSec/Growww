import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  DetachedMultiWindowLayoutManager,
  WebDetachedMultiWindowLayoutManagerAndStateSync,
} from '../src/components/detached_multi_window_layout_manager';

describe('Prompt 663 - Web detached multi-window layout manager & state sync', () => {
  it('instantiates DetachedMultiWindowLayoutManager and alias component', () => {
    expect(DetachedMultiWindowLayoutManager).toBeDefined();
    expect(WebDetachedMultiWindowLayoutManagerAndStateSync).toBeDefined();
    const elem = React.createElement(DetachedMultiWindowLayoutManager);
    expect(elem.type).toBe(DetachedMultiWindowLayoutManager);
  });

  it('renders to HTML string without throwing and verifies multi-monitor topology', () => {
    const html = renderToString(React.createElement(DetachedMultiWindowLayoutManager));
    expect(html).toContain('Detached Multi-Window Layout Manager &amp; Cross-Screen State Sync');
    expect(html).toContain('growww_multiwindow_bus');
    expect(html).toContain('DESK DISPLAY TOPOLOGY');
    expect(html).toContain('SHARED INSTRUMENT SYMBOL');
    expect(html).toContain('Pop-Out New Screen:');
    expect(html).toContain('TradingView 4x Multi-Chart Grid');
    expect(html).toContain('L2/L3 Ultra-DOM Orderbook Ladder');
    expect(html).toContain('Send Heartbeat Pulse');
    expect(html).toContain('Re-Dock to Main');
  });
});

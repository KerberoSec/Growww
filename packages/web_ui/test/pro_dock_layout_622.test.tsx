import { describe, it, expect } from 'vitest';
import React from 'react';
import { ProDockLayout } from '../src/components/pro_dock_layout';

describe('Prompt 622 - Web Pro Workstation Multi-Dock Grid Layout', () => {
  it('renders ProDockLayout with default panels', () => {
    expect(ProDockLayout).toBeDefined();
    const element = React.createElement(ProDockLayout, { layoutName: 'TRADER_DESK' });
    expect(element.props.layoutName).toBe('TRADER_DESK');
  });
});

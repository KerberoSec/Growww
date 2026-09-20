import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MarginPledgePortal, MarginPledgeDepositorySharesPortal } from '../src/components/margin_pledge_portal';

describe('Prompt 652 - Web Margin Pledge Depository Shares Portal', () => {
  it('instantiates MarginPledgePortal and alias component', () => {
    expect(MarginPledgePortal).toBeDefined();
    expect(MarginPledgeDepositorySharesPortal).toBeDefined();
    const element = React.createElement(MarginPledgePortal);
    expect(element.type).toBe(MarginPledgePortal);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(MarginPledgePortal));
    expect(html).toContain('Margin Pledge Portal: Demat Equity');
    expect(html).toContain('RELIANCE');
  });
});

import { describe, it, expect } from 'vitest';
import React from 'react';
import { DeveloperPortal } from '../src/components/developer_portal';

describe('Prompt 609 - Developer Portal & Testnet Faucet', () => {
  it('instantiates DeveloperPortal component', () => {
    expect(DeveloperPortal).toBeDefined();
    const element = React.createElement(DeveloperPortal);
    expect(element.type).toBe(DeveloperPortal);
  });
});

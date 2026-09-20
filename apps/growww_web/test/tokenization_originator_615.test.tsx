import { describe, it, expect } from 'vitest';
import React from 'react';
import { TokenizationOriginator } from '../src/components/tokenization_originator';

describe('Prompt 615 - RWA Asset Issuer & Tokenization Originator Portal', () => {
  it('instantiates TokenizationOriginator component', () => {
    expect(TokenizationOriginator).toBeDefined();
    const element = React.createElement(TokenizationOriginator);
    expect(element.type).toBe(TokenizationOriginator);
  });
});

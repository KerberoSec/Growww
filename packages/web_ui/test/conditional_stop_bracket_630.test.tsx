import { describe, it, expect } from 'vitest';
import React from 'react';
import { ConditionalStopBracket } from '../src/components/conditional_stop_bracket';

describe('Prompt 630 - Web Conditional Stop Trigger & OCO Bracket', () => {
  it('instantiates ConditionalStopBracket component', () => {
    expect(ConditionalStopBracket).toBeDefined();
    const element = React.createElement(ConditionalStopBracket);
    expect(element.type).toBe(ConditionalStopBracket);
  });
});

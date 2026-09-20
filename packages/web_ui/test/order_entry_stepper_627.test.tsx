import { describe, it, expect } from 'vitest';
import React from 'react';
import { OrderEntryStepper } from '../src/components/order_entry_stepper';

describe('Prompt 627 - Order Entry Stepper & Percentage Slider', () => {
  it('instantiates OrderEntryStepper component', () => {
    expect(OrderEntryStepper).toBeDefined();
    const element = React.createElement(OrderEntryStepper);
    expect(element.type).toBe(OrderEntryStepper);
  });
});

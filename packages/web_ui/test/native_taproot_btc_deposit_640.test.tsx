import { describe, it, expect } from 'vitest';
import React from 'react';
import { NativeTaprootBtcDeposit } from '../src/components/native_taproot_btc_deposit';

describe('Prompt 640 - Native Taproot BTC Deposit Screen', () => {
  it('instantiates NativeTaprootBtcDeposit component', () => {
    expect(NativeTaprootBtcDeposit).toBeDefined();
    const element = React.createElement(NativeTaprootBtcDeposit);
    expect(element.type).toBe(NativeTaprootBtcDeposit);
  });
});

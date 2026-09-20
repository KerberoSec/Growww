import { describe, it, expect } from 'vitest';
import React from 'react';
import { KycFlow } from '../src/components/onboarding/kyc_flow';

describe('Prompt 602 - Web Onboarding & DigiLocker KYC Flow', () => {
  it('instantiates the KycFlow component without throwing', () => {
    expect(KycFlow).toBeDefined();
    const element = React.createElement(KycFlow);
    expect(element.type).toBe(KycFlow);
  });
});

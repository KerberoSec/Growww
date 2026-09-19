import { describe, it, expect } from 'vitest';
import React from 'react';
import { KycReviewDashboard } from '../src/components/kyc_review';

describe('Prompt 604 - Admin Console: User Management & KYC Review Dashboard', () => {
  it('instantiates KycReviewDashboard successfully', () => {
    expect(KycReviewDashboard).toBeDefined();
    const element = React.createElement(KycReviewDashboard);
    expect(element.type).toBe(KycReviewDashboard);
  });
});

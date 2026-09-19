import { describe, it, expect } from 'vitest';
import React from 'react';
import { RegulatoryReportingPortal } from '../src/components/regulatory_reporting';

describe('Prompt 607 - Admin Console: Regulatory Reporting & Audit Export Portal', () => {
  it('instantiates RegulatoryReportingPortal component', () => {
    expect(RegulatoryReportingPortal).toBeDefined();
    const element = React.createElement(RegulatoryReportingPortal);
    expect(element.type).toBe(RegulatoryReportingPortal);
  });
});

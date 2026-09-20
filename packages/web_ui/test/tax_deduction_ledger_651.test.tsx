import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { TaxDeductionLedger } from '../src/components/tax_deduction_ledger';

describe('Prompt 651 - Web Tax Deduction Ledger Section 194S & 115BBH', () => {
  it('instantiates TaxDeductionLedger component', () => {
    expect(TaxDeductionLedger).toBeDefined();
    const element = React.createElement(TaxDeductionLedger);
    expect(element.type).toBe(TaxDeductionLedger);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(TaxDeductionLedger));
    expect(html).toContain('Tax Ledger: Section 194S &amp; 115BBH');
    expect(html).toContain('Gross VDA Turnover');
  });
});

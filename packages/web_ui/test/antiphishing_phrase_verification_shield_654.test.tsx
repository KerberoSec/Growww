import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  AntiPhishingShield,
  AntiPhishingPhraseVerificationShield,
  WebAntiPhishingPhraseVerificationShield,
} from '../src/components/antiphishing_phrase_verification_shield';

describe('Prompt 654 - Web Anti-Phishing Phrase Configuration & Verification Shield', () => {
  it('instantiates AntiPhishingShield and aliases', () => {
    expect(AntiPhishingShield).toBeDefined();
    expect(AntiPhishingPhraseVerificationShield).toBeDefined();
    expect(WebAntiPhishingPhraseVerificationShield).toBeDefined();
    const element = React.createElement(AntiPhishingShield);
    expect(element.type).toBe(AntiPhishingShield);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(AntiPhishingShield));
    expect(html).toContain('Anti-Phishing Phrase Configuration');
    expect(html).toContain('GROWWW-ALPHA-9942');
  });
});

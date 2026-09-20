import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { SecurityCenter, WebSecurityCenterTotpPasskeysYubikeyFido2 } from '../src/components/security_center';

describe('Prompt 653 - Web Security Center TOTP Passkeys YubiKey FIDO2', () => {
  it('instantiates SecurityCenter and alias component', () => {
    expect(SecurityCenter).toBeDefined();
    expect(WebSecurityCenterTotpPasskeysYubikeyFido2).toBeDefined();
    const element = React.createElement(SecurityCenter);
    expect(element.type).toBe(SecurityCenter);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(SecurityCenter));
    expect(html).toContain('Security Center: Multi-Factor Authentication');
    expect(html).toContain('Authenticator App (TOTP)');
  });
});

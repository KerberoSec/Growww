import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  ActiveSessionsManager,
  WebActiveSessionsIpGeolocationRemoteLogout,
} from '../src/components/active_sessions_manager';

describe('Prompt 655 - Web Active Sessions Manager & Remote Revocation Portal', () => {
  it('instantiates ActiveSessionsManager and alias component', () => {
    expect(ActiveSessionsManager).toBeDefined();
    expect(WebActiveSessionsIpGeolocationRemoteLogout).toBeDefined();
    const element = React.createElement(ActiveSessionsManager);
    expect(element.type).toBe(ActiveSessionsManager);
  });

  it('renders to HTML string without throwing', () => {
    const html = renderToString(React.createElement(ActiveSessionsManager));
    expect(html).toContain('Active Sessions &amp; Remote Revocation Portal');
    expect(html).toContain('Mumbai, Maharashtra, India');
  });
});

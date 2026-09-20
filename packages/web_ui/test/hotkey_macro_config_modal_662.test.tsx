import { describe, it, expect } from 'vitest';
import React from 'react';
import { renderToString } from 'react-dom/server';
import {
  HotkeyMacroConfigModal,
  WebHighFrequencyHotkeyMacroConfigurationModal,
} from '../src/components/hotkey_macro_config_modal';

describe('Prompt 662 - Web high-frequency hotkey macro configuration modal', () => {
  it('instantiates HotkeyMacroConfigModal and alias component', () => {
    expect(HotkeyMacroConfigModal).toBeDefined();
    expect(WebHighFrequencyHotkeyMacroConfigurationModal).toBeDefined();
    const elem = React.createElement(HotkeyMacroConfigModal);
    expect(elem.type).toBe(HotkeyMacroConfigModal);
  });

  it('renders to HTML string without throwing and verifies key recorder modal elements', () => {
    const html = renderToString(React.createElement(HotkeyMacroConfigModal, { isOpen: true }));
    expect(html).toContain('High-Frequency Hotkey Macro Configuration Modal');
    expect(html).toContain('Sub-millisecond Keyboard Execution Maps');
    expect(html).toContain('Desk Preset:');
    expect(html).toContain('Buy at Best Bid');
    expect(html).toContain('Panic: Cancel All Orders');
    expect(html).toContain('Shift+B');
    expect(html).toContain('Escape');
    expect(html).toContain('Last Hotkey Event Triggered:');
  });

  it('respects isOpen=false prop by rendering empty', () => {
    const html = renderToString(React.createElement(HotkeyMacroConfigModal, { isOpen: false }));
    expect(html).toBe('');
  });
});

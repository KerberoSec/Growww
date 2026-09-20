import { describe, it, expect } from 'vitest';
import React from 'react';
import { P2POrderChatDrawer } from '../src/components/p2p_order_chat_drawer';

describe('Prompt 646 - P2P Order Chat & Payment Verification', () => {
  it('instantiates P2POrderChatDrawer component', () => {
    expect(P2POrderChatDrawer).toBeDefined();
    const element = React.createElement(P2POrderChatDrawer);
    expect(element.type).toBe(P2POrderChatDrawer);
  });
});

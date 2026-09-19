import { describe, it, expect } from 'vitest';
import React from 'react';
import { RwaLaunchpad } from '../src/components/rwa_launchpad';

describe('Prompt 619 - RWA Primary Launchpad', () => {
  it('instantiates RwaLaunchpad component', () => {
    expect(RwaLaunchpad).toBeDefined();
    const element = React.createElement(RwaLaunchpad);
    expect(element.type).toBe(RwaLaunchpad);
  });
});

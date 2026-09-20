import React from 'react';
import { TokenizationOriginator } from '../../components/tokenization_originator';

export const metadata = {
  title: 'RWA Tokenization & Originator Portal | Growww',
  description: 'Deploy compliant ERC-3643 securities tokens backed 1:1 by depository shares.',
};

export default function IssuerPage() {
  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-black text-white">Asset Issuer & Origination Portal</h1>
        <p className="text-xs text-slate-400">Institutional onboarding and permissioned security token issuance.</p>
      </div>
      <TokenizationOriginator />
    </div>
  );
}

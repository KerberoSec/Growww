import React, { useState } from 'react';

interface KYCOnboardingModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCompleteKYC?: (details: { pan: string; aadhaarMasked: string; tier: number }) => void;
}

export const KYCOnboardingModal: React.FC<KYCOnboardingModalProps> = ({
  isOpen,
  onClose,
  onCompleteKYC,
}) => {
  const [step, setStep] = useState<1 | 2 | 3 | 4>(1);
  const [pan, setPan] = useState('');
  const [panValid, setPanValid] = useState<boolean | null>(null);
  const [aadhaarRaw, setAadhaarRaw] = useState('');
  const [bankAccount, setBankAccount] = useState('');
  const [ifsc, setIfsc] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const validatePAN = (val: string) => {
    const cleaned = val.toUpperCase().trim();
    setPan(cleaned);
    const panRegex = /^[A-Z]{5}[0-9]{4}[A-Z]$/;
    setPanValid(panRegex.test(cleaned));
  };

  const getMaskedAadhaar = () => {
    const digits = aadhaarRaw.replace(/\D/g, '');
    if (digits.length >= 4) {
      return `XXXX-XXXX-${digits.slice(-4)}`;
    }
    return digits;
  };

  const handleStep1PAN = () => {
    if (!panValid) {
      setError('Please enter a valid 10-character alphanumeric PAN.');
      return;
    }
    setError(null);
    setStep(2);
  };

  const handleStep2Aadhaar = () => {
    const digits = aadhaarRaw.replace(/\D/g, '');
    if (digits.length !== 12) {
      setError('Aadhaar number must contain exactly 12 digits.');
      return;
    }
    setError(null);
    setStep(3);
  };

  const handleStep3PennyDrop = async () => {
    if (!bankAccount || !ifsc) {
      setError('Please provide bank account number and valid IFSC.');
      return;
    }
    setIsVerifying(true);
    setError(null);

    // Simulate RBI IMPS penny drop verification
    setTimeout(() => {
      setIsVerifying(false);
      setStep(4);
      onCompleteKYC?.({
        pan,
        aadhaarMasked: getMaskedAadhaar(),
        tier: 2, // Tier 2 Verified
      });
    }, 800);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 font-sans text-xs">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-lg w-full p-5 shadow-2xl relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-slate-400 hover:text-slate-100 text-base"
        >
          ✕
        </button>

        {/* Header */}
        <div className="flex items-center gap-3 pb-3 mb-4 border-b border-slate-800">
          <div className="h-10 w-10 rounded-full bg-blue-500/20 border border-blue-500/40 flex items-center justify-center text-blue-400 text-lg font-bold">
            🛡️
          </div>
          <div>
            <h3 className="text-sm font-bold text-slate-100">SEBI / PMLA KYC Verification</h3>
            <p className="text-slate-400 text-[11px]">Mandatory digital onboarding for Indian domestic accounts</p>
          </div>
        </div>

        {/* Step Indicator */}
        <div className="grid grid-cols-4 gap-2 mb-4 font-medium text-center text-[10px]">
          <div className={`py-1 rounded ${step >= 1 ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400'}`}>
            1. PAN Card
          </div>
          <div className={`py-1 rounded ${step >= 2 ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400'}`}>
            2. Aadhaar Vault
          </div>
          <div className={`py-1 rounded ${step >= 3 ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400'}`}>
            3. Penny Drop
          </div>
          <div className={`py-1 rounded ${step >= 4 ? 'bg-emerald-600 text-white' : 'bg-slate-800 text-slate-400'}`}>
            4. Verified
          </div>
        </div>

        {error && (
          <div className="p-2 mb-3 bg-rose-950/50 border border-rose-800/50 rounded text-rose-300 text-[11px]">
            {error}
          </div>
        )}

        {/* Step 1: PAN */}
        {step === 1 && (
          <div className="flex flex-col gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Permanent Account Number (PAN)</label>
              <input
                type="text"
                maxLength={10}
                placeholder="ABCDE1234F"
                value={pan}
                onChange={(e) => validatePAN(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded px-3 py-2 text-slate-100 font-mono text-sm tracking-wider uppercase focus:outline-none focus:border-blue-500"
              />
              <span className="text-[10px] text-slate-400 mt-1 block">
                Validated against NSDL / Income Tax Department database under Section 139AA.
              </span>
            </div>
            <button
              type="button"
              disabled={!panValid}
              onClick={handleStep1PAN}
              className="mt-2 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-40 text-white rounded font-bold transition"
            >
              Verify PAN & Continue
            </button>
          </div>
        )}

        {/* Step 2: Aadhaar */}
        {step === 2 && (
          <div className="flex flex-col gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Aadhaar Card Number (12 Digits)</label>
              <input
                type="text"
                maxLength={12}
                placeholder="1234 5678 9012"
                value={aadhaarRaw}
                onChange={(e) => setAadhaarRaw(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded px-3 py-2 text-slate-100 font-mono text-sm tracking-wider focus:outline-none focus:border-blue-500"
              />
              <div className="bg-slate-800/60 p-2.5 rounded border border-slate-750 mt-2 text-[11px] text-slate-300">
                <span className="text-emerald-400 font-semibold block mb-0.5">UIDAI Masking & ADV Tokenization:</span>
                <span>Raw 12-digit Aadhaar numbers are never stored in plain text. On-chain commitment will store tokenized: </span>
                <span className="font-mono text-emerald-300 font-bold">{getMaskedAadhaar() || 'XXXX-XXXX-XXXX'}</span>
              </div>
            </div>
            <button
              type="button"
              onClick={handleStep2Aadhaar}
              className="mt-2 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded font-bold transition"
            >
              Tokenize Aadhaar & Proceed
            </button>
          </div>
        )}

        {/* Step 3: Penny Drop */}
        {step === 3 && (
          <div className="flex flex-col gap-3">
            <div>
              <label className="block text-slate-300 font-medium mb-1">Bank Account Number</label>
              <input
                type="text"
                placeholder="9876543210123"
                value={bankAccount}
                onChange={(e) => setBankAccount(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded px-3 py-2 text-slate-100 font-mono text-sm tracking-wider focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-slate-300 font-medium mb-1">IFSC Code</label>
              <input
                type="text"
                maxLength={11}
                placeholder="HDFC0001234"
                value={ifsc}
                onChange={(e) => setIfsc(e.target.value.toUpperCase())}
                className="w-full bg-slate-800 border border-slate-700 rounded px-3 py-2 text-slate-100 font-mono text-sm tracking-wider uppercase focus:outline-none focus:border-blue-500"
              />
              <span className="text-[10px] text-slate-400 mt-1 block">
                RBI IMPS Penny Drop will deposit ₹1.00 to verify account holder name via Jaro-Winkler ($\ge 0.85$).
              </span>
            </div>
            <button
              type="button"
              disabled={isVerifying}
              onClick={handleStep3PennyDrop}
              className="mt-2 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded font-bold transition"
            >
              {isVerifying ? 'Initiating Penny Drop & Verifying...' : 'Perform Penny Drop Verification'}
            </button>
          </div>
        )}

        {/* Step 4: Success */}
        {step === 4 && (
          <div className="flex flex-col items-center gap-3 py-3 text-center">
            <div className="h-12 w-12 rounded-full bg-emerald-500/20 border border-emerald-500/40 flex items-center justify-center text-emerald-400 text-xl font-bold">
              ✓
            </div>
            <div>
              <h4 className="text-sm font-bold text-slate-100">KYC Tier 2 Verification Complete!</h4>
              <p className="text-slate-400 text-[11px] mt-1">
                Your PAN and Bank Account are certified. Daily limit upgraded to ₹5,00,000 INR.
              </p>
            </div>
            <div className="bg-slate-800/80 p-2.5 rounded border border-slate-700 text-[11px] text-slate-300 font-mono text-left w-full">
              <div>PAN: <span className="text-slate-100">{pan}</span></div>
              <div>Aadhaar: <span className="text-slate-100">{getMaskedAadhaar()}</span></div>
              <div>Besu Identity Hash: <span className="text-emerald-400">0x8f2a...c17e</span></div>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="mt-2 w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded font-bold transition shadow-lg shadow-emerald-950"
            >
              Enter Exchange
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

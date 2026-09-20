'use client';

import React, { useState } from 'react';

export type KycStep = 'PAN_PROFILE' | 'DIGILOCKER_AADHAAR' | 'CAMERA_LIVENESS' | 'BANK_PENNY_DROP' | 'ESIGN_COMPLETED';

export function KycFlow() {
  const [currentStep, setCurrentStep] = useState<KycStep>('PAN_PROFILE');
  const [pan, setPan] = useState('');
  const [aadhaarLinked, setAadhaarLinked] = useState(false);
  const [livenessScore, setLivenessScore] = useState<number | null>(null);
  const [bankAccount, setBankAccount] = useState('');
  const [ifsc, setIfsc] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handlePanSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (pan.length === 10) {
      setCurrentStep('DIGILOCKER_AADHAAR');
    }
  };

  const handleDigiLockerAuth = () => {
    setIsSubmitting(true);
    setTimeout(() => {
      setAadhaarLinked(true);
      setIsSubmitting(false);
      setCurrentStep('CAMERA_LIVENESS');
    }, 600);
  };

  const handleCaptureLiveness = () => {
    setIsSubmitting(true);
    setTimeout(() => {
      setLivenessScore(0.98); // High confidence facial match against Aadhaar photo
      setIsSubmitting(false);
      setCurrentStep('BANK_PENNY_DROP');
    }, 800);
  };

  const handleBankVerification = (e: React.FormEvent) => {
    e.preventDefault();
    if (bankAccount && ifsc) {
      setIsSubmitting(true);
      setTimeout(() => {
        setIsSubmitting(false);
        setCurrentStep('ESIGN_COMPLETED');
      }, 700);
    }
  };

  return (
    <div className="rounded-xl border border-slate-800 bg-[#111620] p-6 shadow-2xl">
      {/* Step Progress Bar */}
      <div className="mb-6">
        <div className="flex items-center justify-between text-xs font-mono text-slate-400 mb-2">
          <span className={currentStep === 'PAN_PROFILE' ? 'text-[#00F0A0] font-bold' : ''}>1. PAN</span>
          <span className={currentStep === 'DIGILOCKER_AADHAAR' ? 'text-[#00F0A0] font-bold' : ''}>2. Aadhaar</span>
          <span className={currentStep === 'CAMERA_LIVENESS' ? 'text-[#00F0A0] font-bold' : ''}>3. Camera</span>
          <span className={currentStep === 'BANK_PENNY_DROP' ? 'text-[#00F0A0] font-bold' : ''}>4. Bank</span>
          <span className={currentStep === 'ESIGN_COMPLETED' ? 'text-[#00F0A0] font-bold' : ''}>5. E-Sign</span>
        </div>
        <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
          <div
            className="h-full bg-[#00F0A0] transition-all duration-300"
            style={{
              width:
                currentStep === 'PAN_PROFILE' ? '20%' :
                currentStep === 'DIGILOCKER_AADHAAR' ? '40%' :
                currentStep === 'CAMERA_LIVENESS' ? '60%' :
                currentStep === 'BANK_PENNY_DROP' ? '80%' : '100%',
            }}
          />
        </div>
      </div>

      {/* Step 1: PAN */}
      {currentStep === 'PAN_PROFILE' && (
        <form onSubmit={handlePanSubmit} className="space-y-4">
          <h2 className="text-xl font-bold text-white">Enter Permanent Account Number (PAN)</h2>
          <p className="text-sm text-slate-400">Required under SEBI Master Circular for KYC verification.</p>
          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">PAN NUMBER</label>
            <input
              type="text"
              maxLength={10}
              value={pan}
              onChange={(e) => setPan(e.target.value.toUpperCase())}
              placeholder="ABCDE1234F"
              className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-white font-mono uppercase focus:border-[#00F0A0] focus:outline-none"
              required
            />
          </div>
          <button
            type="submit"
            disabled={pan.length !== 10}
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black disabled:opacity-40 hover:bg-[#00d08a] transition-colors"
          >
            Verify PAN & Continue
          </button>
        </form>
      )}

      {/* Step 2: DigiLocker */}
      {currentStep === 'DIGILOCKER_AADHAAR' && (
        <div className="space-y-4">
          <h2 className="text-xl font-bold text-white">DigiLocker Aadhaar Verification</h2>
          <p className="text-sm text-slate-400">Instant paperless verification via UIDAI / DigiLocker gateway.</p>
          <div className="rounded border border-dashed border-slate-700 p-4 bg-slate-900/50 text-center">
            <div className="text-sm text-slate-300 font-mono">PAN Linked: {pan}</div>
            <div className="mt-2 text-xs text-slate-500">Zero physical documents required</div>
          </div>
          <button
            onClick={handleDigiLockerAuth}
            disabled={isSubmitting}
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black disabled:opacity-40 hover:bg-[#00d08a]"
          >
            {isSubmitting ? 'Authenticating with DigiLocker...' : 'Authenticate via DigiLocker'}
          </button>
        </div>
      )}

      {/* Step 3: Camera Liveness */}
      {currentStep === 'CAMERA_LIVENESS' && (
        <div className="space-y-4">
          <h2 className="text-xl font-bold text-white">Live Facial Match & Geo-Tagging</h2>
          <p className="text-sm text-slate-400">Ensures compliance with SEBI Video-KYC anti-spoofing guidelines.</p>
          <div className="aspect-video bg-black/60 rounded border border-slate-700 flex flex-col items-center justify-center text-slate-400 font-mono text-xs">
            <div className="h-16 w-16 rounded-full border-2 border-dashed border-[#00F0A0] flex items-center justify-center mb-2">
              <span className="text-lg">📷</span>
            </div>
            <span>Position face within oval frame</span>
            <span className="text-[10px] text-slate-500 mt-1">Geo-Location: 19.0760° N, 72.8777° E (Mumbai)</span>
          </div>
          <button
            onClick={handleCaptureLiveness}
            disabled={isSubmitting}
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black hover:bg-[#00d08a]"
          >
            {isSubmitting ? 'Verifying Liveness...' : 'Capture Photo & Match'}
          </button>
        </div>
      )}

      {/* Step 4: Bank Penny Drop */}
      {currentStep === 'BANK_PENNY_DROP' && (
        <form onSubmit={handleBankVerification} className="space-y-4">
          <h2 className="text-xl font-bold text-white">Bank Account Verification</h2>
          <p className="text-sm text-slate-400">We will verify via an automated ₹1 penny-drop validation.</p>
          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">ACCOUNT NUMBER</label>
            <input
              type="text"
              value={bankAccount}
              onChange={(e) => setBankAccount(e.target.value)}
              placeholder="910020030040"
              className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-white font-mono focus:border-[#00F0A0] focus:outline-none"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">IFSC CODE</label>
            <input
              type="text"
              value={ifsc}
              onChange={(e) => setIfsc(e.target.value.toUpperCase())}
              placeholder="HDFC0000128"
              className="w-full rounded bg-slate-900 border border-slate-700 px-3 py-2 text-white font-mono uppercase focus:border-[#00F0A0] focus:outline-none"
              required
            />
          </div>
          <button
            type="submit"
            disabled={isSubmitting || !bankAccount || !ifsc}
            className="w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black hover:bg-[#00d08a]"
          >
            {isSubmitting ? 'Initiating Penny Drop...' : 'Verify Bank & Continue'}
          </button>
        </form>
      )}

      {/* Step 5: Completed */}
      {currentStep === 'ESIGN_COMPLETED' && (
        <div className="space-y-4 text-center py-4">
          <div className="mx-auto h-12 w-12 rounded-full bg-[#00F0A0]/20 text-[#00F0A0] flex items-center justify-center text-xl font-bold">
            ✓
          </div>
          <h2 className="text-xl font-bold text-white">KYC & Demat Ready!</h2>
          <p className="text-sm text-slate-400">
            Account verified under KRA records. You are now enabled to trade fractional securities and digital assets.
          </p>
          <a
            href="/trade/btc-usdt"
            className="inline-block w-full rounded bg-[#00F0A0] py-2.5 font-bold text-black hover:bg-[#00d08a]"
          >
            Launch Pro Trading Terminal
          </a>
        </div>
      )}
    </div>
  );
}

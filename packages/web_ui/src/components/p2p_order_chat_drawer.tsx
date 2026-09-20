import React, { useState } from 'react';

export function P2POrderChatDrawer() {
  const [messages, setMessages] = useState([
    { id: 1, sender: 'Seller', text: 'Please send INR via IMPS to my registered HDFC bank account.' },
    { id: 2, sender: 'Buyer', text: 'Transferred ₹50,000. UTR: 491820194820.' },
  ]);
  const [inputText, setInputText] = useState('');

  const sendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputText) return;
    setMessages((prev) => [...prev, { id: Date.now(), sender: 'Buyer', text: inputText }]);
    setInputText('');
  };

  return (
    <div className="p-4 rounded border border-slate-800 bg-[#111620] text-white font-mono text-xs space-y-3">
      <div className="flex justify-between items-center pb-2 border-b border-slate-800">
        <span className="font-bold">Encrypted P2P Order Chat</span>
        <span className="text-[10px] text-[#00F0A0]">E2E Sealed</span>
      </div>
      <div className="h-40 overflow-y-auto space-y-2 p-2 bg-slate-900/60 rounded">
        {messages.map((m) => (
          <div key={m.id} className="text-slate-300">
            <span className="font-bold text-white">{m.sender}: </span>
            <span>{m.text}</span>
          </div>
        ))}
      </div>
      <form onSubmit={sendMessage} className="flex space-x-2">
        <input
          type="text"
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          placeholder="Type message or upload bank payment proof..."
          className="flex-1 bg-slate-900 border border-slate-700 rounded px-2 py-1 text-white"
        />
        <button type="submit" className="px-3 py-1 bg-[#00F0A0] text-black font-bold rounded">
          Send
        </button>
      </form>
    </div>
  );
}

const assert = require('assert');

// 1. Test PAN Validation Logic
function isValidPAN(pan) {
  const panRegex = /^[A-Z]{5}[0-9]{4}[A-Z]$/;
  return panRegex.test(pan.trim().toUpperCase());
}

assert.strictEqual(isValidPAN('ABCDE1234F'), true, 'Valid PAN must pass');
assert.strictEqual(isValidPAN('abcde1234f'), true, 'Lowercase PAN should be accepted when trimmed/uppercased');
assert.strictEqual(isValidPAN('12345ABCDE'), false, 'Invalid format must fail');
assert.strictEqual(isValidPAN('ABCDE12345'), false, '10-digit without trailing letter must fail');
console.log('✓ PAN validation tests passed');

// 2. Test Aadhaar Masking Logic
function maskAadhaar(raw) {
  const digits = raw.replace(/\D/g, '');
  if (digits.length >= 4) {
    return `XXXX-XXXX-${digits.slice(-4)}`;
  }
  return digits;
}

assert.strictEqual(maskAadhaar('1234 5678 9012'), 'XXXX-XXXX-9012', 'Aadhaar masking must preserve only last 4 digits');
assert.strictEqual(maskAadhaar('987654321098'), 'XXXX-XXXX-1098', 'Unspaced Aadhaar must mask properly');
console.log('✓ Aadhaar masking tests passed');

// 3. Test OrderBook Depth Calculation & Zero-Ghost Eviction
function processDelta(currentBook, delta) {
  const bids = new Map(currentBook.bids);
  for (const [p, q] of delta.bids) {
    if (q <= 0) {
      bids.delete(p); // Zero-ghost liquidity eviction
    } else {
      bids.set(p, q);
    }
  }
  return { bids: Array.from(bids.entries()).sort((a, b) => b[0] - a[0]) };
}

const initialBook = {
  bids: [[65000, 1.5], [64900, 2.0], [64800, 3.0]]
};
const deltaUpdate = {
  bids: [[64900, 0.0], [65100, 0.5]] // cancel 64900, add 65100
};
const updatedBook = processDelta(initialBook, deltaUpdate);
const updatedPrices = updatedBook.bids.map(b => b[0]);
assert.deepStrictEqual(updatedPrices, [65100, 65000, 64800], 'Cancelled 64900 must be evicted and 65100 prepended');
console.log('✓ Orderbook zero-ghost eviction delta test passed');

// 4. Test Statutory Section 194S TDS & Zero Platform Fee
function computeTDS(side, notionalINR) {
  if (side === 'SELL') {
    return notionalINR * 0.01; // 1%
  }
  return 0.0;
}

const sellNotional = 100000; // 1 Lakh INR
assert.strictEqual(computeTDS('SELL', sellNotional), 1000.0, '1% TDS on ₹1,00,000 sell must be ₹1,000');
assert.strictEqual(computeTDS('BUY', sellNotional), 0.0, '0% TDS on Buy side');
console.log('✓ Section 194S TDS calculation tests passed');

// 5. Test Section 115BBH Flat 31.2% Tax & Zero Loss Set-Off
function compute115BBHTax(gainINR) {
  if (gainINR <= 0) return 0.0;
  // 30% base + 4% cess = 31.2%
  return gainINR * 0.312;
}

assert.strictEqual(compute115BBHTax(50000), 15600.0, '31.2% statutory tax on ₹50k gain is ₹15,600');
assert.strictEqual(compute115BBHTax(-20000), 0.0, 'Loss generates ₹0 tax liability and cannot offset gains');
console.log('✓ Section 115BBH tax calculation tests passed');

console.log('\nALL 5 WEB COMPONENT & TAX UNIT TESTS PASSED!');

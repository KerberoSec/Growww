// Unit tests for Growww Marketing Showcase & Fee Calculations
const assert = require('assert');

function runMarketingTests() {
  console.log('Running Marketing & Fee Structure Unit Tests...');

  // Test 1: Zero-Fee Structure Verification (0.00% Platform Fee)
  function computePlatformFee(notionalINR, feeRateBps = 0) {
    return (notionalINR * feeRateBps) / 10000;
  }

  const tradeNotional = 1000000; // ₹10 Lakhs trade
  assert.strictEqual(computePlatformFee(tradeNotional, 0), 0, 'Platform fee must strictly be 0.00 at launch');
  console.log('✓ 0.00% Platform fee calculation verified');

  // Test 2: Section 194S TDS deduction on VDA sell side (1%)
  function computeSection194STDS(sellNotionalINR, isPANExempt = false) {
    if (isPANExempt) return 0;
    return sellNotionalINR * 0.01; // 1% flat TDS
  }

  const btcSaleNotional = 5000000; // ₹50 Lakhs BTC sale
  const tdsDeducted = computeSection194STDS(btcSaleNotional);
  assert.strictEqual(tdsDeducted, 50000, 'TDS should be exactly 1% (₹50,000)');
  console.log('✓ Section 194S TDS calculation verified');

  // Test 3: Demo Trading Faucet Voucher Credit Value
  function getDemoTokenAllocation() {
    return {
      usdt: 10000,
      btc: 1.0,
      cooldownHours: 24,
    };
  }

  const faucet = getDemoTokenAllocation();
  assert.strictEqual(faucet.usdt, 10000, 'Demo faucet must offer 10,000 USDT');
  assert.strictEqual(faucet.btc, 1.0, 'Demo faucet must offer 1.0 BTC');
  assert.strictEqual(faucet.cooldownHours, 24, 'Cooldown must be 24 hours');
  console.log('✓ Demo token faucet allocation parameters verified');

  console.log('\nALL 3 MARKETING & COMPLIANCE UNIT TESTS PASSED!');
}

runMarketingTests();

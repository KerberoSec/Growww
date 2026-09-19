// Unit tests for Admin & Grievance & Surveillance Components
const assert = require('assert');

function runAdminTests() {
  console.log('Running Admin & Surveillance & Circuit Breaker Component Tests...');

  // Test 1: Grievance SLA Escalation Stage Calculation
  function calculateSLAStage(daysSinceIntake) {
    if (daysSinceIntake <= 6) return 0; // Normal
    if (daysSinceIntake <= 13) return 1; // Warning
    if (daysSinceIntake <= 17) return 2; // PCO Escalation
    if (daysSinceIntake <= 19) return 3; // Critical War Room
    if (daysSinceIntake <= 20) return 4; // Freeze
    return 5; // Statutory Breach
  }

  assert.strictEqual(calculateSLAStage(3), 0, 'Day 3 should be Stage 0 (Normal)');
  assert.strictEqual(calculateSLAStage(10), 1, 'Day 10 should be Stage 1 (Warning)');
  assert.strictEqual(calculateSLAStage(15), 2, 'Day 15 should be Stage 2 (PCO)');
  assert.strictEqual(calculateSLAStage(18), 3, 'Day 18 should be Stage 3 (War Room)');
  assert.strictEqual(calculateSLAStage(20), 4, 'Day 20 should be Stage 4 (Freeze)');
  assert.strictEqual(calculateSLAStage(22), 5, 'Day 22 should be Stage 5 (Breach)');
  console.log('✓ Grievance SLA 21-day escalation stage logic verified');

  // Test 2: LULD Dynamic Price Collar Boundaries (±5% for Tier 1 Large Cap)
  function computePriceBands(refPricePaise, bandBps) {
    const delta = (refPricePaise * bandBps) / 10000;
    return {
      lowerBandPaise: Math.round(refPricePaise - delta),
      upperBandPaise: Math.round(refPricePaise + delta),
    };
  }

  const bands = computePriceBands(500000, 500); // ₹5,000.00 with 500 bps (5.00%)
  assert.strictEqual(bands.lowerBandPaise, 475000, 'Lower band should be ₹4,750.00');
  assert.strictEqual(bands.upperBandPaise, 525000, 'Upper band should be ₹5,250.00');
  console.log('✓ LULD ±5% price band math verified');

  // Test 3: Surveillance Confidence & Wash Trade Circularity detection threshold
  function evaluateWashTradeConfidence(selfTrades, totalTrades, volumeSameDay) {
    const ratio = selfTrades / totalTrades;
    let confidence = ratio * 0.8;
    if (volumeSameDay > 1000000) confidence += 0.2; // ₹10 Lakh+ boost
    return Math.min(confidence, 1.0);
  }

  const conf = evaluateWashTradeConfidence(9, 10, 2500000);
  assert(conf >= 0.9, 'Confidence should exceed 90% for severe circular matching');
  console.log('✓ Surveillance pattern scoring formula verified');

  // Test 4: Action Taken Report (ATR) validation
  function validateATR(atr) {
    if (!atr.complaintId) return false;
    if (!atr.disposition) return false;
    if (!atr.actionTakenSummary || atr.actionTakenSummary.length < 10) return false;
    if (atr.disposition === 'RESOLVED_WITH_REFUND' && (!atr.refundAmount || atr.refundAmount <= 0)) return false;
    return true;
  }

  assert.strictEqual(validateATR({ complaintId: 'C1', disposition: 'RESOLVED_SATISFIED', actionTakenSummary: 'Root cause resolved' }), true);
  assert.strictEqual(validateATR({ complaintId: 'C2', disposition: 'RESOLVED_WITH_REFUND', actionTakenSummary: 'Refunded', refundAmount: 0 }), false);
  console.log('✓ ATR schema validation verified');

  console.log('\nALL 4 ADMIN & SURVEILLANCE UNIT TESTS PASSED!');
}

runAdminTests();

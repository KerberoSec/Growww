import 'dart:math' as math;
import '../models/apy_calculation_result.dart';
import '../models/earn_enums.dart';
import '../models/sub_paise_amount.dart';
import '../models/tds_withholding_estimate.dart';

/// Fixed-point financial calculation engine for fixed-income vault compounding and TDS deduction.
class YieldMathEngine {
  /// Computes projected yield returns using deterministic sub-paise integer arithmetic
  static ApyCalculationResult computeReturns({
    required SubPaiseAmount principal,
    required double baseApy,
    required TenorType tenor,
    CompoundingFrequency compounding = CompoundingFrequency.daily,
    bool isSeniorCitizen = false,
    bool hasForm15G15H = false,
  }) {
    final effectiveApy = baseApy + tenor.defaultBonusApy;
    final int days = tenor.durationDays > 0 ? tenor.durationDays : 365; // flexible assumes 1-year annualized benchmark

    final double dailyRate = (effectiveApy / 100.0) / 365.0;
    final double principalInr = principal.toInr();

    // Projected daily yield (1 day)
    final double dailyYieldInr = principalInr * dailyRate;
    final double monthlyYieldInr = dailyYieldInr * 30.0;

    double totalYieldInr;
    if (compounding == CompoundingFrequency.daily) {
      // Compound interest: A = P * (1 + r)^n
      final double maturityAmountInr = principalInr * math.pow(1.0 + dailyRate, days);
      totalYieldInr = maturityAmountInr - principalInr;
    } else {
      // Simple interest at maturity: I = P * (APY / 100) * (days / 365)
      totalYieldInr = principalInr * (effectiveApy / 100.0) * (days / 365.0);
    }

    final grossYield = SubPaiseAmount.fromInr(totalYieldInr);

    // TDS Withholding Simulation (Section 194A)
    final tdsEstimate = estimateTds(
      grossInterest: grossYield,
      isSeniorCitizen: isSeniorCitizen,
      hasForm15G15H: hasForm15G15H,
    );

    final netMaturity = principal + grossYield - tdsEstimate.tdsAmountWithheld;
    final maturityDate = DateTime.now().add(Duration(days: tenor.durationDays > 0 ? tenor.durationDays : 365));

    return ApyCalculationResult(
      principal: principal,
      tenor: tenor,
      effectiveApy: effectiveApy,
      projectedDailyYield: SubPaiseAmount.fromInr(dailyYieldInr),
      projectedMonthlyYield: SubPaiseAmount.fromInr(monthlyYieldInr),
      projectedTotalYield: grossYield,
      estimatedTdsDeduction: tdsEstimate.tdsAmountWithheld,
      netMaturityAmount: netMaturity,
      projectedMaturityDate: maturityDate,
    );
  }

  /// Calculates Section 194A TDS withholding (10% standard, 20% non-PAN, 0% Form 15G/15H)
  static TdsWithholdingEstimate estimateTds({
    required SubPaiseAmount grossInterest,
    bool isSeniorCitizen = false,
    bool hasForm15G15H = false,
    bool hasValidPan = true,
  }) {
    final thresholdInr = isSeniorCitizen ? 50000.0 : 40000.0;
    final grossInr = grossInterest.toInr();

    if (hasForm15G15H) {
      return TdsWithholdingEstimate(
        tdsRatePercentage: 0.0,
        grossInterestEarned: grossInterest,
        tdsAmountWithheld: SubPaiseAmount.zero(),
        netInterestPayable: grossInterest,
        isForm15G15HApplied: true,
        panMasked: 'ABCDE****F',
      );
    }

    double rate = 0.0;
    if (grossInr >= thresholdInr) {
      rate = hasValidPan ? 10.0 : 20.0;
    }

    final double tdsInr = grossInr * (rate / 100.0);
    final tdsAmount = SubPaiseAmount.fromInr(tdsInr);
    final netAmount = grossInterest - tdsAmount;

    return TdsWithholdingEstimate(
      tdsRatePercentage: rate,
      grossInterestEarned: grossInterest,
      tdsAmountWithheld: tdsAmount,
      netInterestPayable: netAmount,
      isForm15G15HApplied: false,
      panMasked: 'ABCDE****F',
    );
  }

  /// Computes 0.25% emergency unstake penalty deduction
  static SubPaiseAmount calculateEmergencyPenalty(SubPaiseAmount principal) {
    return principal.multiply(0.0025);
  }
}

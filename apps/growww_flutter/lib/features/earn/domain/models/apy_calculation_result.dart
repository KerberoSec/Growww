import 'earn_enums.dart';
import 'sub_paise_amount.dart';

class ApyCalculationResult {
  final SubPaiseAmount principal;
  final TenorType tenor;
  final double effectiveApy;
  final SubPaiseAmount projectedDailyYield;
  final SubPaiseAmount projectedMonthlyYield;
  final SubPaiseAmount projectedTotalYield;
  final SubPaiseAmount estimatedTdsDeduction;
  final SubPaiseAmount netMaturityAmount;
  final DateTime projectedMaturityDate;

  const ApyCalculationResult({
    required this.principal,
    required this.tenor,
    required this.effectiveApy,
    required this.projectedDailyYield,
    required this.projectedMonthlyYield,
    required this.projectedTotalYield,
    required this.estimatedTdsDeduction,
    required this.netMaturityAmount,
    required this.projectedMaturityDate,
  });

  Map<String, dynamic> toJson() => {
        'principal': principal.toJson(),
        'tenor': tenor.name,
        'effective_apy': effectiveApy,
        'projected_daily_yield': projectedDailyYield.toJson(),
        'projected_monthly_yield': projectedMonthlyYield.toJson(),
        'projected_total_yield': projectedTotalYield.toJson(),
        'estimated_tds_deduction': estimatedTdsDeduction.toJson(),
        'net_maturity_amount': netMaturityAmount.toJson(),
        'projected_maturity_date': projectedMaturityDate.toIso8601String(),
      };

  factory ApyCalculationResult.fromJson(Map<String, dynamic> json) {
    return ApyCalculationResult(
      principal: SubPaiseAmount.fromJson(json['principal'] as Map<String, dynamic>),
      tenor: TenorType.values.byName(json['tenor'] as String),
      effectiveApy: (json['effective_apy'] as num).toDouble(),
      projectedDailyYield: SubPaiseAmount.fromJson(json['projected_daily_yield'] as Map<String, dynamic>),
      projectedMonthlyYield: SubPaiseAmount.fromJson(json['projected_monthly_yield'] as Map<String, dynamic>),
      projectedTotalYield: SubPaiseAmount.fromJson(json['projected_total_yield'] as Map<String, dynamic>),
      estimatedTdsDeduction: SubPaiseAmount.fromJson(json['estimated_tds_deduction'] as Map<String, dynamic>),
      netMaturityAmount: SubPaiseAmount.fromJson(json['net_maturity_amount'] as Map<String, dynamic>),
      projectedMaturityDate: DateTime.parse(json['projected_maturity_date'] as String),
    );
  }
}

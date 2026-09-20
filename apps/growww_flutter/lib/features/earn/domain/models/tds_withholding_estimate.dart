import 'sub_paise_amount.dart';

class TdsWithholdingEstimate {
  final double tdsRatePercentage; // 10% standard, 20% non-PAN, 0% Form 15G/15H
  final SubPaiseAmount grossInterestEarned;
  final SubPaiseAmount tdsAmountWithheld;
  final SubPaiseAmount netInterestPayable;
  final bool isForm15G15HApplied;
  final String panMasked;

  const TdsWithholdingEstimate({
    required this.tdsRatePercentage,
    required this.grossInterestEarned,
    required this.tdsAmountWithheld,
    required this.netInterestPayable,
    required this.isForm15G15HApplied,
    required this.panMasked,
  });

  Map<String, dynamic> toJson() => {
        'tds_rate_percentage': tdsRatePercentage,
        'gross_interest_earned': grossInterestEarned.toJson(),
        'tds_amount_withheld': tdsAmountWithheld.toJson(),
        'net_interest_payable': netInterestPayable.toJson(),
        'is_form_15g_15h_applied': isForm15G15HApplied,
        'pan_masked': panMasked,
      };

  factory TdsWithholdingEstimate.fromJson(Map<String, dynamic> json) {
    return TdsWithholdingEstimate(
      tdsRatePercentage: (json['tds_rate_percentage'] as num).toDouble(),
      grossInterestEarned: SubPaiseAmount.fromJson(json['gross_interest_earned'] as Map<String, dynamic>),
      tdsAmountWithheld: SubPaiseAmount.fromJson(json['tds_amount_withheld'] as Map<String, dynamic>),
      netInterestPayable: SubPaiseAmount.fromJson(json['net_interest_payable'] as Map<String, dynamic>),
      isForm15G15HApplied: json['is_form_15g_15h_applied'] as bool,
      panMasked: json['pan_masked'] as String,
    );
  }
}

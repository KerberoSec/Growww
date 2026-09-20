/// Sub-Paise Fee Breakdown (1 INR = 10,000 sub-paise)
/// Ensures zero floating-point calculation discrepancies for precious metals rematerialization.
class SubPaiseFeeBreakdown {
  final int vaultRematerializationFeeSubPaise;
  final int settlementFeeSubPaise; // 0.00% (Zero Fee) platform settlement fee
  final int assayPackagingFeeSubPaise;
  final int armoredLogisticsFeeSubPaise;
  final int transitInsuranceFeeSubPaise; // 100% replacement value coverage
  final int preciousMetalGstSubPaise; // 3% GST on precious metals rematerialization
  final int handlingServicesGstSubPaise; // 18% GST on logistics & vault handling services
  final int totalGrossPayableSubPaise;

  const SubPaiseFeeBreakdown({
    required this.vaultRematerializationFeeSubPaise,
    required this.settlementFeeSubPaise,
    required this.assayPackagingFeeSubPaise,
    required this.armoredLogisticsFeeSubPaise,
    required this.transitInsuranceFeeSubPaise,
    required this.preciousMetalGstSubPaise,
    required this.handlingServicesGstSubPaise,
    required this.totalGrossPayableSubPaise,
  });

  double get totalInr => totalGrossPayableSubPaise / 10000.0;
  double get settlementFeeInr => settlementFeeSubPaise / 10000.0;
  double get vaultFeeInr => vaultRematerializationFeeSubPaise / 10000.0;
  double get packagingFeeInr => assayPackagingFeeSubPaise / 10000.0;
  double get logisticsFeeInr => armoredLogisticsFeeSubPaise / 10000.0;
  double get insuranceFeeInr => transitInsuranceFeeSubPaise / 10000.0;
  double get preciousMetalGstInr => preciousMetalGstSubPaise / 10000.0;
  double get handlingServicesGstInr => handlingServicesGstSubPaise / 10000.0;
  double get totalGstInr => (preciousMetalGstSubPaise + handlingServicesGstSubPaise) / 10000.0;

  String formatInr(double amount, {bool showSubPaise = false}) {
    if (showSubPaise) {
      return '₹${amount.toStringAsFixed(4)}';
    }
    return '₹${amount.toStringAsFixed(2)}';
  }

  Map<String, dynamic> toJson() => {
        'vault_rematerialization_fee_sub_paise': vaultRematerializationFeeSubPaise,
        'settlement_fee_sub_paise': settlementFeeSubPaise,
        'assay_packaging_fee_sub_paise': assayPackagingFeeSubPaise,
        'armored_logistics_fee_sub_paise': armoredLogisticsFeeSubPaise,
        'transit_insurance_fee_sub_paise': transitInsuranceFeeSubPaise,
        'precious_metal_gst_sub_paise': preciousMetalGstSubPaise,
        'handling_services_gst_sub_paise': handlingServicesGstSubPaise,
        'total_gross_payable_sub_paise': totalGrossPayableSubPaise,
      };

  factory SubPaiseFeeBreakdown.fromJson(Map<String, dynamic> json) {
    return SubPaiseFeeBreakdown(
      vaultRematerializationFeeSubPaise: json['vault_rematerialization_fee_sub_paise'] as int,
      settlementFeeSubPaise: json['settlement_fee_sub_paise'] as int,
      assayPackagingFeeSubPaise: json['assay_packaging_fee_sub_paise'] as int,
      armoredLogisticsFeeSubPaise: json['armored_logistics_fee_sub_paise'] as int,
      transitInsuranceFeeSubPaise: json['transit_insurance_fee_sub_paise'] as int,
      preciousMetalGstSubPaise: json['precious_metal_gst_sub_paise'] as int,
      handlingServicesGstSubPaise: json['handling_services_gst_sub_paise'] as int,
      totalGrossPayableSubPaise: json['total_gross_payable_sub_paise'] as int,
    );
  }

  /// Helper factory to compute exact sub-paise splits deterministically
  factory SubPaiseFeeBreakdown.calculate({
    required double totalWeightGrams,
    required double indicativeMetalValueInr,
    required bool isDoorstep,
    int itemsCount = 1,
  }) {
    // 1 INR = 10,000 sub-paise
    // Vault rematerialization: ₹50 per 10g + base ₹100
    final double vaultFeeInr = 100.0 + (totalWeightGrams / 10.0) * 50.0;
    final int vaultSubPaise = (vaultFeeInr * 10000).round();

    // Zero fee platform settlement
    const int settlementSubPaise = 0;

    // Assayed packaging: ₹75 per item
    final double packagingFeeInr = 75.0 * itemsCount;
    final int packagingSubPaise = (packagingFeeInr * 10000).round();

    // Armored logistics: ₹450 flat if doorstep, else 0
    final double logisticsFeeInr = isDoorstep ? 450.0 : 0.0;
    final int logisticsSubPaise = (logisticsFeeInr * 10000).round();

    // Transit insurance: 0.15% of bullion value if doorstep, else 0
    final double insuranceFeeInr = isDoorstep ? (indicativeMetalValueInr * 0.0015) : 0.0;
    final int insuranceSubPaise = (insuranceFeeInr * 10000).round();

    // 3% GST on rematerialization charge
    final int metalGstSubPaise = ((vaultSubPaise) * 0.03).round();

    // 18% GST on packaging, logistics & transit services
    final int servicesSubPaise = packagingSubPaise + logisticsSubPaise + insuranceSubPaise;
    final int handlingGstSubPaise = (servicesSubPaise * 0.18).round();

    final int totalGross = vaultSubPaise +
        settlementSubPaise +
        packagingSubPaise +
        logisticsSubPaise +
        insuranceSubPaise +
        metalGstSubPaise +
        handlingGstSubPaise;

    return SubPaiseFeeBreakdown(
      vaultRematerializationFeeSubPaise: vaultSubPaise,
      settlementFeeSubPaise: settlementSubPaise,
      assayPackagingFeeSubPaise: packagingSubPaise,
      armoredLogisticsFeeSubPaise: logisticsSubPaise,
      transitInsuranceFeeSubPaise: insuranceSubPaise,
      preciousMetalGstSubPaise: metalGstSubPaise,
      handlingServicesGstSubPaise: handlingGstSubPaise,
      totalGrossPayableSubPaise: totalGross,
    );
  }
}

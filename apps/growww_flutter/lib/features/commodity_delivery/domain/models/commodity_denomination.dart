import 'commodity_enums.dart';

class CommodityDenomination {
  final String denominationId;
  final CommodityType commodityType;
  final double weightGrams;
  final double fineness;
  final String displayName;
  final String barFormFactor; // "Minted Bar", "Cast Bar", "Coin"
  final int availableStockUnits;
  final String assayStandard; // "BIS 1417:2016", "LBMA Good Delivery"
  final String packagingType; // "CertiCard Blister", "Tamper-Evident Capsule"

  const CommodityDenomination({
    required this.denominationId,
    required this.commodityType,
    required this.weightGrams,
    required this.fineness,
    required this.displayName,
    required this.barFormFactor,
    required this.availableStockUnits,
    required this.assayStandard,
    required this.packagingType,
  });

  Map<String, dynamic> toJson() => {
        'denomination_id': denominationId,
        'commodity_type': commodityType.name,
        'weight_grams': weightGrams,
        'fineness': fineness,
        'display_name': displayName,
        'bar_form_factor': barFormFactor,
        'available_stock_units': availableStockUnits,
        'assay_standard': assayStandard,
        'packaging_type': packagingType,
      };

  factory CommodityDenomination.fromJson(Map<String, dynamic> json) {
    return CommodityDenomination(
      denominationId: json['denomination_id'] as String,
      commodityType: CommodityType.values.byName(json['commodity_type'] as String),
      weightGrams: (json['weight_grams'] as num).toDouble(),
      fineness: (json['fineness'] as num).toDouble(),
      displayName: json['display_name'] as String,
      barFormFactor: json['bar_form_factor'] as String,
      availableStockUnits: json['available_stock_units'] as int,
      assayStandard: json['assay_standard'] as String,
      packagingType: json['packaging_type'] as String,
    );
  }
}

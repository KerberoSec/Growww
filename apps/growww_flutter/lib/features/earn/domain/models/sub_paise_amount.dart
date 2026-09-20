/// Fixed-Point Sub-Paise Currency Value Object (1 INR = 10,000 sub-paise, 0.0001 precision).
/// Prevents IEEE 754 floating point drift during high-frequency compound interest accruals.
class SubPaiseAmount {
  final int subPaiseValue;

  const SubPaiseAmount(this.subPaiseValue);

  factory SubPaiseAmount.fromInr(double inr) {
    return SubPaiseAmount((inr * 10000.0).round());
  }

  factory SubPaiseAmount.zero() => const SubPaiseAmount(0);

  double toInr() => subPaiseValue / 10000.0;

  SubPaiseAmount operator +(SubPaiseAmount other) =>
      SubPaiseAmount(subPaiseValue + other.subPaiseValue);

  SubPaiseAmount operator -(SubPaiseAmount other) =>
      SubPaiseAmount(subPaiseValue - other.subPaiseValue);

  SubPaiseAmount multiply(double factor) =>
      SubPaiseAmount((subPaiseValue * factor).round());

  bool operator <(SubPaiseAmount other) => subPaiseValue < other.subPaiseValue;
  bool operator <=(SubPaiseAmount other) => subPaiseValue <= other.subPaiseValue;
  bool operator >(SubPaiseAmount other) => subPaiseValue > other.subPaiseValue;
  bool operator >=(SubPaiseAmount other) => subPaiseValue >= other.subPaiseValue;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is SubPaiseAmount && runtimeType == other.runtimeType && subPaiseValue == other.subPaiseValue;

  @override
  int get hashCode => subPaiseValue.hashCode;

  /// Formats currency according to the Indian Numbering System (Lakhs & Crores)
  String formatInr({bool showSubPaise = false}) {
    final inrVal = toInr();
    final isNegative = inrVal < 0;
    final absVal = inrVal.abs();

    final parts = (showSubPaise
            ? absVal.toStringAsFixed(4)
            : absVal.toStringAsFixed(2))
        .split('.');

    final intPart = parts[0];
    final decPart = parts[1];

    String formattedInt;
    if (intPart.length <= 3) {
      formattedInt = intPart;
    } else {
      final lastThree = intPart.substring(intPart.length - 3);
      final remaining = intPart.substring(0, intPart.length - 3);
      final buffer = StringBuffer();

      for (int i = 0; i < remaining.length; i++) {
        if (i > 0 && (remaining.length - i) % 2 == 0) {
          buffer.write(',');
        }
        buffer.write(remaining[i]);
      }
      buffer.write(',');
      buffer.write(lastThree);
      formattedInt = buffer.toString();
    }

    final sign = isNegative ? '-' : '';
    return '${sign}₹$formattedInt.$decPart';
  }

  Map<String, dynamic> toJson() => {'sub_paise_value': subPaiseValue};

  factory SubPaiseAmount.fromJson(Map<String, dynamic> json) {
    return SubPaiseAmount(json['sub_paise_value'] as int);
  }

  @override
  String toString() => formatInr(showSubPaise: true);
}

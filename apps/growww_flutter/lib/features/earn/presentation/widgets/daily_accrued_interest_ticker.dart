import 'package:flutter/material.dart';
import '../../domain/models/sub_paise_amount.dart';

class DailyAccruedInterestTicker extends StatelessWidget {
  final SubPaiseAmount amount;
  final TextStyle? style;

  const DailyAccruedInterestTicker({
    Key? key,
    required this.amount,
    this.style,
  }) : super(key: key);

  static const Color neonGreen = Color(0xFF00F0A0);

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 8,
          height: 8,
          margin: const EdgeInsets.only(right: 6),
          decoration: const BoxDecoration(
            shape: BoxShape.circle,
            color: neonGreen,
          ),
        ),
        Text(
          amount.formatInr(showSubPaise: true),
          style: style ??
              const TextStyle(
                color: neonGreen,
                fontFamily: 'JetBrains Mono',
                fontWeight: FontWeight.bold,
                fontSize: 16,
                letterSpacing: 0.5,
              ),
        ),
      ],
    );
  }
}

import 'package:flutter/material.dart';

enum TradingEnvironment {
  demo,
  real,
}

class ModeSwitcherNotifier extends ChangeNotifier {
  TradingEnvironment _currentMode = TradingEnvironment.demo;

  TradingEnvironment get currentMode => _currentMode;
  bool get isDemo => _currentMode == TradingEnvironment.demo;
  bool get isReal => _currentMode == TradingEnvironment.real;

  void toggleMode(BuildContext context) {
    if (_currentMode == TradingEnvironment.demo) {
      _showRealModeWarning(context);
    } else {
      _currentMode = TradingEnvironment.demo;
      notifyListeners();
    }
  }

  void _showRealModeWarning(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext ctx) {
        return AlertDialog(
          backgroundColor: const Color(0xFF1E222D),
          title: const Text(
            'Switch to Real-Money Trading?',
            style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
          ),
          content: const Text(
            'You are switching from the simulated demo sandbox to real-money execution. Real funds will be debited on fill.',
            style: TextStyle(color: Color(0xFFB2B5BE)),
          ),
          actions: [
            TextButton(
              child: const Text('Cancel', style: TextStyle(color: Colors.grey)),
              onPressed: () => Navigator.of(ctx).pop(),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF00C087)),
              child: const Text('Confirm Switch', style: TextStyle(color: Colors.black)),
              onPressed: () {
                _currentMode = TradingEnvironment.real;
                notifyListeners();
                Navigator.of(ctx).pop();
              },
            ),
          ],
        );
      },
    );
  }
}

class ModeSwitcherWidget extends StatelessWidget {
  final ModeSwitcherNotifier notifier;

  const ModeSwitcherWidget({Key? key, required this.notifier}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: notifier,
      builder: (context, _) {
        final isDemo = notifier.isDemo;
        return GestureDetector(
          onTap: () => notifier.toggleMode(context),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: isDemo ? const Color(0xFF262932) : const Color(0xFF1B382B),
              borderRadius: BorderRadius.circular(20),
              border: Border.pad(
                BorderSide(
                  color: isDemo ? const Color(0xFF4A4E5A) : const Color(0xFF00C087),
                  width: 1.5,
                ),
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  isDemo ? Icons.science_outlined : Icons.monetization_on,
                  size: 16,
                  color: isDemo ? const Color(0xFFF0B90B) : const Color(0xFF00C087),
                ),
                const SizedBox(width: 6),
                Text(
                  isDemo ? 'DEMO MODE' : 'REAL TRADING',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: isDemo ? const Color(0xFFF0B90B) : const Color(0xFF00C087),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

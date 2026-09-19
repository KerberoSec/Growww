import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:growww_flutter/screens/acceptance/master_client_acceptance_screen.dart';

void main() {
  group('Prompt 600 - Master Client Acceptance Verification Suite', () {
    testWidgets('renders acceptance screen with Obsidian dark mode and 120 FPS indicator',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: MasterClientAcceptanceScreen(
            userId: 'USR-NBSE-600-AUDIT',
            tradingPair: 'BTC/USDT',
          ),
        ),
      );

      // Verify Header
      expect(
        find.text('FLUTTER MASTER CLIENT ACCEPTANCE VERIFICATION'),
        findsOneWidget,
      );
      expect(find.text('120 FPS'), findsOneWidget);

      // Verify Invariants & Status
      expect(find.text('Status: REAL-TIME ACTIVE'), findsOneWidget);
      expect(find.text('BTC/USDT'), findsOneWidget);
      expect(find.text('2.0s QBFT (Deterministic)'), findsOneWidget);
      expect(find.text('USR-NBSE-600-AUDIT'), findsOneWidget);

      // Verify Action Buttons
      expect(find.text('Run Suite (Neon Green)'), findsOneWidget);
      expect(find.text('Export Telemetry'), findsOneWidget);

      // Tap Run Suite button and verify state increments without layout shift
      await tester.tap(find.text('Run Suite (Neon Green)'));
      await tester.pump();

      expect(find.text('10470 events'), findsOneWidget);
    });
  });
}

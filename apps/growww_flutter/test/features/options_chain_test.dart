import 'package:test/test.dart';
import 'package:growww_flutter/features/options_chain/domain/models/options_enums.dart';
import 'package:growww_flutter/features/options_chain/domain/services/options_analytics_engine.dart';
import 'package:growww_flutter/features/options_chain/presentation/controllers/options_chain_controller.dart';

void main() {
  group('Prompt 542 - Options Chain Matrix & Strike Selector Screen', () {
    late OptionsChainController controller;

    setUp(() {
      controller = OptionsChainController(
        symbol: 'BTC/USDT',
        initialSpot: 64250.0,
      );
    });

    tearDown(() {
      controller.dispose();
    });

    test('OptionsAnalyticsEngine resolves ATM strike and classifies moneyness', () {
      final strikes = [62000.0, 63000.0, 64000.0, 65000.0, 66000.0];
      final atm = OptionsAnalyticsEngine.findAtmStrike(64250.0, strikes);
      expect(atm, equals(64000.0));

      final atm2 = OptionsAnalyticsEngine.findAtmStrike(64600.0, strikes);
      expect(atm2, equals(65000.0));

      // Call Moneyness
      final callItm = OptionsAnalyticsEngine.classifyMoneyness(
        strike: 60000.0,
        spotPrice: 64250.0,
        type: OptionType.call,
      );
      expect(callItm.isInTheMoney, isTrue);

      final callOtm = OptionsAnalyticsEngine.classifyMoneyness(
        strike: 70000.0,
        spotPrice: 64250.0,
        type: OptionType.call,
      );
      expect(callOtm, equals(OptionMoneyness.deepOtm));

      // Put Moneyness
      final putItm = OptionsAnalyticsEngine.classifyMoneyness(
        strike: 70000.0,
        spotPrice: 64250.0,
        type: OptionType.put,
      );
      expect(putItm.isInTheMoney, isTrue);

      final putOtm = OptionsAnalyticsEngine.classifyMoneyness(
        strike: 60000.0,
        spotPrice: 64250.0,
        type: OptionType.put,
      );
      expect(putOtm, equals(OptionMoneyness.deepOtm));
    });

    test('OptionsAnalyticsEngine computes PCR and Max Pain strike', () {
      final pcr = OptionsAnalyticsEngine.computePcr(
        totalPutOi: 12000.0,
        totalCallOi: 10000.0,
      );
      expect(pcr, equals(1.2));

      // Zero call OI division guard
      expect(OptionsAnalyticsEngine.computePcr(totalPutOi: 5000.0, totalCallOi: 0.0), equals(0.0));

      final snapshot = OptionsAnalyticsEngine.generateMockChain(
        symbol: 'BTC/USDT',
        spotPrice: 64000.0,
        selectedExpiry: '26 SEP 2026',
        expiries: ['26 SEP 2026'],
      );

      expect(snapshot.strikeRows.length, greaterThanOrEqualTo(5));
      expect(snapshot.atmRow, isNotNull);
      expect(snapshot.atmRow?.strikePrice, equals(64000.0));
      expect(snapshot.maxPainStrike, greaterThan(0.0));
      expect(snapshot.overallPcr, greaterThan(0.0));
    });

    test('OptionsChainController handles expiry switching, filtering, and order staging', () async {
      expect(controller.state.snapshot.selectedExpiry, equals('26 SEP 2026'));
      expect(controller.state.selectedContract, isNull);

      // Change expiry
      controller.selectExpiry('03 OCT 2026');
      expect(controller.state.snapshot.selectedExpiry, equals('03 OCT 2026'));
      expect(controller.state.toastMessage, contains('Loaded expiry 03 OCT 2026'));

      // Change filter and greeks
      controller.setFilter(OptionsViewFilter.callsOnly);
      expect(controller.state.filter, equals(OptionsViewFilter.callsOnly));

      controller.toggleGreeks();
      expect(controller.state.showGreeks, isTrue);

      // Select contract to stage order
      final callContract = controller.state.snapshot.strikeRows.first.call;
      controller.selectContract(callContract, 'BUY');

      expect(controller.state.selectedContract?.symbol, equals(callContract.symbol));
      expect(controller.state.stagedSide, equals('BUY'));
      expect(controller.state.estimatedOrderValue, greaterThan(0.0));

      // Submit staged order
      final executed = await controller.placeStagedOrder();
      expect(executed, isTrue);
      expect(controller.state.selectedContract, isNull);
      expect(controller.state.toastMessage, contains('BUY 1.0x ${callContract.symbol}'));
    });
  });
}

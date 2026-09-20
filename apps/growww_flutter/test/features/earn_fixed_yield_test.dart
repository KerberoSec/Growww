import 'package:test/test.dart';
import 'package:growww_flutter/features/earn/data/repositories/mock_earn_vault_repository.dart';
import 'package:growww_flutter/features/earn/domain/models/earn_enums.dart';
import 'package:growww_flutter/features/earn/domain/models/sub_paise_amount.dart';
import 'package:growww_flutter/features/earn/domain/services/yield_math_engine.dart';
import 'package:growww_flutter/features/earn/presentation/controllers/apy_calculator_controller.dart';
import 'package:growww_flutter/features/earn/presentation/controllers/earn_dashboard_controller.dart';

void main() {
  group('Prompt 539 - Fixed-Income Yield & Staking Vault Suite', () {
    late MockEarnVaultRepository repository;
    late EarnDashboardController dashboardController;

    setUp(() {
      repository = MockEarnVaultRepository();
      dashboardController = EarnDashboardController(
        vaultRepo: repository,
        stakingRepo: repository,
      );
    });

    tearDown(() {
      dashboardController.dispose();
    });

    test('SubPaiseAmount formats Indian Rupee numbers with Lakhs and Crores notation', () {
      final amt1 = SubPaiseAmount.fromInr(1234.56);
      expect(amt1.formatInr(), equals('₹1,234.56'));

      final amtLakh = SubPaiseAmount.fromInr(548291.50);
      expect(amtLakh.formatInr(), equals('₹5,48,291.50'));

      final amtCrore = SubPaiseAmount.fromInr(12500000.00);
      expect(amtCrore.formatInr(), equals('₹1,25,00,000.00'));

      final microAccrual = SubPaiseAmount(14528942); // 1,452.8942 INR
      expect(microAccrual.formatInr(showSubPaise: true), equals('₹1,452.8942'));
    });

    test('SubPaiseAmount arithmetic operations preserve exact integer precision', () {
      final a = SubPaiseAmount.fromInr(100.5025);
      final b = SubPaiseAmount.fromInr(49.4975);

      final sum = a + b;
      expect(sum.toInr(), equals(150.0));
      expect(sum.subPaiseValue, equals(1500000));

      final diff = a - b;
      expect(diff.toInr(), equals(51.0050));

      final multiplied = a.multiply(2.0);
      expect(multiplied.toInr(), equals(201.0050));
    });

    test('YieldMathEngine calculates accurate compound vs maturity returns across tenors', () {
      final principal = SubPaiseAmount.fromInr(100000.0); // ₹ 1 Lakh
      const baseApy = 7.18; // gGSEC base rate

      // 90 Days Lock (Base 7.18 + 0.50 Bonus = 7.68% APY)
      final res90D = YieldMathEngine.computeReturns(
        principal: principal,
        baseApy: baseApy,
        tenor: TenorType.locked90Days,
        compounding: CompoundingFrequency.daily,
      );

      expect(res90D.effectiveApy, equals(7.68));
      expect(res90D.projectedDailyYield.toInr(), closeTo(21.04, 0.05));
      expect(res90D.projectedTotalYield.toInr(), closeTo(1912.0, 2.0));
      expect(res90D.estimatedTdsDeduction.toInr(), equals(0.0)); // Below ₹40,000 threshold

      // 365 Days Lock (Base 7.18 + 1.25 Bonus = 8.43% APY)
      final res1Y = YieldMathEngine.computeReturns(
        principal: SubPaiseAmount.fromInr(1000000.0), // ₹ 10 Lakhs
        baseApy: baseApy,
        tenor: TenorType.locked365Days,
        compounding: CompoundingFrequency.daily,
      );

      expect(res1Y.effectiveApy, equals(8.43));
      expect(res1Y.projectedTotalYield.toInr(), greaterThan(87000.0)); // > ₹40,000
      expect(res1Y.estimatedTdsDeduction.toInr(), greaterThan(8700.0)); // 10% TDS applied
    });

    test('Section 194A TDS withholding rules and Form 15G/15H exemptions', () {
      // Below threshold (₹35,000)
      final belowThreshold = SubPaiseAmount.fromInr(35000.0);
      final tds1 = YieldMathEngine.estimateTds(grossInterest: belowThreshold);
      expect(tds1.tdsRatePercentage, equals(0.0));
      expect(tds1.tdsAmountWithheld.toInr(), equals(0.0));

      // Standard PAN above threshold (₹60,000)
      final aboveThreshold = SubPaiseAmount.fromInr(60000.0);
      final tds2 = YieldMathEngine.estimateTds(grossInterest: aboveThreshold);
      expect(tds2.tdsRatePercentage, equals(10.0));
      expect(tds2.tdsAmountWithheld.toInr(), equals(6000.0));
      expect(tds2.netInterestPayable.toInr(), equals(54000.0));

      // Form 15G applied (Zero TDS)
      final tdsForm15 = YieldMathEngine.estimateTds(
        grossInterest: aboveThreshold,
        hasForm15G15H: true,
      );
      expect(tdsForm15.tdsRatePercentage, equals(0.0));
      expect(tdsForm15.tdsAmountWithheld.toInr(), equals(0.0));
      expect(tdsForm15.isForm15G15HApplied, isTrue);

      // Non-PAN penalty rate (20%)
      final tdsNonPan = YieldMathEngine.estimateTds(
        grossInterest: aboveThreshold,
        hasValidPan: false,
      );
      expect(tdsNonPan.tdsRatePercentage, equals(20.0));
      expect(tdsNonPan.tdsAmountWithheld.toInr(), equals(12000.0));
    });

    test('Emergency unstake penalty calculates strict 0.25% fee', () {
      final principal = SubPaiseAmount.fromInr(100000.0);
      final penalty = YieldMathEngine.calculateEmergencyPenalty(principal);
      expect(penalty.toInr(), equals(250.0));
    });

    test('ApyCalculatorController reacts dynamically to principal and tenor adjustments', () async {
      final vaults = await repository.fetchYieldVaults();
      final vault = vaults.first;

      final calcController = ApyCalculatorController(vault: vault);
      expect(calcController.state.principal.toInr(), equals(25000.0));
      expect(calcController.state.tenor, equals(TenorType.locked90Days));

      // Change principal to ₹ 1,00,000
      calcController.setPrincipal(SubPaiseAmount.fromInr(100000.0));
      expect(calcController.state.principal.toInr(), equals(100000.0));
      expect(calcController.state.result.projectedTotalYield.toInr(), greaterThan(1800.0));

      // Switch tenor to 365D
      calcController.setTenor(TenorType.locked365Days);
      expect(calcController.state.result.effectiveApy, equals(vault.baseApy + 1.25));

      calcController.dispose();
    });

    test('EarnDashboardController manages vault discovery and gasless claim execution', () async {
      await dashboardController.loadDashboard();
      expect(dashboardController.state.vaults.isNotEmpty, isTrue);
      expect(dashboardController.state.activeStakes.isNotEmpty, isTrue);
      expect(dashboardController.state.totalStakedPrincipal.toInr(), equals(50000.0));

      // Filter by G-Sec
      dashboardController.setFilter(YieldAssetType.gGsec);
      await Future.delayed(const Duration(milliseconds: 50));
      expect(dashboardController.state.vaults.every((v) => v.assetType == YieldAssetType.gGsec), isTrue);

      // Claim yield gaslessly
      final stake = dashboardController.state.activeStakes.first;
      final ok = await dashboardController.claimYield(stake.stakeId, ClaimDestination.dematCashLedger);
      expect(ok, isTrue);
      expect(dashboardController.state.successMessage, contains('Yield claimed successfully'));
    });
  });
}

import 'package:test/test.dart';
import 'package:growww_flutter/features/deposit_multichain/data/repositories/mock_multichain_deposit_repository.dart';
import 'package:growww_flutter/features/deposit_multichain/domain/models/deposit_enums.dart';
import 'package:growww_flutter/features/deposit_multichain/domain/services/address_validator_engine.dart';
import 'package:growww_flutter/features/deposit_multichain/presentation/controllers/multichain_deposit_controller.dart';

void main() {
  group('Prompt 565 - Multichain USDT Deposit Modal & Network Selector', () {
    late MockMultichainDepositRepository repository;
    late MultichainDepositController controller;

    setUp(() {
      repository = MockMultichainDepositRepository();
      controller = MultichainDepositController(repository: repository);
    });

    tearDown(() {
      controller.dispose();
    });

    test('AddressValidatorEngine validates Ethereum, Polygon, Arbitrum EVM addresses', () {
      const validEvm = '0x4F8E24C8900A9E71F1166Db9B72F472bC6DEe12A';
      expect(AddressValidatorEngine.validateAddress(validEvm, DepositNetworkType.ethereumErc20), isTrue);
      expect(AddressValidatorEngine.validateAddress(validEvm, DepositNetworkType.polygon), isTrue);
      expect(AddressValidatorEngine.validateAddress(validEvm, DepositNetworkType.arbitrumOne), isTrue);

      // Invalid: missing 0x prefix
      const noPrefix = '4F8E24C8900A9E71F1166Db9B72F472bC6DEe12A';
      expect(AddressValidatorEngine.validateAddress(noPrefix, DepositNetworkType.ethereumErc20), isFalse);

      // Invalid: too short
      const shortEvm = '0x4F8E24C8900A9E71F1166Db9B72';
      expect(AddressValidatorEngine.validateAddress(shortEvm, DepositNetworkType.ethereumErc20), isFalse);

      // Invalid: non-hex characters
      const invalidHex = '0x4F8E24C8900A9E71F1166Db9B72F472bC6DEe12Z';
      expect(AddressValidatorEngine.validateAddress(invalidHex, DepositNetworkType.ethereumErc20), isFalse);
    });

    test('AddressValidatorEngine validates Tron TRC-20 Base58 addresses', () {
      const validTron = 'TJ3q7F8a9NbK4E6x1yPz9mQwLkR4vCdE2s';
      expect(AddressValidatorEngine.validateAddress(validTron, DepositNetworkType.tronTrc20), isTrue);

      // Invalid: doesn't start with T
      const notTron = 'AJ3q7F8a9NbK4E6x1yPz9mQwLkR4vCdE2s';
      expect(AddressValidatorEngine.validateAddress(notTron, DepositNetworkType.tronTrc20), isFalse);

      // Invalid: contains invalid Base58 character (0, O, I, l)
      const invalidBase58 = 'TJ3q7F8a9NbK4E6x1yPz9mQwLkR4vCdE0s'; // Contains 0
      expect(AddressValidatorEngine.validateAddress(invalidBase58, DepositNetworkType.tronTrc20), isFalse);
    });

    test('AddressValidatorEngine validates Solana SPL addresses', () {
      const validSolana = '8Zb1x7Y9kG4mC3wQ6vR9pX2tL5nE8aD1sF4gH7jK9mN';
      expect(AddressValidatorEngine.validateAddress(validSolana, DepositNetworkType.solana), isTrue);

      // Invalid: too short
      const shortSol = '8Zb1x7Y9kG';
      expect(AddressValidatorEngine.validateAddress(shortSol, DepositNetworkType.solana), isFalse);
    });

    test('MultichainDepositController manages network selection and dynamic address retrieval', () async {
      await controller.loadNetworks();
      expect(controller.state.networks.isNotEmpty, isTrue);
      expect(controller.state.selectedNetwork, isNotNull);
      expect(controller.state.depositAddress, isNotNull);

      // Switch to Ethereum
      final ethNet = controller.state.networks
          .firstWhere((n) => n.networkType == DepositNetworkType.ethereumErc20);
      await controller.selectNetwork(ethNet);

      expect(controller.state.selectedNetwork?.networkType, equals(DepositNetworkType.ethereumErc20));
      expect(controller.state.depositAddress?.address, startsWith('0x'));

      // Copy address
      controller.copyAddressToClipboard();
      expect(controller.state.copySuccessNotice, contains('Address copied to clipboard'));
    });
  });
}

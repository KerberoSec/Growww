import 'dart:async';
import '../../domain/interfaces/i_multichain_deposit_repository.dart';
import '../../domain/models/crypto_network_info.dart';
import '../../domain/models/deposit_address_info.dart';
import '../../domain/models/deposit_enums.dart';

class MockMultichainDepositRepository implements IMultichainDepositRepository {
  static final List<CryptoNetworkInfo> _networks = [
    const CryptoNetworkInfo(
      networkId: 'NET-TRON-TRC20',
      networkType: DepositNetworkType.tronTrc20,
      displayName: 'Tron (TRC-20)',
      chainName: 'TRON Mainnet',
      requiredConfirmations: 1,
      estimatedArrivalMinutes: 2,
      depositFeeUsdt: 0.0,
      minDepositUsdt: 10.0,
      contractAddress: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
      explorerUrl: 'https://tronscan.org/#/token20/TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
      congestionLevel: NetworkCongestionLevel.low,
      isEnabled: true,
    ),
    const CryptoNetworkInfo(
      networkId: 'NET-ETH-ERC20',
      networkType: DepositNetworkType.ethereumErc20,
      displayName: 'Ethereum (ERC-20)',
      chainName: 'Ethereum Mainnet',
      requiredConfirmations: 12,
      estimatedArrivalMinutes: 5,
      depositFeeUsdt: 0.0,
      minDepositUsdt: 50.0,
      contractAddress: '0xdac17f958d2ee523a2206206994597c13d831ec7',
      explorerUrl: 'https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7',
      congestionLevel: NetworkCongestionLevel.medium,
      isEnabled: true,
    ),
    const CryptoNetworkInfo(
      networkId: 'NET-POLYGON',
      networkType: DepositNetworkType.polygon,
      displayName: 'Polygon PoS (POL)',
      chainName: 'Polygon Mainnet',
      requiredConfirmations: 64,
      estimatedArrivalMinutes: 3,
      depositFeeUsdt: 0.0,
      minDepositUsdt: 10.0,
      contractAddress: '0xc2132d05d31c914a87c6611c10748aeb04b58e8f',
      explorerUrl: 'https://polygonscan.com/token/0xc2132d05d31c914a87c6611c10748aeb04b58e8f',
      congestionLevel: NetworkCongestionLevel.low,
      isEnabled: true,
    ),
    const CryptoNetworkInfo(
      networkId: 'NET-SOLANA',
      networkType: DepositNetworkType.solana,
      displayName: 'Solana (SPL)',
      chainName: 'Solana Mainnet-Beta',
      requiredConfirmations: 32,
      estimatedArrivalMinutes: 1,
      depositFeeUsdt: 0.0,
      minDepositUsdt: 5.0,
      contractAddress: 'Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB',
      explorerUrl: 'https://solscan.io/token/Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB',
      congestionLevel: NetworkCongestionLevel.low,
      isEnabled: true,
    ),
    const CryptoNetworkInfo(
      networkId: 'NET-ARBITRUM',
      networkType: DepositNetworkType.arbitrumOne,
      displayName: 'Arbitrum One (Rollup)',
      chainName: 'Arbitrum One',
      requiredConfirmations: 20,
      estimatedArrivalMinutes: 2,
      depositFeeUsdt: 0.0,
      minDepositUsdt: 10.0,
      contractAddress: '0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9',
      explorerUrl: 'https://arbiscan.io/token/0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9',
      congestionLevel: NetworkCongestionLevel.low,
      isEnabled: true,
    ),
  ];

  @override
  Future<List<CryptoNetworkInfo>> fetchSupportedNetworks() async {
    return _networks;
  }

  @override
  Future<DepositAddressInfo> fetchDepositAddress({
    required DepositNetworkType networkType,
  }) async {
    String addr;
    switch (networkType) {
      case DepositNetworkType.ethereumErc20:
      case DepositNetworkType.polygon:
      case DepositNetworkType.arbitrumOne:
      case DepositNetworkType.bscBep20:
        addr = '0x4F8E24C8900A9E71F1166Db9B72F472bC6DEe12A';
        break;
      case DepositNetworkType.tronTrc20:
        addr = 'TJ3q7F8a9NbK4E6x1yPz9mQwLkR4vCdE2s';
        break;
      case DepositNetworkType.solana:
        addr = '8Zb1x7Y9kG4mC3wQ6vR9pX2tL5nE8aD1sF4gH7jK9mN';
        break;
    }

    return DepositAddressInfo(
      networkType: networkType,
      address: addr,
      qrPayload: 'ethereum:$addr?contractAddress=usdt',
      isSegregatedVault: true,
      generatedAt: DateTime.now(),
    );
  }

  @override
  Stream<DepositStatus> streamDepositStatus({
    required String address,
  }) {
    return Stream.fromIterable([
      DepositStatus.detecting,
      DepositStatus.confirming,
      DepositStatus.credited,
    ]);
  }
}

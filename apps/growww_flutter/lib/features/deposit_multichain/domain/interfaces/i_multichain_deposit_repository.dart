import '../models/crypto_network_info.dart';
import '../models/deposit_address_info.dart';
import '../models/deposit_enums.dart';

abstract class IMultichainDepositRepository {
  Future<List<CryptoNetworkInfo>> fetchSupportedNetworks();

  Future<DepositAddressInfo> fetchDepositAddress({
    required DepositNetworkType networkType,
  });

  Stream<DepositStatus> streamDepositStatus({
    required String address,
  });
}

import '../models/deposit_enums.dart';

class AddressValidatorEngine {
  static const String _base58Alphabet =
      '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';

  static bool validateAddress(String address, DepositNetworkType networkType) {
    final trimmed = address.trim();
    if (trimmed.isEmpty) return false;

    switch (networkType) {
      case DepositNetworkType.ethereumErc20:
      case DepositNetworkType.polygon:
      case DepositNetworkType.arbitrumOne:
      case DepositNetworkType.bscBep20:
        return _validateEvmAddress(trimmed);

      case DepositNetworkType.tronTrc20:
        return _validateTronAddress(trimmed);

      case DepositNetworkType.solana:
        return _validateSolanaAddress(trimmed);
    }
  }

  static bool _validateEvmAddress(String address) {
    if (address.length != 42) return false;
    if (!address.startsWith('0x') && !address.startsWith('0X')) return false;

    final hexPart = address.substring(2);
    final hexRegex = RegExp(r'^[0-9a-fA-F]{40}$');
    return hexRegex.hasMatch(hexPart);
  }

  static bool _validateTronAddress(String address) {
    if (address.length != 34) return false;
    if (!address.startsWith('T')) return false;

    for (int i = 0; i < address.length; i++) {
      if (!_base58Alphabet.contains(address[i])) return false;
    }
    return true;
  }

  static bool _validateSolanaAddress(String address) {
    if (address.length < 32 || address.length > 44) return false;

    for (int i = 0; i < address.length; i++) {
      if (!_base58Alphabet.contains(address[i])) return false;
    }
    return true;
  }
}

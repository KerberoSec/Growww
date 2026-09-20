import 'dart:async';
import '../../domain/interfaces/i_multichain_deposit_repository.dart';
import '../../domain/models/crypto_network_info.dart';
import '../../domain/models/deposit_address_info.dart';
import '../../domain/models/deposit_enums.dart';

class MultichainDepositState {
  final List<CryptoNetworkInfo> networks;
  final CryptoNetworkInfo? selectedNetwork;
  final DepositAddressInfo? depositAddress;
  final bool isLoading;
  final String? errorMessage;
  final String? copySuccessNotice;

  const MultichainDepositState({
    this.networks = const [],
    this.selectedNetwork,
    this.depositAddress,
    this.isLoading = false,
    this.errorMessage,
    this.copySuccessNotice,
  });

  MultichainDepositState copyWith({
    List<CryptoNetworkInfo>? networks,
    CryptoNetworkInfo? selectedNetwork,
    DepositAddressInfo? depositAddress,
    bool? isLoading,
    String? errorMessage,
    String? copySuccessNotice,
  }) {
    return MultichainDepositState(
      networks: networks ?? this.networks,
      selectedNetwork: selectedNetwork ?? this.selectedNetwork,
      depositAddress: depositAddress ?? this.depositAddress,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      copySuccessNotice: copySuccessNotice,
    );
  }
}

class MultichainDepositController {
  final IMultichainDepositRepository _repository;

  MultichainDepositState _state = const MultichainDepositState();
  final _stateController = StreamController<MultichainDepositState>.broadcast();

  MultichainDepositState get state => _state;
  Stream<MultichainDepositState> get stream => _stateController.stream;

  MultichainDepositController({required IMultichainDepositRepository repository})
      : _repository = repository {
    loadNetworks();
  }

  void _emit(MultichainDepositState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  Future<void> loadNetworks() async {
    _emit(_state.copyWith(isLoading: true));
    try {
      final list = await _repository.fetchSupportedNetworks();
      final defaultNet = list.isNotEmpty ? list.first : null;
      _emit(_state.copyWith(
        networks: list,
        selectedNetwork: defaultNet,
        isLoading: false,
      ));

      if (defaultNet != null) {
        await selectNetwork(defaultNet);
      }
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  Future<void> selectNetwork(CryptoNetworkInfo network) async {
    _emit(_state.copyWith(selectedNetwork: network, isLoading: true));
    try {
      final addr = await _repository.fetchDepositAddress(networkType: network.networkType);
      _emit(_state.copyWith(
        depositAddress: addr,
        isLoading: false,
      ));
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  void copyAddressToClipboard() {
    _emit(_state.copyWith(copySuccessNotice: 'Address copied to clipboard!'));
  }

  void dispose() {
    _stateController.close();
  }
}

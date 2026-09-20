import 'dart:async';
import '../../domain/interfaces/i_commodity_redemption_repository.dart';
import '../../domain/interfaces/i_logistics_tracking_repository.dart';
import '../../domain/models/commodity_denomination.dart';
import '../../domain/models/commodity_enums.dart';
import '../../domain/models/delivery_address.dart';
import '../../domain/models/delivery_receipt.dart';
import '../../domain/models/logistics_tracking_event.dart';
import '../../domain/models/redemption_order.dart';
import '../../domain/models/sub_paise_fee_breakdown.dart';
import '../../domain/models/wdra_vault_location.dart';

class RedemptionWizardState {
  final CommodityType selectedCommodity;
  final double availableBalanceGrams;
  final List<CommodityDenomination> availableDenominations;
  final Map<String, int> selectedQuantities; // denominationId -> qty
  final DeliveryMode deliveryMode;
  final WdraVaultLocation? selectedVault;
  final DeliveryAddress? deliveryAddress;
  final SubPaiseFeeBreakdown? feeBreakdown;
  final bool isLoading;
  final String? errorMessage;
  final RedemptionOrder? activeOrder;
  final DeliveryReceipt? deliveryReceipt;
  final List<LogisticsTrackingEvent> trackingEvents;
  final String? currentOtp;

  const RedemptionWizardState({
    this.selectedCommodity = CommodityType.gold999,
    this.availableBalanceGrams = 25.0,
    this.availableDenominations = const [],
    this.selectedQuantities = const {},
    this.deliveryMode = DeliveryMode.armoredDoorstep,
    this.selectedVault,
    this.deliveryAddress,
    this.feeBreakdown,
    this.isLoading = false,
    this.errorMessage,
    this.activeOrder,
    this.deliveryReceipt,
    this.trackingEvents = const [],
    this.currentOtp,
  });

  double get totalSelectedGrams {
    double total = 0.0;
    for (final denom in availableDenominations) {
      final qty = selectedQuantities[denom.denominationId] ?? 0;
      total += denom.weightGrams * qty;
    }
    return total;
  }

  bool get isValidWeight {
    final weight = totalSelectedGrams;
    return weight >= selectedCommodity.minimumRedemptionGrams &&
        weight <= availableBalanceGrams;
  }

  RedemptionWizardState copyWith({
    CommodityType? selectedCommodity,
    double? availableBalanceGrams,
    List<CommodityDenomination>? availableDenominations,
    Map<String, int>? selectedQuantities,
    DeliveryMode? deliveryMode,
    WdraVaultLocation? selectedVault,
    DeliveryAddress? deliveryAddress,
    SubPaiseFeeBreakdown? feeBreakdown,
    bool? isLoading,
    String? errorMessage,
    RedemptionOrder? activeOrder,
    DeliveryReceipt? deliveryReceipt,
    List<LogisticsTrackingEvent>? trackingEvents,
    String? currentOtp,
  }) {
    return RedemptionWizardState(
      selectedCommodity: selectedCommodity ?? this.selectedCommodity,
      availableBalanceGrams: availableBalanceGrams ?? this.availableBalanceGrams,
      availableDenominations: availableDenominations ?? this.availableDenominations,
      selectedQuantities: selectedQuantities ?? this.selectedQuantities,
      deliveryMode: deliveryMode ?? this.deliveryMode,
      selectedVault: selectedVault ?? this.selectedVault,
      deliveryAddress: deliveryAddress ?? this.deliveryAddress,
      feeBreakdown: feeBreakdown ?? this.feeBreakdown,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      activeOrder: activeOrder ?? this.activeOrder,
      deliveryReceipt: deliveryReceipt ?? this.deliveryReceipt,
      trackingEvents: trackingEvents ?? this.trackingEvents,
      currentOtp: currentOtp ?? this.currentOtp,
    );
  }
}

class CommodityRedemptionController {
  final ICommodityRedemptionRepository _redemptionRepo;
  final ILogisticsTrackingRepository _trackingRepo;

  RedemptionWizardState _state = const RedemptionWizardState();
  final _stateController = StreamController<RedemptionWizardState>.broadcast();

  RedemptionWizardState get state => _state;
  Stream<RedemptionWizardState> get stream => _stateController.stream;

  CommodityRedemptionController({
    required ICommodityRedemptionRepository redemptionRepo,
    required ILogisticsTrackingRepository trackingRepo,
  })  : _redemptionRepo = redemptionRepo,
        _trackingRepo = trackingRepo {
    initialize();
  }

  void _emit(RedemptionWizardState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  Future<void> initialize() async {
    _emit(_state.copyWith(isLoading: true));
    try {
      final denoms = await _redemptionRepo.fetchAvailableDenominations(
        commodityType: _state.selectedCommodity,
      );
      _emit(_state.copyWith(
        availableDenominations: denoms,
        isLoading: false,
      ));
      await recalculateFees();
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  Future<void> selectCommodity(CommodityType type) async {
    final balance = type == CommodityType.silver999 ? 2500.0 : 25.0;
    _emit(_state.copyWith(
      selectedCommodity: type,
      availableBalanceGrams: balance,
      selectedQuantities: {},
      isLoading: true,
    ));
    try {
      final denoms = await _redemptionRepo.fetchAvailableDenominations(commodityType: type);
      _emit(_state.copyWith(
        availableDenominations: denoms,
        isLoading: false,
      ));
      await recalculateFees();
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  Future<void> updateQuantity(String denominationId, int delta) async {
    final current = _state.selectedQuantities[denominationId] ?? 0;
    final next = (current + delta).clamp(0, 50);
    final updated = Map<String, int>.from(_state.selectedQuantities);
    if (next == 0) {
      updated.remove(denominationId);
    } else {
      updated[denominationId] = next;
    }
    _emit(_state.copyWith(selectedQuantities: updated));
    await recalculateFees();
  }

  Future<void> setDeliveryMode(DeliveryMode mode) async {
    _emit(_state.copyWith(deliveryMode: mode));
    await recalculateFees();
  }

  Future<void> selectVault(WdraVaultLocation vault) async {
    _emit(_state.copyWith(selectedVault: vault));
    await recalculateFees();
  }

  Future<void> setDeliveryAddress(DeliveryAddress address) async {
    _emit(_state.copyWith(deliveryAddress: address));
    await recalculateFees();
  }

  List<DenominationOrderItem> _buildOrderItems() {
    final items = <DenominationOrderItem>[];
    for (final denom in _state.availableDenominations) {
      final qty = _state.selectedQuantities[denom.denominationId] ?? 0;
      if (qty > 0) {
        items.add(DenominationOrderItem(denomination: denom, quantity: qty));
      }
    }
    return items;
  }

  Future<void> recalculateFees() async {
    final items = _buildOrderItems();
    if (items.isEmpty) {
      _emit(_state.copyWith(feeBreakdown: null));
      return;
    }
    try {
      final fee = await _redemptionRepo.calculateDeliveryFee(
        commodityType: _state.selectedCommodity,
        items: items,
        deliveryMode: _state.deliveryMode,
        destinationPincode: _state.deliveryAddress?.pincode,
        vaultId: _state.selectedVault?.vaultId,
      );
      _emit(_state.copyWith(feeBreakdown: fee));
    } catch (e) {
      _emit(_state.copyWith(errorMessage: 'Fee calculation failed: $e'));
    }
  }

  Future<RedemptionOrder?> initiateRedemption() async {
    final items = _buildOrderItems();
    if (items.isEmpty || !_state.isValidWeight) {
      _emit(_state.copyWith(errorMessage: 'Invalid redemption weight selection.'));
      return null;
    }

    _emit(_state.copyWith(isLoading: true));
    try {
      final order = await _redemptionRepo.initiateRedemptionIntent(
        commodityType: _state.selectedCommodity,
        items: items,
        deliveryMode: _state.deliveryMode,
        vaultId: _state.selectedVault?.vaultId,
        addressId: _state.deliveryAddress?.addressId,
        idempotencyKey: 'IDEMP-${DateTime.now().millisecondsSinceEpoch}',
      );
      _emit(_state.copyWith(isLoading: false, activeOrder: order));
      return order;
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
      return null;
    }
  }

  Future<bool> authorizeRedemption({
    required String biometricSignature,
    required String totpCode,
  }) async {
    final order = _state.activeOrder;
    if (order == null) return false;

    _emit(_state.copyWith(isLoading: true));
    try {
      final authorized = await _redemptionRepo.authorizeRedemption(
        orderId: order.orderId,
        biometricSignature: biometricSignature,
        totpCode: totpCode,
      );
      _emit(_state.copyWith(isLoading: false, activeOrder: authorized));
      return true;
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
      return false;
    }
  }

  Future<void> loadTrackingInfo(String orderId) async {
    try {
      final events = await _trackingRepo.fetchTrackingHistory(orderId: orderId);
      final otp = await _trackingRepo.generateTimeLockedDeliveryOtp(orderId: orderId);
      _emit(_state.copyWith(trackingEvents: events, currentOtp: otp));
    } catch (e) {
      _emit(_state.copyWith(errorMessage: 'Tracking load failed: $e'));
    }
  }

  Future<void> fetchReceipt(String orderId) async {
    try {
      final receipt = await _redemptionRepo.fetchDeliveryReceipt(orderId: orderId);
      _emit(_state.copyWith(deliveryReceipt: receipt));
    } catch (e) {
      _emit(_state.copyWith(errorMessage: 'Receipt load failed: $e'));
    }
  }

  void dispose() {
    _stateController.close();
  }
}

// Stop-Limit & Conditional Trigger Enums (Prompt 549)

enum TriggerType {
  stopLimit,
  stopMarket,
  takeProfitLimit,
  takeProfitMarket,
  trailingStop,
}

extension TriggerTypeExtension on TriggerType {
  String get displayName {
    switch (this) {
      case TriggerType.stopLimit:
        return 'Stop Limit';
      case TriggerType.stopMarket:
        return 'Stop Market';
      case TriggerType.takeProfitLimit:
        return 'Take Profit Limit';
      case TriggerType.takeProfitMarket:
        return 'Take Profit Market';
      case TriggerType.trailingStop:
        return 'Trailing Stop';
    }
  }

  bool get isLimitOrder =>
      this == TriggerType.stopLimit || this == TriggerType.takeProfitLimit;
}

enum TriggerPriceType {
  lastPrice,
  markPrice,
  indexPrice,
}

extension TriggerPriceTypeExtension on TriggerPriceType {
  String get displayName {
    switch (this) {
      case TriggerPriceType.lastPrice:
        return 'Last Price (LTP)';
      case TriggerPriceType.markPrice:
        return 'Mark Price (Fair)';
      case TriggerPriceType.indexPrice:
        return 'Index Price';
    }
  }

  String get shortCode {
    switch (this) {
      case TriggerPriceType.lastPrice:
        return 'LAST';
      case TriggerPriceType.markPrice:
        return 'MARK';
      case TriggerPriceType.indexPrice:
        return 'INDEX';
    }
  }
}

enum TriggerCondition {
  greaterOrEqual,
  lessOrEqual,
}

extension TriggerConditionExtension on TriggerCondition {
  String get symbolText {
    switch (this) {
      case TriggerCondition.greaterOrEqual:
        return '>=';
      case TriggerCondition.lessOrEqual:
        return '<=';
    }
  }

  String get description {
    switch (this) {
      case TriggerCondition.greaterOrEqual:
        return 'Price Rises To / Above';
      case TriggerCondition.lessOrEqual:
        return 'Price Drops To / Below';
    }
  }
}

enum TriggerOrderStatus {
  pendingTrigger,
  triggered,
  filled,
  cancelled,
  rejected,
}

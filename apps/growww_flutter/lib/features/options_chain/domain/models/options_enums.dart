/// Options chain and derivatives trading enums.
enum OptionType {
  call,
  put;

  String get shortCode => this == OptionType.call ? 'CE' : 'PE';
  String get displayName => this == OptionType.call ? 'Call Option (CE)' : 'Put Option (PE)';
  bool get isCall => this == OptionType.call;
  bool get isPut => this == OptionType.put;
}

enum OptionMoneyness {
  deepItm,
  itm,
  atm,
  otm,
  deepOtm;

  String get displayName {
    switch (this) {
      case OptionMoneyness.deepItm:
        return 'Deep In-the-Money';
      case OptionMoneyness.itm:
        return 'In-the-Money';
      case OptionMoneyness.atm:
        return 'At-the-Money';
      case OptionMoneyness.otm:
        return 'Out-of-the-Money';
      case OptionMoneyness.deepOtm:
        return 'Deep Out-of-the-Money';
    }
  }

  bool get isInTheMoney =>
      this == OptionMoneyness.itm || this == OptionMoneyness.deepItm;
}

enum OptionsViewFilter {
  all,
  callsOnly,
  putsOnly,
  greeksOnly;

  String get label {
    switch (this) {
      case OptionsViewFilter.all:
        return 'Both (CE & PE)';
      case OptionsViewFilter.callsOnly:
        return 'Calls Only';
      case OptionsViewFilter.putsOnly:
        return 'Puts Only';
      case OptionsViewFilter.greeksOnly:
        return 'Greeks Analysis';
    }
  }
}

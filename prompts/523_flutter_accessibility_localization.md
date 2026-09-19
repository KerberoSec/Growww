# 523 - Accessibility & Multi-Language Localization (i18n & a11y) Architecture

## Purpose
Democratizing access to fractional Indian equities and digital security tokens requires supporting India's linguistically diverse investor population as well as international GIFT City investors. Furthermore, SEBI investor charter principles and modern accessibility standards (WCAG 2.1 AA) mandate that digital investment terminals be fully usable by individuals with visual, motor, or cognitive impairments.

This prompt specifies the Internationalization (i18n) and Accessibility (a11y) architecture for the Growww Flutter client. It delivers multi-lingual localization across **English, Hindi (हिन्दी), Marathi (मराठी), Gujarati (ગુજરાતી), Tamil (தமிழ்), Telugu (తెలుగు), Kannada (ಕನ್ನಡ), and Bengali (বাংলা)**. It provides Indian numbering system formatters (Lakhs and Crores: ₹1,23,456.78), IST timezone date-time formatters, full screen-reader semantic tree optimization (TalkBack, VoiceOver, Narrator, Orca), dynamic typography scaling without layout clipping, high-contrast themes, and full keyboard navigation for desktop platforms.

## What You Are Building
A centralized localization and accessibility engine in `lib/core/localization/` and `lib/core/accessibility/`:
- `l10n/app_en.arb`, `app_hi.arb`, `app_mr.arb`, `app_gu.arb`, `app_ta.arb`, `app_te.arb`, `app_kn.arb`, `app_bn.arb`: Application Resource Bundle (ARB) translation files with strict parameter typing, pluralization, and financial terminology glossaries.
- `IndianCurrencyFormatter`: Custom high-performance number formatter formatting fractional INR amounts in standard Indian notation (`1,00,000` = 1 Lakh; `1,00,00,000` = 1 Crore) and dynamic switching to international Million/Billion notation for GIFT City USD accounts.
- `SemanticTradingWidgets`: Reusable accessible UI components (`AccessiblePriceTicker`, `AccessibleOrderPad`, `AccessiblePnlIndicator`) enriched with custom Flutter `Semantics` announcing price movements ("Reliance up by 1.5 percent, current price 2,950 rupees 50 paise").
- `DesktopKeyboardNavigationManager`: Focus traversal policy (`FocusTraversalGroup`, `Shortcuts`, `Actions`) enabling seamless keyboard-only navigation across desktop order books and trading forms.
- `DynamicFontScaler`: Text layout resilience engine ensuring that when users increase system accessibility font sizes (up to 200%), text widgets reflow gracefully without UI overflow errors.

## Scope Boundaries
- **In Scope:**
 - ARB file generation and Flutter `gen-l10n` toolchain integration.
 - 8 core Indian languages + English localization.
 - Indian numbering system (Lakh/Crore) and currency symbol formatting.
 - Screen reader semantic node tree modeling (TalkBack, VoiceOver, Narrator, Orca).
 - High-contrast visual modes and dynamic typography scaling support.
 - Desktop keyboard shortcuts (F1-F12, Enter, Tab, Arrow keys for order entry).
- **Out of Scope / Handled Elsewhere:**
 - Dynamic translation of server-side market news or company announcements (Prompt 221).
 - UI layout implementations of specific trading screens (Prompts 506-512).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_localizations`, `intl` (v0.19+), and Flutter's built-in `Semantics` framework.
- **Justification:** Flutter's native localization system compiles ARB files directly into type-safe Dart classes (`AppLocalizations`) with zero runtime reflection overhead and native integration with platform locale listeners.
- **Dependencies:**
 - `flutter_localizations: sdk`
 - `intl: ^0.19.0`
 - `flutter_riverpod: ^2.5.1`

## Backend / Infra Touchpoints
- **User Preference Service (Prompt 201):** Persisting user language preference (`locale_code`) to server profile for consistent email/SMS dispatch.
- **CMS / Legal Disclosures CDN:** Fetching localized statutory risk disclosures and SEBI Investor Charters in matching regional languages.

## Blockchain Interaction
- **Cryptographic Merkle Proof Explanation in Regional Languages:** The accessibility and localization engine provides plain-language, regional-language explanations and screen-reader vocalizations for complex on-chain proof-of-reserve Merkle trees and DvP blockchain settlement receipts.

## Step-by-Step Build Instructions
1. In `apps/growww_flutter/l10n.yaml`, configure the Flutter localization generator (`arb-dir: lib/l10n`, `template-arb-file: app_en.arb`, `output-localization-file: app_localizations.dart`).
2. Author base `lib/l10n/app_en.arb` containing all UI keys, parameterized placeholders, gender forms, and pluralization rules for trading terms.
3. Create localized ARB files for Hindi (`app_hi.arb`), Marathi (`app_mr.arb`), Gujarati (`app_gu.arb`), Tamil (`app_ta.arb`), Telugu (`app_te.arb`), Kannada (`app_kn.arb`), and Bengali (`app_bn.arb`).
4. Implement `IndianCurrencyFormatter` utility supporting fractional precision (up to 4 decimal places for micro-units), Lakhs/Crores grouping, and negative P&L formatting.
5. Create `LocaleNotifier` using Riverpod to manage active app locale with fallback to system default and persistent storage in `SecureStorageService` (Prompt 521).
6. Configure `MaterialApp.router` in `lib/app.dart` with `localizationsDelegates` and `supportedLocales`.
7. Audit and enhance custom canvas widgets (like charts and order depth ladders) with `CustomPainterSemantics` to emit vocalizable accessibility nodes.
8. Implement `AccessibleOrderPad` wrapping buy/sell buttons with explicit `Semantics(button: true, label: "Confirm Buy Order for Reliance Industries", hint: "Double tap to submit trade")`.
9. Wrap color-coded P&L widgets (Green for Profit, Red for Loss) with semantic labels and directional icons (Up Arrow / Down Arrow) so colorblind users can immediately distinguish gain and loss without relying solely on color hue.
10. Implement `DesktopKeyboardShortcuts` mapping `Ctrl+B` (or `Cmd+B`) to Buy Order, `Ctrl+S` to Sell Order, and `Escape` to close modal sheets.
11. Implement `Directionality` support to ensure seamless layout rendering across LTR and future RTL locales.
12. Ensure all touch targets meet minimum accessible dimensions (48x48 dp on mobile, 32x32 dp on desktop).
13. Write unit tests for `IndianCurrencyFormatter` validating boundary numbers: `0`, `999`, `1000`, `100000`, `10000000`, and negative fractional values.
14. Write widget tests verifying that changing locale dynamically updates UI strings across all screens without reloading.
15. Perform accessibility audits using Android Accessibility Scanner, iOS Accessibility Inspector, and Windows Narrator.

## Interfaces / Contracts

```arb
// lib/l10n/app_en.arb
{
  "@@locale": "en",
  "appTitle": "Growww",
  "buyAction": "Buy",
  "sellAction": "Sell",
  "orderPlacedSuccess": "Order placed successfully for {quantity} units of {symbol}",
  "@orderPlacedSuccess": {
    "description": "Confirmation message after placing an order",
    "placeholders": {
      "quantity": { "type": "String", "example": "0.50" },
      "symbol": { "type": "String", "example": "TCS" }
    }
  },
  "marketStatusClosedAmo": "Market is closed. This order will be queued as an After Market Order (AMO).",
  "proofOfReserveVerified": "On-chain Proof-of-Reserve verified by NSDL/CDSL custody records.",
  "totalPortfolioValue": "Total Portfolio Value",
  "unrealizedPnl": "Unrealized P&L"
}
```

```dart
// lib/core/localization/indian_currency_formatter.dart
import 'package:intl/intl.dart';

class IndianCurrencyFormatter {
  static String formatInr(double amount, {int decimalDigits = 2, bool showSymbol = true}) {
    final symbol = showSymbol ? '₹' : '';
    final isNegative = amount < 0;
    final absAmount = amount.abs();

    final parts = absAmount.toStringAsFixed(decimalDigits).split('.');
    String integerPart = parts[0];
    final decimalPart = parts.length > 1 ? '.${parts[1]}' : '';

    if (integerPart.length > 3) {
      final lastThree = integerPart.substring(integerPart.length - 3);
      final remaining = integerPart.substring(0, integerPart.length - 3);
      
      final buffer = StringBuffer();
      for (int i = 0; i < remaining.length; i++) {
        if (i > 0 && (remaining.length - i) % 2 == 0) {
          buffer.write(',');
        }
        buffer.write(remaining[i]);
      }
      integerPart = '${buffer.toString()},$lastThree';
    }

    final sign = isNegative ? '-' : '';
    return '$sign$symbol$integerPart$decimalPart';
  }

  static String formatLakhCrore(double amount) {
    if (amount.abs() >= 10000000) {
      return '₹${(amount / 10000000).toStringAsFixed(2)} Cr';
    } else if (amount.abs() >= 100000) {
      return '₹${(amount / 100000).toStringAsFixed(2)} L';
    } else if (amount.abs() >= 1000) {
      return '₹${(amount / 1000).toStringAsFixed(2)} k';
    }
    return formatInr(amount);
  }
}
```

## Security & Compliance Notes
- **Legal Text Accuracy:** Translations for statutory risk disclosures, SEBI investor warnings, and terms of service must be certified by professional legal translators; AI or machine-translated approximations are strictly prohibited for statutory compliance texts.
- **Number Discrepancies:** Formatting bugs in currency presentation can lead to grave financial misunderstandings. All currency formatters must be backed by exhaustive property-based unit tests.
- **Screen Reader Privacy:** Sensitive masked values (e.g. `****1234` for bank accounts) must be vocalized as "ending in one two three four" rather than reading literal asterisk characters.

## Acceptance Criteria
- [ ] Application dynamically switches between all 8 supported Indian languages and English in real time.
- [ ] Number and currency formatters strictly conform to the Indian Lakhs/Crores numbering system.
- [ ] Screen readers (TalkBack, VoiceOver, Narrator, Orca) vocalize financial quotes, order pads, and portfolio summaries accurately.
- [ ] Text widgets scale cleanly up to 200% system font size without pixel overflow or overlapping elements.
- [ ] Desktop keyboard shortcuts allow complete order placement without mouse interaction.
- [ ] Colorblind-accessible visual cues (icons and text indicators) accompany all red/green financial P&L metrics.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 503 (Design System & Theming), Prompt 521 (Local Secure Storage).
- **Parallel Tasks:** Prompt 509 (Order Placement Flow), Prompt 510 (Portfolio Holdings Screen).

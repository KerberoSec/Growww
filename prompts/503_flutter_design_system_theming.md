# 503 - Flutter Design System & Theming (Light/Dark, Platform-Adaptive)

## Purpose
Provides a cohesive, compliant, and accessible design system and theming engine for the Growww multi-platform application. Financial applications demand extreme clarity, high visual contrast, intuitive color semantics (e.g., gain green vs loss red), and seamless adaptability across mobile touch interfaces and desktop mouse/keyboard environments. This design system ensures consistent branding, regulatory badge rendering, and rapid UI development.

## What You Are Building
A comprehensive UI component and theming library located under `apps/growww_flutter/lib/core/design_system/` featuring:
- **Design Tokens:** Strongly-typed color palettes (Light & Dark), semantic typography (Inter & JetBrains Mono for financial figures), spacing scales, corner radii, and shadow elevations.
- **Theme Extensions:** Custom `ThemeExtension` classes (`GrowwwColors`, `GrowwwTypography`, `GrowwwFinancialTheme`) accessible via `Theme.of(context)`.
- **Platform-Adaptive Core Widgets:** Custom buttons, numeric keypad inputs, interactive data tables, bottom sheets / desktop dialogs, form inputs, and shimmer skeletons.
- **Regulatory & Blockchain Verification Badges:** Standardized UI widgets for "Proof-of-Reserve Verified", "DvP Atomic Settlement", "SEBI Regulated", and "GIFT City IFSCA Approved".

## Scope Boundaries
- **In Scope:**
 - Design tokens, typography styles, light/dark themes, theme extensions, atomic UI components, responsive layout breakpoint utilities, and blockchain badge widgets.
- **Out of Scope / Handled Elsewhere:**
 - Full screen orchestration and flow coordination (handled in Prompts 504-515).
 - TradingView candlestick chart engine (handled in Prompt 508).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ Material 3 theming combined with adaptive styling.
 - *Justification:* Flutter's `ThemeData` combined with `ThemeExtension` provides zero-cost theming switching, automatic dark mode resolution from OS settings, and strict type safety without third-party styling bloat.
- **Typography:** `google_fonts` (Inter for UI copy, JetBrains Mono with tabular figures `fontFeatures: [FontFeature.tabularFigures()]` for stock quotes and balance numbers).
- **Vector Graphics:** `flutter_svg` for crisp, resolution-independent iconography and regulatory emblems.

## Backend / Infra Touchpoints
- **Asset Bundle:** Bundled vector icons, logos, and certified SEBI / NSDL / CDSL trust badges stored in `assets/icons/` and `assets/images/`.
- **Dynamic Theme Mode:** Persists user theme preference (System, Light, Dark) via local storage (Prompt 521).

## Blockchain Interaction
- **Trust Badges & Visual Proofs:** Provides specialized, tamper-evident visual indicators:
 - `ProofOfReserveBadge`: Interactive badge displaying green shield with tooltip showing latest on-chain block hash and custodial backing ratio (100.0% backed).
 - `DvPSettlementBadge`: Pill badge illustrating Delivery-versus-Payment escrow security status.
 - `BlockchainTxLinkWidget`: Styled hash display with one-tap copy and block explorer browser launch.

## Step-by-Step Build Instructions
1. Create design system structure in `lib/core/design_system/`: `tokens/`, `theme/`, `components/`, `badges/`, `layout/`.
2. Define base color tokens in `tokens/colors.dart` (Primary Teal/Green `#00D09C`, Surface Dark `#121418`, Surface Light `#FFFFFF`, Loss Red `#EB5B56`, Profit Green `#00B386`, Besu Blue `#3B82F6`, Warning Amber `#FFB703`).
3. Define typography tokens in `tokens/typography.dart` enforcing `tabularFigures` on all price and balance text styles to prevent layout jitter on real-time price updates.
4. Implement `GrowwwThemeExtension` classes for custom semantic colors and financial indicators.
5. Construct `AppTheme.light` and `AppTheme.dark` ThemeData instances configuring Material 3 color schemes, input decoration themes, app bar themes, and elevation overlays.
6. Build responsive layout helper (`ResponsiveLayout`, `Breakpoints`) handling Mobile (<600dp), Tablet (600-1024dp), and Desktop (>1024dp) viewports.
7. Build atomic component: `GrowwwButton` (Primary, Secondary, Danger, Ghost) with built-in loading spinners and disabled states.
8. Build atomic component: `GrowwwTextField` with floating labels, error states, and currency prefix (`₹` / `$`).
9. Build atomic component: `GrowwwNumericKeypad` optimized for fast PIN entry and fractional trade quantity entry.
10. Build atomic component: `GrowwwCard` and `GrowwwBottomSheet` with desktop modal dialog fallback.
11. Implement `ProofOfReserveBadge` and `DvPSettlementBadge` with animated glow effects and tap-to-inspect popups.
12. Build `GrowwwShimmer` widget for skeletonized loading states on financial summaries.
13. Create a widgetbook / design system preview catalog for developer testing and visual regression checks.

## Interfaces / Contracts
```dart
// lib/core/design_system/theme/growww_colors.dart
import 'package:flutter/material.dart';

@immutable
class GrowwwColors extends ThemeExtension<GrowwwColors> {
  final Color profit;
  final Color loss;
  final Color blockchainVerified;
  final Color sebiCompliant;
  final Color cardBackground;
  final Color divider;

  const GrowwwColors({
    required this.profit,
    required this.loss,
    required this.blockchainVerified,
    required this.sebiCompliant,
    required this.cardBackground,
    required this.divider,
  });

  @override
  GrowwwColors copyWith({
    Color? profit,
    Color? loss,
    Color? blockchainVerified,
    Color? sebiCompliant,
    Color? cardBackground,
    Color? divider,
  }) {
    return GrowwwColors(
      profit: profit ?? this.profit,
      loss: loss ?? this.loss,
      blockchainVerified: blockchainVerified ?? this.blockchainVerified,
      sebiCompliant: sebiCompliant ?? this.sebiCompliant,
      cardBackground: cardBackground ?? this.cardBackground,
      divider: divider ?? this.divider,
    );
  }

  @override
  GrowwwColors lerp(ThemeExtension<GrowwwColors>? other, double t) {
    if (other is! GrowwwColors) return this;
    return GrowwwColors(
      profit: Color.lerp(profit, other.profit, t)!,
      loss: Color.lerp(loss, other.loss, t)!,
      blockchainVerified: Color.lerp(blockchainVerified, other.blockchainVerified, t)!,
      sebiCompliant: Color.lerp(sebiCompliant, other.sebiCompliant, t)!,
      cardBackground: Color.lerp(cardBackground, other.cardBackground, t)!,
      divider: Color.lerp(divider, other.divider, t)!,
    );
  }
}
```

## Security & Compliance Notes
- **Accessibility & Contrast:** Color combinations must meet WCAG 2.1 AA contrast ratio (minimum 4.5:1 for normal text, 3:1 for large text).
- **Tabular Numerics for Accuracy:** All financial values must use fixed-width tabular figures to prevent misreading moving decimal points during high-volatility market events.
- **Unambiguous Regulatory Styling:** Regulatory warnings (SEBI risk alerts) must be rendered with distinctive high-visibility amber/red styling and cannot be obscured.

## Acceptance Criteria
- [ ] Light and Dark themes switch dynamically without UI glitches or restart.
- [ ] `GrowwwColors` and `GrowwwTypography` are fully accessible via context extensions.
- [ ] All numeric financial widgets use monospace/tabular figures with zero horizontal jitter during price ticks.
- [ ] `ProofOfReserveBadge` renders correctly in both compact (mobile) and expanded (desktop) layouts.
- [ ] Responsive layout breakpoints smoothly transition between mobile single-column and desktop multi-column views.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 502 (App Architecture).
- **Parallel Tasks:** Prompt 523 (Accessibility & Localization).
- **Enables:** Prompts 504-515 (All screen implementations).

# Cross-Platform Design System & Obsidian Visual Tokens Specification

**Specification ID:** SPEC-ARCH-043-DESIGN-SYSTEM  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Client UI/UX, Design Tokens & Typography Architecture  
**Target Platforms:** Web Pro Terminal (Next.js/Tailwind), Desktop Thick Client (macOS, Windows, Linux Flutter), Mobile (iOS & Android)  
**Last Updated:** September 2026  

---

## 1. Executive Overview

The Growww / NBSE trading design system enforces a sovereign, high-density financial interface optimized for microsecond visual perception, zero eye strain during multi-hour trading sessions, and zero layout shift.

The visual foundation is anchored by the **Deep Obsidian Palette** (`#0B0E14`), paired with high-contrast, color-blind accessible neon trading accents: Emerald Green (`#00F0A0`) for bids/profits and Ruby Crimson (`#FF3B56`) for asks/losses.

---

## 2. Universal Color Palette Tokens

```
+-------------------------------------------------------------------------------------------------------+
|                                    OBSIDIAN DESIGN TOKEN HIERARCHY                                    |
|                                                                                                       |
|  [ Canvas Ground ]   ---> #0B0E14  (Deepest Obsidian, zero-emission OLED black)                     |
|  [ Elevated Card ]   ---> #141822  (Surface layer for panels, orderbooks, and modular cards)         |
|  [ Popover / Modal ] ---> #1B2130  (Floating dialogue surfaces, dropdowns, and context menus)        |
|  [ Subtle Border ]   ---> #1F2636  (1px structural dividers and dockview splitters)                  |
|  [ Hover / Active ]  ---> #2A3348  (Interactive row highlights and active tabs)                      |
|                                                                                                       |
|  TRADING ACCENTS (HIGH-CONTRAST NEON):                                                                |
|  [ Bid / Up / Buy ]  ---> #00F0A0  (Neon Emerald, 4.5:1+ contrast on dark background)                 |
|  [ Ask / Down / Sell]---> #FF3B56  (Neon Ruby, 4.5:1+ contrast on dark background)                   |
|  [ Warning / Amber ] ---> #FFD600  (Volatility collar and circuit breaker warnings)                   |
|  [ Info / Cyan ]     ---> #00E5FF  (ERC-4337 Sponsored Paymaster gas badge)                          |
|  [ Neutral Text ]    ---> #FFFFFF (Primary 100%), #94A3B8 (Secondary 60%), #475569 (Muted 40%)       |
+-------------------------------------------------------------------------------------------------------+
```

### 2.1 CSS Variables & Tailwind Configuration (Web Pro Terminal)
```css
/* apps/growww_web/src/styles/tokens.css */
:root {
  --color-canvas: #0B0E14;
  --color-surface: #141822;
  --color-surface-elevated: #1B2130;
  --color-border: #1F2636;
  --color-border-subtle: #2A3348;

  --color-bid: #00F0A0;
  --color-bid-subtle: rgba(0, 240, 160, 0.12);
  --color-bid-glow: rgba(0, 240, 160, 0.25);

  --color-ask: #FF3B56;
  --color-ask-subtle: rgba(255, 59, 86, 0.12);
  --color-ask-glow: rgba(255, 59, 86, 0.25);

  --color-warning: #FFD600;
  --color-gas-sponsored: #00E5FF;
  --color-zero-fee: #00F0A0;

  --font-mono-tabular: "JetBrains Mono", "Roboto Mono", monospace;
  --font-sans-ui: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
```

### 2.2 Flutter ThemeExtension (Desktop & Mobile)
```dart
// apps/growww_flutter/lib/core/theme/obsidian_theme_tokens.dart
import 'package:flutter/material.dart';

@immutable
class ObsidianThemeTokens extends ThemeExtension<ObsidianThemeTokens> {
  final Color canvas;
  final Color surface;
  final Color surfaceElevated;
  final Color border;
  final Color bid;
  final Color bidSubtle;
  final Color ask;
  final Color askSubtle;
  final Color zeroFeeBadge;
  final Color gasSponsoredBadge;

  const ObsidianThemeTokens({
    required this.canvas,
    required this.surface,
    required this.surfaceElevated,
    required this.border,
    required this.bid,
    required this.bidSubtle,
    required this.ask,
    required this.askSubtle,
    required this.zeroFeeBadge,
    required this.gasSponsoredBadge,
  });

  static const dark = ObsidianThemeTokens(
    canvas: Color(0xFF0B0E14),
    surface: Color(0xFF141822),
    surfaceElevated: Color(0xFF1B2130),
    border: Color(0xFF1F2636),
    bid: Color(0xFF00F0A0),
    bidSubtle: Color(0x1F00F0A0),
    ask: Color(0xFFFF3B56),
    askSubtle: Color(0x1FFF3B56),
    zeroFeeBadge: Color(0xFF00F0A0),
    gasSponsoredBadge: Color(0xFF00E5FF),
  );

  @override
  ObsidianThemeTokens copyWith({
    Color? canvas,
    Color? surface,
    Color? surfaceElevated,
    Color? border,
    Color? bid,
    Color? bidSubtle,
    Color? ask,
    Color? askSubtle,
    Color? zeroFeeBadge,
    Color? gasSponsoredBadge,
  }) {
    return ObsidianThemeTokens(
      canvas: canvas ?? this.canvas,
      surface: surface ?? this.surface,
      surfaceElevated: surfaceElevated ?? this.surfaceElevated,
      border: border ?? this.border,
      bid: bid ?? this.bid,
      bidSubtle: bidSubtle ?? this.bidSubtle,
      ask: ask ?? this.ask,
      askSubtle: askSubtle ?? this.askSubtle,
      zeroFeeBadge: zeroFeeBadge ?? this.zeroFeeBadge,
      gasSponsoredBadge: gasSponsoredBadge ?? this.gasSponsoredBadge,
    );
  }

  @override
  ObsidianThemeTokens lerp(ThemeExtension<ObsidianThemeTokens>? other, double t) {
    if (other is! ObsidianThemeTokens) return this;
    return ObsidianThemeTokens(
      canvas: Color.lerp(canvas, other.canvas, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      surfaceElevated: Color.lerp(surfaceElevated, other.surfaceElevated, t)!,
      border: Color.lerp(border, other.border, t)!,
      bid: Color.lerp(bid, other.bid, t)!,
      bidSubtle: Color.lerp(bidSubtle, other.bidSubtle, t)!,
      ask: Color.lerp(ask, other.ask, t)!,
      askSubtle: Color.lerp(askSubtle, other.askSubtle, t)!,
      zeroFeeBadge: Color.lerp(zeroFeeBadge, other.zeroFeeBadge, t)!,
      gasSponsoredBadge: Color.lerp(gasSponsoredBadge, other.gasSponsoredBadge, t)!,
    );
  }
}
```

---

## 3. Typography & Tabular Numbers Standard (CLS = 0)

In financial trading applications, fluctuating price and quantity numerals must never cause the layout to vibrate or shift (Cumulative Layout Shift must equal exactly 0.00).

1. **Tabular Figures Mandate:**
   - On Web: `font-feature-settings: "tnum" 1, "zero" 1;` applied to all price displays, volume counts, and timestamps.
   - In Flutter: `fontFeatures: [FontFeature.tabularFigures(), FontFeature.slashedZero()]` applied to all `Text` and `RichText` widgets rendering market data.
2. **Typeface Pairing:**
   - **Financial Data:** `JetBrains Mono` or `Roboto Mono` with strict character cell dimensions.
   - **UI Labels & Chrome:** `Inter` on Web/Windows/Linux, `SF Pro Text` on macOS/iOS.

---

## 4. Component Visual Specifications

### 4.1 Level 2 Depth Bars
- Horizontal visual fill bars representing cumulative bid/ask volume across each price rung.
- Bid depth fill: Linear gradient from `rgba(0, 240, 160, 0.18)` at the left to `rgba(0, 240, 160, 0.04)` at the right edge.
- Ask depth fill: Linear gradient from `rgba(255, 59, 86, 0.18)` at the right to `rgba(255, 59, 86, 0.04)` at the left edge.
- Price update flash: 150ms ease-out flash highlighting microsecond price revisions (`#00F0A0` flash on uptick, `#FF3B56` flash on downtick).

### 4.2 DOM Price Ladder Rungs
- Static vertical price scale with fixed height of 24px per rung.
- Bid column: Left-aligned with subtle green hover border (`#00F0A0` 1px).
- Ask column: Right-aligned with subtle red hover border (`#FF3B56` 1px).
- Working order pills: Solid pill badges (`#00F0A0` for active buy, `#FF3B56` for active sell) showing order quantity and a right-click cancel trigger.

### 4.3 Strict 0.00% Zero-Fee & 0 Gas Badges
Across every order entry sheet, trade confirmation modal, and execution blotter:
- **Zero-Fee Badge:**
  - Background: `rgba(0, 240, 160, 0.12)`
  - Border: `1px solid rgba(0, 240, 160, 0.40)`
  - Text: `#00F0A0`, 11px uppercase bold monospaced: `0.00% ZERO-FEE (₹0.00 COMM)`
- **Gas Sponsored Badge:**
  - Background: `rgba(0, 229, 255, 0.12)`
  - Border: `1px solid rgba(0, 229, 255, 0.40)`
  - Text: `#00E5FF`, 11px uppercase bold monospaced: `[SPONSORED: ₹0.00 GAS]`

---

## 5. Accessibility & Ergonomics (WCAG 2.1 AAA)

1. **Contrast Ratios:**
   - Text on `#0B0E14`: Minimum contrast ratio of 7.0:1 for standard text and 4.5:1 for secondary numbers.
   - Green `#00F0A0` on `#0B0E14`: Contrast ratio of 12.4:1 (Exceeds WCAG AAA).
   - Red `#FF3B56` on `#0B0E14`: Contrast ratio of 5.8:1 (Exceeds WCAG AA, compliant for bold numerals).
2. **Reduced Motion Mode:**
   - When `@media (prefers-reduced-motion: reduce)` is detected, price tick flashes and spring animations disable instantly, falling back to immediate state switches.
3. **Screen Reader Support:**
   - Every financial ladder level includes semantic text descriptions (e.g. `aria-label="Bid price 92,450.00 USDT, volume 1.45 BTC"`).

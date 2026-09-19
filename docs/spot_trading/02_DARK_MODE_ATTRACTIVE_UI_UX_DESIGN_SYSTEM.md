# Dark Mode UI/UX Design System Specification
## Spot Trading Cockpit, Aesthetics, and Interactive Dynamics

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved Architecture Specification  
**Classification:** Proprietary / Trading Platform Experience  
**Target Environments:** Web (Desktop, High-DPI Displays), Mobile (iOS, Android Flutter), Electron Desktop Client  

---

## Executive Overview & Design Vision

The Growww Spot Trading Platform UI/UX design system provides an institutional-grade, visually magnetic, dark-mode-first environment modeled after premier global execution platforms (Binance, OKX, Bybit) while maintaining the clean ergonomics of top-tier consumer fintech applications (Zerodha Kite, Groww).

Modern high-velocity traders operate across multi-hour sessions where visual fatigue, layout instability, and cognitive latency directly impair execution quality. This specification establishes the foundations, token hierarchies, component anatomies, and micro-interaction mechanics required to deliver an ultra-responsive, addictive, and visually flawless trading terminal.

---

## 1. Visual Design Philosophy & Addictive UI Dynamics

### 1.1 The Core Aesthetic: Neo-Brutalist Precision Meets Cyber-Chic Ergonomics
The visual language balances two complementary design philosophies:
1. **Financial Neo-Brutality:** Crisp structural boundaries, razor-sharp 1px border gridlines, zero decorative skeuomorphic fluff, high information density, and strict geometrical alignment. Every pixel serves a data-driven purpose.
2. **Cyber-Chic Ambient Depth:** Layered elevations utilizing deep obsidian bases, subtle charcoal cards, selective radiant glows, and vivid phosphor accents (emerald green and crimson red) that guide peripheral attention without causing retina burn.

### 1.2 Addictive UI Mechanics and Dopamine Loops
High-engagement trading interfaces capitalize on clear, instant sensory reward loops:
- **Instantaneous Feedback Loop (sub-16ms response):** Every user action (tab click, slider drag, order submission) yields immediate optical and physical feedback. There is zero perceived latency between intention and interface acknowledgement.
- **Dynamic Liquidity Real-Time Pulses:** The order book and trade tape act as living, breathing streams. Subtle background depth washes pulse rhythmically as matching engine events arrive, immersing the trader in active market liquidity.
- **Variable Reward Affirmation:** Order execution is not merely logged as a table row; it is punctuated with tactile, acoustic, and luminous celebrations proportional to the fill state.
- **Sensory Mastery & Flow State:** Predictable hotkey ergonomics, zero layout displacement (Cumulative Layout Shift = 0), and unified spatial positioning allow veteran traders to enter a state of uninterrupted operational flow.

### 1.3 Progressive Disclosure & Cognitive Load Budgeting
Information density is calibrated in tiers:
- **Primary Peripheral Zone (Always Visible):** Current Mark Price, 24h Delta, Live Order Book Depth, Real-Time Chart, and Order Entry Pad.
- **Secondary On-Demand Zone (Single Interaction):** Open Orders, Working Trigger Orders, Real-Time Positions, Order History, and Execution Details.
- **Tertiary Deep-Dive Zone (Drawer/Modal):** Market Depth Charts, Order Book Settings, Advanced Algorithmic Slicing (TWAP/VWAP configs), and Audit Logs.

---

## 2. Dark Mode Palette Specification

The color architecture is built from ground-up for OLED and high-refresh-rate IPS displays. It completely eliminates pure white glare, balances ambient low-light contrast ratios according to WCAG 2.1 AAA specifications for data values, and implements strict chromatic semantics.

### 2.1 Core Neutral Palette & Surface Elevations

```
+-------------------+----------------+-----------------------------------------------------+
| Token Name        | Hex Code       | Semantic Role & Surface Elevation                  |
+-------------------+----------------+-----------------------------------------------------+
| $bg-obsidian-0    | #0B0E14        | Base root canvas, viewport backdrop, OLED black     |
| $surface-charcoal | #151922        | Elevation 1: Card containers, docked panels, tables |
| $surface-elevated | #1E232E        | Elevation 2: Input fields, sub-tabs, dropdown menus |
| $surface-overlay  | #282E3D        | Elevation 3: Hover states, active chips, dialogs    |
| $border-subtle    | #1E2430        | Internal 1px panel dividers and table row borders   |
| $border-contrast  | #2B313F        | Outer window frames, interactive control boundaries |
| $border-focus     | #5C6BC0        | Active focus ring, input highlight, selection frame |
+-------------------+----------------+-----------------------------------------------------+
```

### 2.2 Chromatic Semantic Accents

```
+-------------------+----------------+-----------------------------------------------------+
| Token Name        | Hex Code       | Semantic Role & Interaction Behavior                |
+-------------------+----------------+-----------------------------------------------------+
| $bull-emerald     | #00C087        | Buy side, Ask matching, positive PnL, bullish bars  |
| $bull-glow        | rgba(0,192,135)| 12% to 20% alpha background fill for bid depth bars |
| $bear-crimson     | #FF3B30        | Sell side, Bid matching, negative PnL, bearish bars |
| $bear-glow        | rgba(255,59,48)| 12% to 20% alpha background fill for ask depth bars |
| $demo-amber       | #F5A623        | Simulation badge, paper trade ribbons, mock notices |
| $demo-amber-bg    | rgba(245,166,35| 10% alpha ambient header wash for demo cockpit      |
| $system-indigo    | #5C6BC0        | Primary CTA interactive accent, active navigation   |
| $system-cyan      | #00E5FF        | Algorithmic triggers, websocket heartbeat pulse     |
+-------------------+----------------+-----------------------------------------------------+
```

### 2.3 Typography & Content Contrast Tiers

```
+-------------------+----------------+------------+----------------------------------------+
| Token Name        | Hex Code       | Opacity    | Semantic Application                   |
+-------------------+----------------+------------+----------------------------------------+
| $text-primary     | #EAECEF        | 100%       | Numerical prices, quantities, balances |
| $text-secondary   | #B7BDC6        | 80%        | Column headers, active tab labels      |
| $text-muted       | #848E9C        | 55%        | Units, timestamps, inactive states     |
| $text-disabled    | #474D57        | 35%        | Disabled inputs, offline indicators    |
+-------------------+----------------+------------+----------------------------------------+
```

### 2.4 Contrast Ratio Matrix & Accessibility Conformance
All text and data elements meet or exceed rigorous contrast requirements against their respective backgrounds:
- `$text-primary` (#EAECEF) on `$bg-obsidian-0` (#0B0E14): **15.2:1** (Exceeds WCAG AAA 7:1 limit)
- `$bull-emerald` (#00C087) on `$surface-charcoal` (#151922): **8.4:1** (Exceeds WCAG AAA)
- `$bear-crimson` (#FF3B30) on `$surface-charcoal` (#151922): **5.9:1** (Exceeds WCAG AA 4.5:1)
- `$demo-amber` (#F5A623) on `$bg-obsidian-0` (#0B0E14): **9.8:1** (Exceeds WCAG AAA)

---

## 3. Monospaced Tabular Financial Typography

Visual stability is paramount in high-frequency trading. When tick updates arrive every 50 to 100 milliseconds, proportional fonts cause numerical values to shift horizontally (jitter), introducing severe eye strain and user misclicks.

### 3.1 Font Stack Selection
- **Primary Numerical Engine:** `JetBrains Mono`, `Roboto Mono`, `SF Mono`, `monospace`
- **Primary Interface / Label Engine:** `Inter`, `SF Pro Display`, `-apple-system`, `sans-serif`

### 3.2 CSS OpenType Numeric Configuration
Every financial numeric container enforces tabular lining figures:

```css
.tabular-financial-nums {
  font-family: 'JetBrains Mono', 'Roboto Mono', monospace;
  font-feature-settings: 'tnum' 1, 'lnum' 1, 'zero' 1;
  font-variant-numeric: tabular-nums lining-nums slashed-zero;
  letter-spacing: -0.02em;
  text-align: right;
  white-space: nowrap;
}
```

### 3.3 Zero-Layout-Shift (CLS = 0) Grid Architecture
1. **Fixed Column Dimensions:** All order book, trade tape, and ledger columns utilize explicit CSS grid or flex basis dimensions (e.g., `width: 96px; min-width: 96px; max-width: 96px`).
2. **Right-Hand Decimal Alignment:** Numerical strings are padded and aligned to the right, ensuring integer and fractional digits line up on vertical axes.
3. **Muted Trailing Zeros:** Trailing zeros representing unfilled precision slots are styled using `$text-muted` (55% opacity), maintaining visual width while drawing focus to significant digits:
   - Display representation: `64,250.8000` -> `64,250.8` (bright) followed by `000` (muted).

### 3.4 Typographic Scale Hierarchy

```
+---------------+-----------+-------------+------------+-----------------------------------+
| Style Key     | Size (px) | Line Height | Weight     | Context / Usage                   |
+---------------+-----------+-------------+------------+-----------------------------------+
| Display Large | 28px      | 34px        | Bold 700   | Header Last Traded Price (LTP)    |
| Heading 1     | 18px      | 24px        | SemiBold   | Cockpit module titles, modals     |
| Heading 2     | 14px      | 20px        | Medium 500 | Panel sub-headers, widget tabs    |
| Body Data     | 13px      | 18px        | Medium 500 | Level-2 rows, tape prices, orders |
| Caption Label | 11px      | 14px        | Normal 400 | Form field tags, table headers    |
| Micro Metric  | 10px      | 12px        | SemiBold   | Status badges, fee tier markers   |
+---------------+-----------+-------------+------------+-----------------------------------+
```

---

## 4. Screen Anatomy: Spot Trading Cockpit Architecture

The trading cockpit layout employs an interconnected modular dock system that scales seamlessly from ultra-wide multi-monitor setups down to compact tablet and smartphone viewports.

### 4.1 Desktop Grid Dock Layout (4-Column Institutional Split)

```
+---------------------------------------------------------------------------------------------------------+
| TOP HEADER: Symbol Selector | 24h Tickers | Engine Status | Latency | Demo/Live Switcher | User Profile  |
+------------------------------------+------------------------------------+-------------------------------+
| LEFT PANEL (20% W)                 | CENTER STAGE (55% W)               | RIGHT PANEL (25% W)           |
|                                    |                                    |                               |
| [ LEVEL-2 ORDER BOOK ]             | [ TRADINGVIEW ADVANCED CHART ]     | [ ORDER ENTRY SHEET ]         |
| Asks (Crimson Depth Ladder)        | Candlesticks, Multi-Timeframe bar, | Buy (Emerald) / Sell (Crimson)|
| Spread Display & BPS Tracker       | Technical Indicators, Drawing Dock | Limit / Market / Stop / OCO   |
| Bids (Emerald Depth Ladder)        | ---------------------------------- | Price, Size, % Sliders        |
| ---------------------------------- | [ WORKING CAPITAL & TICKER STRIP ] | Estimated Fees, STT, TDS Calc |
| [ LIVE RECENT TRADES TAPE ]        | 24h High/Low, Vol, Index Mark      | Execute CTA Button            |
| Price | Size | Millisecond Time    |                                    | Post-Trade Balance Preview    |
+------------------------------------+------------------------------------+-------------------------------+
| BOTTOM CONSOLE (100% W, Collapsible 240px Height)                                                       |
| Tabs: Open Orders (4) | Trigger Orders (1) | Order History | Trade Fills | Asset Holdings | API Diagnostics |
+---------------------------------------------------------------------------------------------------------+
```

### 4.2 Module 1: Level-2 Order Book Depth Ladder
The Order Book renders the continuous matching engine state with deterministic visual clarity:
- **Dual Visual Modes:**
  - Traditional Split (Top Asks in Crimson, Bottom Bids in Emerald, Spread in center).
  - Single Cumulative Ladder (Continuous vertical view with zero-point spread).
- **Dynamic Relative Depth Bars:** Behind each numeric price/size row, an animated horizontal bar illustrates cumulative volume:
  - Bids: Horizontal fill anchored from right to left using `$bull-glow` (`rgba(0, 192, 135, 0.16)`).
  - Asks: Horizontal fill anchored from right to left using `$bear-glow` (`rgba(255, 59, 48, 0.16)`).
- **Spread & Basis Points Telemetry:** The central dividing bar dynamically shows:
  - `Spread: 0.50 USDT (0.78 bps)` accompanied by real-time directional chevrons indicating micro-momentum.
- **Tick Size Grouping Selector:** Interactive drop-down enabling instant price aggregation: `0.01`, `0.1`, `1.0`, `10.0`.
- **My Orders Alignment Marker:** Small glowing geometric pip markers on the outer left gutter indicating user limit orders resting inside the order book queue.

### 4.3 Module 2: Live Recent Trades Stream (The Tape)
- High-velocity tape displaying the last 50 fills directly from Kafka matching-engine consumer threads.
- Column structure: `Price (USDT) | Amount (BTC) | Time (HH:mm:ss.SSS)`.
- Flash highlight dynamics:
  - Buyer-maker (Sell fill against resting bid): Text rendered in `$bear-crimson`, brief 150ms subtle glow.
  - Seller-maker (Buy fill against resting ask): Text rendered in `$bull-emerald`, brief 150ms subtle glow.
- Large volume whale icon indicator for transactions exceeding 99th percentile size.

### 4.4 Module 3: TradingView Advanced Chart Integration
- Custom Dark Theme synchronization strictly bound to platform palette tokens:
  - Canvas Background: `$bg-obsidian-0` (`#0B0E14`).
  - Gridlines: Thin 1px `$border-subtle` (`#1E2430`) with 20% opacity.
  - Bullish Candlestick Body: `$bull-emerald` (`#00C087`), Wick: `#00C087`, Border: `#00C087`.
  - Bearish Candlestick Body: `$bear-crimson` (`#FF3B30`), Wick: `#FF3B30`, Border: `#FF3B30`.
- **Top Ribbon Controls:** Timeframe toggles (`1s`, `1m`, `5m`, `15m`, `1h`, `4h`, `1D`, `1W`), Candlestick style (Candles, Heikin Ashi, Line, Footprint), Indicator Drawer button.
- **Interactive Working Order Overlays:** Pending limit orders and stop triggers appear directly on the chart canvas as interactive draggable horizontal lines with quick cancel [x] action handles.

### 4.5 Module 4: Buy / Sell Order Entry Sheet
- **Directional Segmented Switcher:** 
  - Massive dual-segmented toggle at the top of the ticket: `BUY` (Emerald) vs `SELL` (Crimson).
  - Activating `BUY` floods the active tab indicator with `$bull-emerald` and sets the submit CTA to `$bull-emerald`.
  - Activating `SELL` floods the active tab indicator with `$bear-crimson` and sets the submit CTA to `$bear-crimson`.
- **Order Type Ribbon:** `Limit`, `Market`, `Stop-Limit`, `Trailing Stop`, `OCO`.
- **Input Fields Anatomy:**
  - Container: `$surface-elevated` (`#1E232E`) with 1px `$border-contrast` (`#2B313F`) border.
  - Label pinned left in `$text-secondary`; numerical value entered right in `$text-primary`.
  - Interactive "BEST BID" / "BEST ASK" quick-fill shortcut buttons.
- **Tactile Proportion Allocation Slider & Quick Chips:**
  - Slider bar with magnetic notches at `25%`, `50%`, `75%`, `100%`.
  - Discrete percentage pill chips beneath the slider for instant one-click allocation.
- **Statutory & Fee Transparency Box:**
  - Real-time pre-trade breakdown:
    - Maker/Taker Exchange Fee: `0.0750%`
    - Statutory STT / TDS (Section 194S 1% compliance deduction for Indian users): Computed real-time.
    - Total Margin / Capital Required.
    - Post-Execution Balance Estimate.
- **Double-Tap Safety Toggle:** Optional setting requiring double confirmation click for orders exceeding specified risk thresholds.

---

## 5. Mode Switcher Demarcation: Demo Mode vs Real Mode

To guarantee regulatory compliance, prevent costly trader mistakes, and deliver crystal-clear platform awareness, the interface features unmissable visual demarcation between Real Capital Trading and Simulated Paper Trading.

### 5.1 High-Contrast Persistent Demo Mode Banner
When Demo Mode is engaged:
1. **Viewport Top Ambient Header Banner:** A persistent 32px height banner pinned to the very top edge:
   - Background: Striped amber hazard pattern or solid `$demo-amber` (`#F5A623`) with high-contrast obsidian typography.
   - Text: `[ DEMO SIMULATION MODE - NO REAL CAPITAL AT RISK - MOCK SETTLEMENT ACTIVE ]`
2. **Chart Watermark:** A subtle 8% opacity SVG watermark centered behind the TradingView candles: `DEMO ACCOUNT - VIRTUAL FUNDS`.
3. **Wallet Balance Badge:** Capital displays are framed in `$demo-amber` borders with the clear label `VIRTUAL USDT`.
4. **Order Button Accentuation:** Primary CTA switches from traditional Emerald/Crimson to Cyber Amber with text: `SUBMIT SIMULATED ORDER`.
5. **Faucet Quick-Reset Tool:** A dedicated instant-action button in the header: `Reset Paper Capital ($100,000 USDT)`.

### 5.2 Real Mode Stealth Protocol
When Real Mode is active:
1. **Status Ribbon:** The ambient amber ribbon disappears entirely; the header collapses to standard 56px height.
2. **Institutional "LIVE MARKET" Beacon:** A small glowing emerald indicator (pulsing green dot `#00C087`) in the top navigation bar accompanied by `LIVE MARKET - AUDITED CUSTODY`.
3. **Pristine Obsidian Canvas:** Full visual clarity with zero watermarks or distracting warning bars.

### 5.3 Mode Switching Interlock Modal
Transitioning between Demo and Real requires a safe two-step handshake:
- Trader toggles the switch in the top header.
- An elevation-3 modal appears requiring explicit confirmation:
  - If entering Real Mode: "You are entering Real Trading with Live Capital. All executed orders will settle against real liquidity." [CONFIRM TO GO LIVE].
  - Prevents accidental execution crossovers.

---

## 6. Micro-Interactions, Haptics, and Celebratory Feedback

The trading experience achieves tactile addiction through hyper-responsive visual physics and sensory micro-interactions.

### 6.1 Real-Time Price Flash Mechanics
Every inbound WebSocket message triggers subtle, performant CSS keyframe animations:
- **Price Uptick:** 
  - Text color transitions to `$bull-emerald` instantly.
  - Background container flashes `rgba(0, 192, 135, 0.22)` and smoothly decays back to transparent over 350 milliseconds (`cubic-bezier(0.25, 1, 0.5, 1)`).
- **Price Downtick:** 
  - Text color transitions to `$bear-crimson` instantly.
  - Background container flashes `rgba(255, 59, 48, 0.22)` and smoothly decays back to transparent over 350 milliseconds.

### 6.2 Celebratory Trade Fill Feedback (Audio & Haptics)
Upon receiving a WebSocket execution report (`ORDER_FILLED`):
1. **Mobile Taptic Engine Mapping (iOS / Android):**
   - **Order Placed / Resting:** Light impact haptic tick (`UIImpactFeedbackGenerator(style: .light)`).
   - **Partial Fill:** Double light tick in rapid succession (80ms spacing).
   - **Complete (100%) Fill:** Crisp rigid impact (`UIImpactFeedbackGenerator(style: .rigid)`) followed by a micro-success celebration pattern.
2. **Visual Fill Radiance:**
   - A non-disruptive, localized micro-particle spark or expanding ring radiates from the status badge in the bottom console.
   - The executed trade row in the "Trade Fills" tab flashes an emerald glow for 800 milliseconds.
3. **Acoustic Sound Bites (User Toggleable in Settings):**
   - *Placement Click:* High-frequency mechanical shutter sound (12ms duration, crisp pitch).
   - *Execution Chime:* Subtle harmonic chime (880Hz soft bell decay).
   - *Cancellation Tone:* Low-frequency damped thud.

### 6.3 Interactive Slider Physics
- Dragging the percentage allocation slider provides magnetic snap resistance at exactly `25%`, `50%`, `75%`, and `100%`.
- On mobile touch devices, crossing each milestone triggers a sharp transient haptic tick.

---

## 7. Complete Settings, Profile, Security, and Navigation Hierarchy

The terminal framework organizes complex multi-asset capabilities into an intuitive, zero-clutter navigation structure.

### 7.1 Global Header Anatomy (Desktop)
```
+---------------------------------------------------------------------------------------------------------+
| [GROWWW SPOT] | Markets | Trade v | Derivatives | Earn | Portfolio | Search (Ctrl+K) | [DEMO/REAL] | [USER] |
+---------------------------------------------------------------------------------------------------------+
```
- **Brand & Primary Navigation (Left):** Direct access to core product verticals.
- **Global Command Palette (Center):** Universal search trigger (`Ctrl + K` or `Cmd + K`) opening an instant-find palette for pairs (e.g., `BTC/USDT`), navigation routes, and hotkey lookups.
- **Mode Switcher & Utility Dock (Right):**
  - High-visibility `DEMO / REAL` toggle switch.
  - WebSocket Latency Telemetry: `24ms` with green status dot.
  - Notification Bell (with unread badge count).
  - User Profile & Security Shield Avatar.

### 7.2 Bottom Navigation Hierarchy (Mobile App Form Factor)
For mobile devices, the interface condenses into a high-ergonomics thumb zone navigation dock:
1. **Home / Overview:** Market movers, personal balance summary, quick deposit.
2. **Markets:** Watchlists, sector heatmaps, top gainers/losers/volume.
3. **Trade (Cockpit):** The dedicated spot trading terminal detailed in Section 4.
4. **Futures / Derivatives:** Leverage trading terminal.
5. **Wallets / Assets:** Double-entry ledger balances, deposit/withdrawal flows.

### 7.3 Settings Drawer & Customization Suite
Accessible via user preferences menu:
- **Display Themes:**
  - *Obsidian Dark (Default):* Deep #0B0E14 base optimized for OLED.
  - *Charcoal Night:* Slate #151922 base for lower-contrast environments.
  - *Color Vision Accessibility Mode:* Replaces Green/Red with Blue/Orange color-blind friendly palette.
- **Trading Ergonomics:**
  - Order Confirmation Dialogs: Toggle on/off for Market, Limit, and Stop orders.
  - One-Click Execution Mode: Instant submission without confirmation modals.
  - Audio & Haptic Feedback: Discrete toggles for placement sounds, fill chimes, and haptic intensity.
- **Dock Layout Customization:** Save and restore workspace layouts (e.g., "Full Chart View", "Order Book Scalper View", "Multi-Window Tape").

### 7.4 Security Hub & Account Governance
- **Security Health Score:** Dynamic ring progress meter (e.g., `95% - Excellent`).
- **Two-Factor Authentication (2FA) Management:**
  - Hardware Security Keys: WebAuthn / FIDO2 (YubiKey) support.
  - Time-based One-Time Password (TOTP): Authenticator app link status.
  - SMS / Biometric Passkeys: Fingerprint and Face Unlock credentials.
- **Anti-Phishing Code:** Custom user-defined secret string rendered inside all genuine system emails and login notification popups.
- **Session & Device Audit Grid:** Real-time table showing active sessions, IP addresses, geolocation flags, and one-click "Terminate All Other Sessions" emergency revoke button.
- **API Key Management:** Cryptographic key generation, IP address whitelisting, and strict permission scopes (Read-Only, Trade-Enabled, Withdrawal-Disabled).

### 7.5 User Profile & Tiered Verification Architecture
- **Identity Tier Badge:** 
  - Tier 1: Basic Email / Phone (Read-only / Simulated mode).
  - Tier 2: Full National Identity KYC (Aadhaar / PAN verification, Daily limits unlocked).
  - Tier 3: Institutional Enterprise (Corporate entity verification, FIX API access enabled).
- **Daily Withdrawal Limit Progress Bar:** Graphical capacity indicator showing remaining 24-hour liquidity limits in fiat and crypto terms.

---

## 8. Summary Checklist for Engineering Implementation

```
[ ] Palette Hex & Alpha Tokens registered in global CSS variables / Flutter theme constants.
[ ] OpenType Tabular Figures (tnum, lnum) applied to all numeric financial components.
[ ] Fixed column widths enforced to achieve Cumulative Layout Shift of zero (CLS = 0).
[ ] Order book relative depth background bars rendering with correct opacity gradients.
[ ] TradingView canvas theme overridden to match #0B0E14 background and platform colors.
[ ] Persistent high-contrast amber banner and watermark active whenever Demo Mode is engaged.
[ ] Haptic feedback and subtle tick flash animations integrated into WebSocket event handlers.
[ ] Keyboard shortcut manager active with Cmd+K command palette and instant hotkey routing.
```

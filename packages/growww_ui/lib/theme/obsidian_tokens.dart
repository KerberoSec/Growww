import 'package:flutter/material.dart';

/// Obsidian High-Contrast Dark Mode Visual Tokens for Growww / NBSE Pro Workstation
class ObsidianTokens {
  // Background Hierarchy
  static const Color backgroundVoid = Color(0xFF08090C); // Pitch black backdrop
  static const Color backgroundPanel = Color(0xFF12141A); // Secondary cards/panels
  static const Color backgroundCard = Color(0xFF1A1D24); // Elevated drawers & popovers
  static const Color backgroundElevated = Color(0xFF222630); // Hover & active states

  // Market Direction Colors (High-Contrast Financial Ergonomics)
  static const Color greenBull = Color(0xFF00E676); // Neon Green for Asks / Buy / Up Ticks
  static const Color redBear = Color(0xFFFF1744); // Crimson Red for Bids / Sell / Down Ticks
  static const Color greenSubdued = Color(0xFF1B382B);
  static const Color redSubdued = Color(0xFF381B22);

  // Accent & Brand Tokens
  static const Color goldAccent = Color(0xFFF0B90B); // Binance-grade Gold for Warnings & VIP
  static const Color cyanAccent = Color(0xFF00E5FF); // Cyberpunk Cyan for Demo Sandbox
  static const Color purpleSecurity = Color(0xFF7C4DFF); // Passkey & Biometric indicators

  // Typography Palette
  static const Color textPrimary = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0xFF9096A2);
  static const Color textTertiary = Color(0xFF5A606D);
  static const Color textDisabled = Color(0xFF383D48);

  // Border & Divider Tokens
  static const Color borderSubtle = Color(0xFF232732);
  static const Color borderFocused = Color(0xFF3A4152);

  // Monospaced Typography Style for High-Frequency Price Ticks
  static const TextStyle tabularTick = TextStyle(
    fontFamily: 'JetBrainsMono',
    fontFeatures: [FontFeature.tabularFigures()],
    fontSize: 13.0,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.2,
  );

  static const TextStyle headerPro = TextStyle(
    fontFamily: 'Inter',
    fontSize: 16.0,
    fontWeight: FontWeight.w700,
    color: textPrimary,
  );
}

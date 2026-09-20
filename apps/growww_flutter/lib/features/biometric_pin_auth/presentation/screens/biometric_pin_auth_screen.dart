import 'package:flutter/material.dart';
import '../../domain/models/pin_auth_enums.dart';
import '../controllers/biometric_pin_auth_controller.dart';

/// Flutter Biometric Authentication & 6-Digit PIN Screen
/// Institutional-grade cyber-aesthetic security pad with FaceID, Fingerprint,
/// animated dot indicators, custom dial pad, and brute-force throttling lockout.
class BiometricPinAuthScreen extends StatefulWidget {
  final BiometricPinAuthController controller;
  final VoidCallback? onAuthenticated;

  const BiometricPinAuthScreen({
    Key? key,
    required this.controller,
    this.onAuthenticated,
  }) : super(key: key);

  @override
  State<BiometricPinAuthScreen> createState() => _BiometricPinAuthScreenState();
}

class _BiometricPinAuthScreenState extends State<BiometricPinAuthScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonCyan = Color(0xFF00E5FF);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<BiometricPinAuthState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;

        if (state.isAuthenticated && widget.onAuthenticated != null) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            widget.onAuthenticated!();
          });
        }

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            leading: IconButton(
              icon: const Icon(Icons.arrow_back_ios, color: Colors.white, size: 18),
              onPressed: () => Navigator.of(context).maybePop(),
            ),
            title: const Text(
              'SOVEREIGN BIOMETRIC AUTH',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.2,
              ),
            ),
          ),
          body: SafeArea(
            child: Column(
              children: [
                const Spacer(flex: 1),
                _buildHeader(state),
                const SizedBox(height: 28),
                _buildPinDots(state),
                const SizedBox(height: 20),
                if (state.errorMessage != null)
                  _buildErrorAlert(state.errorMessage!)
                else if (state.isLocked)
                  _buildLockoutBanner(state.lockoutRemainingSeconds),
                const Spacer(flex: 2),
                _buildNumericKeypad(state),
                const SizedBox(height: 24),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildHeader(BiometricPinAuthState state) {
    return Column(
      children: [
        Container(
          width: 60,
          height: 60,
          decoration: BoxDecoration(
            color: surfaceCard,
            shape: BoxShape.circle,
            border: Border.all(
              color: state.isLocked ? neonRed : (state.isAuthenticated ? neonGreen : neonCyan),
              width: 1.5,
            ),
          ),
          child: Icon(
            state.isLocked
                ? Icons.lock_clock
                : (state.isAuthenticated ? Icons.verified_user : Icons.security),
            color: state.isLocked ? neonRed : (state.isAuthenticated ? neonGreen : neonCyan),
            size: 28,
          ),
        ),
        const SizedBox(height: 16),
        Text(
          state.authPurpose.title,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 6),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Text(
            state.authPurpose.subtitle,
            textAlign: TextAlign.center,
            style: const TextStyle(color: textMuted, fontSize: 13),
          ),
        ),
      ],
    );
  }

  Widget _buildPinDots(BiometricPinAuthState state) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(state.pinLength, (index) {
        final isFilled = index < state.filledDots;
        final isErr = state.entryState == PinEntryState.error;
        final isLocked = state.isLocked;

        Color dotColor;
        if (isLocked || isErr) {
          dotColor = neonRed;
        } else if (isFilled) {
          dotColor = neonGreen;
        } else {
          dotColor = Colors.transparent;
        }

        return Container(
          margin: const EdgeInsets.symmetric(horizontal: 8),
          width: 16,
          height: 16,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: dotColor,
            border: Border.all(
              color: isLocked || isErr
                  ? neonRed
                  : (isFilled ? neonGreen : Colors.white30),
              width: 2,
            ),
            boxShadow: isFilled && !isErr && !isLocked
                ? [
                    BoxShadow(
                      color: neonGreen.withOpacity(0.5),
                      blurRadius: 8,
                      spreadRadius: 1,
                    )
                  ]
                : null,
          ),
        );
      }),
    );
  }

  Widget _buildErrorAlert(String message) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 28),
      child: Text(
        message,
        textAlign: TextAlign.center,
        style: const TextStyle(color: neonRed, fontSize: 12, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _buildLockoutBanner(int remainingSec) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 24),
      padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
      decoration: BoxDecoration(
        color: neonRed.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: neonRed),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.timer_outlined, color: neonRed, size: 16),
          const SizedBox(width: 8),
          Text(
            'Brute-force throttle active: Retry in ${remainingSec}s',
            style: const TextStyle(color: neonRed, fontSize: 12, fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }

  Widget _buildNumericKeypad(BiometricPinAuthState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 36),
      child: Column(
        children: [
          _buildKeypadRow([1, 2, 3], state),
          const SizedBox(height: 16),
          _buildKeypadRow([4, 5, 6], state),
          const SizedBox(height: 16),
          _buildKeypadRow([7, 8, 9], state),
          const SizedBox(height: 16),
          _buildKeypadBottomRow(state),
        ],
      ),
    );
  }

  Widget _buildKeypadRow(List<int> digits, BiometricPinAuthState state) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: digits.map((digit) {
        return _buildKeypadButton(
          label: digit.toString(),
          onTap: state.isLocked ? null : () => widget.controller.pressDigit(digit),
        );
      }).toList(),
    );
  }

  Widget _buildKeypadBottomRow(BiometricPinAuthState state) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        // Biometric Trigger Button (FaceID / Fingerprint)
        _buildActionKey(
          icon: state.biometricType == BiometricPromptType.faceId
              ? Icons.face
              : Icons.fingerprint,
          color: neonCyan,
          onTap: state.isLocked ? null : () => widget.controller.triggerBiometricAuth(),
        ),
        // Digit 0
        _buildKeypadButton(
          label: '0',
          onTap: state.isLocked ? null : () => widget.controller.pressDigit(0),
        ),
        // Backspace Button
        _buildActionKey(
          icon: Icons.backspace_outlined,
          color: textMuted,
          onTap: state.isLocked ? null : () => widget.controller.pressBackspace(),
        ),
      ],
    );
  }

  Widget _buildKeypadButton({required String label, VoidCallback? onTap}) {
    return SizedBox(
      width: 68,
      height: 68,
      child: Material(
        color: surfaceCard,
        shape: const CircleBorder(),
        child: InkWell(
          customBorder: const CircleBorder(),
          splashColor: neonGreen.withOpacity(0.2),
          highlightColor: neonGreen.withOpacity(0.1),
          onTap: onTap,
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: onTap != null ? Colors.white : Colors.white24,
                fontSize: 22,
                fontFamily: 'JetBrains Mono',
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildActionKey({
    required IconData icon,
    required Color color,
    VoidCallback? onTap,
  }) {
    return SizedBox(
      width: 68,
      height: 68,
      child: Material(
        color: Colors.transparent,
        shape: const CircleBorder(),
        child: InkWell(
          customBorder: const CircleBorder(),
          splashColor: color.withOpacity(0.2),
          onTap: onTap,
          child: Center(
            child: Icon(
              icon,
              color: onTap != null ? color : color.withOpacity(0.3),
              size: 26,
            ),
          ),
        ),
      ),
    );
  }
}

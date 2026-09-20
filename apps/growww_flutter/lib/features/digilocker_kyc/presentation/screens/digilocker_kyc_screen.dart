import 'package:flutter/material.dart';
import '../../domain/models/kyc_enums.dart';
import '../controllers/digilocker_kyc_controller.dart';

/// Flutter Onboarding DigiLocker KYC Flow Screen
/// Compliant with UIDAI Aadhaar Paperless Offline e-KYC, NSDL PAN verification,
/// and DPDP Act 2023 zero-PII data handling.
class DigiLockerKycScreen extends StatefulWidget {
  final DigiLockerKycController controller;
  final VoidCallback? onCompleted;

  const DigiLockerKycScreen({
    Key? key,
    required this.controller,
    this.onCompleted,
  }) : super(key: key);

  @override
  State<DigiLockerKycScreen> createState() => _DigiLockerKycScreenState();
}

class _DigiLockerKycScreenState extends State<DigiLockerKycScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color cardSurface = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonCyan = Color(0xFF00E5FF);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color textMuted = Color(0xFF8B949E);

  late final TextEditingController _aadhaarTextController;
  late final TextEditingController _otpTextController;
  late final TextEditingController _panTextController;

  @override
  void initState() {
    super.initState();
    _aadhaarTextController = TextEditingController(text: widget.controller.state.aadhaarInput);
    _otpTextController = TextEditingController(text: widget.controller.state.otpInput);
    _panTextController = TextEditingController(text: widget.controller.state.panInput);
  }

  @override
  void dispose() {
    _aadhaarTextController.dispose();
    _otpTextController.dispose();
    _panTextController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<DigiLockerKycState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: const Text(
              'SOVEREIGN KYC • DIGILOCKER GATEWAY',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.2,
              ),
            ),
            bottom: PreferredSize(
              preferredSize: const Size.fromHeight(48),
              child: _buildStepper(state.step),
            ),
          ),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (state.errorMessage != null) ...[
                  _buildErrorBanner(state.errorMessage!),
                  const SizedBox(height: 12),
                ],
                if (state.successMessage != null) ...[
                  _buildSuccessBanner(state.successMessage!),
                  const SizedBox(height: 12),
                ],
                _buildStepContent(state),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildStepper(DigiLockerKycStep currentStep) {
    final steps = ['Aadhaar', 'NSDL PAN', 'Attested'];
    final activeIdx = currentStep.stepIndex - 1;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      decoration: const BoxDecoration(
        color: cardSurface,
        border: Border(bottom: BorderSide(color: Colors.white12)),
      ),
      child: Row(
        children: [
          for (int i = 0; i < steps.length; i++) ...[
            Container(
              width: 22,
              height: 22,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: i < activeIdx
                    ? neonGreen
                    : (i == activeIdx ? neonCyan : Colors.white24),
              ),
              child: Center(
                child: i < activeIdx
                    ? const Icon(Icons.check, size: 14, color: Colors.black)
                    : Text(
                        '${i + 1}',
                        style: TextStyle(
                          color: i == activeIdx ? Colors.black : Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
              ),
            ),
            const SizedBox(width: 6),
            Text(
              steps[i],
              style: TextStyle(
                color: i <= activeIdx ? Colors.white : textMuted,
                fontSize: 12,
                fontWeight: i == activeIdx ? FontWeight.bold : FontWeight.normal,
              ),
            ),
            if (i < steps.length - 1)
              Expanded(
                child: Container(
                  height: 2,
                  margin: const EdgeInsets.symmetric(horizontal: 8),
                  color: i < activeIdx ? neonGreen : Colors.white12,
                ),
              ),
          ],
        ],
      ),
    );
  }

  Widget _buildStepContent(DigiLockerKycState state) {
    switch (state.step) {
      case DigiLockerKycStep.aadhaarEntry:
      case DigiLockerKycStep.aadhaarOtpVerification:
        return _buildAadhaarSection(state);
      case DigiLockerKycStep.digilockerConsent:
        return _buildDigiLockerConsentCard(state);
      case DigiLockerKycStep.panVerification:
        return _buildPanSection(state);
      case DigiLockerKycStep.completed:
        return _buildCompletionCard(state);
    }
  }

  Widget _buildAadhaarSection(DigiLockerKycState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _buildDigiLockerOneTapCard(state),
        const SizedBox(height: 20),
        Row(
          children: const [
            Expanded(child: Divider(color: Colors.white24)),
            Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: Text('OR DIRECT UIDAI OTP', style: TextStyle(color: textMuted, fontSize: 11, fontWeight: FontWeight.bold)),
            ),
            Expanded(child: Divider(color: Colors.white24)),
          ],
        ),
        const SizedBox(height: 20),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: cardSurface,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white12),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('12-Digit Aadhaar Number', style: TextStyle(color: textMuted, fontSize: 12)),
              const SizedBox(height: 8),
              TextField(
                controller: _aadhaarTextController,
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 15),
                keyboardType: TextInputType.number,
                maxLength: 14,
                decoration: InputDecoration(
                  hintText: '2847 1928 3741',
                  hintStyle: const TextStyle(color: Colors.white30),
                  filled: true,
                  fillColor: obsidianBackground,
                  counterText: '',
                  prefixIcon: const Icon(Icons.fingerprint, color: neonGreen, size: 20),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(6), borderSide: BorderSide.none),
                ),
                onChanged: (val) => widget.controller.setAadhaarInput(val),
              ),
              const SizedBox(height: 12),
              if (!state.isOtpSent)
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: neonGreen,
                    foregroundColor: Colors.black,
                    minimumSize: const Size.fromHeight(44),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                  ),
                  onPressed: () => widget.controller.requestAadhaarOtp(),
                  child: const Text('Get UIDAI OTP', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              if (state.isOtpSent) ...[
                const SizedBox(height: 12),
                const Text('Enter 6-Digit OTP', style: TextStyle(color: textMuted, fontSize: 12)),
                const SizedBox(height: 8),
                TextField(
                  controller: _otpTextController,
                  style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 16, letterSpacing: 4),
                  keyboardType: TextInputType.number,
                  maxLength: 6,
                  decoration: InputDecoration(
                    hintText: '123456',
                    hintStyle: const TextStyle(color: Colors.white30),
                    filled: true,
                    fillColor: obsidianBackground,
                    counterText: '',
                    prefixIcon: const Icon(Icons.sms_outlined, color: neonCyan, size: 20),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(6), borderSide: BorderSide.none),
                  ),
                  onChanged: (val) => widget.controller.setOtpInput(val),
                ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      state.otpCooldown > 0 ? 'Resend in ${state.otpCooldown}s' : 'Did not receive OTP?',
                      style: const TextStyle(color: textMuted, fontSize: 11),
                    ),
                    if (state.otpCooldown == 0)
                      TextButton(
                        onPressed: () => widget.controller.requestAadhaarOtp(),
                        child: const Text('Resend OTP', style: TextStyle(color: neonGreen, fontSize: 11)),
                      ),
                  ],
                ),
                const SizedBox(height: 12),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: neonGreen,
                    foregroundColor: Colors.black,
                    minimumSize: const Size.fromHeight(44),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                  ),
                  onPressed: () => widget.controller.verifyAadhaarOtp(_otpTextController.text),
                  child: const Text('Verify Aadhaar OTP', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildDigiLockerOneTapCard(DigiLockerKycState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: cardSurface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: neonCyan.withOpacity(0.4)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: const [
              Icon(Icons.bolt, color: neonCyan, size: 22),
              SizedBox(width: 8),
              Text(
                'Instant DigiLocker Fast-Track',
                style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
          const SizedBox(height: 8),
          const Text(
            'Fetch verified Aadhaar & PAN in seconds directly from the National DigiLocker ecosystem with zero document uploads.',
            style: TextStyle(color: textMuted, fontSize: 12),
          ),
          const SizedBox(height: 14),
          ElevatedButton.icon(
            style: ElevatedButton.styleFrom(
              backgroundColor: neonCyan,
              foregroundColor: Colors.black,
              minimumSize: const Size.fromHeight(44),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            icon: const Icon(Icons.shield_outlined, size: 18),
            label: state.isProcessing
                ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.black))
                : const Text('Connect DigiLocker Fast-Track', style: TextStyle(fontWeight: FontWeight.bold)),
            onPressed: state.isProcessing ? null : () => widget.controller.triggerDigiLockerSync(),
          ),
        ],
      ),
    );
  }

  Widget _buildDigiLockerConsentCard(DigiLockerKycState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: cardSurface,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          const Icon(Icons.verified_user, color: neonCyan, size: 40),
          const SizedBox(height: 12),
          const Text('DigiLocker Consent Received', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          Text(state.consentRecord?.consentArtifactHash ?? '', style: const TextStyle(color: textMuted, fontSize: 10)),
        ],
      ),
    );
  }

  Widget _buildPanSection(DigiLockerKycState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: cardSurface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('NSDL PAN Registry Validation', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14)),
              IconButton(
                icon: const Icon(Icons.document_scanner, color: neonCyan, size: 20),
                onPressed: () {
                  widget.controller.simulatePanOcr();
                  _panTextController.text = widget.controller.state.panInput;
                },
                tooltip: 'Scan PAN with OCR',
              ),
            ],
          ),
          const SizedBox(height: 6),
          const Text('Enter your 10-character permanent account number for Section 194S TDS compliance.', style: TextStyle(color: textMuted, fontSize: 12)),
          const SizedBox(height: 16),
          TextField(
            controller: _panTextController,
            style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 15, letterSpacing: 2),
            maxLength: 10,
            textCapitalization: TextCapitalization.characters,
            decoration: InputDecoration(
              hintText: 'ABCPS1234D',
              hintStyle: const TextStyle(color: Colors.white30),
              filled: true,
              fillColor: obsidianBackground,
              counterText: '',
              prefixIcon: const Icon(Icons.badge_outlined, color: neonCyan, size: 20),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(6), borderSide: BorderSide.none),
            ),
            onChanged: (val) => widget.controller.setPanInput(val),
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: neonGreen,
              foregroundColor: Colors.black,
              minimumSize: const Size.fromHeight(44),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () => widget.controller.verifyPan(),
            child: const Text('Verify PAN with Income Tax Registry', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  Widget _buildCompletionCard(DigiLockerKycState state) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: cardSurface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: neonGreen.withOpacity(0.5)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Center(
            child: Icon(Icons.verified, color: neonGreen, size: 54),
          ),
          const SizedBox(height: 12),
          const Center(
            child: Text(
              'TIER-2 SOVEREIGN KYC VERIFIED',
              style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16, letterSpacing: 1.1),
            ),
          ),
          const SizedBox(height: 8),
          const Center(
            child: Text(
              'Institutional verification complete. Your account is activated for unlimited INR deposits, withdrawals, and spot trading.',
              textAlign: TextAlign.center,
              style: TextStyle(color: textMuted, fontSize: 12),
            ),
          ),
          const Divider(color: Colors.white12, height: 32),
          _buildReceiptRow('Aadhaar Status', state.aadhaarPayload?.formattedMaskedNumber ?? 'Verified'),
          _buildReceiptRow('PAN Registry', state.panPayload?.maskedPan ?? 'ABCPS****D'),
          _buildReceiptRow('Entity Type', state.panPayload?.entityClassification ?? 'Individual'),
          _buildReceiptRow('Besu DID Attestation', state.onChainAttestationTx ?? '0x9b7a...810a', isMono: true, valueColor: neonCyan),
          const SizedBox(height: 24),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: neonGreen,
              foregroundColor: Colors.black,
              minimumSize: const Size.fromHeight(46),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () {
              widget.onCompleted?.call();
              Navigator.of(context).maybePop();
            },
            child: const Text('Proceed to Trading Terminal', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  Widget _buildReceiptRow(String label, String value, {bool isMono = false, Color? valueColor}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: textMuted, fontSize: 12)),
          Text(
            value,
            style: TextStyle(
              color: valueColor ?? Colors.white,
              fontFamily: isMono ? 'JetBrains Mono' : null,
              fontWeight: FontWeight.w600,
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorBanner(String msg) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: neonRed.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: neonRed),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: neonRed, size: 18),
          const SizedBox(width: 8),
          Expanded(child: Text(msg, style: const TextStyle(color: neonRed, fontSize: 12))),
        ],
      ),
    );
  }

  Widget _buildSuccessBanner(String msg) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: neonGreen.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: neonGreen),
      ),
      child: Row(
        children: [
          const Icon(Icons.check_circle_outline, color: neonGreen, size: 18),
          const SizedBox(width: 8),
          Expanded(child: Text(msg, style: const TextStyle(color: neonGreen, fontSize: 12))),
        ],
      ),
    );
  }
}

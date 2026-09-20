import 'package:flutter/material.dart';
import '../../domain/models/wdra_vault_location.dart';
import '../controllers/commodity_redemption_controller.dart';

class WdraVaultLocatorScreen extends StatefulWidget {
  final CommodityRedemptionController controller;

  const WdraVaultLocatorScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<WdraVaultLocatorScreen> createState() => _WdraVaultLocatorScreenState();
}

class _WdraVaultLocatorScreenState extends State<WdraVaultLocatorScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  String _selectedCityFilter = 'All';
  List<WdraVaultLocation> _vaults = [];
  bool _isLoading = true;

  final List<String> _cities = ['All', 'Mumbai', 'New Delhi', 'Ahmedabad', 'Bengaluru'];

  @override
  void initState() {
    super.initState();
    _loadVaults();
  }

  Future<void> _loadVaults() async {
    setState(() => _isLoading = true);
    // In mock, fetch repository vaults
    _vaults = [
      const WdraVaultLocation(
        vaultId: 'VAULT-MUM-001',
        repositoryName: 'National E-Repository Limited (NERL)',
        repositoryRegistrationNo: 'WDRA/REG/2021/MUM/048',
        operatorName: 'Sequel Vaulting Logistics Pvt Ltd',
        facilityName: 'Bandra-Kurla Complex Bullion Hub',
        addressLine: 'Plot C-59, G-Block, BKC, Bandra East',
        city: 'Mumbai',
        state: 'Maharashtra',
        pincode: '400051',
        latitude: 19.0657,
        longitude: 72.8687,
        operatingHours: '10:00 AM - 05:00 PM (Mon-Fri)',
        supportedCommodities: [],
        pickupRequirements: ['Original PAN Card', 'Aadhaar Biometric Check', 'DigiLocker Gate Pass'],
        isOperational: true,
      ),
      const WdraVaultLocation(
        vaultId: 'VAULT-DEL-002',
        repositoryName: 'CDSL Commodity Repository Limited (CCRL)',
        repositoryRegistrationNo: 'WDRA/REG/2020/DEL/012',
        operatorName: 'MMTC-PAMP India Pvt Ltd',
        facilityName: 'Connaught Place Central Depository',
        addressLine: 'Barakhamba Road, Statesman House, Ground Floor',
        city: 'New Delhi',
        state: 'Delhi',
        pincode: '110001',
        latitude: 28.6289,
        longitude: 77.2285,
        operatingHours: '09:30 AM - 04:30 PM (Mon-Fri)',
        supportedCommodities: [],
        pickupRequirements: ['Original PAN Card', 'Aadhaar Biometric Check', 'Gate Pass OTP'],
        isOperational: true,
      ),
      const WdraVaultLocation(
        vaultId: 'VAULT-AHM-003',
        repositoryName: 'National E-Repository Limited (NERL)',
        repositoryRegistrationNo: 'WDRA/REG/2022/GUJ/089',
        operatorName: 'Brink\'s India Secure Logistics',
        facilityName: 'GIFT City Bullion Vault',
        addressLine: 'Zone 1, GIFT SEZ, Gandhinagar',
        city: 'Ahmedabad',
        state: 'Gujarat',
        pincode: '382355',
        latitude: 23.1601,
        longitude: 72.6841,
        operatingHours: '10:00 AM - 06:00 PM (Mon-Fri)',
        supportedCommodities: [],
        pickupRequirements: ['Original PAN Card', 'SEZ Visitor Pass', 'DigiLocker QR'],
        isOperational: true,
      ),
      const WdraVaultLocation(
        vaultId: 'VAULT-BLR-004',
        repositoryName: 'National E-Repository Limited (NERL)',
        repositoryRegistrationNo: 'WDRA/REG/2023/KAR/104',
        operatorName: 'BVC Secure Logistics Pvt Ltd',
        facilityName: 'MG Road Depository Center',
        addressLine: '45 MG Road, Ashok Nagar',
        city: 'Bengaluru',
        state: 'Karnataka',
        pincode: '560001',
        latitude: 12.9756,
        longitude: 77.6066,
        operatingHours: '10:00 AM - 05:00 PM (Mon-Sat)',
        supportedCommodities: [],
        pickupRequirements: ['Original PAN Card', 'Aadhaar Card', 'Delivery Token'],
        isOperational: true,
      ),
    ];
    setState(() => _isLoading = false);
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _selectedCityFilter == 'All'
        ? _vaults
        : _vaults.where((v) => v.city == _selectedCityFilter).toList();

    return Scaffold(
      backgroundColor: obsidianBackground,
      appBar: AppBar(
        backgroundColor: obsidianBackground,
        elevation: 0,
        title: const Text('WDRA Vault Locator', style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold)),
      ),
      body: Column(
        children: [
          // City filter chips
          Container(
            height: 50,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: _cities.length,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                final city = _cities[index];
                final isSelected = city == _selectedCityFilter;
                return ChoiceChip(
                  label: Text(city, style: TextStyle(color: isSelected ? Colors.black : Colors.white, fontSize: 12)),
                  selected: isSelected,
                  selectedColor: neonGreen,
                  backgroundColor: surfaceCard,
                  onSelected: (selected) {
                    if (selected) setState(() => _selectedCityFilter = city);
                  },
                );
              },
            ),
          ),
          const SizedBox(height: 8),
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator(color: neonGreen))
                : ListView.separated(
                    padding: const EdgeInsets.all(16),
                    itemCount: filtered.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (context, index) {
                      final vault = filtered[index];
                      final isSelected = widget.controller.state.selectedVault?.vaultId == vault.vaultId;
                      return _buildVaultCard(vault, isSelected);
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildVaultCard(WdraVaultLocation vault, bool isSelected) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: isSelected ? neonGreen : Colors.white12, width: isSelected ? 1.5 : 1.0),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(
                  vault.facilityName,
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: neonGreen.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: const Text('WDRA Accredited', style: TextStyle(color: neonGreen, fontSize: 10, fontWeight: FontWeight.bold)),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(vault.operatorName, style: const TextStyle(color: textMuted, fontSize: 12)),
          const SizedBox(height: 6),
          Text('${vault.addressLine}, ${vault.city} - ${vault.pincode}', style: const TextStyle(color: Colors.white70, fontSize: 11)),
          const SizedBox(height: 4),
          Text('Hours: ${vault.operatingHours}', style: const TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(height: 10),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: isSelected ? neonGreen : surfaceCard,
                foregroundColor: isSelected ? Colors.black : Colors.white,
                side: BorderSide(color: isSelected ? neonGreen : Colors.white24),
              ),
              onPressed: () {
                widget.controller.selectVault(vault);
                Navigator.of(context).pop();
              },
              child: Text(isSelected ? 'Selected Vault' : 'Select This Vault'),
            ),
          ),
        ],
      ),
    );
  }
}

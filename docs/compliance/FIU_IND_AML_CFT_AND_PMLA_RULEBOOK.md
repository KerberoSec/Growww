# FIU-IND Anti-Money Laundering (AML), CFT & PMLA Compliance Rulebook

## 1. Statutory Mandate & Registration

- **Enforcement Authority:** Financial Intelligence Unit - India (FIU-IND), Ministry of Finance, Government of India.
- **Governing Legislation:** Prevention of Money Laundering Act 2002 (PMLA) and Prevention of Money-Laundering (Maintenance of Records) Rules 2005.
- **Entity Classification:** Reporting Entity (RE) registered with FIU-IND under PMLA Section 12.
- **Principal Officer:** Designated senior compliance officer responsible for regulatory filings, suspicious transaction reports, and direct liaison with FIU-IND.

---

## 2. Customer Due Diligence (CDD) & KYC Standards

### 2.1 Resident Indian Investors (Domestic)
1. **Central KYC (CKYC) Registry Ingestion:** Mandatory query against the CKYCR database using 14-digit CKYC identification number.
2. **Aadhaar e-KYC (UIDAI):** Biometric or OTP-based authentication with automated cryptographic XML signature verification.
3. **PAN Verification:** Real-time query against Income Tax Department (NSDL/Protean database) verifying name, date of birth, and Aadhaar-PAN seeding status.
4. **Bank Account Penny Drop:** Instant verification of beneficiary account name via UPI/IMPS penny drop; mandatory 100% string match against KYC records. Third-party deposits are rejected and returned immediately.

### 2.2 Politically Exposed Persons (PEP) & Sanctions Screening
- Automated screening against UN Consolidated List, OFAC SDN List, EU Financial Sanctions Database, and MHA UAPA banned entities.
- Daily automated delta re-screening of the entire active investor database. Immediate account freeze upon positive sanctions match.

---

## 3. Transaction Monitoring & Statutory Reporting

### 3.1 Suspicious Transaction Reports (STR)
The compliance surveillance daemon (`services/kyc-aml-service`) monitors trading activity in real time. An STR is automatically flagged and compiled for submission to the FINnet 2.0 portal within **7 working days** of suspicion under scenarios including:
- Rapid structuring of deposits just below the ₹10,00,000 threshold.
- Suden, unexplained high-velocity trading activity inconsistent with declared income slab.
- Deposit of cryptocurrency from high-risk mixers or sanctioned wallet clusters.
- Rapid circular trading between connected accounts without change in beneficial ownership.

### 3.2 Cash Transaction Reports (CTR) & Non-Profit Organization Reports (NTR)
- Automatic aggregation and filing of statutory monthly CTRs for cross-border currency conversions exceeding ₹10,00,000 equivalent.

---

## 4. FATF Travel Rule Compliance for Crypto Transfers

For all incoming and outgoing virtual digital asset (VDA) transfers (BTC, USDT):
- **Threshold:** Full originator and beneficiary information captured for all transfers exceeding ₹50,000 equivalent.
- **Payload Metadata:**
  - Originator Name, Account ID, and Physical Address / National ID.
  - Beneficiary Name and Destination Wallet Address.
  - Inter-VASP Messaging: Transmitted over secure, encrypted IVMS 101 protocol before broadcast.
- **Quarantine Policy:** Deposits originating from unverified VASPs or darknet clusters are held in quarantined cold escrow pending manual compliance clearance.

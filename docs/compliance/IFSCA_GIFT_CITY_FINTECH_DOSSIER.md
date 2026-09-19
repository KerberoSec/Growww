# IFSCA GIFT City FinTech Regulatory Dossier: Global Capital Ingress & Multi-Currency Gateway

## 1. Executive Summary & Jurisdiction

- **Regulatory Authority:** International Financial Services Centres Authority (IFSCA), GIFT City, Gandhinagar, Gujarat, India.
- **Applicant:** Growww International IFSC Private Limited.
- **Regulatory License Category:** IFSCA FinTech Regulatory Sandbox / International Capital Markets Intermediary.
- **Core Mandate:** Enabling non-resident investors, foreign portfolio investors (FPIs), and non-resident Indians (NRIs) to access tokenized Indian real-world assets (RWA) and multi-currency digital liquidity (USD, EUR, USDT, e₹) within a tax-neutral, sovereign offshore financial framework.

---

## 2. Cross-Border Capital Architecture & Two-Entity Model

```
+------------------------------------+                  +------------------------------------+
| DOMESTIC INDIAN JURISDICTION       |                  | GIFT CITY IFSC JURISDICTION        |
| (SEBI / RBI Regulated)             |                  | (IFSCA Regulated)                  |
+------------------------------------+                  +------------------------------------+
| • Growww Technologies India P. Ltd |                  | • Growww International IFSC P. Ltd |
| • Domestic retail investors (INR)  |  Encrypted mTLS  | • Foreign investors & NRIs (USD)   |
| • Primary Indian Equities & G-Secs |<---------------->| • Global multi-currency funding    |
| • NSDL / CDSL Depository Custody   |   Cross-Entity   | • Offshore banking escrow rails    |
| • 100% Tax Compliant (STT / TDS)   |   Saga Gateway   | • Tax-neutral capital gains (Sec 47|
+------------------------------------+                  +------------------------------------+
```

### Key Separation Rules:
1. **Zero Direct Inter-Entity Database Sharing:** All communications occur via asynchronous, cryptographically signed gRPC channels with mutual TLS.
2. **Strict Currency Segregation:** Domestic entity operates strictly in INR and e₹; IFSC entity operates in freely convertible foreign currencies (USD, EUR, GBP) and permitted stablecoins (USDT/USDC).
3. **FEMA & LRS Compliance:** Indian residents investing abroad via the IFSC gateway strictly comply with the RBI Liberalised Remittance Scheme (LRS) annual limit ($250,000 USD).

---

## 3. Product Features & Foreign Ingress Rails

1. **Foreign Investor Onboarding:** Centralized digital KYC compliant with FATF standards, passport verification, proof of address, and automatic screening against UN, OFAC, and INTERPOL sanctions lists.
2. **Multi-Currency Escrow Accounts:** Integrated with IFSC Banking Units (IBUs) including ICICI Bank IBU, HDFC Bank IBU, and HSBC GIFT City branch.
3. **Tokenized Indian Depository Receipts (T-IDRs):** Foreign investors purchase USD-denominated fractional tokens representing underlying Indian blue-chip equities held in domestic NSDL/CDSL escrow.
4. **24/7 Global Trading Desk:** Continuous secondary trading matched against internal liquidity pools without waiting for domestic Indian market opening hours.

---

## 4. Tax Neutrality & Statutory Benefits

Under Section 47(viiab) of the Indian Income Tax Act 1961:
- Zero capital gains tax on transactions in foreign-currency denominated tokenized securities executed on an recognized stock exchange in IFSC by non-resident investors.
- Zero Securities Transaction Tax (STT) and zero Stamp Duty for non-resident trades within GIFT City jurisdiction.
- Minimum Alternate Tax (MAT) capped at 9% for offshore corporate entities operating within the IFSC.

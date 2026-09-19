# 813 - High-Fidelity NSDL/CDSL Depository Simulator & RBI UPI/e-Rupee Banking Mock Sandbox Engine

## Purpose
In the Indian financial and securities ecosystem, integrating with traditional market infrastructure (NSDL, CDSL, Clearing Corporations, Core Banking Systems, NPCI UPI switches, and RBI CBDC/e-Rupee rails) presents severe development bottlenecks. Live production test environments are restricted, slow, batch-bound, and lack controllable edge-case simulation for system failures.

This prompt specifies the architecture, data models, and protocols for Growww's High-Fidelity Depository & Banking Mock Sandbox Suite. The platform provides an autonomous, deterministically reproducible sandbox engine that simulates NSDL/CDSL depository participant operations (BOID registry, demat holding ledgers, pledge/unpledge, corporate actions, and encrypted SFTP batch settlements) alongside RBI/NPCI banking rails (UPI 2.0 collect/pay, mandate autopay, IMPS/NEFT/RTGS, ISO 20022 financial messaging, and RBI Digital Rupee CBDC tokenized wallets) with programmable chaos and latency injection.

## What You Are Building
An autonomous, high-performance depository and banking sandbox simulation platform:
- `services/mock-depository/`: High-fidelity NSDL (DPM) and CDSL (CDAS) depository engine managing BOID master data, demat accounts, ISIN holding ledgers, pledge/freeze locks, and automated settlement batches.
- `services/mock-sftp-server/`: Embedded secure SFTP server with automated PGP/GPG encryption/decryption, processing scheduled batch file exchanges (`.BENPOS`, `.DIS`, `.SOD`, `.EOD`, `.RECON`).
- `services/mock-banking-engine/`: Multi-rail banking simulator providing UPI 2.0 (NPCI switch emulation, dynamic QR, intent flows, auto-mandates, callback webhooks), IMPS, NEFT, and RTGS APIs.
- `services/mock-cbdc-engine/`: RBI Digital Rupee (e-Rupee CBDC) tokenized wallet engine simulating Retail (CBDC-R) and Wholesale (CBDC-W) mint, burn, transfer, and atomic DvP operations.
- `pkg/iso20022/`: High-performance ISO 20022 financial messaging engine parsing and serializing `pacs.008`, `pacs.002`, `camt.053`, `camt.054`, `pain.001`, and `pain.002` XML/JSON documents.
- `services/sandbox-chaos-injector/`: Configurable failure injection middleware simulating NPCI switch 500s, bank network timeouts, corrupted SFTP payloads, duplicate transactions, and out-of-order execution.
- `web/sandbox-admin-portal/`: Interactive web console and CLI for inspecting ledger balances, overriding mock states, triggering ad-hoc EOD settlement runs, and injecting synthetic faults.

## Scope Boundaries
- **In Scope:**
  - Full simulation of NSDL (16-digit alphanumeric starting with 'IN') and CDSL (16-digit numeric) depository account structures.
  - SFTP batch settlement lifecycle with fixed-width and comma-delimited file formats (`.BENPOS`, `.DIS`, `.EOD`).
  - Automated PGP key-ring encryption and decryption of depository batch drops.
  - UPI 2.0 NPCI switch emulation: VPA validation, pay/collect requests, dynamic QR strings, autopay mandate authorization, and real-time asynchronous callbacks.
  - RBI Digital Rupee (CBDC) simulation: cryptographic token issuance, hierarchical deterministic (HD) wallet addressing, and two-tier distribution model (RBI -> Bank -> Customer).
  - ISO 20022 financial message validation against SWIFT/NPCI XSD schemas.
  - Dynamic fault and latency injection engine (0ms to 60000ms latency, configurable error percentages).
  - Web UI and REST/gRPC administration APIs for state reset and synthetic entity provisioning.
- **Out of Scope / Handled Elsewhere:**
  - Live production Custodian Depository adapter (Prompt 213).
  - Live production Payment Gateway integration (Prompt 212).
  - Production Trade Settlement orchestration (Prompt 208).
  - Automated reconciliation engine execution (Prompt 215).

## Technology to Use
- **Go (Golang 1.22+)**: Primary language for high-throughput mock depository and banking switch services. Justification: Ultra-low latency, native concurrency for handling thousands of mock transactions per second, and compact binary footprint.
- **SFTPGo / Embedded Go SFTP (`golang.org/x/crypto/ssh`)**: Integrated SFTP server with automated virtual filesystem routing and PGP crypto hooks.
- **PostgreSQL 16**: Relational storage for depository accounts, holding ledgers, banking virtual accounts, and transaction logs.
- **Redis 7**: High-speed memory store for VPA routing tables, pending UPI mandates, and transient chaos injection configurations.
- **React + Vite & TailwindCSS**: Lightweight sandbox administration dashboard.
- **OpenPGP (`golang.org/x/crypto/openpgp`)**: Standards-compliant PGP encryption engine for depository batch file processing.

## Backend / Infra Touchpoints
- **Custodian Depository Integration Service (Prompt 213)**: Connects to mock SFTP and depository REST/gRPC endpoints.
- **Payment Gateway Service (Prompt 212)**: Connects to mock UPI and ISO 20022 banking endpoints.
- **Trade Settlement Service (Prompt 208)**: Triggers depository settlement batches and CBDC atomic DvP instructions.
- **Reconciliation Service (Prompt 215)**: Ingests mock `.EOD` and `camt.053` bank statements for automated end-of-day reconciliation.
- **Docker Compose & Kubernetes**: Packaged as standard containerized services for local development and shared staging environments.

## Blockchain Interaction
The sandbox suite bridges traditional Indian financial infrastructure simulations with on-chain settlement contracts:
- **Proof of Reserve Synchronization**: The mock depository generates synthetic `.BENPOS` files reflecting custodian equity holdings, which the Custodian Adapter (Prompt 213) and Proof of Reserve service (Prompt 306) ingest to verify on-chain token backing on Hyperledger Besu.
- **Atomic CBDC-to-Token DvP Emulation**: The mock e-Rupee engine exposes cryptographic signature verification APIs, allowing `SettlementDvP.sol` and the settlement relayer (Prompt 208) to execute atomic Delivery versus Payment where fractional equity tokens transfer on-chain simultaneously with mock CBDC payment finality.
- **Depository Lock/Pledge Synchronization**: When shares are pledged or frozen in the mock depository, the simulator emits webhook events that trigger on-chain compliance token freeze methods via `ComplianceRegistry.sol`.

## Step-by-Step Build Instructions
1. Scaffold project directory structure: `services/mock-depository/`, `services/mock-banking-engine/`, `services/mock-cbdc-engine/`, `services/mock-sftp-server/`, `pkg/iso20022/`, and `web/sandbox-admin-portal/`.
2. Define PostgreSQL schemas for depository accounts (BOID, PAN, DP ID, demat balance, pledged balance, frozen balance) and banking accounts (VPA, Account Number, IFSC, balance, CBDC wallet address).
3. Implement `pkg/iso20022/` parser and serializer for `pacs.008.001.08`, `pacs.002.001.10`, `camt.053.001.08`, and `pain.001.001.09` with XML schema validation.
4. Build `services/mock-depository/` core engine: implement demat account creation, credit/debit instructions, corporate action distribution (dividends, splits, bonus shares), and holding freeze operations.
5. Implement `services/mock-sftp-server/` with automated PGP key ring management. Configure virtual SFTP directories: `/inbound/dis/`, `/inbound/pledge/`, `/outbound/benpos/`, `/outbound/eod/`.
6. Author the batch settlement processor: parse incoming `.DIS` files, validate signatures and account balances, execute transfers, and generate encrypted `.BENPOS` and `.EOD` response files according to NSDL/CDSL fixed-width specs.
7. Build `services/mock-banking-engine/` implementing UPI 2.0 endpoints (`/upi/v2/validate-vpa`, `/upi/v2/pay`, `/upi/v2/collect`, `/upi/v2/mandate/create`, `/upi/v2/mandate/execute`).
8. Implement asynchronous callback worker in mock banking engine that dispatches webhook notifications to configured gateway endpoints with simulated network delays.
9. Build `services/mock-cbdc-engine/` implementing RBI Digital Rupee (e-Rupee) APIs: wallet balance query, mint synthetic CBDC tokens, transfer between VPAs/wallets, and atomic DvP escrow hold/release endpoints.
10. Implement `services/sandbox-chaos-injector/` HTTP/gRPC middleware: allows setting failure probabilities (e.g. 15% rate limit errors, 10% timeout, 5% corrupted XML response) via admin REST API.
11. Build `web/sandbox-admin-portal/`: dashboard displaying active accounts, live transaction feeds, SFTP file browser, chaos configuration sliders, and manual "Trigger EOD Settlement" button.
12. Package the entire suite into `docker-compose.sandbox.yml` and Kubernetes Helm charts with automated health check probes.

## Interfaces / Contracts

```xml
<!-- pkg/iso20022/samples/pacs.008.001.08.xml (Financial Customer Credit Transfer) -->
<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.08">
  <FIToFICstmrCdtTrf>
    <GrpHdr>
      <MsgId>GROWWW-UPI-20260918-0091823</MsgId>
      <CreDtTm>2026-09-18T16:00:00+05:30</CreDtTm>
      <NbOfTxs>1</NbOfTxs>
      <SttlmInf>
        <SttlmMtd>CLRG</SttlmMtd>
        <ClrSys>
          <Prtry>NPCI_UPI</Prtry>
        </ClrSys>
      </SttlmInf>
    </GrpHdr>
    <CdtTrfTxInf>
      <PmtId>
        <EndToEndId>E2E-20260918-7749102</EndToEndId>
        <TxId>TXN-NPCI-88392019482</TxId>
      </PmtId>
      <IntrBkSttlmAmt Ccy="INR">50000.00</IntrBkSttlmAmt>
      <Dbtr>
        <Nm>AARAV SHARMA</Nm>
        <Id>
          <OrgId>
            <Othr>
              <Id>aarav@okhdfcbank</Id>
              <SchmeNm><Prtry>VPA</Prtry></SchmeNm>
            </Othr>
          </OrgId>
        </Id>
      </Dbtr>
      <Cdtr>
        <Nm>GROWWW CUSTODY ESCROW</Nm>
        <Id>
          <OrgId>
            <Othr>
              <Id>growww.escrow@icici</Id>
              <SchmeNm><Prtry>VPA</Prtry></SchmeNm>
            </Othr>
          </OrgId>
        </Id>
      </Cdtr>
    </CdtTrfTxInf>
  </FIToFICstmrCdtTrf>
</Document>
```

```text
# Fixed-Width NSDL/CDSL Depository .BENPOS File Format (Example Row Layout)
# Field Layout: RecordType(02) | BOID(16) | PAN(10) | ISIN(12) | FreeQty(15,3) | PledgeQty(15,3) | FreezeQty(15,3) | Timestamp(14)
01IN30012610984921ABCDE1234FINE002A01018000000000100.000000000000000.000000000000000.00020260918160000
011208160019482019BCDEF2345GINE002A01018000000000050.500000000000010.000000000000000.00020260918160000
```

```go
// services/mock-banking-engine/api/upi_handler.go (Excerpt)
package api

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type UPIPaymentRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	PayerVPA      string  `json:"payer_vpa" binding:"required"`
	PayeeVPA      string  `json:"payee_vpa" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,eq=INR"`
	CallbackURL   string  `json:"callback_url" binding:"required,url"`
}

type UPIPaymentResponse struct {
	Status        string    `json:"status"` // SUCCESS, PENDING, FAILED
	NPCIRefNumber string    `json:"npci_ref_number"`
	ResponseCode  string    `json:"response_code"` // "00" for Success, "U16" for Insufficient Funds
	Timestamp     time.Time `json:"timestamp"`
}

func HandleUPIPay(c *gin.Context) {
	var req UPIPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process simulated payment and return immediate ack + async webhook dispatch
	resp := UPIPaymentResponse{
		Status:        "SUCCESS",
		NPCIRefNumber: "NPCI" + time.Now().Format("20060102150405"),
		ResponseCode:  "00",
		Timestamp:     time.Now(),
	}
	c.JSON(http.StatusOK, resp)
}
```

## Security & Compliance Notes
- **Synthetic Data Generation**: The sandbox exclusively operates on synthetic PAN numbers (e.g. starting with `ABCDE`), virtual dummy Aadhaar IDs, and reserved mock Bank IFSCs (`GROW0000001`) to prevent production PII leakage.
- **PGP Encryption Enforced**: All SFTP depository batch drops require valid PGP signatures and encryption using 4096-bit RSA keys, matching production NSDL/CDSL cybersecurity mandates.
- **Audit Log Trail**: Every simulated depository credit, debit, freeze, and UPI payment is assigned a cryptographically hashed audit record conforming to SEBI regulatory log-retention standards (Prompt 807).
- **Idempotency Standards**: All mock banking endpoints enforce `X-Idempotency-Key` headers, rejecting or returning cached responses on duplicate requests.

## Acceptance Criteria
- [ ] Mock Depository generates compliant NSDL and CDSL `.BENPOS` and `.EOD` batch files.
- [ ] Embedded SFTP server handles automated PGP encrypted file uploads, downloads, and directory polling.
- [ ] UPI 2.0 switch emulator successfully executes collect, pay, VPA validation, and autopay mandate flows with async callbacks.
- [ ] RBI e-Rupee CBDC simulator supports token minting, wallet balances, transfers, and atomic DvP escrow locks.
- [ ] ISO 20022 messaging library correctly validates and generates `pacs.008`, `pacs.002`, `camt.053`, and `pain.001` XML.
- [ ] Sandbox Chaos Injector allows runtime configuration of latency delays and failure rates via Admin API.
- [ ] Admin Web Console allows visual ledger inspection, synthetic balance credit, and manual settlement triggers.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 104 (Event Schemas & Kafka Topics), Prompt 212 (Payment Gateway Service), Prompt 213 (Custodian Depository Integration Service).
- **Parallel Tasks:** Prompt 811 (Dual-Environment CI/CD Pipeline), Prompt 901 (Unit & Integration Testing Strategy).
- **Next Steps:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch Plan).

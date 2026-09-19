# 742 - HSM FIPS 140-3 Level 3 & Level 4 Transition Roadmap

## Purpose
Establishes the cryptographic transition roadmap, hardware security module validation standards, and regulatory compliance mapping for migration from legacy FIPS 140-2 Level 3 to NIST FIPS 140-3 Level 3/4 across the Growww / NBSE platform.

This specification details physical tamper envelopes, side-channel attack countermeasures, zeroization firmware policies, and mTLS remote signing daemon architectures compliant with SEBI CSCRF and RBI Cyber Security Guidelines.

## Key Capabilities
- **FIPS 140-3 Level 3 Baseline**: Enforced across all validator consensus signing, MPC key share nodes, and settlement relayer daemons.
- **Physical Security & Tamper Detection**: Active zeroization circuitry on enclosure breach, environmental monitoring (voltage, temperature, radiation).
- **Remote Attestation & mTLS Mesh**: Dedicated gRPC `CryptoSignerService` on TLS 1.3 (`TLS_AES_256_GCM_SHA384`) with SPIFFE/SPIRE workload attestation.
- **Strict Anti-Equivocation**: Local RocksDB write-ahead log preventing double-signing on QBFT consensus.

## Hardware Security Specification
```protobuf
syntax = "proto3";

package growww.crypto.hsm.v1;

option go_package = "growww/packages/proto/growww/crypto/hsm/v1;hsmv1";

enum FipsStandard {
  FIPS_STANDARD_UNSPECIFIED = 0;
  FIPS_STANDARD_140_2_LEVEL_3 = 1; // Deprecated by NIST
  FIPS_STANDARD_140_3_LEVEL_3 = 2; // Mandatory Production Baseline
  FIPS_STANDARD_140_3_LEVEL_4 = 3; // GIFT City High-Security Vaults
}

message HsmKeyMetadata {
  string key_alias = 1;
  string key_arn = 2;
  FipsStandard standard = 3;
  string curve = 4; // "secp256k1", "ed25519", "bls12-381"
  bool extractable = 5; // Strictly false
  uint64 creation_timestamp_ms = 6;
}

message SignDigestRequest {
  string key_alias = 1;
  bytes digest_32 = 2;
  string request_id = 3;
}

message SignDigestResponse {
  string request_id = 1;
  bytes signature = 2;
  uint32 v = 3;
  FipsStandard validation_level = 4;
}

service HsmSignerService {
  rpc SignDigest(SignDigestRequest) returns (SignDigestResponse);
  rpc GetKeyMetadata(SignDigestRequest) returns (HsmKeyMetadata);
}
```

## Migration Milestones
1. **Milestone 1**: CloudHSM firmware upgrade to FIPS 140-3 Level 3 across AWS Mumbai, GCP Frankfurt, and Azure Singapore clusters.
2. **Milestone 2**: Hardware attestation integration with SPIFFE/SPIRE agents verifying hardware identity tokens.
3. **Milestone 3**: Physical air-gapped GIFT City cold storage upgrade to FIPS 140-3 Level 4 tamper-sensing active chassis.

# ADR-0002: Polyglot Monorepo Workspace Layout

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Architect, Platform Lead  

---

## 1. Context
Growww consists of 24 microservices across Go, Python, and Rust, 4 frontend/mobile applications across Next.js and Flutter, Solidity smart contracts, and shared protobuf definitions. Managing multiple independent repositories creates version drift across wire protocols, breaking changes in schemas, and operational complexity.

---

## 2. Decision
Adopt a single **unified monorepo** managed with language-native workspace tooling:
- **Go:** `go.work` linking all 16 Go microservices and `packages/go_common`.
- **Rust:** `Cargo.toml` virtual workspace linking `services/matching-engine` and `services/audit-service`.
- **JavaScript / TypeScript:** `pnpm` workspaces with Turborepo (`turbo.json`).
- **Protobuf:** Buf CLI (`buf.yaml` v2) managing single-source-of-truth wire contracts.

---

## 3. Alternatives Considered
- **Multi-Repo (Polyrepo):** Rejected due to high overhead in cross-service protocol updates, synchronization lag, and difficult atomic refactoring.

---

## 4. Consequences
- **Positive:** Atomic changes across protobuf schemas, backend services, and frontends; unified CI/CD pipelines.
- **Trade-offs:** Requires tooling to prune build scopes and manage independent build caches.

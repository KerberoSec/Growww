# Engineering Standards, Provenance & Governance Framework

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Engineering Excellence & Security Governance Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Cryptographic Provenance & Source Control Governance (G-01)

As an institutional financial platform operating under SEBI, RBI, and IFSCA oversight, the provenance of all software is strictly auditable and cryptographically verified.

### 1.1 Commit Signing & Author Verification
1. **Mandatory Signing:** Every commit pushed to the repository must be signed using a verified SSH key or GPG key linked to the developer's corporate identity.
2. **Identity Enforcement:** Commit author emails must match the corporate domain (`@growww.in`). Commits from personal email providers are rejected at the pre-receive git hook barrier.
3. **Linear History & Branch Protection:**
   - Direct pushes to `main` and release branches are permanently disabled.
   - Merges require a pull request with at least 2 approving reviews, including designated CODEOWNERS.
   - All CI verification workflows (SAST, unit tests, linting, formatting, schema checks) must pass.

---

## 2. Toolchain Pinning & Reproducibility (G-02)

To prevent compiler drift and ensure deterministic builds across developer workstations and CI runners, all toolchains are explicitly pinned:

```
# .tool-versions (mise / asdf)
golang   1.22.5
python   3.11.9
nodejs   20.14.0
rust     1.78.0
flutter  3.22.2
poetry   1.8.3
```

```toml
# rust-toolchain.toml
[toolchain]
channel = "1.78.0"
components = ["rustfmt", "clippy"]
targets = ["x86_64-unknown-linux-musl"]
```

---

## 3. Architecture Decision Records (ADR) Governance (G-03)

Any consequential design choice that is expensive to reverse requires a formal Architecture Decision Record stored under `docs/adr/`.

### ADR Template Standard
- **Title:** `ADR-XXXX: <Short Decision Summary>`
- **Status:** Proposed | Accepted | Deprecated | Superseded
- **Context:** Architectural problem, constraints, and regulatory requirements.
- **Decision:** Chosen architecture and design patterns.
- **Alternatives Considered:** Comparison matrix detailing trade-offs and rationale for rejection.
- **Consequences:** Positive benefits, accepted operational trade-offs, and failure modes.
- **Revisit Trigger:** Explicit measurable criteria that would trigger revisiting the decision.

---

## 4. Operational Release & Image Tagging Standards (G-06, G-07)

1. **Semantic Versioning:** All shared packages and services follow SemVer (`vMAJOR.MINOR.PATCH`).
2. **Immutable Container Image Tags:** Production container deployments strictly prohibit `:latest` tags. Images are tagged immutably using the commit SHA and build timestamp:
   $$\text{registry.growww.in/services/order-service:sha-a1b2c3d-20260918}$$
3. **Pre-Commit Security Checks (G-09):** Local git hooks execute `gitleaks` secret scanning, formatting, and linting prior to every commit.

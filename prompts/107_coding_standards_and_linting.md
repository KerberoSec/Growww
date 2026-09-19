# 107 - Polyglot Coding Standards, Code Formatting & Linting Suite

## Purpose
Establishes the enterprise-grade coding standards, static analysis rules, formatting configurations, and automated pre-commit quality gates across all 6 core programming languages used in the Growww platform: Go, Rust, Python, Dart/Flutter, TypeScript/Next.js, and Solidity.

In a mission-critical financial and blockchain infrastructure, code consistency, type safety, memory security, and vulnerability prevention are non-negotiable. This prompt provides developers with the canonical linter configurations, static application security testing (SAST) rules, and pre-commit automation to ensure zero lint warnings, deterministic code styling, and automated vulnerability detection before code enters the repository.

## What You Are Building
A comprehensive coding standards specification (`docs/standards/coding_standards.md`), unified `.editorconfig`, language-specific configuration files, and a pre-commit hook suite:
- Unified Editor Configuration (`.editorconfig`) enforcing charset, line endings, and indentation rules.
- Go Standards: `golangci-lint` configuration (`.golangci.yml`) enforcing error handling, concurrency safety, and formatting.
- Rust Standards: `rustfmt.toml` and `clippy.toml` with strict financial math and memory safety rules.
- Python Standards: `pyproject.toml` configuring Ruff (linter + formatter) and strict Mypy type-checking.
- Dart / Flutter Standards: `analysis_options.yaml` with strict pedantic lints for client-side state management and memory safety.
- TypeScript / Next.js Standards: `.eslintrc.js` and `prettier.config.js` enforcing strict TypeScript and React performance patterns.
- Solidity / Smart Contract Standards: `.solhint.json` and Slither static analysis configuration.
- Pre-Commit Hook Suite: `.pre-commit-config.yaml` orchestrating multi-language local validation.

## Scope Boundaries
- **In Scope:**
 - Formatting, styling, and naming conventions for Go, Rust, Python, Dart, TypeScript, and Solidity.
 - Linter configuration files and security static analysis rule definitions.
 - Pre-commit hook orchestration and local developer CLI scripts.
 - Zero-warning CI gating policy.
- **Out of Scope / Handled Elsewhere:**
 - Monorepo directory structure (handled in Prompt 106).
 - CI/CD pipeline automation and build matrix (handled in Prompt 803).
 - Automated security scanning in CI/CD (handled in Prompt 905).

## Technology to Use
- **Linters & Formatters:**
 - *Go:* `golangci-lint` (v1.57+), `gofumpt`, `govet`, `errcheck`, `gosec`.
 - *Rust:* `rustfmt`, `clippy` (enforcing `#![deny(clippy::all)]` and `#![deny(clippy::pedantic)]`).
 - *Python:* `ruff` (v0.3+ for lightning-fast linting/formatting) and `mypy` (v1.9+ in strict mode).
 - *Dart/Flutter:* `flutter_lints`, `custom_lint`, `dart format`.
 - *TypeScript:* ESLint (v8.57+), Prettier (v3.2+), TypeScript strict compiler checks.
 - *Solidity:* `solhint` (v4.1+), `forge fmt`, and `slither` (v0.10+ static analyzer).
- **Git Hooks:** `pre-commit` framework (v3.7+).

## Backend / Infra Touchpoints
- **IDE Integration:** Shared VS Code and IntelliJ IDEA recommended settings (`.vscode/settings.json`, `.vscode/extensions.json`).
- **Git Pre-Commit Hook:** Local execution preventing non-compliant commits.
- **CI Build Gating:** Automated PR checks failing on any linter error or code formatting diff.

## Blockchain Interaction
Establishes strict security and static analysis rules for smart contracts:
- **Solidity Coding Standards:** Enforces adherence to Solidity Style Guide (0.8.24+), NatSpec documentation comments on all public/external functions, and explicit visibility specifiers.
- **Slither Security Static Analysis:** Checks for reentrancy vulnerabilities, uninitialized storage pointers, dangerous delegatecalls, precision loss in integer division, and unchecked ERC-20/ERC-3643 return values.
- **Zero Inline Assembly Invariant:** Inline Yul assembly is prohibited in production business logic contracts unless explicitly approved via Architecture Decision Record and audited.

## Step-by-Step Build Instructions
1. Initialize documentation file `docs/standards/coding_standards.md`.
2. Create `.editorconfig` in repository root standardizing UTF-8 encoding, LF line endings, and trimmed trailing whitespace.
3. Configure Go static analysis:
 - Create `.golangci.yml` enabling linters: `gofumpt`, `govet`, `errcheck`, `gosec`, `staticcheck`, `revive`, `gocritic`.
 - Configure strict error-check rules requiring explicit handling of all returned `error` values.
4. Configure Rust static analysis:
 - Create `rustfmt.toml` configuring 100-character line width and 4-space indentation.
 - Configure `clippy.toml` with strict arithmetic overflow warnings, zero unwrap in production, and mandatory documentation comments.
5. Configure Python static analysis:
 - Update `pyproject.toml` with `[tool.ruff]` (enabling Pyflakes, pycodestyle, isort, flake8-bugbear, bandit) and `[tool.mypy]` (`strict = true`, `disallow_untyped_defs = true`).
6. Configure Dart / Flutter static analysis:
 - Create `analysis_options.yaml` including package `flutter_lints` with additional rules: `avoid_print`, `always_declare_return_types`, `prefer_const_constructors`, `unawaited_futures`.
7. Configure TypeScript / Next.js static analysis:
 - Create `.eslintrc.js` extending `@typescript-eslint/recommended-requiring-type-checking` and `next/core-web-vitals`.
 - Create `.prettierrc` configuring 2-space indentation, single quotes, and trailing commas.
8. Configure Solidity static analysis:
 - Create `.solhint.json` enforcing ERC-20/ERC-3643 naming rules, compiler version pragma pinning (`^0.8.24`), and NatSpec requirements.
 - Create `slither.config.json` defining automated vulnerability detectors.
9. Create `.pre-commit-config.yaml` linking all language formatters and linters into a single local validation hook.
10. Test pre-commit hook execution locally across sample files in each language.
11. Document IDE configuration guides for VS Code and IntelliJ in `docs/standards/coding_standards.md`.
12. Establish the "Zero Warning Policy": CI PR builds treat all linter warnings as fatal errors (`--warnings-as-errors`).

## Interfaces / Contracts

### Pre-Commit Configuration (`.pre-commit-config.yaml`)
```yaml
repos:
 - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
 - id: check-added-large-files
        args: ['--maxkb=500']
 - id: check-merge-conflict
 - id: end-of-file-fixer
 - id: trailing-whitespace

 - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.3.4
    hooks:
 - id: ruff
        args: [--fix]
 - id: ruff-format

 - repo: local
    hooks:
 - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run --new-from-rev=HEAD~1
        language: system
        types: [go]
        pass_filenames: false

 - id: cargo-clippy
        name: cargo-clippy
        entry: cargo clippy --workspace -- -D warnings
        language: system
        types: [rust]
        pass_filenames: false

 - id: flutter-analyze
        name: flutter analyze
        entry: bash -c "cd apps/client_flutter && flutter analyze"
        language: system
        types: [dart]
        pass_filenames: false

 - id: solhint
        name: solhint
        entry: solhint 'contracts/**/*.sol'
        language: system
        files: \.sol$
```

### Go Linter Configuration (`.golangci.yml`)
```yaml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters:
  enable:
 - gofumpt
 - govet
 - errcheck
 - staticcheck
 - gosec
 - revive
 - gocritic
 - bodyclose
 - unconvert

linters-settings:
  govet:
    check-shadowing: true
  gosec:
    severity: medium
    confidence: medium
```

### Python Configuration (`pyproject.toml`)
```toml
[tool.ruff]
target-version = "py311"
line-length = 100
select = ["E", "F", "W", "I", "B", "S", "UP", "PL", "RUF"]
ignore = ["E501"]

[tool.ruff.isort]
known-first-party = ["growww"]

[tool.mypy]
python_version = "3.11"
strict = true
disallow_untyped_defs = true
disallow_incomplete_defs = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true
warn_return_any = true
```

## Security & Compliance Notes
- **Security-Focused Static Analysis:** Automated inclusion of `gosec` (Go), `bandit` (Python via Ruff), `cargo audit` (Rust), and `slither` (Solidity) detects common vulnerabilities (SQL injection, unsafe deserialization, integer overflow, hardcoded secrets) before code review.
- **Strict Error Handling Enforcement:** In financial services, silent error swallowing is catastrophic. Linter rules (`errcheck` in Go, `no-unused-vars` in TS, and `#[must_use]` in Rust) strictly prevent unhandled error returns.
- **Deterministic Math Precision:** Linters enforce that all financial monetary calculations use integer or fixed-point representations (paise / cents), completely forbidding floating-point math (`float32`, `float64`, `double`) in financial calculations.

## Acceptance Criteria
- [ ] Comprehensive coding standards document (`docs/standards/coding_standards.md`) published.
- [ ] All 6 language configuration files (`.golangci.yml`, `rustfmt.toml`, `pyproject.toml`, `analysis_options.yaml`, `.eslintrc.js`, `.solhint.json`) configured in repository.
- [ ] `.pre-commit-config.yaml` created and verified executing across all language toolchains.
- [ ] Zero-warning CI policy enforced in all language build tasks.
- [ ] Security static analysis rules (gosec, bandit, slither) active and passing on all template files.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 106 (Monorepo Layout & Tooling Strategy).
- **Parallel Work:** Prompt 108 (Environment Strategy).
- **Blocks:** Prompt 803 (CI Pipeline Design), Prompt 905 (Security Testing Automation), All service implementation prompts.

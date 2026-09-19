# Shared Packages & Common Libraries

## Executive Overview
The `packages/` directory provides reusable, shared libraries across the polyglot monorepo, ensuring consistent Protobuf contracts, UI design tokens, cryptographic algorithms, and domain models.

## Package Inventory
1. `proto/`: Canonical Protocol Buffers v3 service definitions and data schemas for gRPC and Kafka.
2. `growww_ui/`: Flutter Obsidian dark-theme design kit, CustomPainter depth charts, and haptic widgets.
3. `web_ui/`: Tailwind CSS and React component library for Next.js web workstations.
4. `go_common/`: Shared Go utilities for OpenTelemetry tracing, structured logging, mTLS, and health probes.
5. `crypto_utils/`: Cryptographic utilities for EIP-712 structured signing, Merkle tree verification, and SPV proofs.
6. `domain_types/`: Shared canonical financial types (Order, Trade, Symbol, FixedPoint8) preventing conversion drift.

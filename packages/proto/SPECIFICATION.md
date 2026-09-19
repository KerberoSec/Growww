# Protocol Buffers Governance & Compatibility Specification

## Schema Guidelines
- All message fields must use explicit Proto3 typing with explicit numeric tags.
- Field tags 1 through 15 are strictly reserved for high-frequency repeated fields (order ID, price, quantity).
- Backward and forward compatibility strictly enforced using the `buf` CLI linter.

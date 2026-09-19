# Dedicated BTC/USDT Spot Order Gateway Specification

## Architecture Overview
A high-throughput Go microservice dedicated to validating, sequencing, and routing client spot orders for the primary Bitcoin/Tether trading pair.

## Pre-Execution Validation Pipeline
1. Check authenticated user session and EIP-712 cryptographic signature.
2. Validate order price against the +/- 10% fat-finger collar derived from the active mark price.
3. Reserve required quote balance (USDT for Buy) or base balance (BTC for Sell) in memory.
4. Assign monotonic 64-bit sequence number and publish order event to Kafka topic `market.orders.v1`.

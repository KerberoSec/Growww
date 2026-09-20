#!/usr/bin/env bash
# ==============================================================================
# Growww / NBSE Hardware IRQ Affinity & Core Pinning Script
# Isolates Cores 2-15 for the Matching Engine by routing all IRQs to Cores 0-1
# ==============================================================================

set -euo pipefail

# Housekeeping cores mask: Cores 0 and 1 = binary 11 = 0x3
HOUSEKEEPING_MASK="3"

echo "[INFO] Re-routing hardware IRQs to housekeeping cores (mask 0x${HOUSEKEEPING_MASK})..."

if [ -d /proc/irq ]; then
    for irq_smp in /proc/irq/*/smp_affinity; do
        if [ -w "${irq_smp}" ]; then
            echo "${HOUSEKEEPING_MASK}" > "${irq_smp}" 2>/dev/null || true
        fi
    done
    echo "[INFO] SMP affinity updated across available hardware IRQs."
else
    echo "[WARN] /proc/irq not found or not writable in this environment."
fi

# Disable irqbalance if active
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop irqbalance 2>/dev/null || true
    systemctl disable irqbalance 2>/dev/null || true
    echo "[INFO] Stopped irqbalance service."
fi

echo "[INFO] IRQ affinity configuration complete."

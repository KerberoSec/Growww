#!/usr/bin/env bash
# ==============================================================================
# Growww / NBSE DPDK & Hugepages Initialization Script
# Configures 1GB/2MB hugepages and sets up hugetlbfs for zero-copy ring buffers
# ==============================================================================

set -euo pipefail

HUGEPAGE_DIR="/mnt/huge"
NUM_2MB_PAGES=8192

echo "[INFO] Configuring system hugepages for exchange matching engine..."

# Ensure mount point exists
if [ ! -d "${HUGEPAGE_DIR}" ]; then
    echo "[INFO] Creating hugepages mount point ${HUGEPAGE_DIR}..."
    mkdir -p "${HUGEPAGE_DIR}"
fi

# Allocate 2MB hugepages if sysfs is writable
if [ -w /sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages ]; then
    echo "${NUM_2MB_PAGES}" > /sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages
    echo "[INFO] Allocated ${NUM_2MB_PAGES} x 2MB hugepages."
else
    echo "[WARN] /sys/kernel/mm/hugepages not writable (container/sandbox environment). Skipping sysfs write."
fi

# Mount hugetlbfs if not already mounted
if ! mountpoint -q "${HUGEPAGE_DIR}"; then
    if [ "$(id -u)" -eq 0 ]; then
        mount -t hugetlbfs nodev "${HUGEPAGE_DIR}"
        echo "[INFO] Mounted hugetlbfs on ${HUGEPAGE_DIR}"
    else
        echo "[WARN] Non-root execution: cannot mount hugetlbfs. Ensure this is run with sudo in production."
    fi
else
    echo "[INFO] Hugetlbfs already mounted on ${HUGEPAGE_DIR}"
fi

# Set permissions
if [ -d "${HUGEPAGE_DIR}" ]; then
    chmod 775 "${HUGEPAGE_DIR}" || true
fi

echo "[INFO] Hugepages configuration complete."

#!/usr/bin/env python3
"""
Growww / NBSE Ultra-Low Latency System Tuning Validation Suite
Verifies that kernel bootloader arguments, tuned profile configurations,
sysctl parameters, and IRQ isolation meet institutional trading standards.
"""

import os
import re
import sys
import configparser
from typing import Dict, List, Tuple


class TuningValidator:
    REQUIRED_SYSCTL_SETTINGS = {
        "kernel.sched_rt_runtime_us": "-1",
        "kernel.sched_migration_cost_ns": "5000000",
        "kernel.watchdog": "0",
        "net.core.rmem_max": "67108864",
        "net.core.wmem_max": "67108864",
        "net.core.busy_read": "50",
        "net.core.busy_poll": "50",
        "vm.swappiness": "0",
        "vm.dirty_ratio": "10",
        "vm.dirty_background_ratio": "5",
        "vm.nr_hugepages": "8192",
        "fs.file-max": "20971520",
    }

    REQUIRED_BOOT_ARGS = [
        "isolcpus=",
        "nohz_full=",
        "rcu_nocbs=",
        "irqaffinity=",
        "intel_idle.max_cstate=0",
        "processor.max_cstate=0",
        "idle=poll",
        "intel_pstate=performance",
        "transparent_hugepage=never",
        "tsc=reliable",
    ]

    def __init__(self, tuned_conf_path: str, sysctl_conf_path: str):
        self.tuned_conf_path = tuned_conf_path
        self.sysctl_conf_path = sysctl_conf_path

    def validate_tuned_conf(self) -> Tuple[bool, List[str]]:
        errors = []
        if not os.path.isfile(self.tuned_conf_path):
            return False, [f"Tuned configuration file not found at {self.tuned_conf_path}"]

        config = configparser.ConfigParser()
        try:
            config.read(self.tuned_conf_path)
        except Exception as e:
            return False, [f"Failed to parse tuned.conf: {e}"]

        # Check sections
        expected_sections = ["main", "cpu", "bootloader", "sysctl", "vm", "disk"]
        for sec in expected_sections:
            if not config.has_section(sec):
                errors.append(f"Missing required section [{sec}] in tuned.conf")

        # Check CPU governor
        if config.has_section("cpu"):
            gov = config.get("cpu", "governor", fallback="")
            if gov != "performance":
                errors.append(f"CPU governor must be 'performance', got '{gov}'")

        # Check Bootloader CMDLINE
        if config.has_section("bootloader"):
            cmdline = config.get("bootloader", "cmdline", fallback="")
            for arg in self.REQUIRED_BOOT_ARGS:
                if arg not in cmdline:
                    errors.append(f"Missing required kernel bootloader arg '{arg}' in cmdline")

        # Check sysctl values in tuned.conf
        if config.has_section("sysctl"):
            for k, expected_v in self.REQUIRED_SYSCTL_SETTINGS.items():
                actual_v = config.get("sysctl", k, fallback=None)
                if actual_v is None:
                    errors.append(f"Missing sysctl key '{k}' in [sysctl] section of tuned.conf")
                elif actual_v.strip() != expected_v.strip():
                    errors.append(f"Sysctl key '{k}' expected '{expected_v}', got '{actual_v}'")

        return len(errors) == 0, errors

    def validate_sysctl_conf(self) -> Tuple[bool, List[str]]:
        errors = []
        if not os.path.isfile(self.sysctl_conf_path):
            return False, [f"Sysctl configuration file not found at {self.sysctl_conf_path}"]

        parsed_sysctl: Dict[str, str] = {}
        with open(self.sysctl_conf_path, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or line.startswith(";"):
                    continue
                if "=" in line:
                    k, v = line.split("=", 1)
                    parsed_sysctl[k.strip()] = v.strip()

        for k, expected_v in self.REQUIRED_SYSCTL_SETTINGS.items():
            if k not in parsed_sysctl:
                errors.append(f"Missing sysctl key '{k}' in {self.sysctl_conf_path}")
            elif parsed_sysctl[k] != expected_v:
                errors.append(f"Sysctl key '{k}' expected '{expected_v}', got '{parsed_sysctl[k]}'")

        return len(errors) == 0, errors

    def run_all_checks(self) -> bool:
        print("=== Growww Ultra-Low Latency System Tuning Validation ===")
        all_passed = True

        tuned_ok, tuned_errs = self.validate_tuned_conf()
        if tuned_ok:
            print("[PASS] tuned.conf profile format, CPU governor, bootloader cmdline, and sysctls verified.")
        else:
            print("[FAIL] tuned.conf validation failed:")
            for err in tuned_errs:
                print(f"  - {err}")
            all_passed = False

        sysctl_ok, sysctl_errs = self.validate_sysctl_conf()
        if sysctl_ok:
            print("[PASS] 99-exchange-latency.conf sysctl definitions verified.")
        else:
            print("[FAIL] sysctl configuration validation failed:")
            for err in sysctl_errs:
                print(f"  - {err}")
            all_passed = False

        if all_passed:
            print("\n>>> ALL TUNING SPECIFICATIONS VALIDATED SUCCESSFULLY (100% PASS) <<<\n")
        return all_passed


if __name__ == "__main__":
    base_dir = os.path.dirname(os.path.abspath(__file__))
    tuned_path = os.path.join(base_dir, "tuned-profiles", "ultra-low-latency-exchange", "tuned.conf")
    sysctl_path = os.path.join(base_dir, "sysctl.d", "99-exchange-latency.conf")

    validator = TuningValidator(tuned_path, sysctl_path)
    success = validator.run_all_checks()
    sys.exit(0 if success else 1)

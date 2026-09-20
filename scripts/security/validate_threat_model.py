#!/usr/bin/env python3
"""
Automated Threat Model Validator (Prompt 701)
Validates that docs/security/threat_model.md contains zero unmitigated High/Critical risks
and adheres strictly to SEBI CSCRF STRIDE guidelines.
"""

import os
import re
import sys

def parse_threat_model(file_path: str):
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"Threat model file not found: {file_path}")

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # Match markdown table rows: | **TH-xxx** | Category | Component | ...
    pattern = re.compile(
        r"\|\s*\*\*(TH-\d{3})\*\*\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|"
    )

    threats = []
    for match in pattern.finditer(content):
        t_id, category, component, desc, cvss, dread, mitigation, prompt_ref, status = match.groups()
        threats.append({
            "id": t_id.strip(),
            "category": category.strip(),
            "component": component.strip(),
            "description": desc.strip(),
            "cvss": cvss.strip(),
            "dread": dread.strip(),
            "mitigation": mitigation.strip(),
            "prompt": prompt_ref.strip(),
            "status": status.strip().replace("*", "")
        })

    return threats

def validate_threats(threats):
    valid_stride = {"Spoofing", "Tampering", "Repudiation", "Information Disclosure", "Denial of Service", "Elevation of Privilege"}
    errors = []

    if len(threats) == 0:
        errors.append("No threats parsed from threat model matrix.")

    for t in threats:
        # Check STRIDE category
        if t["category"] not in valid_stride:
            errors.append(f"Threat {t['id']}: Invalid STRIDE category '{t['category']}'")

        # Check mitigation status
        if t["status"].lower() != "mitigated":
            errors.append(f"Threat {t['id']}: Status is '{t['status']}', expected 'Mitigated'")

        # Check prompt reference
        if not re.search(r"Prompt\s+\d+", t["prompt"]):
            errors.append(f"Threat {t['id']}: Missing implementation prompt reference in '{t['prompt']}'")

        # Check mitigation description length
        if len(t["mitigation"]) < 10:
            errors.append(f"Threat {t['id']}: Insufficient mitigation detail")

    return errors

def main():
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
    threat_doc = os.path.join(repo_root, "docs", "security", "threat_model.md")

    try:
        threats = parse_threat_model(threat_doc)
        print(f"[ThreatModelValidator] Parsed {len(threats)} STRIDE threat definitions from {threat_doc}")
        errors = validate_threats(threats)

        if errors:
            print(f"[ThreatModelValidator] FAILED with {len(errors)} errors:")
            for err in errors:
                print(f"  - {err}")
            sys.exit(1)
        else:
            print("[ThreatModelValidator] SUCCESS: All threats validated. 0 unmitigated risks found.")
            sys.exit(0)

    except Exception as e:
        print(f"[ThreatModelValidator] FATAL: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main()

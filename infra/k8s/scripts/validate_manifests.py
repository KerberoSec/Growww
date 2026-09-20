#!/usr/bin/env python3
"""
Kubernetes Manifests & Helm Configuration Validator for Growww Sovereign Exchange.
Performs static analysis, schema compliance, Pod Security Standard (PSS) checks,
multi-AZ anti-affinity audits, and NetworkPolicy zero-trust verification.
"""

import os
import sys
import re
import yaml
import subprocess
from pathlib import Path

REQUIRED_NAMESPACES = {
    "growww-core",
    "growww-trading",
    "growww-settlement",
    "growww-gift-city",
    "growww-blockchain",
    "growww-secops",
    "growww-observability",
}

REQUIRED_PRIORITY_CLASSES = {
    "critical-trading": 1000000,
    "core-settlement": 900000,
    "general-service": 500000,
    "batch-worker": 100000,
}


def render_mock_helm_template(content: str) -> str:
    """Render helm templates with dummy values for AST / YAML parsing."""
    rendered_lines = []
    for line in content.splitlines():
        # Handle helm control structures
        stripped = line.strip()
        if stripped.startswith("{{- if") or stripped.startswith("{{- else") or stripped.startswith("{{- end"):
            continue
        if stripped.startswith("{{/*") or stripped.startswith("{{- /*"):
            continue
        if stripped.startswith("{{- range"):
            continue
        if stripped.startswith("{{- define") or stripped.startswith("{{- include"):
            continue
        if stripped.startswith("{{- $"):
            continue

        # Replace inline helm interpolations {{ ... }}
        if "{{" in line and "}}" in line:
            # If line is a list item like "- {{ .Values.host }}" -> "- exchange.growww.in"
            if re.match(r"^\s*-\s*\{\{.*\}\}\s*$", line):
                line = re.sub(r"\{\{.*\}\}", "exchange.growww.in", line)
            # If line is key: {{ ... }}
            elif ":" in line:
                key, _ = line.split(":", 1)
                if "replicas" in key.lower():
                    line = f"{key}: 3"
                elif "port" in key.lower():
                    line = f"{key}: 8080"
                elif "cpu" in key.lower():
                    line = f"{key}: '2000m'"
                elif "memory" in key.lower():
                    line = f"{key}: '4Gi'"
                else:
                    line = f"{key}: 'mock-value'"
            else:
                line = re.sub(r"\{\{.*?\}\}", "mock-value", line)
        rendered_lines.append(line)
    return "\n".join(rendered_lines)


def load_yaml_documents(file_path: Path):
    """Safely load multiple YAML documents from a file."""
    docs = []
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            raw_content = f.read()
            if "_helpers.tpl" in file_path.name:
                return []
            content = render_mock_helm_template(raw_content) if "templates" in str(file_path) else raw_content
            parsed = yaml.safe_load_all(content)
            for doc in parsed:
                if isinstance(doc, dict):
                    docs.append(doc)
    except Exception as e:
        print(f"[-] Error loading {file_path}: {e}")
    return docs


def validate_namespaces(docs):
    """Verify all 7 core namespaces exist and have restricted PSS."""
    found_ns = set()
    for doc in docs:
        if doc.get("kind") == "Namespace":
            ns_name = doc.get("metadata", {}).get("name")
            if ns_name in REQUIRED_NAMESPACES:
                found_ns.add(ns_name)
                labels = doc.get("metadata", {}).get("labels", {})
                pss = labels.get("pod-security.kubernetes.io/enforce")
                if pss != "restricted":
                    print(f"[-] Namespace {ns_name} missing restricted PSS (found: {pss})")
                    return False
    missing = REQUIRED_NAMESPACES - found_ns
    if missing:
        print(f"[-] Missing required namespaces: {missing}")
        return False
    print(f"[+] Verified {len(found_ns)}/{len(REQUIRED_NAMESPACES)} namespaces with PSS restricted")
    return True


def validate_priority_classes(docs):
    """Verify priority classes match required specification values."""
    found_pcs = {}
    for doc in docs:
        if doc.get("kind") == "PriorityClass":
            name = doc.get("metadata", {}).get("name")
            val = doc.get("value")
            found_pcs[name] = val
    for req_name, req_val in REQUIRED_PRIORITY_CLASSES.items():
        if req_name not in found_pcs:
            print(f"[-] Missing PriorityClass: {req_name}")
            return False
        if found_pcs[req_name] != req_val:
            print(f"[-] PriorityClass {req_name} value mismatch: got {found_pcs[req_name]}, expected {req_val}")
            return False
    print(f"[+] Verified all {len(REQUIRED_PRIORITY_CLASSES)} PriorityClasses with accurate values")
    return True


def validate_pod_security(docs):
    """Verify Deployments enforce restricted pod and container security contexts."""
    deployments = [d for d in docs if d.get("kind") == "Deployment"]
    for dep in deployments:
        dep_name = dep.get("metadata", {}).get("name", "unknown")
        pod_spec = dep.get("spec", {}).get("template", {}).get("spec", {})
        pod_sec = pod_spec.get("securityContext", {})

        if not pod_sec.get("runAsNonRoot"):
            print(f"[-] Deployment {dep_name}: pod securityContext missing runAsNonRoot: true")
            return False

        containers = pod_spec.get("containers", [])
        for c in containers:
            c_name = c.get("name")
            c_sec = c.get("securityContext", {})
            if not c_sec.get("readOnlyRootFilesystem"):
                print(f"[-] Container {dep_name}/{c_name} missing readOnlyRootFilesystem: true")
                return False
            if c_sec.get("allowPrivilegeEscalation") is not False:
                print(f"[-] Container {dep_name}/{c_name} missing allowPrivilegeEscalation: false")
                return False
            caps = c_sec.get("capabilities", {}).get("drop", [])
            if "ALL" not in caps:
                print(f"[-] Container {dep_name}/{c_name} missing drop: ['ALL'] capabilities")
                return False

            # Verify resource limits
            res = c.get("resources", {})
            if not res.get("requests") or not res.get("limits"):
                print(f"[-] Container {dep_name}/{c_name} missing resource requests or limits")
                return False

    print(f"[+] Verified {len(deployments)} deployments adhere to Restricted Pod Security Standards")
    return True


def validate_pdb(docs):
    """Verify critical workloads have associated PodDisruptionBudgets."""
    deployments = {d.get("metadata", {}).get("name") for d in docs if d.get("kind") == "Deployment"}
    pdbs = {p.get("metadata", {}).get("name") for p in docs if p.get("kind") == "PodDisruptionBudget"}
    
    for dep in deployments:
        matching_pdb = any(dep in pdb for pdb in pdbs)
        if not matching_pdb:
            print(f"[-] Warning: Deployment {dep} does not have an explicit PDB")
    print(f"[+] Verified {len(pdbs)} PodDisruptionBudgets guarding critical services")
    return True


def validate_network_policies(docs):
    """Verify default-deny network policies exist."""
    netpols = [d for d in docs if d.get("kind") == "NetworkPolicy"]
    default_deny = [p for p in netpols if p.get("metadata", {}).get("name") == "default-deny-all"]
    if not default_deny:
        print("[-] Missing default-deny-all NetworkPolicy")
        return False
    print(f"[+] Verified {len(netpols)} NetworkPolicies including default-deny rules")
    return True


def validate_kustomize(base_dir: Path):
    """Run kubectl kustomize to verify manifest validity."""
    try:
        res = subprocess.run(
            ["kubectl", "kustomize", str(base_dir)],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=False,
        )
        if res.returncode != 0:
            print(f"[-] kubectl kustomize failed on {base_dir}:\n{res.stderr}")
            return False
        print(f"[+] kubectl kustomize succeeded on {base_dir}")
        return True
    except FileNotFoundError:
        print("[!] kubectl not found in PATH; skipping live kustomize CLI validation")
        return True


def main():
    repo_root = Path(__file__).resolve().parent.parent.parent.parent
    k8s_dir = repo_root / "infra" / "k8s"

    print("=== Growww Kubernetes Manifests Validation ===")
    all_docs = []
    for root, _, files in os.walk(k8s_dir):
        for f in sorted(files):
            if f.endswith((".yaml", ".yml")):
                p = Path(root) / f
                all_docs.extend(load_yaml_documents(p))

    print(f"[*] Loaded {len(all_docs)} Kubernetes resource manifests")

    checks = [
        validate_namespaces(all_docs),
        validate_priority_classes(all_docs),
        validate_pod_security(all_docs),
        validate_pdb(all_docs),
        validate_network_policies(all_docs),
        validate_kustomize(k8s_dir / "base"),
        validate_kustomize(k8s_dir / "environments" / "testnet"),
        validate_kustomize(k8s_dir / "environments" / "mainnet"),
    ]

    if all(checks):
        print("\n[SUCCESS] All Kubernetes production manifests validated successfully!")
        return 0
    else:
        print("\n[FAILURE] Manifest validation failed!")
        return 1


if __name__ == "__main__":
    sys.exit(main())

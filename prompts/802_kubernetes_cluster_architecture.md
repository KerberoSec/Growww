# 802 - Multi-Tenant Kubernetes Cluster Architecture & Namespace Segmentation

## Purpose
Production-grade financial and regulated blockchain platforms require ironclad isolation, zero-trust network policy enforcement, predictable computing tiers, and resilient high-availability orchestration. This prompt establishes the master Kubernetes (EKS / GKE) cluster architecture for Growww across all deployment tiers (`dev`, `staging`, `prod-mumbai`, `prod-gift-city`).

By segmenting workloads across strictly isolated namespaces - separating domestic SEBI-regulated custody integrations, high-frequency trading matching engines, international IFSCA gateways, sensitive key management services, and permissioned Hyperledger Besu consortium nodes - the architecture guarantees zero cross-tenant contamination, enforces sub-millisecond network QoS for trading pipelines, and meets regulatory compliance mandates for financial infrastructure resilience.

## What You Are Building
Declarative Kubernetes cluster topologies, Helm umbrella charts, and governance manifests:
- `deployments/k8s/namespaces/`: Declarative namespace definitions with Pod Security Standards (`enforce: restricted`) and resource boundaries (`ResourceQuota`, `LimitRange`).
- `deployments/k8s/network-policies/`: Zero-trust Calico / Cilium eBPF NetworkPolicies enforcing default-deny ingress and egress across all namespaces, with explicit whitelist rules for inter-service communication.
- `deployments/helm/growww-platform/`: Unified Helm v3 umbrella chart orchestrating core microservices, sidecars, and stateful operators.
- `deployments/k8s/node-pools/`: Node pool topology specifications, taints, tolerations, and node affinity rules mapping specialized workloads (e.g. ultra-low latency trading engine pinned to isolated CPU cores; Besu nodes pinned to high-IOPS NVMe instances).
- `deployments/k8s/autoscaling/`: KEDA (Kubernetes Event-driven Autoscaling) ScaledObjects and Horizontal Pod Autoscalers (HPA) driven by custom metrics (Kafka consumer group lag, order queue depth, RPC request rates).
- `deployments/k8s/reliability/`: PodDisruptionBudgets (PDBs) and PriorityClasses guaranteeing zero downtime during node drain/upgrade operations.

## Scope Boundaries
- **In Scope:**
 - Namespace segmentation: `growww-core`, `growww-trading`, `growww-settlement`, `growww-gift-city`, `growww-blockchain`, `growww-secops`, and `growww-observability`.
 - Zero-trust NetworkPolicy definitions restricting all namespace-to-namespace and external egress traffic.
 - Dedicated node pool topologies (General compute, Memory-optimized, CPU-pinned trading, NVMe blockchain storage).
 - High availability patterns: Pod anti-affinity across Availability Zones (AZs), PDBs, and PriorityClasses.
 - Custom event-driven autoscaling via KEDA and metric-based HPA.
- **Out of Scope / Handled Elsewhere:**
 - Cloud infrastructure provisioning (VPC, EKS/GKE cluster creation via Terraform - Prompt 805).
 - GitOps automated synchronization via ArgoCD (Prompt 804).
 - OpenTelemetry distributed collector configuration (Prompt 806).

## Technology to Use
- **Kubernetes 1.30+**: Enterprise container orchestration platform. Justification: Provides mature API stability, native Pod Security Admission (`restricted` profile), Gateway API support, and robust multi-AZ failover mechanics.
- **Helm v3**: Package manager for templated, reproducible microservice chart deployments across environments.
- **Cilium CNI (v1.15+) / Calico**: eBPF-based high-performance networking and network security. Justification: Cilium provides wire-speed eBPF packet filtering, transparent mTLS encryption via WireGuard, and fine-grained L7 DNS/HTTP egress policies with zero iptables overhead.
- **KEDA v2.14+**: Event-driven autoscaler. Justification: Allows scaling matching engine consumers and settlement workers directly on Kafka consumer lag and queue depths rather than lagging CPU metrics.

## Backend / Infra Touchpoints
- **AWS EKS / GCP GKE**: Managed control planes running across 3 Availability Zones (AZs) in Mumbai (`ap-south-1`) and GIFT City local zones.
- **Storage Classes**: AWS `gp3` (standard microservices), Local NVMe / `io2` (PostgreSQL, Kafka, Hyperledger Besu blockchain nodes).
- **AWS ALB / NGINX Ingress Controller**: Ingress gateway handling external SSL termination and routing.
- **HashiCorp Vault Agent Injector**: Sidecar injector providing ephemeral secrets to pods in `growww-core` and `growww-trading`.

## Blockchain Interaction
The `growww-blockchain` namespace hosts dedicated stateful Hyperledger Besu validator nodes and RPC relayers:
- **Node Pinning & Storage**: Deploys as a `StatefulSet` with node affinity targeting `node-type=blockchain-validator`, utilizing local NVMe `StorageClass` with `VolumeBindingMode: WaitForFirstConsumer` for maximum I/O throughput.
- **Consensus Isolation**: QBFT consensus P2P port (`30303`) is exposed via a headless Kubernetes Service with explicit Calico network policies permitting P2P ingress exclusively from authorized consortium validator IP ranges.
- **RPC Ingress Control**: JSON-RPC port (`8545`) and WebSocket port (`8546`) are locked down to accept connections only from pods inside `growww-settlement` and `growww-observability` namespaces.

## Step-by-Step Build Instructions
1. Scaffold directory structure `deployments/k8s/{namespaces,network-policies,helm,node-pools,autoscaling,reliability}`.
2. Define declarative namespaces with metadata labels and Pod Security Standard annotations (`pod-security.kubernetes.io/enforce: restricted`).
3. Define PriorityClasses: `critical-trading` (value 1000000), `core-settlement` (value 900000), `general-service` (value 500000), `batch-worker` (value 100000).
4. Implement baseline Calico/Cilium `NetworkPolicy` for every namespace with default deny-all ingress and egress.
5. Create targeted egress rules allowing pods in `growww-trading` and `growww-settlement` to connect to RDS PostgreSQL, Redis, and internal Kafka brokers over specific ports.
6. Create cross-namespace whitelist rules allowing `growww-settlement` pods to access Besu JSON-RPC in `growww-blockchain`.
7. Configure `StorageClass` definitions for high-IOPS NVMe (`local-nvme-block`) and general block storage (`aws-ebs-gp3`).
8. Create `ResourceQuota` and `LimitRange` manifests for each namespace to avoid noisy-neighbor resource starvation.
9. Implement PodDisruptionBudgets (PDBs) ensuring `minAvailable: 66%` or `maxUnavailable: 1` across all critical stateful and stateless services.
10. Define KEDA `ScaledObject` templates triggering horizontal scaling on Kafka lag thresholds (`lagThreshold: 50`).
11. Build the master Helm chart `deployments/helm/growww-platform` with environment-specific values files (`values-dev.yaml`, `values-staging.yaml`, `values-prod.yaml`).
12. Configure node affinity and pod anti-affinity rules across availability zones (`topologyKey: topology.kubernetes.io/zone`).
13. Validate cluster dry-run with `kubeconform` and `polaris` static validation tools.
14. Test pod eviction simulation to verify PDBs and zero-downtime rolling upgrades.

## Interfaces / Contracts
```yaml
# deployments/k8s/network-policies/settlement-to-blockchain-netpol.yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-settlement-to-besu-rpc
  namespace: growww-blockchain
spec:
  endpointSelector:
    matchLabels:
      app.kubernetes.io/name: besu-validator
  ingress:
 - fromEndpoints:
 - matchLabels:
            "k8s:io.kubernetes.pod.namespace": growww-settlement
            app.kubernetes.io/name: trade-settlement-service
      toPorts:
 - ports:
 - port: "8545"
              protocol: TCP
 - port: "8546"
              protocol: TCP
 - fromCIDR:
 - "10.100.20.0/24" # Consortium Validator P2P Peering Subnet
      toPorts:
 - ports:
 - port: "30303"
              protocol: TCP
 - port: "30303"
              protocol: UDP
```

```yaml
# deployments/k8s/autoscaling/settlement-keda-scaler.yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: settlement-worker-scaler
  namespace: growww-settlement
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: trade-settlement-worker
  minReplicaCount: 3
  maxReplicaCount: 20
  cooldownPeriod: 60
  triggers:
 - type: kafka
      metadata:
        bootstrapServers: kafka-cluster-kafka-bootstrap.growww-core:9092
        consumerGroup: settlement-orchestrator-group
        topic: trades.executed.v1
        lagThreshold: "25"
```

## Security & Compliance Notes
- All pods must comply with the Kubernetes Restricted Pod Security Standard: `runAsNonRoot: true`, `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, and `capabilities: { drop: ["ALL"] }`.
- Egress to external third parties (NSDL, CDSL, Payment Aggregators) must be restricted to dedicated egress gateway proxies (`growww-egress-gateway`) equipped with static NAT Elastic IPs registered with SEBI/depository firewalls.
- No direct internet ingress is permitted into any namespace except through the hardened AWS WAF-fronted Ingress controller.

## Acceptance Criteria
- [ ] All 7 core namespaces are provisioned with enforced `restricted` Pod Security Standards.
- [ ] Calico/Cilium network policies block all unauthorized inter-namespace traffic while permitting verified microservice pathways.
- [ ] Pod Disruption Budgets (PDBs) and multi-AZ anti-affinity rules prevent service outages during cluster node upgrades.
- [ ] KEDA autoscalers dynamically scale consumers in response to simulated Kafka topic load spikes.
- [ ] Hyperledger Besu validator StatefulSets successfully mount NVMe storage and isolate P2P and RPC traffic per specifications.
- [ ] Cluster manifests pass `kubeconform` validation without schema errors.

## Suggested Order / Dependencies
- Prerequisites: Prompt 101 (Architecture Overview), Prompt 102 (Service Boundaries), Prompt 805 (Terraform EKS Provisioning).
- Parallel Tasks: Prompt 804 (GitOps CD), Prompt 806 (Observability Stack).

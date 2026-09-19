# 809 - FinOps Cloud Cost Monitoring, Allocation & Resource Optimization

## Purpose
Operating high-frequency financial platforms, multi-region Kubernetes clusters, real-time message buses, and dedicated blockchain validator infrastructure incurs significant cloud and hardware expenditure. Without strict FinOps governance, cloud costs can grow uncontrollably, and expenses cannot be accurately attributed between the domestic SEBI-regulated entity and the international GIFT City IFSCA gateway entity.

This prompt establishes Growww's FinOps cost monitoring, financial attribution, automated resource right-sizing, and infrastructure optimization framework. By deploying Kubecost / OpenCost inside Kubernetes, enforcing strict cost-allocation tagging via Terraform, automating dynamic node provisioning via Karpenter, and shifting cold blockchain and audit data into tiered storage classes, the system maximizes infrastructure efficiency while maintaining ironclad SLA and regulatory performance standards.

## What You Are Building
A comprehensive FinOps governance and automated cost optimization framework:
- `deployments/helm/kubecost/`: Helm deployment for Kubecost / OpenCost providing real-time in-cluster cost allocation down to the Kubernetes namespace, deployment, and pod level.
- `infra/terraform/modules/cost-governance/`: Terraform module enforcing AWS/GCP Cost Allocation Tags (`CostCenter`, `Environment`, `LegalEntity`, `ServiceTier`, `Owner`) across all cloud resources.
- `deployments/k8s/autoscaling/karpenter-nodepools.yaml`: Karpenter NodePool configurations dynamically provisioning compute instances with intelligent CPU/memory bin-packing, Graviton (ARM64) instance utilization, and spot instance integration for stateless non-critical batch workers.
- `scripts/finops/storage-lifecycle-manager.sh`: Automated storage tiering script transitioning Kafka logs, historical PostgreSQL partitions, and Hyperledger Besu block archives from high-cost NVMe/gp3 storage to AWS S3 Standard, S3 Infrequent Access (IA), and Glacier Deep Archive.
- `.github/workflows/ci-infracost.yml`: Pull request CI action using Infracost to calculate and display infrastructure cost diffs directly on Terraform PRs before merging.
- `docs/finops/cost_attribution_model.md`: Official financial cost-sharing agreement and attribution model between the Domestic Custody entity and the GIFT City Gateway entity.

## Scope Boundaries
- **In Scope:**
 - In-cluster Kubernetes cost monitoring (Kubecost) and Prometheus metrics integration.
 - Multi-entity cost allocation between Domestic SEBI Entity and GIFT City IFSCA Entity.
 - Karpenter just-in-time compute provisioning and node right-sizing.
 - Storage tiering automation for databases, Kafka logs, and blockchain archival blocks.
 - Automated PR-level infrastructure cost estimation via Infracost.
 - AWS/GCP Budgets and Cost Anomaly Detection alerting to Slack.
- **Out of Scope / Handled Elsewhere:**
 - Platform investor fee calculation engine (Prompt 210).
 - Base Terraform cloud resource provisioning (Prompt 805).
 - Production database replication topologies (Prompt 401, 406).

## Technology to Use
- **Kubecost / OpenCost (v2.x)**: Open-source Kubernetes cost monitoring standard. Justification: Accurately maps cloud provider billing rates (AWS CUR / GCP Cloud Billing) to in-cluster Kubernetes CPU, RAM, GPU, storage, and network egress metrics in real time.
- **Karpenter v0.37+**: High-performance, declarative Kubernetes node autoscaler. Justification: Provisions just-in-time, correctly sized compute instances in seconds directly from the AWS/GCP API, eliminating the slow multi-minute warmup and wasted headroom of traditional Auto Scaling Groups (Cluster Autoscaler).
- **Infracost**: Shift-left cloud cost estimation tool for Terraform.
- **AWS Cost & Usage Report (CUR) / AWS Budgets**: Native cloud cost attribution and anomaly detection services.

## Backend / Infra Touchpoints
- **AWS Billing / GCP Billing API**: Ingests actual negotiated cloud rates and savings plans.
- **Kubernetes API**: Scraped by Kubecost and managed by Karpenter.
- **Slack `#finops-alerts`**: Channel receiving cost anomaly notifications and weekly spend summaries.
- **AWS S3 / EFS / EBS**: Storage volumes subject to automated lifecycle tiering policies.

## Blockchain Interaction
Optimizes compute and storage footprints for the Hyperledger Besu consortium network without compromising consensus integrity:
- **Separation of Hot vs Cold Chain Data**: Hyperledger Besu validator nodes require ultra-high IOPS NVMe storage for the active world state trie and recent block headers. The storage lifecycle automation snapshots and prunes historical block bodies older than 180 days into compressed cold S3 storage archives, reducing high-cost `io2`/NVMe EBS footprints by over 70%.
- **Validator Node Tiering Guarantee**: FinOps policies strictly forbid placing Hyperledger Besu validator nodes or high-frequency order matching engines on Spot/Preemptible instances. Validators are permanently pinned to dedicated On-Demand or 3-Year Reserved Instances (Savings Plans) to guarantee 100% consensus uptime and zero node eviction.
- **RPC Relayer Right-Sizing**: Non-validating read-only Besu JSON-RPC read replicas are deployed on auto-scaling Graviton ARM64 instances (`c7g.xlarge`), dynamically scaling up during market trading hours (9:15 AM - 3:30 PM IST) and scaling down during overnight hours.

## Step-by-Step Build Instructions
1. Scaffold directory `deployments/helm/kubecost/`, `infra/terraform/modules/cost-governance/`, and `deployments/k8s/autoscaling/`.
2. Configure AWS Cost Allocation Tags activation in AWS Management Console via Terraform (`LegalEntity`, `CostCenter`, `Environment`, `ServiceTier`).
3. Deploy Kubecost Helm chart in `growww-observability` namespace with AWS Cost and Usage Report (CUR) S3 integration.
4. Configure Kubecost multi-tenant reports segmenting costs between domestic services (`growww-core`, `growww-settlement`, `growww-trading`) and GIFT City services (`growww-gift-city`).
5. Deploy Karpenter controller into the EKS cluster with IAM roles and instance profiles.
6. Author Karpenter `NodePool` manifests defining instance families (`c7g`, `m7g`, `r7g` Graviton instances for 20% better price-performance), AZ distribution, and disruption policies (`consolidationPolicy: WhenUnderutilized`).
7. Create separate Karpenter NodePools: (a) `general-spot` for batch analytics workers, (b) `dedicated-ondemand` for trading matching engine and blockchain validator nodes.
8. Implement Vertical Pod Autoscaler (VPA) in recommendation mode to analyze historical CPU/memory requests and identify over-provisioned microservices.
9. Write `scripts/finops/storage-lifecycle-manager.sh` to automate S3 lifecycle transitions: S3 Standard (30 days) -> S3 Standard-IA (90 days) -> S3 Glacier Flexible (365 days) -> S3 Glacier Deep Archive (7 years).
10. Integrate Infracost in `.github/workflows/ci-infracost.yml` to post Terraform PR cost comments and enforce budget guardrails (failing PRs that increase monthly spend beyond authorized thresholds without manager approval).
11. Configure AWS Cost Anomaly Detection to alert `#finops-alerts` on Slack whenever daily spend deviates by >$150 from historical baselines.
12. Establish the official cost attribution model spreadsheet and document cross-entity intercompany invoicing rules.
13. Execute a simulated workload scale test to verify Karpenter consolidates empty nodes and terminates unused compute instances within 5 minutes.
14. Review initial Kubecost reports and adjust Kubernetes container resource requests and limits to eliminate idle waste.

## Interfaces / Contracts
```yaml
# deployments/k8s/autoscaling/karpenter-nodepools.yaml (Excerpt)
apiVersion: karpenter.sh/v1beta1
kind: NodePool
metadata:
  name: growww-general-compute
spec:
  template:
    spec:
      requirements:
 - key: "karpenter.k8s.aws/instance-category"
          operator: In
          values: ["c", "m", "r"]
 - key: "karpenter.k8s.aws/instance-cpu"
          operator: In
          values: ["4", "8", "16"]
 - key: "karpenter.k8s.aws/instance-generation"
          operator: Gt
          values: ["6"]
 - key: "kubernetes.io/arch"
          operator: In
          values: ["arm64", "amd64"]
 - key: "karpenter.sh/capacity-type"
          operator: In
          values: ["on-demand"]
      nodeClassRef:
        name: default-nodeclass
  disruption:
    consolidationPolicy: WhenUnderutilized
    consolidateAfter: 30s
    expireAfter: 720h # 30 days node recycling
```

```yaml
# .github/workflows/ci-infracost.yml (Excerpt)
name: Infracost Terraform PR Check

on:
  pull_request:
    paths:
 - 'infra/terraform/**'

jobs:
  infracost:
    runs-on: ubuntu-latest
    steps:
 - uses: actions/checkout@v4
 - uses: infracost/actions/setup@v3
        with:
          api-key: ${{ secrets.INFRACOST_API_KEY }}
 - name: Generate Infracost Cost Diff
        run: |
          infracost breakdown --path=infra/terraform/environments/prod-mumbai \
                              --format=json \
                              --out-file=infracost.json
          infracost comment github --path=infracost.json \
                                  --repo=$GITHUB_REPOSITORY \
                                  --github-token=${{ secrets.GITHUB_TOKEN }} \
                                  --pull-request=${{ github.event.pull_request.number }} \
                                  --behavior=update
```

## Security & Compliance Notes
- Strict Non-Intermingling: Cost optimization must never co-locate domestic SEBI-regulated workloads and international GIFT City workloads on the same physical underlying compute instances where network namespace escape vulnerabilities could pose compliance breaches.
- Production Stability Over Savings: Financial trading engines, settlement workers, and Besu validator nodes must never run on Spot/Preemptible instances.
- Data Retention Adherence: Automated storage tiering to Glacier Deep Archive must respect the 7-year WORM compliance retention lock (Prompt 807).

## Acceptance Criteria
- [ ] Kubecost is operational and accurately attributes costs per Kubernetes namespace and legal entity.
- [ ] 100% of cloud resources provisioned via Terraform contain valid Cost Allocation Tags.
- [ ] Karpenter dynamically provisions Graviton and x86 compute nodes, consolidating underutilized nodes within 5 minutes.
- [ ] Infracost CI workflow posts automated monthly cost impact breakdowns on all Terraform pull requests.
- [ ] AWS Cost Anomaly Detection triggers real-time alerts to Slack upon simulated unexpected resource spikes.
- [ ] Storage lifecycle automation tiers cold Kafka and blockchain data without impacting active queries.

## Suggested Order / Dependencies
- Prerequisites: Prompt 108 (Environment Strategy), Prompt 802 (Kubernetes Architecture), Prompt 805 (Terraform Cloud Infra).
- Parallel Tasks: Prompt 804 (GitOps CD), Prompt 810 (Release Management).

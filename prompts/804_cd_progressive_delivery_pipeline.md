# 804 - GitOps Continuous Delivery & Progressive Canary Rollouts

## Purpose
Financial exchange systems and regulated settlement engines cannot tolerate deployment downtime, unexpected regressions, or operational disruptions. The Continuous Delivery (CD) system implements a declarative GitOps model powered by ArgoCD and Argo Rollouts, enforcing zero-downtime, progressive canary and blue-green deployments across all Growww environments (`dev`, `staging`, `prod-mumbai`, `prod-gift-city`).

By pairing automated metric-driven canary verification (evaluating real-time HTTP error rates, order latency percentiles, and Kafka consumer lag) with audited promotion gates, this architecture ensures that new releases are proven healthy on small percentages of production traffic before full cutover, automatically rolling back within seconds if anomalies occur.

## What You Are Building
A production GitOps delivery and progressive rollout pipeline:
- `deployments/gitops/root-application.yaml`: ArgoCD "App-of-Apps" master declaration orchestrating infrastructure and microservice applications.
- `deployments/gitops/applicationsets/`: Dynamic ArgoCD ApplicationSets generating cluster-specific application instances from Git branch/directory matrices.
- `deployments/k8s/rollouts/`: Argo Rollouts Custom Resource Definitions (`Rollout` replacing standard `Deployment`) for critical trading and settlement services.
- `deployments/k8s/analysis/`: Prometheus-driven `AnalysisTemplate` resources querying Prometheus / Mimir metrics to automatically validate canary step health.
- `deployments/k8s/overlays/`: Kustomize environment overlays managing image tags, replica counts, ingress endpoints, and environment variables across `dev`, `staging`, `prod-mumbai`, and `prod-gift-city`.
- `scripts/gitops/promote-release.sh`: Automated GitOps promotion tool updating target image tags in environment repositories following approved change management tickets.

## Scope Boundaries
- **In Scope:**
 - GitOps repository architecture and ArgoCD synchronization engines.
 - Progressive canary delivery with step weights (5% -> 20% -> 50% -> 100%) and automated rollback triggers.
 - Blue-Green deployment strategy for non-canary workloads (e.g. database schema migrations and stateful workers).
 - Metric analysis queries (HTTP 5xx rate < 0.05%, p99 latency < 25ms, Kafka consumer lag < 100).
 - Multi-cluster synchronization between domestic Mumbai AWS/GCP clusters and GIFT City IFSCA nodes.
- **Out of Scope / Handled Elsewhere:**
 - Continuous Integration builds and container image signing (Prompt 803).
 - Base cloud infrastructure provisioning via Terraform (Prompt 805).
 - Formal CAB approval workflows and release scheduling governance (Prompt 810).

## Technology to Use
- **ArgoCD v2.11+**: Enterprise GitOps continuous delivery tool for Kubernetes. Justification: Implements single-source-of-truth Git operations, automated drift detection and self-healing, granular RBAC, and seamless multi-cluster management.
- **Argo Rollouts v1.7+**: Advanced deployment controller supporting Canary and Blue-Green deployment strategies with automated traffic routing via Cilium / AWS ALB. Justification: Native integration with Prometheus metrics for automated rollback and fine-grained canary traffic shifting.
- **Kustomize v5**: Native Kubernetes configuration management without complex templating languages.
- **Prometheus / Mimir**: Real-time metrics engine feeding Prometheus metrics to Argo Rollout Analysis runs.

## Backend / Infra Touchpoints
- **GitOps Config Repository**: `growww-cd-gitops` repository maintaining declarative environment states.
- **ArgoCD Control Plane**: Hosted in management cluster, managing target production clusters via mTLS.
- **Prometheus Metrics Endpoints**: Scraped by Argo Rollouts controller to evaluate canary health.
- **Slack / PagerDuty**: Webhook destinations receiving deployment notifications, canary promotions, and automated rollback alerts.

## Blockchain Interaction
Orchestrates the deployment and maintenance of Hyperledger Besu validator nodes, RPC relays, and transaction relayer microservices:
- **Blue-Green Relayer Upgrades**: Upgrades for the blockchain transaction relayer service (Prompt 208/309) use blue-green deployments with zero-downtime queue handoff to ensure no on-chain settlement transactions or event subscriptions are dropped.
- **Besu Node Rolling Upgrades**: Blockchain node updates are orchestrated as ordered StatefulSet rollouts with pre-stop hooks validating peer count and local block synchronization before proceeding to subsequent consortium nodes.

## Step-by-Step Build Instructions
1. Structure GitOps repository layout: `deployments/gitops/{applicationsets,analysis,rollouts,overlays/{dev,staging,prod-mumbai,prod-gift-city}}`.
2. Deploy ArgoCD and Argo Rollouts controllers to the cluster via official Helm charts.
3. Configure ArgoCD RBAC, SSO integration (OIDC with Google Workspace / Okta), and repository credentials via Kubernetes secrets.
4. Define the `AnalysisTemplate` resources in `deployments/k8s/analysis/` for success rate, p99 latency, and Kafka lag.
5. Convert standard `Deployment` manifests for core microservices (e.g. `order-matching-engine`, `trade-settlement-service`) to Argo `Rollout` manifests.
6. Configure Cilium / ALB ingress traffic routing rules in the `Rollout` manifest to enable precise percentage-based canary traffic splitting.
7. Configure canary steps: Step 1 (5% traffic, 10 min analysis) -> Step 2 (20% traffic, 15 min analysis) -> Step 3 (50% traffic, 15 min analysis) -> Step 4 (100% full promotion).
8. Define Blue-Green rollout strategy for asynchronous background workers requiring strict single-instance execution.
9. Create ArgoCD `ApplicationSet` using Git directory generator to automatically discover and synchronize new microservices.
10. Configure automated sync policies with `selfHeal: true` and `prune: true` for dev/staging, and manual sync / PR-based promotion for production.
11. Implement Slack / PagerDuty notification webhooks on ArgoCD for `SyncFailed`, `SyncSucceeded`, `RolloutDegraded`, and `RolloutAborted`.
12. Write `scripts/gitops/promote-release.sh` to automate updating target image tags in environment Kustomize overlays.
13. Test automated rollback simulation by deploying a mock faulty container image (simulating 5% HTTP 500 errors) and verifying automatic rollback within 60 seconds.
14. Document the progressive delivery operating manual and emergency manual override commands (`kubectl argo rollouts abort/promote`).

## Interfaces / Contracts
```yaml
# deployments/k8s/analysis/http-success-rate-analysis.yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: http-success-rate-analysis
  namespace: growww-trading
spec:
  metrics:
 - name: success-rate
      interval: 30s
      successCondition: result[0] >= 0.9995
      failureLimit: 2
      provider:
        prometheus:
          address: http://prometheus-k8s.growww-observability:9090
          query: |
            sum(rate(http_requests_total{app="order-service", status!~"5.*"}[2m]))
            /
            sum(rate(http_requests_total{app="order-service"}[2m]))
```

```yaml
# deployments/k8s/rollouts/order-service-rollout.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: order-service
  namespace: growww-trading
spec:
  replicas: 10
  strategy:
    canary:
      analysis:
        templates:
 - templateName: http-success-rate-analysis
        args:
 - name: service-name
            value: order-service
      steps:
 - setWeight: 5
 - pause: { duration: 10m }
 - setWeight: 20
 - pause: { duration: 15m }
 - setWeight: 50
 - pause: { duration: 15m }
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
 - name: order-service
          image: ghcr.io/growww/order-service:v1.4.2
          ports:
 - containerPort: 8080
```

## Security & Compliance Notes
- GitOps Audit Trail: All production deployments must originate from merged Git commits signed by authorized release engineers; direct `kubectl apply` access to production clusters is blocked via IAM.
- Segregation of Environments: Production clusters in Mumbai and GIFT City operate on separate IAM roles and Git repository branches/directories to prevent accidental cross-environment state pollution.
- Automated Rollback Audit: Any automated canary rollback generates an immutable audit incident record sent to the compliance logging lake.

## Acceptance Criteria
- [ ] ArgoCD ApplicationSets automatically synchronize infrastructure and microservices across all environments.
- [ ] Progressive canary rollout shifts traffic incrementally (5% -> 20% -> 50% -> 100%) while executing Prometheus analysis queries.
- [ ] Simulated service regressions (elevated error rate or latency) trigger automated rollbacks in <60 seconds without human intervention.
- [ ] Blue-green rollouts for settlement workers switch traffic with zero dropped Kafka messages or duplicate transactions.
- [ ] ArgoCD UI and notifications provide real-time visibility into sync status, diffs, and health metrics.

## Suggested Order / Dependencies
- Prerequisites: Prompt 108 (Environment Strategy), Prompt 802 (Kubernetes Architecture), Prompt 803 (CI Pipeline), Prompt 806 (Observability Stack).
- Parallel Tasks: Prompt 805 (Terraform Infra), Prompt 810 (Release Management).

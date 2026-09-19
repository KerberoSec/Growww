# Kubernetes Cluster Mesh & Autoscaling Specification

## Cluster Orchestration
- Managed AWS EKS Multi-AZ cluster deployed across `ap-south-1a`, `ap-south-1b`, and `ap-south-1c`.
- **Networking**: Cilium eBPF CNI delivering ultra-high throughput and sub-millisecond pod-to-pod latency.
- **Autoscaling**: Karpenter provisioner managing dynamic node scaling; Horizontal Pod Autoscaler (HPA) scaling microservices up to 500 pods under burst load.
- **Zero-Trust**: Istio service mesh with strict mTLS 1.3 and SPIFFE workload authentication.

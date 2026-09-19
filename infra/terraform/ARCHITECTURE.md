# Terraform Multi-Region Infrastructure-as-Code Specification

## Infrastructure Topology
- Multi-region infrastructure covering AWS Mumbai (`ap-south-1`) as Primary and Hyderabad (`ap-south-2`) as Hot Standby.
- Dedicated Equinix PoP module for the GIFT City (IFSCA) international trading zone.
- Automated Route 53 Application Recovery Controller (ARC) routing with health check failover triggers.

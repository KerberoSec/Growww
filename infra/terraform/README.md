# Terraform Multi-Region Cloud Infrastructure

## Purpose & Scope
The `infra/terraform` module manages deployment, networking, and runtime environments for the Growww / NBSE exchange.

## Architectural Responsibilities
Infrastructure-as-Code modules for AWS Mumbai (ap-south-1), Hyderabad (ap-south-2), and GIFT City international financial zone.

## Local Testing & Scale Profile
- **Local Workstation**: Supports lightweight execution for developers running on Linux/macOS laptops.
- **Enterprise Scale**: Hardened for multi-AZ high-availability supporting up to 1 Crore (10 Million) concurrent connections.

# 0005: Lab provider abstraction (no Vagrant dependency)

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 shelled out directly to `vagrant`/`VBoxManage` with string-built VM names/paths (MASTER_PLAN §2.3 #7, #9) and depended on a stale, now-BSL-licensed toolchain (Vagrant/Packer moved off OSS licensing in 2023). The "twin" was also never config-faithful — just a stock image plus a dropped-in JSON file (§2.3 #5).

## Decision

Introduce a `LabProvider` port (MASTER_PLAN §5.8) with swappable adapters: `fake` (built first, used in CI/demo/Phase 1), `hyperv` (primary real adapter on Windows, via PowerShell Direct + an isolated internal-only virtual switch), with `libvirt`/`proxmox` as later adapters. VirtualBox is not a target adapter. All executor calls use `exec.Command(name, args...)` with allow-listed executables and validated, typed arguments (never shell-string interpolation) — directly closing MASTER_PLAN §2.3 #7 and §6.3's shell-injection rule.

Image building uses Packer (optional) or Microsoft evaluation media + Ansible/PowerShell-Direct provisioning; CyberDitto never redistributes Windows images.

## Rejected alternatives

- Continuing with Vagrant — BSL licensing risk, and it adds a layer (Vagrant → VBoxManage) that buys nothing once we talk to Hyper-V/libvirt APIs directly.

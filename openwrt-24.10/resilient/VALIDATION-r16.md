# Resilient r16 validation

Validated on 2026-09-27 before publication. The previous release report is
[VALIDATION-r15.md](VALIDATION-r15.md).

## Inputs

- Application/core source: `955ea9943fdbd8004686392721befb824c22b8a5`.
- Packaging and VM pipeline: `fcf82a06afd15b22a852cafb9e671383fb63d964`.
- Service/core package: `2.5.7-resilient.10-r16.resilient1`.
- LuCI package: `26.268.0-r16.resilient1`.
- Package architecture: `aarch64_cortex-a53`.

The embedded GUI passed 225 tests across 45 files, TypeScript checking, the
six-locale key check, ESLint with zero errors and a production Vite build.
Service and core cross-builds and Go vet passed. The RoutingA Russia preset was
also parsed by RoutingA v1.0.2; 43 GUI test files with 215 tests passed on the
clean upstream-based feature branch.

## OpenWrt 24.10.4 VM

The pipeline booted the official `armsr/armv8` OpenWrt 24.10.4 EFI image in
QEMU, installed release-native dependencies and all three r16 IPKs, and
confirmed the installed package versions, matching service/core binary
versions, LuCI files, embedded GUI and `/api/version` with
`coreVersionValid: true`. All package-pipeline checks passed.

The installed core also accepted a configuration referring to
`geosite:category-ru`: `v2raya_core run -test` reported `Configuration OK`.
This verifies the RoutingA preset's principal geosite category against the
geosite database shipped as an OpenWrt package.

The disposable VM uses a 512 MiB `/usr` disk because the generic image's root
partition is too small for both static binaries. It adds a test-only
`aarch64_cortex-a53` opkg architecture entry to the `aarch64_generic` image.
Neither adjustment is part of a physical-router installation. The proxy-mode
regression suite passed on a separate reproducible OpenWrt 24.10.4 x86 VM.

## Signed feed

`Packages` was signed with OpenWrt `usign` in the disposable VM and verified
against the published public key (`SIGNATURE_OK`). The private key is not in
this repository. `SHA256SUMS` identifies the three published IPKs, package
index and build metadata. The feed key fingerprint is `9478c50315c92021`;
the public-key SHA-256 is
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.

The r16 feed is offered for OpenWrt 24.10.0–24.10.8 on
`aarch64_cortex-a53` routers. This release was installed and tested on
24.10.4 in a VM; other patch releases and physical routers have not been
tested for r16.

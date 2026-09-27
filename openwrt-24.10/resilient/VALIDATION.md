# Resilient r17 validation

Validated on 2026-09-27 before publication. The previous report is
[VALIDATION-r16.md](VALIDATION-r16.md).

## Inputs

- Application/core source: `dae1496fb4b57e77456ab0d9120bb7e1d52b1638`.
- Packaging and VM pipeline: `fcf82a06afd15b22a852cafb9e671383fb63d964`.
- Service/core package: `2.5.7-resilient.11-r17.resilient1`.
- LuCI package: `26.268.0-r17.resilient1`.
- Package architecture: `aarch64_cortex-a53`.

The only application change since r16 sets `font-family: inherit` on dialog
action buttons so a link such as RoutingA's Help and manual matches adjacent
button text. The integrated fork passed 225 GUI tests across 45 files,
TypeScript checking, six-locale key checking, ESLint with zero errors, and
a production Vite build. The upstream-based one-line PR branch passed 213
GUI tests across 43 files, typecheck, lint, and a production build.

## OpenWrt 24.10.4 VM

The pipeline booted the official `armsr/armv8` OpenWrt 24.10.4 EFI image in
QEMU, installed release-native dependencies and all three r17 IPKs, and
confirmed the package versions, matching service/core binary versions, LuCI
files, embedded GUI, and `/api/version` with `coreVersionValid: true`.
The installed web server served the production CSS containing the font fix.
All package-pipeline checks passed.

The disposable VM uses a 512 MiB `/usr` disk because the generic image's
root partition is too small for both static binaries. It adds a test-only
`aarch64_cortex-a53` opkg architecture entry to the `aarch64_generic` image.
Neither adjustment is part of a physical-router installation.

## Signed feed

`Packages` was signed with OpenWrt `usign` in the disposable VM and verified
against the published public key (`SIGNATURE_OK`). The private key is not in
this repository. `SHA256SUMS` covers the three IPKs, package index, and build
metadata. The feed key fingerprint is `9478c50315c92021`; the public-key
SHA-256 is
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.

The r17 feed is offered for OpenWrt 24.10.0–24.10.8 on
`aarch64_cortex-a53` routers. This release was installed and tested on
24.10.4 in a VM; the other patch releases and physical routers have not
been tested for r17.

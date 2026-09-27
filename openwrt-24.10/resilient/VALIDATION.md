# Resilient r14 validation

Validated on 2026-09-27 before publication. The previous release report is
[VALIDATION-r13.md](VALIDATION-r13.md).

## Inputs

- Application/core source: `1d1dd1d7c645192238da1279b49f42b62979826b`.
- Packaging and VM pipeline: `d37b9b30ea2ce1cc797eb1452336475060f335fe`.
- Service/core package: `2.5.7-resilient.8-r14.resilient1`.
- LuCI package: `26.268.0-r14.resilient1`.
- Package architecture: `aarch64_cortex-a53`.

The embedded GUI passed 223 unit tests across 45 files, TypeScript checking,
the six-locale key check and the production Vite build. Go tests for the group
controller, worker, configuration and generated core template passed. Four
controller tests and four core lifecycle tests use Linux executable fixtures
that cannot run on the Windows build host; the applicable tests were run with
those fixtures excluded. The real ARM64 binaries were then exercised in Linux.

## OpenWrt 24.10.4 VM

The pipeline booted the official `armsr/armv8` OpenWrt 24.10.4 EFI image in
QEMU, installed release-native dependencies and all three r14 IPKs, and
confirmed the installed package versions, matching service/core binary
versions, LuCI files, embedded GUI and `/api/version` with
`coreVersionValid: true`.

The disposable VM uses a 512 MiB `/usr` disk because the generic image's root
partition is too small for both static binaries. It adds a test-only
`aarch64_cortex-a53` opkg architecture entry to the `aarch64_generic` image.
Neither adjustment is part of a physical-router installation.

Two local SOCKS5 fixtures (A and B) supplied controlled traffic and a probe
URL. The VM ran `tools/test-proxy-modes-openwrt.sh` and emitted all markers:

```text
__RESILIENT_FIXED_TRAFFIC_OK__
__RESILIENT_KEEP_CURRENT_TRAFFIC_OK__
__RESILIENT_REFRESH_NO_DROP_OK__
__RESILIENT_LEAST_LATENCY_TRAFFIC_OK__
__RESILIENT_FIRST_AVAILABLE_TRAFFIC_OK__
__RESILIENT_RANDOM_TRAFFIC_OK__
__RESILIENT_ROUND_ROBIN_TRAFFIC_OK__
__RESILIENT_SUBSCRIPTION_REORDER_NO_DROP_OK__
```

The checks make repeated HTTP requests through the proxy, verify the selected
fixture, confirm keep-current failover and retention, and compare the main
core PID across membership refresh and subscription reordering. The round-robin
test confirms both healthy fixtures serve traffic. Throughput threshold and
fallback behavior are covered by Go tests; the local SOCKS fixtures do not
simulate bandwidth. This finite VM test cannot guarantee uninterrupted
connectivity on every physical network.

## Signed feed and upgrade helper

`Packages` was signed with OpenWrt `usign` and verified using the public key.
The VM then fetched `Packages.gz` and `Packages.sig` over its test feed;
`opkg update` reported `Signature check passed`. The staged upgrade helper
downloaded all three matching packages, checked SHA-256 digests and versions,
and completed successfully against that signed feed. The helper was exercised
with r14 already installed; its r13-to-r14 migration path was not run on a
physical router.

The feed key fingerprint is `9478c50315c92021` and the public-key SHA-256 is
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.
The private key is not in this repository. `SHA256SUMS` identifies the three
published IPKs, package index and build metadata.

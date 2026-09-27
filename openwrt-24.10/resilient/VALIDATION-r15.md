# Resilient r15 validation

Validated on 2026-09-27 before publication. The previous release report is
[VALIDATION-r14.md](VALIDATION-r14.md).

## Inputs

- Application/core source: `deeb7b41a266eb6a47d1ed7110f104fbb07a674a`.
- Packaging and VM pipeline: `fcf82a06afd15b22a852cafb9e671383fb63d964`.
- Service/core package: `2.5.7-resilient.9-r15.resilient1`.
- LuCI package: `26.268.0-r15.resilient1`.
- Package architecture: `aarch64_cortex-a53`.

The embedded GUI passed 223 unit tests across 45 files, TypeScript checking,
the six-locale key check and the production Vite build. ESLint reported zero
errors; the Windows checkout has pre-existing CRLF formatting warnings. The
new Go controller tests confirm that a cached automatic group does not hold
up manual start and that an empty group is checked before its first start.
Service and core builds and vet passed. The full Go test suite does not pass
on Windows because some tests invoke Unix commands or executable fixtures;
the targeted controller tests passed and the ARM64 binaries were exercised in
Linux.

## OpenWrt 24.10.4 VM

The pipeline booted the official `armsr/armv8` OpenWrt 24.10.4 EFI image in
QEMU, installed release-native dependencies and all three r15 IPKs, and
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
__RESILIENT_MANUAL_START_WITH_CACHED_GROUP_OK__
```

The checks make repeated HTTP requests through the proxy, verify the selected
fixture, confirm keep-current failover and retention, and compare the main
core PID across membership refresh and subscription reordering. The random
test checks that repeated requests keep using the same healthy fixture; the
round-robin test confirms both healthy fixtures serve traffic. The manual
start test stops and restarts the core with cached automatic group members,
then confirms the saved route carries traffic. Throughput
threshold and fallback behavior are covered by Go tests; the local SOCKS
fixtures do not simulate bandwidth. This finite VM test cannot guarantee uninterrupted
connectivity on every physical network.

## Signed feed and upgrade helper

`Packages` was signed with OpenWrt `usign` and verified using the public key.
The VM then fetched `Packages.gz` and `Packages.sig` over its test feed;
`opkg update` reported `Signature check passed`. The staged upgrade helper
downloaded all three matching packages, checked SHA-256 digests and versions,
and completed successfully against that signed feed. The helper was exercised
with r15 already installed; its r14-to-r15 migration path was not run on a
physical router.

The feed key fingerprint is `9478c50315c92021` and the public-key SHA-256 is
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.
The private key is not in this repository. `SHA256SUMS` identifies the three
published IPKs, package index and build metadata.

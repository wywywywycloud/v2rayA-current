# Resilient r13 validation

Validated on 2026-09-27 before publication.

## Inputs

- Application/core commit: `0de21b1b86d9609145e429d2772219d69c1502e6`.
- Packaging commit: `7ee2123614c167646b12761d512508255327ebe5`.
- Service/core version: `2.5.7-resilient.7-r13.resilient1`.
- LuCI version: `26.268.0-r13.resilient1`.
- Target package architecture: `aarch64_cortex-a53`.

The production GUI build passed all 222 unit tests, TypeScript checking, the
six-locale key check and ESLint without errors. Targeted Go tests cover fixed
selection, automatic handoff, URL/throughput eligibility and group strategies.
The application and matching core were cross-compiled as static Linux ARM64
binaries. `SHA256SUMS` records exact digests for all three IPKs and their index.

The dashboard was exercised in a browser with 12 mock members. The run checked
automatic labels and wrapping, selection of a concrete member, the pin marker,
the automatic transition to **Do not switch**, disabled automatic controls and
the card-level group settings button.

## OpenWrt 24.10.4 ARM64 VM

The package pipeline booted the official
`openwrt-24.10.4-armsr-armv8-generic-ext4-combined-efi.img.gz` image in QEMU.
The image matched OpenWrt's published SHA-256
`2129ab59e79dff64537779f24befbc6bee241b72998b9ea434a1166f317dc30c`.

The clean VM accepted signed official package lists, installed release-native
`kmod-nft-tproxy`, geodata and all three r13 packages, then checked:

```text
v2raya-resilient             2.5.7-resilient.7-r13.resilient1
v2raya-resilient-core        2.5.7-resilient.7-r13.resilient1
luci-app-v2raya-resilient    26.268.0-r13.resilient1
v2rayA                       2.5.7-resilient.7
V2RAYA_CORE                  2.5.7-resilient.7
```

The service started, the embedded GUI returned its HTML, the LuCI page was
present and `/api/version` returned matching service/core versions with
`coreVersionValid: true`.

The generic image's root partition is too small for both static Go binaries,
so the test attaches a temporary 512 MiB `/usr` disk. It also adds a test-only
`aarch64_cortex-a53` opkg architecture entry; the VM itself is
`aarch64_generic`. Neither adjustment is part of a physical-router install.

## Signed feed

`Packages` is signed with OpenWrt `usign` and verified before publication.
The feed key fingerprint is `9478c50315c92021`; its public-key SHA-256 is
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.
The private key is never committed to the repository.

The reproducible VM entry point is
`tools/test-openwrt-24.10.4-arm64.ps1` on the packaging branch. Its serial log
contains all r13 package, binary, service, GUI and API success markers.

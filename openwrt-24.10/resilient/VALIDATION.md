# Resilient r11 validation

Validated on 2026-09-27 before publication.

## Inputs

- Application/core commit: `23c5259bace38fc806c6b9a790f1fe0e6c9a8d1f`.
- Packaging commit: `c0fde21b86908c3f6ee3529e5b0168bd829b11f1`.
- Service/core version: `2.5.7-resilient.5-r11.resilient1`.
- LuCI version: `26.268.0-r11.resilient1`.
- Target package architecture: `aarch64_cortex-a53`.

The service and matching core were cross-compiled as static Linux ARM64
binaries after a production GUI build. The package index records exact SHA-256
digests for all three IPKs.

## OpenWrt 24.10.4 ARM64 VM

The package pipeline booted the official
`openwrt-24.10.4-armsr-armv8-generic-ext4-combined-efi.img.gz` image in QEMU.
The downloaded image matched OpenWrt's published SHA-256
`2129ab59e79dff64537779f24befbc6bee241b72998b9ea434a1166f317dc30c`.

The clean VM accepted signed official package lists, installed release-native
`kmod-nft-tproxy`, geodata and all three r11 packages, then reported:

```text
v2raya-resilient             2.5.7-resilient.5-r11.resilient1
v2raya-resilient-core        2.5.7-resilient.5-r11.resilient1
luci-app-v2raya-resilient    26.268.0-r11.resilient1
v2rayA                       2.5.7-resilient.5
V2RAYA_CORE                  2.5.7-resilient.5 (based on xray-core 26.7.28)
```

The service started, the embedded GUI returned its HTML, the LuCI page was
present and `/api/version` returned `serviceValid: true`,
`coreVersionValid: true`, version `2.5.7-resilient.5` for both components and
variant `V2rayaCore`.

The generic image's root partition is too small for both static Go binaries,
so the test attaches a temporary 512 MiB `/usr` disk. It also adds a test-only
`aarch64_cortex-a53` opkg architecture entry; the VM itself is
`aarch64_generic`. Neither adjustment is part of a physical-router install.

The VM run exposed and prevented two Windows-host packaging defects before
publication: CRLF in the init script and backslash directory names in the LuCI
tar archive. The published IPKs contain LF shell files and POSIX archive paths.

## Signed feed

`Packages` was signed and verified with OpenWrt `usign` inside an OpenWrt
24.10.4 VM. Resilient r11 uses public-key fingerprint `9478c50315c92021` and
public-key SHA-256
`7a5f9426bdbacd25d1e89ad7e5e65579b2c33bcd278b0489772f3ab017df8787`.
The private key is not present in the repository.

This r11 key differs from the previous Resilient feed key. Existing
installations must run `add-resilient-feed.sh` once before updating package
lists. The script validates the new public key and index signature before
installing the key; it does not change v2rayA packages or settings.

The reproducible ARM64 test entry point is
`tools/test-openwrt-24.10.4-arm64.ps1` on the packaging branch.

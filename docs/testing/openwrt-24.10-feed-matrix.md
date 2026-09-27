# OpenWrt 24.10 Resilient feed matrix

The published `2.5.7-resilient.11-r17.resilient1` service/core pair and
`26.268.0-r17.resilient1` LuCI package were installed from the signed Resilient
feed in disposable QEMU VMs for OpenWrt **24.10.0 through 24.10.8**. Each VM
used its own official `armsr/armv8` EFI image and release-native dependency
feeds. All nine passed `opkg download`, installation, service start, embedded
GUI, LuCI files and `/api/version` with `coreVersionValid: true`.
Each also started the proxy core and passed four HTTP requests through its
proxy port to a controlled local SOCKS5 node, while the service stayed up.
This does not test any external VPN provider.

The reproducible harness is in the [packaging fork's tools directory](https://github.com/wywywywycloud/v2raya-openwrt-current/tree/release/resilient-openwrt-24.10/tools):

- `prepare-openwrt-24.10-arm64.ps1` verifies the official image SHA-256 and
  decompresses it.
- `test-openwrt-24.10-feed-matrix.ps1` runs up to three versioned QEMU guests
  concurrently and records `matrix-results.json` plus serial logs.
- `test-openwrt-24.10.4-arm64.ps1` is the per-VM runner despite its historical
  filename; `-OpenWrtVersion` accepts all nine patch releases.
- `local-socks-fixture.py` and `test-basic-traffic-openwrt.sh` provide a
  controlled end-to-end traffic check when `-TrafficSmoke` is set.

The generic ARM64 images identify as `aarch64_generic`. The runner adds a
test-only `aarch64_cortex-a53` opkg architecture entry to install the release
package; it also mounts an extra `/usr` disk because the generic image is small.
Neither change is part of installation on a physical router. These VMs do not
establish wireless, flash-upgrade or hardware-offload compatibility for every
device, and r17 has not been verified on a physical router.

The [signed feed validation report](https://github.com/wywywywycloud/v2rayA-current/blob/openwrt-feed/openwrt-24.10/resilient/VALIDATION.md)
records the exact package versions and results.

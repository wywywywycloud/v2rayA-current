# Resilient r17 validation

The original r17 package checks ran before publication; the nine-version
published-feed matrix ran on 2026-09-27. The previous report is
[VALIDATION-r16.md](VALIDATION-r16.md).

## Inputs

- Application/core source: `dae1496fb4b57e77456ab0d9120bb7e1d52b1638`.
- Package build source: `fcf82a06afd15b22a852cafb9e671383fb63d964`.
- Post-publication VM matrix pipeline: `15c7d5970c4ca272193dea08a6f7287372cf4efd`.
- Service/core package: `2.5.7-resilient.11-r17.resilient1`.
- LuCI package: `26.268.0-r17.resilient1`.
- Package architecture: `aarch64_cortex-a53`.

The only application change since r16 sets `font-family: inherit` on dialog
action buttons so a link such as RoutingA's Help and manual matches adjacent
button text. The integrated fork passed 225 GUI tests across 45 files,
TypeScript checking, six-locale key checking, ESLint with zero errors, and
a production Vite build. The upstream-based one-line PR branch passed 213
GUI tests across 43 files, typecheck, lint, and a production build.

## OpenWrt 24.10.0–24.10.8 VM matrix

After publication, the pipeline booted the official `armsr/armv8` EFI image
for **each** of OpenWrt 24.10.0, .1, .2, .3, .4, .5, .6, .7 and .8 in QEMU.
Each guest ran the published signed-feed setup script, used `opkg download`
for the LuCI IPK, then installed LuCI and its service/core dependencies from
the feed. Each guest resolved release-native dependencies from its own official
OpenWrt feeds. All nine installs passed the package-version, binary-version,
LuCI-file, embedded-GUI, service and `/api/version` checks, including
`coreVersionValid: true`. The installed 24.10.4 web server also served the
production CSS containing the font fix.

The matrix then installed release-native `curl`, imported a controlled local
SOCKS5 node, started the proxy core through the application's API and made
four HTTP requests through its proxy port in each guest. Every response came
from the selected fixture node, and the service remained running. This checks
the package's end-to-end proxy path, not access to an external VPN provider.

| OpenWrt | Signed feed and `opkg` | Service, GUI, LuCI, core version | Proxied requests |
| --- | --- | --- | --- |
| 24.10.0 | Passed | Passed | 4/4 |
| 24.10.1 | Passed | Passed | 4/4 |
| 24.10.2 | Passed | Passed | 4/4 |
| 24.10.3 | Passed | Passed | 4/4 |
| 24.10.4 | Passed | Passed | 4/4 |
| 24.10.5 | Passed | Passed | 4/4 |
| 24.10.6 | Passed | Passed | 4/4 |
| 24.10.7 | Passed | Passed | 4/4 |
| 24.10.8 | Passed | Passed | 4/4 |

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
`aarch64_cortex-a53` routers. All nine patch releases have been installed and
tested in generic ARM64 VMs. Physical routers have not been tested for r17.

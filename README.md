<div align="center">

<img src="gui/public/static/v2raya-icon.svg" width="96" alt="x-Ray VPN">

# x-Ray VPN

Experimental edition based on v2rayA-resilient.

Includes fixes for probe and DNS cancellation, duplicate probes, process cleanup, asset downloads, XHTTP uploads and REALITY startup validation.

TLS session resumption and variable POST sizes remain off by default. Chrome 133 matching is partial; no current-browser equivalence or general CPU/RAM improvement is claimed.

[Experimental release notes](EXPERIMENTAL.md)

**A web client for its own Xray-based core, with transparent proxy on Linux, Windows and macOS.**

English · [简体中文](README_zh.md) · [Русский](README_ru.md)

[Requirements](#requirements) • [Install](#install) • [First run](#first-run) • [Transparent proxy](#transparent-proxy) • [Data and upgrades](#data-and-upgrades) • [Support](#support)

</div>

x-Ray VPN runs as a service and is used from a browser, on the machine itself or on a router or NAS. It imports subscriptions and share links for VMess, VLESS, Shadowsocks, Trojan, Hysteria2, TUIC, [Juicity](https://github.com/juicity), AnyTLS, WireGuard, SOCKS5 and HTTP(S), groups nodes and selects the member with the lowest measured latency or a pinned one, and splits traffic with rules written in RoutingA. ShadowsocksR is not supported.

## Requirements

| Component | Requirement |
| --- | --- |
| Core | `v2raya_core` of the **same version** as `v2raya`, next to it or in `PATH`. The packages and installers below ship both; a hand-installed pair must match. |
| Rule data | `geoip.dat` and `geosite.dat` in `/usr/share/v2raya` or `/usr/local/share/v2raya`. The packages ship them; a hand install downloads them from GitHub on the first start and exits if that fails. |
| Linux transparent proxy | root; `iptables` or `nftables` for `redirect` and `tproxy`; `/dev/net/tun` and `ip` from iproute2 for `tun` |
| Windows | Windows 10 build 14393 or later; administrator for `tun`; started without elevation it runs in lite mode and can still set the current user's system proxy |
| macOS | root for `tun`; started as a user it runs in lite mode with the system proxy instead |
| Browser | a current Chrome, Edge, Firefox or Safari |

## Install

The experimental target is OpenWrt **24.10.4 ARM64** (`aarch64_cortex-a53`). Use the matching three IPKs attached to the [experimental release](https://github.com/wywywywycloud/v2rayA-resilient/releases/tag/experimental-2026.09.28.1). See [reproducible build and packaging instructions](install/openwrt-experimental/README.md) and that release’s validation notes for the actual artifact checks. Install or roll back service and core together.

Upstream distribution packages and the earlier signed Resilient feed are different builds; their installation matrices do not validate this experimental release.

## First run

The service listens on `0.0.0.0:2017`, so on a router or NAS the interface is at `http://<device-address>:2017`. The first account registered becomes the administrator and registration needs no login, so restrict access before registering on a shared network, or start with `--address 127.0.0.1:2017` for local use only.

Open the interface and create the administrator. The tutorial then walks through importing a subscription or a share link, adding the nodes to a group and choosing the routing rules; start the core with Start on the dashboard or the status button at the top of the page.

Once the core runs, point an application at SOCKS5 `127.0.0.1:20170` or HTTP `127.0.0.1:20171` (`20172` applies the routing rules to HTTP), or turn on transparent proxy on the dashboard and pick a mode from the table below. A group with several connected members uses the one with the lowest measured latency; pinning a member routes the group through it alone. Subscriptions update from the subscriptions page or on the interval set in the settings.

<img src="docs/images/screenshot.png" alt="The dashboard, light and dark" width="100%">

## Transparent proxy

| Platform | Modes |
| --- | --- |
| Linux | `redirect`, `tproxy`, `tun`; with `--lite`, the system proxy on GNOME and KDE |
| macOS | `tun`; with `--lite`, the system proxy |
| Windows | `tun`, system proxy |

The system proxy modes set the desktop's proxy settings and cover applications that honour them. The other modes intercept traffic.

With `tun`, the core opens a TUN device and assigns its address; with automatic routing on (the default), x-Ray VPN installs the routes and the DNS setting, otherwise your own script does. The core's own connections, the direct outbound and the DNS module's upstream queries never enter the TUN (by socket mark on Linux, by binding to the physical interface on Windows and macOS). Plain DNS on port 53 that reaches the TUN is answered by the core's DNS module; encrypted DNS is not intercepted. x-Ray VPN and the core are always excluded, and other processes can be excluded by executable name in the settings. More specific connected and static routes bypass the TUN.

Known limitation: on Windows and macOS an application that queries a LAN resolver directly still bypasses the TUN. The system resolver is covered: Windows points it at the TUN gateway, macOS at the core's listener on `127.0.0.1`, which needs port 53 free.

## Routing rules

<img src="docs/images/routinga.png" alt="The RoutingA editor, light and dark" width="100%">

RoutingA has a list mode with a form per rule and a text mode with line numbers, colouring and per-line checks. The syntax reference sits beside the editor. Import or export rules as a text file.

## Data and upgrades

The SQLite database `v2raya.db` and the generated core config live in the configuration directory: `/etc/v2raya` on Linux and macOS, `%ProgramData%\SYSTEM\v2rayA` for the Windows service, the user's config directory with `--lite`, or the path given by `--config`. Logs go to `/var/log/v2raya/v2raya.log` under the systemd and OpenRC units, or to `--log-file` / `V2RAYA_LOG_FILE`. Network requests are made for subscription updates, rule-data downloads, latency probes, DNS, a release check against GitHub at start and weekly, and an NTP query (`ntp.aliyun.com`) when a VMess node fails, to tell a wrong clock from a bad node.

Stop the service and back up the configuration directory before upgrading. Upgrading from a release before 2.4 migrates the BoltDB database on the first start; the old file is kept as `bolt.db.bak` (`bolt.db.bak.1` and up when that name is taken), and accounts have to be registered again. Upgrade `v2raya` and `v2raya_core` together: the dashboard shows the core version, and a mismatch is reported in a banner.

## Support

Ask questions in the [discussions](https://github.com/wywywywycloud/v2rayA-resilient/discussions) and report bugs in the [issues](https://github.com/wywywywycloud/v2rayA-resilient/issues).

Do not use this project for anything illegal.

## Credits

Founded by [@mzz2017](https://github.com/mzz2017). The Material Design 3 interface, the in-core TUN and the 2.5 service rework are by [@Zakkaus](https://github.com/Zakkaus). The OpenRC files come from the [gentoo-zh community](https://gentoozh.org). Routing data from [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community), [v2fly/geoip](https://github.com/v2fly/geoip) and, for the GFWList mode, [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat); the transparent-proxy rules were learnt from [zfl9/ss-tproxy](https://github.com/zfl9/ss-tproxy) and [hq450/fancyss](https://github.com/hq450/fancyss).

## License

[AGPL-3.0-only](LICENSE). The core is a fork of [Xray-core](https://github.com/XTLS/Xray-core) (MPL-2.0).

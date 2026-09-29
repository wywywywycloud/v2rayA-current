# x-Ray VPN experimental

Install **xray-proxy-client-experimental** from this signed feed. The metapackage pulls the exact matching service, core and LuCI version `2026.09.29-experimental.1`, plus their declared dependencies from your official OpenWrt repositories.

If this Resilient feed is already configured, click **System > Software > Update lists**, then search for **xray-proxy-client-experimental** and install it. From SSH:

```sh
opkg update
opkg install xray-proxy-client-experimental
```

For a new feed setup, run the existing `add-resilient-feed.sh`; the feed URL remains unchanged. The signing key was rotated on 2026-09-29 (old `9478c50315c92021` lost, new `1c624cadb7ee79e7`): existing installations must re-run the script once before `opkg update`. Keep your device's official repositories enabled. Kernel modules must come from the repository matching your installed kernel; never force dependency or architecture checks.

This replaces the existing Resilient application using its compatible package names and configuration paths. It is not a parallel installation. Since these same package names now point to the newer experimental version, they also appear as updates for installed Resilient packages. Back up your configuration before updating.

The release was tested on OpenWrt 24.10.4 ARM64 (`aarch64_cortex-a53`). No new nine-version or physical-router test is claimed. Existing package files remain available for rollback. The complete dependency chain and validation are in [the experimental release](https://github.com/wywywywycloud/v2raya-openwrt-current/releases/tag/experimental-2026.09.29.1).

---
# v2rayA Resilient for OpenWrt 24.10

This is the personal fork distribution, separate from the upstream pull requests.
Supported package architecture: **aarch64_cortex-a53**. Do not add a different
architecture to a physical router to bypass opkg checks.

## Supported OpenWrt releases

The current r21 feed is offered for **OpenWrt 24.10.0 through 24.10.8**,
inclusive, on **aarch64_cortex-a53** routers. The feed setup script accepts all
nine releases. r13 was installed across that entire version range; r17 was
installed from the published signed feed on the official `armsr/armv8`
VM images for **all nine releases**. The package, service, embedded GUI and LuCI
checks passed on each version. Physical routers remain to be tested for r21.
Each VM also passed four proxied HTTP requests against a controlled local
SOCKS5 test node.
Use the newest 24.10 security update available for the router.

OpenWrt 25.12 is outside this distribution: that series replaced opkg/IPK with
apk/APK and requires separate native packages and a signed APK repository. No
end-of-life OpenWrt series is listed as supported.

## Names

| Package | Purpose | Version |
| --- | --- | --- |
| **luci-app-v2raya-resilient** | Install this one in LuCI; pulls the complete fork | 26.268.0-r21.resilient1 |
| v2raya-resilient | Service and embedded v2rayA web interface | 2.5.7-resilient.14-r21.resilient1 |
| v2raya-resilient-core | Exact matching proxy core | 2.5.7-resilient.14-r21.resilient1 |

r21 keeps subscription reordering and renaming from restarting the core, preserves
every automatic policy across server changes, and keeps transparent interception
installed through group-switch startup and rollback failures. Keep-current only
replaces a reachable server when the catalog really changes; low or unknown
speed alone does not evict it. Proxy-tagged UDP DNS uses its configured transport.
The one-probe-core bound, first-install defaults and manual subscription bypass
confirmation are retained.

The menu is **Services > v2rayA Resilient**. Configuration and service paths retain
`v2raya` so existing settings and accounts can survive replacement. This is a
replacement for the original packages, not a second concurrent instance.
No separate xray-core installation is needed by this distribution.

## Add the signed source once

Run over SSH on the router:

```sh
wget -O /tmp/add-resilient-feed.sh https://raw.githubusercontent.com/wywywywycloud/v2rayA-current/openwrt-feed/add-resilient-feed.sh
sh /tmp/add-resilient-feed.sh
```

The script validates OpenWrt version, architecture, the pinned public-key digest
and feed signature before adding the source. It does **not** install the fork or
change its settings. It replaces this project's previous feed entries and keeps
all other sources. The previous source configuration is saved as
`/etc/opkg/customfeeds.conf.before-resilient` when changed.

Public key fingerprint: `1c624cadb7ee79e7`.
Public key SHA256: `0f4cd283f4885d7c32b13c4a2451e6acb9101d0b2324256bc3dad043a81edc9d`.

Release r11 rotates the Resilient feed key. Existing installations must run
the setup command above once before `opkg update`; package settings and the
installed service are not changed by the setup script.

On 2026-09-29 the key was rotated again (old `9478c50315c92021` lost, new
`1c624cadb7ee79e7`). Existing installations must re-run the setup command
once before `opkg update`.

The source line, also visible under Software > Configure opkg, is:

```text
src/gz v2raya_resilient https://raw.githubusercontent.com/wywywywycloud/v2rayA-current/openwrt-feed/openwrt-24.10/resilient/aarch64_cortex-a53
```

Adding this line alone does not install the signing key; use the setup above.
Keep signature verification enabled.

## Install using the OpenWrt GUI

1. Open **System > Software**, click **Update lists**, then filter **resilient**.
2. Click **Install** beside **luci-app-v2raya-resilient** and confirm the dependency
   dialog. Leave **Allow overwriting conflicting package files** unchecked.
3. Refresh LuCI. Open **Services > v2rayA Resilient**. On a clean installation,
   enable the service and click **Save & Apply**, then open its web interface.
4. Confirm the application and core both report `2.5.7-resilient.13`.

If Software is absent, install the official `luci-app-package-manager` first.
The setup needs normal working Internet access. Stop a broken transparent proxy
before adding/updating sources if it prevents downloads.

## Existing installations

Back up `/etc/config/v2raya` and `/etc/v2raya/` before migration. The branded
packages declare replacement of `v2raya`, `v2raya-core` and `luci-app-v2raya`.
The previous personal package names are also replaced automatically.
Modified UCI configuration is retained. opkg may report that the new template
was saved as `/etc/config/v2raya-opkg`; this message is expected and does not mean
the preserved configuration was overwritten.

If the early experimental **v2raya-fork** metapackage is installed, remove that
metapackage in Software first: it pins the old 2.2.x packages. Do not remove the
configuration directory. Migration of an arbitrary production database should
always retain an off-router backup.

## Configure automatic groups and updates

Enable **Automatically add all servers** in the selected group's settings to
copy the entire catalog, including unavailable nodes, without probing them.
The initial PROXY group defaults to **3000s** between scheduled checks.

- **Keep current until failure** retains a reachable current server even when
  its speed is low or unknown. Three consecutive failed reachability checks
  trigger replacement. A real membership or connection-parameter change also
  triggers a fresh least-latency, speed-qualified choice.
- **Least latency** orders candidates by TCP latency and stops at the first
  successful URL check and complete 256 KiB speed sample at 100 KiB/s or above.
- **Random** shuffles candidates at each scheduled selection and checks them
  sequentially using the same availability/speed criteria.
- **Round robin** balances the members that passed those checks.
- **Do not switch** keeps an explicitly pinned server.

A server change never changes the selected policy. Pinning a server pauses an
automatic policy; **Auto** clears that pin and resumes the same policy. Reordering
or renaming a subscription leaves the selected connection and running core intact.
New candidates without an eligible speed sample are excluded; if no candidate
qualifies, proxy-assigned traffic is blocked and membership is retained.

During a real group switch, unchanged TPROXY/REDIRECT interception stays in place,
including startup and rollback failure. RoutingA's direct/proxy split is retained.
A configuration whose interception cannot be preserved (including TUN) rejects
the switch before stopping the running core. Existing TCP sessions can end on a
real server/core change. Explicit service stop or a service crash is outside this
group-switch protection.

For each subscription, choose one **Automatic subscription update** mode:

- **Disabled** updates only when requested manually.
- **On service start** refreshes once whenever v2rayA starts.
- **At an interval** refreshes on startup and at the required interval in minutes.
- **At an interval with fail-safe recovery** also tests the saved servers at the
  required failure interval and refreshes until at least one server works.

Both interval fields accept whole minutes from 1 to 525600 when their mode uses
them. Failed or empty downloads retain the saved nodes. Upgrades preserve the
old global startup/interval behavior without enabling fail-safe recovery.
The retired per-subscription auto-select option enables automatic membership for
the `PROXY` group once when at least one old subscription used it. Otherwise,
automatic group membership remains disabled until enabled explicitly.

For an existing Resilient installation, use the staged upgrade helper over SSH:

```sh
wget -O /tmp/upgrade-resilient.sh https://raw.githubusercontent.com/wywywywycloud/v2raya-openwrt-current/release/resilient-openwrt-24.10/tools/upgrade-resilient.sh
sh /tmp/upgrade-resilient.sh
```

The helper updates the signed feed index, checks flash and temporary space,
downloads and verifies all three IPKs, then installs the matching core, service
and LuCI package together. If there is too little free flash, it stops before
upgrading any of them. The three package versions are checked afterward. opkg
does not provide an atomic transaction, so retain a configuration backup and
inspect the package status if an installation fails for another reason.
Upgrade an existing r20 installation to r21 to receive these fixes.

## Sources and validation

- Application/core: [Resilient source](https://github.com/wywywywycloud/v2rayA-current/tree/main), application commit `283daeb0`, with a packaging-only defaults patch.
- Packaging/LuCI: [Resilient packaging branch](https://github.com/wywywywycloud/v2raya-openwrt-current/tree/release/resilient-openwrt-24.10).
- [Signed feed and validation report](https://github.com/wywywywycloud/v2rayA-current/tree/openwrt-feed/openwrt-24.10/resilient).

The service embeds its GUI and uses its matching v2raya_core. Packages are
assembled by `tools/build-current.sh` / `tools/package-current.py` using static
Linux ARM64 binaries. This is not a claim of a full OpenWrt SDK build.

The r17 packages installed from the published feed on the official OpenWrt
24.10.0–24.10.8 `armsr/armv8` images with release-native kernel modules and
dependencies. The test starts the
service, checks both `2.5.7-resilient.11` versions, the embedded GUI, LuCI and
`coreVersionValid`. A test-only `/usr` disk gives the small generic image enough
room for both static binaries. The generic ARM64 image accepts the Cortex-A53
package through a **test-only** opkg architecture alias. Physical hardware and
other architectures are not covered by these VM runs.

r18 was also installed on a fresh OpenWrt 24.10.4 ARM64 VM before publication.
The service, matching core, LuCI menu, embedded GUI and API started. Its live
settings matched the reference router's public defaults, including RoutingA,
TPROXY, HTTP+TLS sniffing, PROXY keep-current/auto-add/3000s and a newly
imported test subscription's 60-minute/one-minute failover schedule. The
published signed feed was then installed on a second clean 24.10.4 VM with
`opkg install luci-app-v2raya-resilient`; the same checks passed. See
[the r18 validation report](https://github.com/wywywywycloud/v2rayA-current/blob/openwrt-feed/openwrt-24.10/resilient/VALIDATION-r18.md).


r19 passed the clean OpenWrt 24.10.4 ARM64 package and first-install defaults
checks. Live API tests exercised stopped-core manual import and update in both
proxy and PAC modes, with and without the one-request bypass flag. Both routes
remained saved, and the core stayed stopped. See [r19 validation](VALIDATION-r19.md).


r20 was tested on one 256 MiB OpenWrt 24.10.4 VM. Membership refresh, speed-based
failover and the two-core process bound passed. See [r20 validation](https://github.com/wywywywycloud/v2rayA-current/blob/openwrt-feed/openwrt-24.10/resilient/VALIDATION-r20.md).

r21 reproduces subscription churn and startup failures before applying the fix.
See [the r21 validation report](VALIDATION-r21.md) for test conditions and limits.

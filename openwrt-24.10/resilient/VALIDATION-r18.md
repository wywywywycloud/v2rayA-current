# Resilient r18 validation

Release `2.5.7-resilient.11-r18.resilient1` packages the application at
`7d9fcfca05070226066f7f7d16e67aa99c08c3aa` with the OpenWrt-only
first-install defaults patch from packaging commit `b8eb734`. The application
fork's `main` branch and upstream pull requests were not changed.

The reference router was read through SSH and authenticated read-only API
calls. Only settings safe for a public package were copied. Its account,
subscription URL, server links and current pin are not in the release. An old
router-specific `dnsfix.sh` hook was left out because this application build
already has native DNS redirect; an existing UCI conffile keeps the hook on
upgrade.

Before publication, Go tests passed for `db/configure` and `server/service`.
The GUI and matching Linux ARM64 service/core were built from the isolated
checkout. A clean official OpenWrt 24.10.4 `armsr/armv8` VM installed the
three IPKs and passed the live defaults test.

After publication, a **new clean VM** ran the public `add-resilient-feed.sh`,
verified the signed opkg index, listed and downloaded the r18 LuCI package,
then installed all three packages with `opkg install
luci-app-v2raya-resilient`. The service, matching core, LuCI menu, embedded
GUI and version API passed. Its API reported RoutingA, TPROXY, HTTP+TLS
sniffing, IP forwarding and port sharing; the PROXY group reported keep-current,
auto-add and a 3000-second probe interval. The RoutingA rules matched the
reference router's SHA-256 digest
`b69c11abacc4466393876bcf249603f7958ce9e0aa960f0dae41e0b2799ebb27`.
A newly imported disposable subscription reported `interval_failsafe`, a
60-minute normal interval, a one-minute failure interval and direct recovery.
The UCI service was enabled on the clean install. All checks exited zero.

The VM is generic ARM64 with a test-only `aarch64_cortex-a53` opkg architecture
alias and an extra `/usr` disk. This release has not been installed on the
physical reference router, and the r18 check above covers OpenWrt 24.10.4;
the preceding r17 release was tested on 24.10.0–24.10.8. No physical-router
or other-version claim is implied for r18.

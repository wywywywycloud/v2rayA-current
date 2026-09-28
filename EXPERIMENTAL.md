# x-Ray VPN experimental

x-Ray VPN is an experimental edition of v2rayA-resilient using the existing Xray-core and x/net forks. Install or roll back the service and matching core together. Binary names (`v2raya`, `v2raya_core`), API, configuration directories and service identities remain compatible.

Included corrections cover cancellation of probes and shared DNS work, duplicate probe ownership, process/output cleanup, streaming asset downloads, immutable TLS reload snapshots, XHTTP connection ownership and slow-upload progress, preservation of TLS record boundaries, and early validation of incompatible REALITY fingerprints. H2 SETTINGS GREASE applies only to the resolved Chrome 133 profile.

TLS session resumption stays disabled by default. The default XHTTP POST limit stays fixed at 1,000,000 bytes; variable ranges require explicit configuration. These options have demonstrated functional benefits, but their costs depend on the workload.

The measured resumption fixture reduced complete TLS bytes by 8.71%, while a separate 300-request ARM emulation measurement increased client CPU from 7.71 to 7.93 seconds. Short variable POSTs also increased CPU in earlier measurements. There is no general CPU, RAM, latency or energy improvement claim. Lower asset-downloader allocations do not establish lower whole-process RAM usage.

Chrome 133 matching is partial and does not imply equivalence to current Chrome or resistance to traffic identification. Saturated narrow-link cases failed in both baseline and candidate. Raw TCP half-close and prompt propagation of HTTP/1 client cancellation to every remote origin remain limitations. External ECH tests were not validated; tests with bounded workloads do not establish indefinite leak freedom.

The experimental selection adds no new routing policy, DNS cache architecture, global worker limit or resource controller. Existing functionality was checked in bounded local and OpenWrt ARM64 tests; see the release's validation notes for checks of the actual published artifacts.

## OpenWrt packaging

The self-contained packaging source and invocation are in [install/openwrt-experimental](install/openwrt-experimental/README.md). Package identities remain `v2raya-resilient`, `v2raya-resilient-core` and `luci-app-v2raya-resilient`; user-visible names are x-Ray VPN. Core and service package versions must match exactly.

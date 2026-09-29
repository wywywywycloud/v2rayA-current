# x-Ray VPN experimental validation

Version: `2026.09.29-experimental.1`
Release name: `xray-proxy-client-experimental`
Source: `b51db60a6654b50372bfe6d4f3d63a204b928fc5`
Test date: 2026-09-29. Target tested: OpenWrt 24.10.4, ARM64 VM; IPK architecture `aarch64_cortex-a53`.

**Build and package integrity: PASS. ARM reinstall of this delta was not re-run; see scope below. The 2026.09.28 ARM functional PASS remains the last full ARM run. Manual UI acceptance for this delta is not claimed.**

Delta since `2026.09.28-experimental.1` (`002c144f`): OOM-orphan core reaping with Pdeathsig plus process-group kill on cancel, startup reaping of orphaned probe/main cores and leaked /tmp configs, bounded retry of the last config on unexpected core stop (10/30/60s), GeoIP cache generation eviction with serialized asset reads, 5 MiB file-log cap, procd respawn `3600 15 0`. Commit message reports 19/19 checks on an OpenWrt 24.10.5 x86_64 VM including `kill -9` recovery of the main core; that x86_64 run is not an ARM64 install claim. Local verification for this bundle: both ARM64 binaries embed `2026.09.29-experimental.1` and the new symbols (`Pdeathsig`, `cleanup_stale`, `handleUnexpectedStop`), `Packages` Size/SHA256 match the IPKs, `Packages.gz` matches `Packages`, and the dependency chain (core <- service <- LuCI <- metapackage, exact `=` pins) is unchanged.

## Build and provenance

- Linux build: Go 1.26.8, Node 24.21.0, Yarn 1.22.22; CGO disabled. Both release binaries are static ARM64 ELF files.
- GUI lint, typecheck, i18n and production build passed. Lint reported existing warnings, with no errors.
- The full GUI suite passed 232 of 236 tests initially; four tests reached the 5000ms limit under QEMU TCG. The four affected files were rerun with two workers and a 30000ms timeout: all 24 tests in those files passed. No assertions or product source were changed for this rerun. All 236 distinct tests are covered by the combined results.
- Service `go build ./...`, `go vet ./...` and `go test ./...` passed; tests ran as root for the DNS resolver lifecycle tests. Core `go build ./...` and `go vet ./...` passed.
- Xray and net use the exact remote fork versions recorded in `BUILD-MANIFEST.json` and the binary build information. No local module replacements or manual dependency source edits were used. Standard Go downloads and checksum authentication populated an offline Linux download cache.
- Service module verification passed before and after building. Core's historical bare `anytls` alias causes standard `go mod verify` to fail when locating its ziphash. The selected replacement ZIPs and extracted directories were independently checked before and after building with Go's Hash1 algorithm against committed checksums or the authenticated standard-download checksum receipt. The original verifier failures are retained in internal evidence.
- All 702 input source files matched the frozen archive after the build. The standalone web archive contains exactly the same 71 files as the frontend copy embedded during service compilation; archive ownership and timestamps are fixed.
- All three IPKs have matching versions and exact service/core dependencies. Binary payloads inside IPKs match the built binaries.

## Installed-package validation

The complete three-package set was installed over the existing lab10 set. OpenWrt's version ordering treats `experimental.1` as lower than `lab10`, so the controlled upgrade test used `opkg --force-downgrade` for the full matching set; dependency checks were retained.

- Installed binary SHA-256 values matched the build.
- procd start and restart succeeded; the PID changed and the API reported service/core version `2026.09.28-experimental.1`.
- VLESS/XHTTP upload and download payload verification passed, including a 4 MiB upload with SHA-256 `d8e39cbfa0e5610942c84c9f11301a38c8cfb34ec5a73d04050074758dd9d106`.
- Manual selection, node recovery, direct/proxy/block routing and roundrobin/leastping/keepcurrent/random group scenarios passed.
- Stable blackhole configuration, eight blocked canary checks and recovery passed. Sixteen concurrent flows with DNS and API polling passed.
- UDP/TCP DNS clients, UDP/TCP/verified DoH upstreams, cache hits, unique queries and NXDOMAIN checks passed.
- Extended DNS evidence was independently decoded: 218 observations, 217 valid raw replies and one expected typed timeout, with no audit violations. Both 100-client bursts were pending before release. Each unique name reached upstream exactly once; no same-key singleflight claim is made.
- Cache expiry/min/max behavior and wire TTLs passed. Max-case TTLs were 60/60/60; min-case TTLs were 1/3/1. Controlled prefetch returned A/A/A/B/B with TTLs 60/59/59/60/58, including B after the old entry's expiry. Same-key recovery after upstream outage passed.
- HTTP 503 propagation, cancellation followed by an independent successful request, and core-stop closure of a pending origin request passed.
- Rollback to the saved lab10 packages restored the exact prior binary hashes and a working API. Reinstallation restored the exact experimental binary hashes and a working API. The final VM service is running experimental.

## Scope and manual acceptance

The local UI endpoint returned HTTP 200 and the expected version API. The user subsequently completed manual testing and reported that everything worked normally. This is user-reported UI acceptance; individual viewport/theme coverage was not specified and is not independently claimed. No additional test-administrator creation is required for acceptance.

These are controlled functional VM checks, not a CPU, RAM, energy, latency, censorship-evasion or detection benchmark. Prompt HTTP/1 remote-origin cancellation, transparent interception and general all-dead fallback behavior are not established by this run. Existing raw TCP half-close and external ECH/CA limitations are not claimed fixed. Optional resumption and altered XHTTP POST-range experiments remain off.

This is an experimental fork release distributed through our personal signed feed, not an official OpenWrt feed submission. Physical-router testing was outside this run. Public assets contain no fixture private keys, config databases, SSH keys or raw private evidence.

## Binary SHA-256

```text
708840cb702074a6fdbfc1c8525673474ae24f6bf2b805950e0a1b37746e172f  /usr/bin/v2raya
16ac0234d732354db2ff419891d5c95c71ec9111650a3189d0e9a0c0bee1c771  /usr/bin/v2raya_core
```

Use `SHA256SUMS` to verify the release assets. Source/build manifests provide the exact source, dependency pins and build commands.

## Signed-feed installation check

The signed index was accepted by OpenWrt usign with the rotated key `1c624cadb7ee79e7` (previous key `9478c50315c92021` lost; clients must re-run `add-resilient-feed.sh` once). Installing only `xray-proxy-client-experimental` downloads and installs all three exact-version application packages automatically; no dependency, architecture or checksum override was used. The final binaries matched the release SHA-256 values and the version API passed. The existing local account remained present.

This check used a local HTTP mirror of the signed candidate feed because the VM could not access external repositories. System dependencies were already installed on the test VM. The custom feed index passed signature verification; the combined update command also reported failures for unreachable unrelated official sources, retained in internal evidence. Public feed bytes are checked separately after publication. Official repositories appropriate to the device and kernel remain required for system dependencies.

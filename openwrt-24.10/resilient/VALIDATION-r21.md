# Resilient r21: subscription stability and protected group switching

Validated 2026-09-28 on a disposable OpenWrt 24.10.4 ARM64 VM with 512 MiB RAM.
Service/core: `2.5.7-resilient.14-r21.resilient1`; LuCI: `26.268.0-r21.resilient1`.
Application source: `283daeb0` (merged in fork PR #1 as `991aa7d8`). Packaging
source: `8546ed6`, with `tools/apply-resilient-defaults.py` applied only to the
disposable release checkout. The main source and upstream PR do not carry those
OpenWrt-specific defaults. See `BUILD.txt` for exact commits and compiler flags.

## Reproduction before the fix

The previous r20 binaries were installed on the same isolated VM. Controlled
SOCKS fixtures supplied reordered subscriptions, slow but reachable servers,
one-shot health failures and unavailable nodes. No production credentials or
subscription contents were used in the fixtures.

- A subscription reorder changed the main core PID and terminated an active
  proxied stream despite an unchanged selected endpoint.
- Pinning and clearing the selection in keep-current returned `leastping`.
- A two-second pre-start core hook exposed a gap in transparent interception.
  A postrouting counter observed **75 direct-bypass packets** during one
  successful switch, with the interception table missing in **47/55 samples**.
  Startup plus rollback failure left the table absent and increased the counter
  to **146 cumulatively** (71 additional packets; table absent in 46/55 samples).
- A DNS fixture answering UDP but rejecting TCP returned SERVFAIL and dispatcher
  EOF errors with the old core's forced TCP transport.

The TEST-NET fixture subnet was removed from the VM's deliberate direct-route
exceptions, and a successful transparent HTTP request was checked before the
interception measurement. Measurements taken while the subnet was exempt are
excluded from these results.

## Corrected behavior

- Seven reorder/rename updates preserved the main core PID and a live HTTP
  stream. Connection identity ignores display names and catalog ordering.
- Manual pin/unpin preserved keep-current and automatic membership. Only the
  pinned member was reported active. A slow but reachable current member stayed
  selected without restarting the core.
- Adding a real member initiated the least-ping, speed-qualified replacement
  search while the configured policy remained keep-current.
- A single health failure retained the route. A confirmed failure selected an
  available replacement without changing policy.
- Least latency, random and round robin retained their policies after updates.
- Repeated catalog changes before worker completion retain a persistent pending
  selection invalidation; a unit regression covers changing back to the original
  catalog. Catalog, connected references and invalidation commit atomically.
- With RoutingA assigning the fixture destination to proxy, both **TPROXY and
  REDIRECT** kept the interception table in every sample and recorded **zero
  bypass packets**, for successful switches and startup/rollback failure.
- UDP DNS returned NOERROR with the expected A record through the dispatcher.
  Unit tests verify that fallback remains on the selected outbound.

The source-candidate run sampled 1,042 process states: maximum **two** core
processes and minimum MemAvailable 291,728 KiB. The installed release-package
rerun sampled **872** states: maximum **two**, minimum MemAvailable **346,884 KiB**.
This bounds observed probe concurrency; it does not establish a universal memory
limit or rule out unrelated core memory growth.

## Build, tests and installation

- Fork GUI: lint (zero errors; existing line-ending warnings), typecheck,
  six-locale check, **236 tests** and production build passed.
- Upstream PR branch: the same checks and **222 GUI tests** passed.
- Both service and core modules built and passed `go vet ./...`.
- Full service tests ran on Linux ARM64. Changed configure/controller/service
  packages were rerun after the final invalidation fix. The upstream branch's
  full service suite also passed. The deliberate over-256-MiB allocation test
  ran separately on the Windows host and passed for both branches.
- Core DNS transport tests passed. Browser checks covered the group dialog,
  saved policy, 390 px mobile width, and light/dark themes.
- All nine fork PR checks passed, including Linux/macOS/Windows amd64/arm64
  builds, GUI checks and the OpenWrt strategy job, before merging the fork.
- The package index was signed with the existing feed key and verified using
  OpenWrt `usign`. All three r21 IPKs upgraded the VM's r20 package installation;
  both binaries reported `2.5.7-resilient.14`, and the existing UCI configuration
  was retained. The complete live subscription suite, both interception-failure scenarios and
  the UDP-only DNS check passed again on those IPKs.

## Reproduction tools and limits

Use the packaging repository's `tools/local-stability-fixture.py`,
`tools/test-subscription-stability.py` and
`tools/test-interception-retention.py`, following `tools/README.md`. They mutate
only a disposable local VM, including its routing rules and lifecycle hook.
Never run the test drivers against a production router.

The release keeps existing feed compatibility declarations for OpenWrt
24.10.0–24.10.8, but this change was tested on 24.10.4; a new nine-version matrix
and physical-router installation are not claimed. TPROXY/REDIRECT group switches
preserve interception. Unsupported preservation, including TUN, rejects a switch
before stopping the core. Explicit stop and service crashes are outside this
group-switch protection. A real server/core switch may terminate existing TCP
sessions; metadata-only updates leave them intact.

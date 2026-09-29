package v2ray

import (
	"os"
	"path/filepath"
	"time"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// staleProbeTempAge bounds how old an unreferenced probe config may be.
// A probe lives at most CoreStartupTimeout plus the probe timeout (tens of
// seconds), so anything older was leaked by a SIGKILLed run whose defers
// never executed. The files live in os.TempDir (/tmp, i.e. tmpfs/shmem on
// OpenWrt), where every leak shrinks the 256 MiB box a little more.
const staleProbeTempAge = 10 * time.Minute

var staleProbeTempPatterns = []string{
	"v2raya-subscription-*.json",
	"tun_route_*",
}

// CleanupStaleProbeTempFiles removes probe/TUN configs leaked by runs that
// died before their deferred os.Remove executed. Called once at startup,
// when no probe of this instance can exist yet; the age guard additionally
// protects against racing a concurrently starting instance.
func CleanupStaleProbeTempFiles() {
	now := time.Now()
	for _, pattern := range staleProbeTempPatterns {
		matches, err := filepath.Glob(filepath.Join(os.TempDir(), pattern))
		if err != nil {
			continue
		}
		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil || now.Sub(info.ModTime()) < staleProbeTempAge {
				continue
			}
			if err := os.Remove(path); err == nil {
				log.Warn("removed stale probe temp file %s left by a previous run", path)
			}
		}
	}
}

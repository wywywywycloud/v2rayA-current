//go:build !linux

package v2ray

import "os/exec"

// ConfigureChildDeathSig is a no-op outside Linux: Pdeathsig is Linux-only
// and probe/core orphans there are reaped by the normal context cancel.
func ConfigureChildDeathSig(cmd *exec.Cmd) {}

// CleanupStaleProbeCores only applies to Linux /proc reaping.
func CleanupStaleProbeCores() {}

//go:build linux

package v2ray

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// ConfigureChildDeathSig makes a child die with its parent.
//
// exec.CommandContext only kills the child when the parent cancels the
// context. An OOM SIGKILL (or any abrupt parent death) runs no cleanup,
// leaving v2raya_core and subscription probe cores orphaned. Pdeathsig asks
// the kernel to SIGKILL the child when the parent thread dies. Setpgid plus
// a group-killing Cancel covers grandchildren that outlive a normal cancel.
func ConfigureChildDeathSig(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
	cmd.SysProcAttr.Setpgid = true
	// WaitDelay reaps the child if Cancel kills it after the context ends.
	if cmd.WaitDelay <= 0 {
		cmd.WaitDelay = 2 * time.Second
	}
	if cmd.Cancel == nil {
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return nil
			}
			// Negative pid = whole process group started by this child.
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				return cmd.Process.Kill()
			}
			return nil
		}
	}
}

// CleanupStaleProbeCores kills cores orphaned by a previous OOM run: probe
// cores (cmdline references /tmp/v2raya-subscription-*.json) and main cores
// (cmdline references the stable config path). A SIGKILLed v2raya leaves
// both behind; a stale main core additionally holds the API/inbound ports,
// which would fail the next Start's port check and keep the proxy down.
// Only orphans reparented to init (ppid 1) are touched, never live children
// of this or any running instance.
func CleanupStaleProbeCores() {
	mainConfig := asset.GetV2rayConfigPath()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}
	for _, e := range entries {
		if !isPidDir(e.Name()) {
			continue
		}
		cmdline, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil {
			continue
		}
		// cmdline is NUL-separated; a substring match is enough.
		kind := ""
		switch {
		case strings.Contains(string(cmdline), "v2raya-subscription-"):
			kind = "probe"
		case mainConfig != "" && strings.Contains(string(cmdline), mainConfig):
			kind = "main"
		default:
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 1 {
			continue
		}
		if !isOrphanPid(pid) {
			continue
		}
		log.Warn("killing stale %s core pid %d left by a previous run", kind, pid)
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}

func isPidDir(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isOrphanPid reports whether pid's parent is init (ppid 1), i.e. it was
// reparented after its v2raya parent died instead of being our live child.
func isOrphanPid(pid int) bool {
	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return false
	}
	// comm may contain spaces inside parens; ppid follows ") ".
	rest := stat
	if i := strings.LastIndex(string(stat), ") "); i >= 0 {
		rest = stat[i+2:]
	}
	fields := strings.Fields(string(rest))
	if len(fields) < 2 {
		return false
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return false
	}
	return ppid == 1
}

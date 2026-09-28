package v2ray

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestNewProcessStartupCleanup(t *testing.T) {
	for _, mode := range []string{"poststart-error", "startup-timeout", "clean-exit"} {
		t.Run(mode, func(t *testing.T) {
			env := conf.GetEnvironmentConfig()
			previous := *env
			t.Cleanup(func() { *env = previous })
			env.Config = t.TempDir()
			env.V2rayAssetsDirectory = env.Config
			env.CoreStartupTimeout = 1
			env.V2rayBin = filepath.Join(env.Config, "fake-core")
			pidfile := filepath.Join(env.Config, "pid")
			body := "#!/bin/sh\necho $$ > '" + pidfile + "'\nexec sleep 60\n"
			if mode == "clean-exit" {
				body = "#!/bin/sh\nexit 0\n"
			}
			if err := os.WriteFile(env.V2rayBin, []byte(body), 0700); err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			listener.Close()
			closed := 0
			tmpl := &Template{API: &coreObj.APIObject{}, ApiPort: port, ApiCloses: []func(){func() { closed++ }}}
			postErr := errors.New("poststart failed")
			post := func() error {
				if mode != "poststart-error" {
					return nil
				}
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					if data, err := os.ReadFile(pidfile); err == nil {
						if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
							return postErr
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
				return fmt.Errorf("helper did not start")
			}
			process, err := NewProcess(tmpl, func() error { return nil }, post, func(*Process) {})
			if process != nil || err == nil {
				t.Fatalf("unexpected startup: process=%v err=%v", process, err)
			}
			if closed != 1 {
				t.Fatalf("template closed %d times", closed)
			}
			if mode == "poststart-error" && !errors.Is(err, postErr) {
				t.Fatal(err)
			}
			if mode == "startup-timeout" && !strings.Contains(err.Error(), "did not open its API port") {
				t.Fatal(err)
			}
			if mode == "clean-exit" {
				if !strings.Contains(err.Error(), "exited right after starting") {
					t.Fatal(err)
				}
				return
			}
			data, readErr := os.ReadFile(pidfile)
			if readErr != nil {
				t.Fatal(readErr)
			}
			pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			if killErr := syscall.Kill(pid, 0); !errors.Is(killErr, syscall.ESRCH) {
				t.Fatalf("child %d not reaped: %v", pid, killErr)
			}
		})
	}
}

package v2ray

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCoreCommandHelper(t *testing.T) {
	switch os.Getenv("V2RAYA_PROCESS_HELPER") {
	case "exit":
		os.Exit(0)
	case "wait":
		time.Sleep(time.Minute)
		os.Exit(0)
	}
}

func commandWatchers() int {
	var stacks bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&stacks, 2)
	return strings.Count(stacks.String(), "os/exec.(*Cmd).watchCtx(")
}

func TestCoreCommandReapsContextWatchers(t *testing.T) {
	baseline := commandWatchers()
	for i := range 32 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd, err := RunWithLog(ctx, os.Args[0], []string{os.Args[0], "-test.run=^TestCoreCommandHelper$"}, "",
			append(os.Environ(), "V2RAYA_PROCESS_HELPER=exit"))
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		state, err := cmd.Wait()
		cancel()
		if err != nil || state == nil || !state.Success() {
			t.Fatalf("iteration %d: state=%v err=%v", i, state, err)
		}
		// Multiple owners observe one completed wait; they never call Process.Wait again.
		if again, err := cmd.Wait(); err != nil || again != state {
			t.Fatalf("second wait: %v, %v", again, err)
		}
	}
	runtime.GC()
	if got := commandWatchers(); got > baseline {
		t.Fatalf("context watchers leaked: before=%d after=%d", baseline, got)
	}
}

func TestCoreCommandCancellationReapsProcess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, err := RunWithLog(ctx, os.Args[0], []string{os.Args[0], "-test.run=^TestCoreCommandHelper$"}, "",
		append(os.Environ(), "V2RAYA_PROCESS_HELPER=wait"))
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-cmd.done:
	case <-time.After(10 * time.Second):
		t.Fatal("cancel did not reap child")
	}
	if state, err := cmd.Wait(); state == nil || err == nil {
		t.Fatalf("canceled command: state=%v err=%v", state, err)
	}
}

func TestCoreCommandStartFailure(t *testing.T) {
	cmd, err := RunWithLog(context.Background(), "v2raya-nonexistent-test-executable", nil, "", os.Environ())
	if err == nil || cmd != nil {
		t.Fatalf("start failure: cmd=%v err=%v", cmd, err)
	}
}

func TestProcessCloseWaitsForCommandCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, err := RunWithLog(ctx, os.Args[0], []string{os.Args[0], "-test.run=^TestCoreCommandHelper$"}, "",
		append(os.Environ(), "V2RAYA_PROCESS_HELPER=wait"))
	if err != nil {
		t.Fatal(err)
	}
	process := &Process{proc: cmd.Process, command: cmd, procCancel: cancel, template: &Template{}, done: make(chan struct{})}
	var managerMu sync.Mutex
	managerMu.Lock()
	go func() {
		_, _ = cmd.Wait()
		managerMu.Lock()
		defer managerMu.Unlock()
		close(process.done)
	}()
	closed := make(chan error, 1)
	go func() { closed <- process.Close() }()
	select {
	case err := <-closed:
		managerMu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		managerMu.Unlock()
		t.Fatal("Close blocked on the monitor callback's manager lock")
	}
	select {
	case <-cmd.done:
	default:
		t.Fatal("Close returned before command cleanup completed")
	}
	if err := process.Close(); err != nil {
		t.Fatalf("repeated Close: %v", err)
	}
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()
	if err := process.WaitUntilExit(waitCtx); err != nil {
		t.Fatal(err)
	}
}

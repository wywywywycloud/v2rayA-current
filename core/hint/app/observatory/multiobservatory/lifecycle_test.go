package multiobservatory

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xtls/xray-core/app/dispatcher"
	"github.com/xtls/xray-core/app/proxyman"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
)

func TestMultiObservatoryRealClose(t *testing.T) {
	var active, connections, requests atomic.Int64
	started := make(chan struct{}, 100)
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		active.Add(1)
		defer active.Add(-1)
		started <- struct{}{}
		<-r.Context().Done()
	}))
	s.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
		if state == http.StateClosed || state == http.StateHijacked {
			connections.Add(-1)
		}
	}
	s.Start()
	defer s.Close()
	defer s.CloseClientConnections()
	v, err := core.New(&core.Config{
		App:      []*serial.TypedMessage{serial.ToTypedMessage(&dispatcher.Config{}), serial.ToTypedMessage(&proxyman.OutboundConfig{})},
		Outbound: []*core.OutboundHandlerConfig{{Tag: "probe", ProxySettings: serial.ToTypedMessage(&freedom.Config{})}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = v.Start(); err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	for i := 0; i < 10; i++ {
		obj, err := core.CreateObject(v, &Config{Observers: []*ObserverConfig{{Tag: "group", ProbeUrl: s.URL, ProbeInterval: int64(time.Hour), SubjectSelector: []string{"probe"}}}})
		if err != nil {
			t.Fatal(err)
		}
		m := obj.(*MultiObservatory)
		if err = m.Start(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			m.Close()
			t.Fatal("adapter did not start child probe")
		}
		if _, err = m.GetObservationByTag("group", context.Background()); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		m.Close()
		for time.Since(start) < time.Second && (active.Load() != 0 || connections.Load() != 0) {
			time.Sleep(5 * time.Millisecond)
		}
		if active.Load() != 0 || connections.Load() != 0 {
			t.Fatalf("cycle %d: active=%d connections=%d", i, active.Load(), connections.Load())
		}
		t.Logf("cycle=%d cancel_settle=%s active=%d connections=%d", i, time.Since(start), active.Load(), connections.Load())
	}
	count := requests.Load()
	time.Sleep(100 * time.Millisecond)
	if requests.Load() != count {
		t.Fatal("request after adapter Close")
	}
	t.Logf("cycles=10 requests=%d", count)
}

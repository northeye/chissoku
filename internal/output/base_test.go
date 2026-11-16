package output

import (
	"context"
	"testing"
	"time"
)

func TestBaseEmitsFirstThenInterval(t *testing.T) {
	t.Parallel()
	b := &Base{Interval: 1}
	got := make(chan string, 10)
	b.start(context.Background(), func(d []byte) { got <- string(d) })
	defer b.Close()

	b.Output([]byte("first"))
	if s := recv(t, got, time.Second); s != "first" {
		t.Fatalf("got %q, want first", s)
	}
	// only the latest data is emitted on the next tick
	b.Output([]byte("second"))
	b.Output([]byte("third"))
	if s := recv(t, got, 2*time.Second); s != "third" {
		t.Fatalf("got %q, want third", s)
	}
}

func TestBaseZeroInterval(t *testing.T) {
	t.Parallel()
	b := &Base{Interval: 0}
	got := make(chan string, 10)
	b.start(context.Background(), func(d []byte) { got <- string(d) })
	defer b.Close()

	for _, want := range []string{"a", "b"} {
		b.Output([]byte(want))
		if s := recv(t, got, time.Second); s != want {
			t.Fatalf("got %q, want %q", s, want)
		}
	}
}

func TestBaseCloseIsIdempotent(t *testing.T) {
	t.Parallel()
	b := &Base{Interval: 60}
	b.Close() // before start
	b.start(context.Background(), func([]byte) {})
	b.Close()
	b.Close()
	b.Output([]byte("after close")) // must not block or panic
}

func TestStdoutDeactivatesOnIterations(t *testing.T) {
	t.Parallel()
	s := &Stdout{Base: Base{Interval: 0}, Iterations: 2}
	deactivated := make(chan string, 1)
	ctx := WithDeactivator(context.Background(), func(o Outputter) { deactivated <- o.Name() })
	if err := s.Initialize(ctx); err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	s.Output([]byte(`{"n":1}`))
	time.Sleep(50 * time.Millisecond)
	s.Output([]byte(`{"n":2}`))
	if name := recv(t, deactivated, time.Second); name != "stdout" {
		t.Fatalf("deactivated %q", name)
	}
}

func recv[T any](t *testing.T, ch <-chan T, timeout time.Duration) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatal("timed out")
	}
	var zero T
	return zero
}

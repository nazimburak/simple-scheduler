package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	tickInterval = 100 * time.Millisecond
}

func TestRunAtJob(t *testing.T) {
	s := New()
	defer s.Stop()

	var runCount int32
	runAt := time.Now()

	err := s.Register("once", nil, &runAt, func(ctx context.Context) error {
		atomic.AddInt32(&runCount, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	s.Start()

	time.Sleep(200 * time.Millisecond)
	if n := atomic.LoadInt32(&runCount); n != 1 {
		t.Fatalf("expected job to run once, but ran %d times", n)
	}
}

func TestRegister_InvalidCronExpr(t *testing.T) {
	s := New()
	defer s.Stop()

	cronExpr := "invalid"
	err := s.Register("bad", &cronExpr, nil, func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatalf("expected invalid cron expression error, got nil")
	}
}

func TestRegister_NilTask(t *testing.T) {
	s := New()
	defer s.Stop()

	runAt := time.Now().Add(time.Second)
	err := s.Register("nil", nil, &runAt, nil)
	if err == nil {
		t.Fatalf("expected task cannot be nil error, got nil")
	}
}

func TestRegister_NoCronNoRunAt(t *testing.T) {
	s := New()
	defer s.Stop()

	err := s.Register("missing", nil, nil, func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatalf("expected error for missing schedule, got nil")
	}
}

func TestDuplicateJob(t *testing.T) {
	s := New()
	defer s.Stop()

	runAt := time.Now().Add(1 * time.Second)
	task := func(ctx context.Context) error { return nil }

	if err := s.Register("dummy", nil, &runAt, task); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if err := s.Register("dummy", nil, &runAt, task); err == nil {
		t.Fatalf("expected duplicate job error, got nil")
	}
}

func TestStopIsIdempotent(t *testing.T) {
	s := New()
	s.Start()

	s.Stop()
	s.Stop()
}

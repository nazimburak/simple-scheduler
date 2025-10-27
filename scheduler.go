package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nazimburak/cronkit"
)

var tickInterval = 5 * time.Second

type job struct {
	name           string
	parsedCronExpr *cronkit.Expression
	task           func(context.Context) error
	nextRun        time.Time
	running        bool
}

type Scheduler struct {
	mu     sync.Mutex
	jobs   map[string]*job
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func New() *Scheduler {
	return &Scheduler{
		jobs:   make(map[string]*job),
		stopCh: make(chan struct{}),
	}
}

// Register adds a job to the scheduler. Either cronExpr or runAt must be provided.
func (s *Scheduler) Register(name string, cronExpr *string, runAt *time.Time, task func(context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job already exists: %s", name)
	}
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	job := &job{
		name: name,
		task: task,
	}

	switch {
	case cronExpr != nil:
		expr, err := cronkit.Parse(*cronExpr)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		job.parsedCronExpr = expr
		job.nextRun = expr.Next(time.Now())
	case runAt != nil:
		job.nextRun = *runAt
	default:
		return fmt.Errorf("either cronExpr or runAt must be provided")
	}

	s.jobs[name] = job
	return nil
}

// Start begins the scheduler loop.
// It checks every 5 seconds if any job is due and runs it in a separate goroutine.
func (s *Scheduler) Start() {
	go func() {
		for {
			select {
			case <-s.stopCh:
				return
			default:
				now := time.Now()

				s.mu.Lock()
				for _, job := range s.jobs {
					if job.running {
						continue
					}
					if !job.nextRun.IsZero() && !now.Before(job.nextRun) {
						job.running = true
						s.wg.Add(1)
						go s.run(job)
					}
				}
				s.mu.Unlock()

				time.Sleep(tickInterval)
			}
		}
	}()
}

// run executes a single job and schedules its next run (if cron-based).
func (s *Scheduler) run(job *job) {
	defer s.wg.Done()
	defer func() {
		s.mu.Lock()
		job.running = false
		s.mu.Unlock()
	}()

	ctx := context.Background()
	if err := job.task(ctx); err != nil {
		log.Printf("[scheduler] job %s failed: %v", job.name, err)
	}

	if job.parsedCronExpr != nil {
		job.nextRun = job.parsedCronExpr.Next(time.Now())
	} else {
		job.nextRun = time.Time{}
	}
}

// Stop gracefully stops the scheduler and waits for running jobs to finish.
func (s *Scheduler) Stop() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
	s.wg.Wait()
}

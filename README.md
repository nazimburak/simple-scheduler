# simple-scheduler

A simple Go package for scheduling and running background jobs. 
You can **register** jobs with a name, schedule them using either a **cron pattern** or a **specific time**, **start** the scheduler to run them automatically, and **stop** it gracefully when shutting down your application.

It runs each job concurrently in its own goroutine and depends only on [cronkit](https://github.com/nazimburak/cronkit) for cron parsing.

---

## Features

- Register jobs dynamically at runtime
- Supports both **cron-based recurring** and **one-time** jobs 
- Thread-safe registration and execution
- Minimal dependency

---

## Usage

```
s := scheduler.New()

// Example 1: Run at a specific time
runAt := time.Now().Add(2 * time.Second)
_ = s.Register("once", nil, &runAt, func(ctx context.Context) error {
    fmt.Println("One-time job executed at", time.Now())
    return nil
})

// Example 2: Run every minute (cron-based)
expr := "*/1 * * * *"
_ = s.Register("minute-job", &expr, nil, func(ctx context.Context) error {
    fmt.Println("Cron job executed at", time.Now())
    return nil
})

s.Start()
```

---

## Behavior & Notes

| Behavior                      | Description                                                                                     |
|-------------------------------|-------------------------------------------------------------------------------------------------|
| **Scheduler Execution** | `Start()` is non-blocking, it automatically launches the scheduler in a separate goroutine. |
| **Job Identification** | Each job must have a unique name. Duplicate names cause registration to fail.                   |
| **Scheduling**       | Either `cronExpr` or `runAt` must be provided (not both nil).                                   |
| **Tick Interval**          | Default check interval is 5 seconds                                                             |
| **Graceful Stop**  | `Stop()` waits for all active jobs to finish before returning.                                  |

---

## Planned Futures

- **Retry mechanism** `(WithRetry)`: Configurable retry count for failed jobs
- **Overlap policy** `(WithOverlapPolicy)`: Control whether a new job run can start while the previous one is still running (skip, wait, or run concurrently)
- **Context cancellation support**: Allow jobs to respect shutdown signals or deadlines
package background

import (
	"context"
	"testing"
	"time"
)

func TestRunnerWaitsForJobs(t *testing.T) {
	var r Runner
	finished := false
	r.Go("job", func(context.Context) {
		time.Sleep(20 * time.Millisecond)
		finished = true
	})

	if !r.Wait(context.Background()) {
		t.Fatal("Wait returned false")
	}
	if !finished {
		t.Error("Wait returned before the job finished")
	}
}

func TestRunnerRecoversPanics(t *testing.T) {
	var r Runner
	r.Go("panics", func(context.Context) { panic("boom") })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !r.Wait(ctx) {
		t.Error("a panicking job should still count as finished")
	}
}

func TestRunnerWaitStopsAtDeadline(t *testing.T) {
	var r Runner
	release := make(chan struct{})
	defer close(release)
	r.Go("slow", func(context.Context) { <-release })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if r.Wait(ctx) {
		t.Error("Wait should report the job as unfinished")
	}
}

func TestRunnerGivesJobsAContextWithADeadline(t *testing.T) {
	var r Runner
	hasDeadline := false
	r.Go("job", func(ctx context.Context) { _, hasDeadline = ctx.Deadline() })
	r.Wait(context.Background())
	if !hasDeadline {
		t.Error("job context has no deadline")
	}
}

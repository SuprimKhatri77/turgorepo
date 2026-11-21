package background

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerWaitsForJobs(t *testing.T) {
	var r Runner
	finished := false
	r.Go("job", func(context.Context) {
		time.Sleep(20 * time.Millisecond)
		finished = true
	})

	require.True(t, r.Wait(context.Background()), "Wait returned false")
	assert.True(t, finished, "Wait returned before the job finished")
}

func TestRunnerRecoversPanics(t *testing.T) {
	var r Runner
	r.Go("panics", func(context.Context) { panic("boom") })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.True(t, r.Wait(ctx), "a panicking job should still count as finished")
}

func TestRunnerWaitStopsAtDeadline(t *testing.T) {
	var r Runner
	release := make(chan struct{})
	defer close(release)
	r.Go("slow", func(context.Context) { <-release })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	assert.False(t, r.Wait(ctx), "Wait should report the job as unfinished")
}

func TestRunnerGivesJobsAContextWithADeadline(t *testing.T) {
	var r Runner
	hasDeadline := false
	r.Go("job", func(ctx context.Context) { _, hasDeadline = ctx.Deadline() })
	r.Wait(context.Background())
	assert.True(t, hasDeadline, "job context has no deadline")
}

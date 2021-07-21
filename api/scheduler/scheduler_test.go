package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_CanStartAndTerminate(t *testing.T) {
	s := NewScheduler(context.Background())
	s.StartJobEvery(1*time.Minute, func() { fmt.Println("boop") })

	err := s.Shutdown()
	assert.NoError(t, err, "Shutdown should return no errors")
	assert.Empty(t, s.crontab.Entries(), "all jobs should have been removed")
}

func Test_CanTerminateByCancellingContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := NewScheduler(ctx)
	s.StartJobEvery(1*time.Minute, func() { fmt.Println("boop") })

	cancel()

	for i := 0; i < 100; i++ {
		if len(s.crontab.Entries()) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("all jobs are expected to be cleaned by now; it might be a timing issue, otherwise implementation defect")
}

func Test_StartAndStopJob(t *testing.T) {
	s := NewScheduler(context.Background())
	defer s.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	var jobOne string
	var workDone bool
	jobOne = s.StartJobEvery(time.Second, func() {
		assert.Equal(t, 1, len(s.crontab.Entries()), "scheduler should have one active job")
		workDone = true

		s.StopJob(jobOne)
		cancel()
	})

	<-ctx.Done()
	assert.True(t, workDone, "value should been set in the job")
	assert.Equal(t, 0, len(s.crontab.Entries()), "scheduler should have no active jobs")

}

package scheduler

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	crontab     *cron.Cron
	shutdownCtx context.Context
}

func NewScheduler(ctx context.Context) *Scheduler {
	crontab := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))
	crontab.Start()

	s := &Scheduler{
		crontab: crontab,
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			s.Shutdown()
		}()
	}

	return s
}

// Shutdown stops the scheduler and waits for it to stop if it is running; otherwise does nothing.
func (s *Scheduler) Shutdown() error {
	if s.crontab == nil {
		return nil
	}

	log.Println("[DEBUG] Stopping scheduler")
	ctx := s.crontab.Stop()
	<-ctx.Done()

	for _, j := range s.crontab.Entries() {
		s.crontab.Remove(j.ID)
	}

	err := ctx.Err()
	if err == context.Canceled {
		return nil
	}
	return err
}

// StopJob stops the job from being run in the future
func (s *Scheduler) StopJob(jobID string) error {
	id, err := strconv.Atoi(jobID)
	if err != nil {
		return errors.Wrapf(err, "failed convert jobID %q to int", jobID)
	}
	s.crontab.Remove(cron.EntryID(id))

	return nil
}

// StartJobEvery schedules a new periodic job with a given duration.
// Returns job id that could be used to stop the given job
func (s *Scheduler) StartJobEvery(duration time.Duration, job func()) string {
	entryId := s.crontab.Schedule(cron.Every(duration), cron.FuncJob(job))
	return strconv.Itoa(int(entryId))
}

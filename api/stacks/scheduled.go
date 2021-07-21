package stacks

import (
	"log"
	"time"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/scheduler"
)

func StartStackSchedules(scheduler *scheduler.Scheduler, stackdeployer StackDeployer, datastore portainer.DataStore, gitService portainer.GitService) error {
	stacks, err := datastore.Stack().RefreshableStacks()
	if err != nil {
		return errors.Wrap(err, "failed to fetch refreshable stacks")
	}
	for _, stack := range stacks {
		d, err := time.ParseDuration(stack.AutoUpdate.Interval)
		if err != nil {
			return errors.Wrap(err, "Unable to parse auto update interval")
		}
		jobID := scheduler.StartJobEvery(d, func() {
			if err := RedeployWhenChanged(stack.ID, stackdeployer, datastore, gitService); err != nil {
				log.Printf("[ERROR] %s\n", err)
			}
		})

		stack.AutoUpdate.JobID = jobID
		if err := datastore.Stack().UpdateStack(stack.ID, &stack); err != nil {
			return errors.Wrap(err, "failed to update stack job id")
		}
	}
	return nil
}

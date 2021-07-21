package stacks

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
)

func validateStackAutoUpdate(autoUpdate *portainer.StackAutoUpdate) error {
	if autoUpdate == nil {
		return nil
	}
	if autoUpdate.Webhook != "" && !govalidator.IsUUID(autoUpdate.Webhook) {
		return errors.New("invalid Webhook format")
	}
	if autoUpdate.Interval != "" {
		if _, err := time.ParseDuration(autoUpdate.Interval); err != nil {
			return errors.New("invalid Interval format")
		}
	}
	return nil
}

package migrator

import (
	portainer "github.com/portainer/portainer/api"
)

func (m *Migrator) migrateDBVersionTo33() error {
	err := migrateStackEntryPoint(m.stackService)
	if err != nil {
		return err
	}

	return nil
}

func migrateStackEntryPoint(stackService portainer.StackService) error {
	stacks, err := stackService.Stacks()
	if err != nil {
		return err
	}
	for i := range stacks {
		stack := &stacks[i]
		if stack.GitConfig == nil {
			continue
		}
		stack.GitConfig.ConfigFilePath = stack.EntryPoint
		if err := stackService.UpdateStack(stack.ID, stack); err != nil {
			return err
		}
	}
	return nil
}

package daemon

import (
	"fmt"

	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/runconfig"
)

func (container *Container) Checkpoint(opts *runconfig.CriuConfig) error {
	if err := container.daemon.Checkpoint(container, opts); err != nil {
		return err
	}

	if opts.LeaveRunning == false {
		container.ReleaseNetwork()
	}
	return nil
}

func (container *Container) Restore(opts *runconfig.CriuConfig, forceRestore bool) error {
	var err error
	container.Lock()
	defer container.Unlock()

	defer func() {
		if err != nil {
			container.setError(err)
			// if no one else has set it, make sure we don't leave it at zero
			if container.ExitCode == 0 {
				container.ExitCode = 128
			}
			container.toDisk()
			container.cleanup()
		}
	}()

	if err := container.Mount(); err != nil {
		return err
	}
	if err = container.initializeNetworking(true); err != nil {
		return err
	}
	linkedEnv, err := container.setupLinkedContainers()
	if err != nil {
		return err
	}
	if err = container.setupWorkingDirectory(); err != nil {
		return err
	}

	env := container.createDaemonEnvironment(linkedEnv)
	if err = populateCommand(container, env); err != nil {
		return err
	}

	mounts, err := container.setupMounts()
	if err != nil {
		return err
	}

	container.command.Mounts = mounts
	return container.waitForRestore(opts, forceRestore)
}

func (container *Container) waitForRestore(opts *runconfig.CriuConfig, forceRestore bool) error {
	container.monitor = newContainerMonitor(container, container.hostConfig.RestartPolicy)

	// After calling promise.Go() we'll have two goroutines:
	// - The current goroutine that will block in the select
	//   below until restore is done.
	// - A new goroutine that will restore the container and
	//   wait for it to exit.
	select {
	case <-container.monitor.restoreSignal:
		if container.ExitCode != 0 {
			return fmt.Errorf("restore process failed")
		}
	case err := <-promise.Go(func() error { return container.monitor.Restore(opts, forceRestore) }):
		return err
	}

	return nil
}

package daemon

import (
	"fmt"

	"github.com/docker/docker/runconfig"
)

// Checkpoint a running container.
func (daemon *Daemon) ContainerCheckpoint(name string, opts *runconfig.CriuConfig) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}
	if !container.IsRunning() {
		return fmt.Errorf("Container %s not running", name)
	}
	if err := container.Checkpoint(opts); err != nil {
		return fmt.Errorf("Cannot checkpoint container %s: %s", name, err)
	}

	container.LogEvent("checkpoint")
	return nil
}

// Restore a checkpointed container.
func (daemon *Daemon) ContainerRestore(name string, opts *runconfig.CriuConfig, forceRestore bool) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}

	if !forceRestore {
		// TODO: It's possible we only want to bypass the checkpointed check,
		// I'm not sure how this will work if the container is already running
		if container.IsRunning() {
			return fmt.Errorf("Container %s already running", name)
		}

		if !container.IsCheckpointed() {
			return fmt.Errorf("Container %s is not checkpointed", name)
		}
	} else {
		if !container.HasBeenCheckpointed() && opts.ImagesDirectory == "" {
			return fmt.Errorf("You must specify an image directory to restore from %s", name)
		}
	}

	if err = container.Restore(opts, forceRestore); err != nil {
		container.LogEvent("die")
		return fmt.Errorf("Cannot restore container %s: %s", name, err)
	}

	container.LogEvent("restore")
	return nil
}

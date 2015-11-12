package daemon

import (
	"fmt"
	"path/filepath"

	derr "github.com/docker/docker/errors"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/runconfig"
)

// ContainerRestore restores the process in a container with CRIU
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

	if opts.ImagesDirectory == "" {
		opts.ImagesDirectory = filepath.Join(container.root, "criu.image")
	}

	if opts.WorkDirectory == "" {
		opts.WorkDirectory = filepath.Join(container.root, "criu.work")
	}

	if err = daemon.containerRestore(container, opts, forceRestore); err != nil {
		return fmt.Errorf("Cannot restore container %s: %s", name, err)
	}

	return nil
}

// containerRestore prepares the container to be restored by setting up
// everything the container needs, just like containerStart, such as
// storage and networking, as well as links between containers.
// The container is left waiting for a signal that restore has finished
func (daemon *Daemon) containerRestore(container *Container, opts *runconfig.CriuConfig, forceRestore bool) error {
	var err error
	container.Lock()
	defer container.Unlock()

	if container.Running {
		return nil
	}

	if container.removalInProgress || container.Dead {
		return derr.ErrorCodeContainerBeingRemoved
	}

	// if we encounter an error during start we need to ensure that any other
	// setup has been cleaned up properly
	defer func() {
		if err != nil {
			container.setError(err)
			// if no one else has set it, make sure we don't leave it at zero
			if container.ExitCode == 0 {
				container.ExitCode = 128
			}
			container.toDisk()
			daemon.Cleanup(container)
			daemon.LogContainerEvent(container, "die")
		}
	}()

	if err := daemon.conditionalMountOnStart(container); err != nil {
		return err
	}

	// Make sure NetworkMode has an acceptable value. We do this to ensure
	// backwards API compatibility.
	container.hostConfig = runconfig.SetDefaultNetModeIfBlank(container.hostConfig)

	if err := daemon.initializeNetworking(container, true); err != nil {
		return err
	}
	linkedEnv, err := daemon.setupLinkedContainers(container)
	if err != nil {
		return err
	}
	if err := container.setupWorkingDirectory(); err != nil {
		return err
	}
	env := container.createDaemonEnvironment(linkedEnv)
	if err := daemon.populateCommand(container, env); err != nil {
		return err
	}

	if !container.hostConfig.IpcMode.IsContainer() && !container.hostConfig.IpcMode.IsHost() {
		if err := daemon.setupIpcDirs(container); err != nil {
			return err
		}
	}

	mounts, err := daemon.setupMounts(container)
	if err != nil {
		return err
	}
	mounts = append(mounts, container.ipcMounts()...)

	container.command.Mounts = mounts
	return daemon.waitForRestore(container, opts, forceRestore)
}

func (daemon *Daemon) waitForRestore(container *Container, opts *runconfig.CriuConfig, forceRestore bool) error {
	container.monitor = daemon.newContainerMonitor(container, container.hostConfig.RestartPolicy)

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

package daemon

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/runconfig"
)

// Checkpoint checkpoints the running container, saving its state afterwards
func (container *Container) Checkpoint(opts *runconfig.CriuConfig) error {
	if err := container.daemon.Checkpoint(container, opts); err != nil {
		return err
	}

	if opts.LeaveRunning == false {
		container.cleanup()
	}

	// commit the filesystem as well
	commitCfg := &ContainerCommitConfig{
		Pause:  true,
		Config: container.Config,
	}
	img, err := container.daemon.Commit(container, commitCfg)
	if err != nil {
		return err
	}

	// Update the criu image path and image ID of the container
	criuImagePath := opts.ImagesDirectory
	if criuImagePath == "" {
		criuImagePath = filepath.Join(container.daemon.configStore.ExecRoot, "execdriver", container.daemon.configStore.ExecDriver, container.ID, "criu.image")
	}
	container.CriuimagePaths[criuImagePath] = img.ID

	// Update image layer of the committed container
	container.ImageID = img.ID

	if err := container.toDisk(); err != nil {
		return fmt.Errorf("Cannot update config for container: %s", err)
	}

	return nil
}

// Restore restores the container's process from images on disk
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

	if err := container.daemon.createRootfs(container); err != nil {
		return err
	}

	if err := container.Mount(); err != nil {
		return err
	}

	// Make sure NetworkMode has an acceptable value. We do this to ensure
	// backwards API compatibility.
	container.hostConfig = runconfig.SetDefaultNetModeIfBlank(container.hostConfig)
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

	if !container.hostConfig.IpcMode.IsContainer() && !container.hostConfig.IpcMode.IsHost() {
		if err := container.setupIpcDirs(); err != nil {
			return err
		}
	}

	mounts, err := container.setupMounts()
	if err != nil {
		return err
	}
	mounts = append(mounts, container.ipcMounts()...)

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

package daemon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/runconfig"
)

// ContainerCheckpoint checkpoints the process running in a container with CRIU
func (daemon *Daemon) ContainerCheckpoint(name string, opts *runconfig.CriuConfig) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}
	if !container.IsRunning() {
		return fmt.Errorf("Container %s not running", name)
	}

	if opts.ImagesDirectory == "" {
		opts.ImagesDirectory = filepath.Join(container.root, "criu.image")
		if err := os.MkdirAll(opts.ImagesDirectory, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	}

	if opts.WorkDirectory == "" {
		opts.WorkDirectory = filepath.Join(container.root, "criu.work")
		if err := os.MkdirAll(opts.WorkDirectory, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	}

	if err := daemon.Checkpoint(container, opts); err != nil {
		return fmt.Errorf("Cannot checkpoint container %s: %s", name, err)
	}

	container.SetCheckpointed(opts.LeaveRunning)

	if opts.LeaveRunning == false {
		daemon.Cleanup(container)
	}

	// commit the filesystem as well, support AUFS only
	commitCfg := &ContainerCommitConfig{
		Pause:  true,
		Config: container.Config,
	}
	img, err := daemon.Commit(name, commitCfg)
	if err != nil {
		return err
	}
	// Update the criu image path and image ID of the container
	criuImagePath := opts.ImagesDirectory
	container.CriuimagePaths[criuImagePath] = img.ID
	// Update image layer of the committed container
	container.ImageID = img.ID

	if err := container.toDisk(); err != nil {
		return fmt.Errorf("Cannot update config for container: %s", err)
	}

	return nil
}
